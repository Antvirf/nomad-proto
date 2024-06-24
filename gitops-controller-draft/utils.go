package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

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

		nomad_job_object_items := GitRepositoryObjectItems{}
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			ErrorUnset:       true, // omission of any keys in the Nomad var will cause error
			ErrorUnused:      true, // randomly added keys in Nomad vars will cause error
			Result:           &nomad_job_object_items,
			WeaklyTypedInput: true,
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
			Items:            nomad_job_object_items,
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
