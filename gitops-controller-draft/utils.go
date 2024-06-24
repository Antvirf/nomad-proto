package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/nomad/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

func getEnv(key, defaultValue string) string {
	logger = zap.L()
	value := os.Getenv(key)
	if len(value) == 0 {
		fmt.Println("config: setting configurable variable from its default:", key, "=", defaultValue)
		return defaultValue
	}
	fmt.Println("config: setting configurable variable from environment", key, "=", value)
	return value
}

func ConvertVariableToNomadJobGroupStruct(variables []api.Variable) []NomadJobGroupObject {
	logger := zap.L()
	nomad_job_objects := []NomadJobGroupObject{}

	for _, variable := range variables {
		nomad_job_object_items := NomadJobGroupObjectItems{}
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			ErrorUnset:       true, // omission of any keys in the Nomad var will cause error
			ErrorUnused:      true, // randomly added keys in Nomad vars will cause error
			Result:           &nomad_job_object_items,
			WeaklyTypedInput: true,
			TagName:          "hcl", // Share the struct tag with HCL
		})
		err := decoder.Decode(variable.Items)
		if err != nil {
			logger.Error("failed to decode variable's items block to expected format",
				zap.Error(err))
			continue
		}

		// Convert the object's Items to a NomadObjectItems struct
		nomad_job_objects = append(nomad_job_objects, NomadJobGroupObject{
			Namespace:        variable.Namespace,
			Path:             variable.Path,
			Items:            nomad_job_object_items,
			OriginalVariable: &variable,
		},
		)
	}

	return nomad_job_objects
}

func ConvertVariableToGitRepositoryStruct(variables []api.Variable) []GitRepositoryObject {
	nomad_job_objects := []GitRepositoryObject{}
	logger := zap.L()

	for _, variable := range variables {

		// PATCHERS: Add empty values for status_ fields if not set in items currently
		if _, exists := variable.Items["status_current_commit"]; !exists {
			variable.Items["status_current_commit"] = ""
		}

		git_repository_object_items := GitRepositoryObjectItems{}
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			ErrorUnset:       true, // omission of any keys in the Nomad var will cause error
			ErrorUnused:      true, // randomly added keys in Nomad vars will cause error
			Result:           &git_repository_object_items,
			WeaklyTypedInput: true,
			TagName:          "hcl", // Share the struct tag with HCL
		})
		err := decoder.Decode(variable.Items)
		if err != nil {
			logger.Error("failed to decode variable's items block to expected format",
				zap.Error(err))
			continue
		}

		// Convert the object's Items to a NomadObjectItems struct
		nomad_job_objects = append(nomad_job_objects, GitRepositoryObject{
			Namespace:        variable.Namespace,
			Path:             variable.Path,
			Items:            git_repository_object_items,
			OriginalVariable: &variable,
		},
		)
	}

	return nomad_job_objects
}

func GetGitRepositoryForNomadJobGroup(job NomadJobGroupObject, repositories *[]GitRepositoryObject) (GitRepositoryObject, error) {
	for _, repo := range *repositories {
		if repo.Path == job.Items.GitRepositoryName {
			return repo, nil
		}
	}
	return GitRepositoryObject{}, errors.New("no GitRepo found for given NomadJobGroup")
}

func GetPathForRepository(repo GitRepositoryObject) string {
	// Compute base64 hash of the GitRepository object Path (=name), so that we have no collisisions
	// This might be needed if several sources target the same repository but e.g. different branch
	hashed_path_name := base64.RawURLEncoding.EncodeToString([]byte(repo.Path))
	repo_name := path.Base(repo.Items.Url)
	base_path_plus_hash := filepath.Join(controller_git_clone_base_path, hashed_path_name, repo_name)
	return base_path_plus_hash
}

func FilterFilePathsFromGivenDirectoryAndRegex(directory string, regex string) ([]os.DirEntry, error) {
	logger := zap.L()
	potential_files_to_apply := []os.DirEntry{}
	files_in_dir, err := os.ReadDir(
		directory,
	)
	if (err != nil) || len(files_in_dir) == 0 {
		return potential_files_to_apply, errors.New("directory is inaccessible or empty")
	}

	// Filter filepaths with given regex
	for _, file_path := range files_in_dir {
		matches_path_filter, err := regexp.MatchString(regex, file_path.Name())
		if err != nil {
			logger.Error("regex error when matching path filter to a filename",
				zap.Error(err), zap.String("fileName", file_path.Name()),
			)
		}
		if matches_path_filter {
			logger.Debug("file found matching filter", zap.String("directory", directory), zap.String("regex", regex), zap.String("fileName", file_path.Name()))
			potential_files_to_apply = append(potential_files_to_apply, file_path)
		}
	}
	return potential_files_to_apply, nil
}

// CopyFile and CopyDir from https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
/* MIT License
 *
 * Copyright (c) 2017 Roland Singer [roland.singer@desertbit.com]
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
