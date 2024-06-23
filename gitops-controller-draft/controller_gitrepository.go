package main

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"
)

func ControllerGitRepository(clientConfig *api.Config) {
	logger.Info("starting controller: GitRepository")
	logger = zap.L()
	//  Initialize Nomad Client
	client, err := api.NewClient(clientConfig)
	if err != nil {
		logger.Error("failed to initialize Nomad client", zap.Error(err))
	}

	git_repositories := FetchGitRepositoriesForController(client)

	// Main loop - get GitRepositories, clone them to local filesystem
	for _, repo := range git_repositories {

		base_path_plus_hash := GetPathForRepository(repo)

		// Check if the directory exists already, if yes - fetch instead of clone
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

		repository, err := git.PlainClone(base_path_plus_hash, false, &git.CloneOptions{
			URL:           repo.Items.Url,
			Progress:      nil,
			ReferenceName: plumbing.ReferenceName(repo.Items.Branch),
			SingleBranch:  true, // only fetch the desired ref, getting everything is unnecessary
			Depth:         1,    // only fetch one commit, history is unnecessary

		})
		if err != nil {
			logger.Error("failed to clone Git repository", zap.String("gitRepository", repo.Path), zap.Error(err))
			continue
		}
		current_revision, _ := repository.ResolveRevision(plumbing.Revision(string(repo.Items.Branch)))
		logger.Info("successfully cloned Git Repository", zap.String("gitRepository", repo.Path), zap.String("commit", current_revision.String()))

		// Update the Nomad variable with status current commit (last refresh/modify time is stored as part of the variable itself)
		repo.OriginalVariable.Items["status_current_commit"] = current_revision.String()
		repo.OriginalVariable, _, _ = client.Variables().Update(repo.OriginalVariable, &api.WriteOptions{})
		if err != nil {
			logger.Error("failed to update `status_current_commit` field back to Nomad Variables for GitRepository", zap.String("gitRepository", repo.Path))
		}
	}
}
