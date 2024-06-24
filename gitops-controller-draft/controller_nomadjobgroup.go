package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"
)

func ControllerNomadJobGroup(clientConfig *api.Config) {
	logger.Info("starting controller: NomadJobGroup")

	//  Initialize Nomad Client
	client, err := api.NewClient(clientConfig)
	if err != nil {
		logger.Error("failed to initialize Nomad client", zap.Error(err))
	}

	nomad_jobs := FetchNomadJobGroupsForController(client)
	git_repositories := FetchGitRepositoriesForController(client)

	// NomadJobGroups to more NomadJobGroups / First loop
	for _, job := range nomad_jobs {
		repo, err := GetGitRepositoryForNomadJobGroup(job, &git_repositories)
		if err != nil {
			logger.Error("failed to reconcile NomadJobGroup due to missing repository",
				zap.Error(err),
				zap.String("jobReferenceToGitRepository", job.Items.GitRepositoryName))
		}
		base_path_plus_hash := GetPathForRepository(repo)
		repo_job_path := filepath.Join(base_path_plus_hash, job.Items.NomadJobGroupRelativePath)

		potential_files_to_apply, err := FilterFilePathsFromGivenDirectoryAndRegex(repo_job_path, job.Items.NomadJobGroupRegexPathFilter)
		if err != nil {
			logger.Error("failed to get or filter filepaths from input directory", zap.String("directory", repo_job_path), zap.String("gitRepository", repo.Path), zap.Error(err))
		}

		// Loop through list of files
		for _, nomad_job_group_file_path := range potential_files_to_apply {

			file_contents_bytes, err := os.ReadFile(filepath.Join(repo_job_path, nomad_job_group_file_path.Name()))
			if err != nil {
				logger.Error("failed to read file", zap.Error(err), zap.String("fileName", nomad_job_group_file_path.Name()))
				continue // if we fail to read the file, skip it.
			}

			// Parse bytes to internal HCL
			variable_hcl, diagnostics := hclparse.NewParser().ParseHCL(file_contents_bytes, nomad_job_group_file_path.Name())
			if diagnostics.HasErrors() {
				logger.Error("failed to parse file as HCL Job", zap.String("error", diagnostics.Error()), zap.String("fileName", nomad_job_group_file_path.Name()))
				continue // if we failed to parse, skip this file.
			}

			// Parse internal HCL to something usable, decoding its contents to a NomadJobGroup object
			var nomadjobgroup_hcl NomadJobGroupObject
			decodeDiags := gohcl.DecodeBody(variable_hcl.Body, nil, &nomadjobgroup_hcl)
			if decodeDiags.HasErrors() {
				logger.Error("failed to decode NomadJobGroup HCL file", zap.String("error", decodeDiags.Error()), zap.String("fileName", nomad_job_group_file_path.Name()))

				continue // if we failed to decode its contents, skip this file.
			}

			// Push the job spec to Nomad Variables
			_, _, err = client.Variables().Create(&api.Variable{
				Path:      nomadjobgroup_hcl.Path,
				Namespace: nomadjobgroup_hcl.Namespace,
				Items: api.VariableItems{
					"spec":                              nomadjobgroup_hcl.Items.Spec,
					"status":                            nomadjobgroup_hcl.Items.Status,
					"controller_name":                   nomadjobgroup_hcl.Items.ControllerName,
					"git_repository_name":               nomadjobgroup_hcl.Items.GitRepositoryName,
					"nomad_job_relative_path":           nomadjobgroup_hcl.Items.NomadJobRelativePath,
					"nomad_job_regex_path_filter":       nomadjobgroup_hcl.Items.NomadJobRegexPathFilter,
					"nomad_job_group_relative_path":     nomadjobgroup_hcl.Items.NomadJobGroupRelativePath,
					"nomad_job_group_regex_path_filter": nomadjobgroup_hcl.Items.NomadJobGroupRegexPathFilter,
				},
			}, &api.WriteOptions{})
			if err != nil {
				logger.Error("failed to create NomadJobGroup variable", zap.Error(err))
			}
			logger.Info("successfully created/updated NomadJobGroup variable", zap.String("nomadJobGroup", nomadjobgroup_hcl.Path))
		}
	}

	// NomadJobGroup to Nomad Jobs / Main loop - get the repo for this job, find the file(s), apply the jobs
	for _, job := range nomad_jobs {
		repo, err := GetGitRepositoryForNomadJobGroup(job, &git_repositories)
		if err != nil {
			logger.Error("failed to reconcile NomadJobGroup due to missing repository",
				zap.Error(err),
				zap.String("jobReferenceToGitRepository", job.Items.GitRepositoryName))
		}

		base_path_plus_hash := GetPathForRepository(repo)
		repo_job_path := filepath.Join(base_path_plus_hash, job.Items.NomadJobRelativePath)

		potential_files_to_apply, err := FilterFilePathsFromGivenDirectoryAndRegex(repo_job_path, job.Items.NomadJobRegexPathFilter)
		if err != nil {
			logger.Error("failed to get or filter filepaths from input directory", zap.String("directory", repo_job_path), zap.String("gitRepository", repo.Path), zap.Error(err))
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

			// Add meta information to each Job
			job_hcl.SetMeta("nomad_gitops_managed", "true")
			job_hcl.SetMeta("nomad_gitops_current_commit", repo.Items.StatusCurrentCommit)
			job_hcl.SetMeta("nomad_gitops_last_reconciliation_timestamp", time.Now().Format(time.RFC3339))
			job_hcl.SetMeta("nomad_gitops_nomad_job_group", job.Path)
			job_hcl.SetMeta("nomad_gitops_git_repository", repo.Path)
			job_hcl.SetMeta("nomad_gitops_controller_name", controller_name)
			job_hcl.SetMeta("nomad_gitops_controller_namespace", controller_namespace)

			hcl_job_specs = append(hcl_job_specs, job_hcl)
		}

		// Go through HCL job specs, register each job
		for _, job_spec := range hcl_job_specs {
			register_result, _, err := client.Jobs().Register(job_spec, &api.WriteOptions{})
			if err != nil {
				logger.Error("failed to register job", zap.String("jobName", *job_spec.Name), zap.Error(err))
			}
			logger.Info("registered job successfully", zap.String("jobName", *job_spec.Name), zap.String("evalId", register_result.EvalID))
		}
	}
}
