package main

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"
)

func ControllerNomadJob(clientConfig *api.Config) {
	logger = zap.L()
	logger.Info("starting controller: NomadJob")

	//  Initialize Nomad Client
	client, err := api.NewClient(clientConfig)
	if err != nil {
		logger.Error("failed to initialize Nomad client", zap.Error(err))
	}

	nomad_jobs := FetchNomadJobsForController(client)
	git_repositories := FetchGitRepositoriesForController(client)

	// Main loop - get the repo for this job, find the file(s), apply the jobs
	for _, job := range nomad_jobs {
		repo, err := GetGitRepositoryForNomadJob(job, &git_repositories)
		if err != nil {
			logger.Error("failed to reconcile NomadJob due to missing repository",
				zap.Error(err),
				zap.String("jobReferenceToGitRepository", job.Items.GitRepositoryName))
		}

		base_path_plus_hash := GetPathForRepository(repo)
		repo_job_path := filepath.Join(base_path_plus_hash, repo.Items.RelativePath)
		files_in_dir, err := os.ReadDir(
			repo_job_path,
		)
		if (err != nil) || len(files_in_dir) == 0 {
			logger.Error("given directory is either unavailable or empty", zap.String("directory", repo_job_path), zap.Error(err))
		}

		// Filter filepaths with given regex
		potential_files_to_apply := []os.DirEntry{}
		for _, file_path := range files_in_dir {
			matches_path_filter, err := regexp.MatchString(repo.Items.RegexPathFilter, file_path.Name())
			if err != nil {
				logger.Error("failed to regex match path filter to a filename",
					zap.Error(err), zap.String("fileName", file_path.Name()),
				)
			}
			if matches_path_filter {
				logger.Info("file found matching filter", zap.String("gitRepository", repo.Path), zap.String("fileName", file_path.Name()))
				potential_files_to_apply = append(potential_files_to_apply, file_path)
			}
		}

		// Go through the job files, parse the HCL and add to next list if valid
		hcl_job_specs := []*api.Job{}
		for _, job_spec_file := range potential_files_to_apply {
			file_contents_bytes, err := os.ReadFile(filepath.Join(repo_job_path, job_spec_file.Name()))
			if err != nil {
				logger.Error("failed to read file", zap.Error(err), zap.String("fileName", job_spec_file.Name()))
			}
			job_hcl, err := client.Jobs().ParseHCL(string(file_contents_bytes), true)
			if err != nil {
				logger.Error("failed to parse file as HCL Job", zap.Error(err), zap.String("fileName", job_spec_file.Name()))
				continue
			}
			logger.Info("successfully parsed Job specification", zap.String("fileName", job_spec_file.Name()))
			hcl_job_specs = append(hcl_job_specs, job_hcl)
		}

		// Go through HCL job specs, register each job
		for _, job_spec := range hcl_job_specs {
			logger.Info(*job_spec.Name)
			register_result, _, err := client.Jobs().Register(job_spec, &api.WriteOptions{})
			if err != nil {
				logger.Error("failed to register job", zap.String("jobName", *job_spec.Name), zap.Error(err))
			}
			logger.Info("registered job successfully", zap.String("jobName", *job_spec.Name), zap.String("evalId", register_result.EvalID))
		}
	}
}