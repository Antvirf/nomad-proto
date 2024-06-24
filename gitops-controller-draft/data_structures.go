package main

import "github.com/hashicorp/nomad/api"

type GitRepositoryObjectItems struct {
	ControllerName      string `mapstructure:"controller_name"`
	Url                 string `mapstructure:"url"`
	Branch              string `mapstructure:"branch"`
	StatusCurrentCommit string `mapstructure:"status_current_commit"`
}

type GitRepositoryObject struct {
	OriginalVariable *api.Variable
	Items            GitRepositoryObjectItems `mapstructure:"items"`
	Namespace        string                   `mapstructure:"namespace"`
	Path             string                   `mapstructure:"path"`
}

type NomadJobGroupObjectItems struct {
	Spec              string `mapstructure:"spec"`
	Status            string `mapstructure:"status"`
	ControllerName    string `mapstructure:"controller_name"`
	GitRepositoryName string `mapstructure:"git_repository_name"`
	RelativePath      string `mapstructure:"relative_path"`
	RegexPathFilter   string `mapstructure:"regex_path_filter"`
	Recurse           bool   `mapstructure:"recurse"`
}

type NomadJobGroupObject struct {
	OriginalVariable *api.Variable
	Items            NomadJobGroupObjectItems `mapstructure:"items"`
	Namespace        string                   `mapstructure:"namespace"`
	Path             string                   `mapstructure:"path"`
}
