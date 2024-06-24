package main

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"
)

func ControllerGitRepository(client *api.Client) {
	logger.Info("starting controller: GitRepository")

	git_repositories := FetchGitRepositoriesForController(client)

	// Main loop - get GitRepositories, clone them to local filesystem
	for _, repo := range git_repositories {

		base_path_plus_hash := GetPathForRepository(repo)

		// Check if the directory exists already, if yes - clean it up
		if _, err := os.Stat(base_path_plus_hash); !os.IsNotExist(err) {
			err = os.RemoveAll(base_path_plus_hash)
			if err != nil {
				logger.Error("failed to clean up files prior to cloning directory", zap.Error(err))
				continue
			}
		}
		// Make the directory in advance
		err := os.MkdirAll(base_path_plus_hash, os.ModePerm)
		if err != nil {
			logger.Error("failed to create controller base path for cloning directories", zap.Error(err))
			continue
		}

		// If type of repo is local-directory, we just clone that local dir using the `url` to the right place
		if repo.Items.Type == "local-directory" {
			logger.Info("copying GitRepository from a 'local-directory'", zap.String("localPath", repo.Items.Url), zap.String("destination", base_path_plus_hash), zap.String("gitRepository", repo.Path))
			os.Remove(base_path_plus_hash)
			err := CopyDir(repo.Items.Url, base_path_plus_hash)
			if err != nil {
				logger.Error("failed to copy local directory", zap.Error(err))
			}
			continue // end this loop here for this GitRepository instance - for `local-directory` we are done.
		}

		// Handle cloning
		repository, err := git.PlainClone(base_path_plus_hash, false, &git.CloneOptions{
			URL:           repo.Items.Url,
			Progress:      nil,
			ReferenceName: plumbing.ReferenceName(repo.Items.Branch),
			SingleBranch:  true, // only fetch the desired ref, getting everything is unnecessary
			Depth:         1,    // only fetch one commit, history is unnecessary

		})
		if err != nil {
			logger.Error("failed to clone Git repository", zap.String("gitRepository", repo.Path), zap.String("url", repo.Items.Url), zap.String("destination", base_path_plus_hash), zap.Error(err))
			continue // If failed to clone, move on to the next repository.
		}
		current_revision, _ := repository.ResolveRevision(plumbing.Revision(string(repo.Items.Branch)))
		logger.Info("successfully cloned Git Repository", zap.String("gitRepository", repo.Path), zap.String("commit", current_revision.String()))

		// Update the Nomad variable with status current commit (last refresh/modify time is stored as part of the variable itself)
		repo.OriginalVariable.Items["status_current_commit"] = current_revision.String()
		repo.OriginalVariable, _, err = client.Variables().Update(repo.OriginalVariable, &api.WriteOptions{})
		if err != nil {
			logger.Error("failed to update `status_current_commit` field back to Nomad Variables for GitRepository", zap.String("gitRepository", repo.Path))
		}
	}
}
