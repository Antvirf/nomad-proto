package main

import "github.com/hashicorp/nomad/api"

type GitRepositoryObjectItems struct {
	ControllerName      string `hcl:"controller_name"`
	Url                 string `hcl:"url"`
	Type                string `hcl:"type"`
	Branch              string `hcl:"branch"`
	StatusCurrentCommit string `hcl:"status_current_commit"`
}

type GitRepositoryObject struct {
	OriginalVariable *api.Variable
	Items            GitRepositoryObjectItems `hcl:"items"`
	Namespace        string                   `hcl:"namespace"`
	Path             string                   `hcl:"path"`
}

type NomadJobGroupObjectItems struct {
	Spec                         string `hcl:"spec"`
	Status                       string `hcl:"status"`
	ControllerName               string `hcl:"controller_name"`
	GitRepositoryName            string `hcl:"git_repository_name"`
	NomadJobRelativePath         string `hcl:"nomad_job_relative_path"`
	NomadJobRegexPathFilter      string `hcl:"nomad_job_regex_path_filter"`
	NomadJobGroupRelativePath    string `hcl:"nomad_job_group_relative_path"`
	NomadJobGroupRegexPathFilter string `hcl:"nomad_job_group_regex_path_filter"`
}

type NomadJobGroupObject struct {
	OriginalVariable *api.Variable
	Items            NomadJobGroupObjectItems `hcl:"items,block"`
	Namespace        string                   `hcl:"namespace"`
	Path             string                   `hcl:"path"`
}
