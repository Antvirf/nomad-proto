package main

import "github.com/hashicorp/nomad/api"

type GitRepositoryObjectItems struct {
	ControllerName      string `mapstructure:"controller_name"`
	Url                 string `mapstructure:"url"`
	Branch              string `mapstructure:"branch"`
	RelativePath        string `mapstructure:"relative_path"`
	RegexPathFilter     string `mapstructure:"regex_path_filter"`
	StatusCurrentCommit string `mapstructure:"status_current_commit"`
	Recurse             bool   `mapstructure:"recurse"`
}

type GitRepositoryObject struct {
	OriginalVariable *api.Variable
	Items            GitRepositoryObjectItems `mapstructure:"items"`
	Namespace        string                   `mapstructure:"namespace"`
	Path             string                   `mapstructure:"path"`
}

// Nomad Job object structs
type NomadJobObjectItems struct {
	Spec              string `mapstructure:"spec"`
	Status            string `mapstructure:"status"`
	ControllerName    string `mapstructure:"controller_name"`
	GitRepositoryName string `mapstructure:"git_repository_name"`
}

type NomadJobObject struct {
	OriginalVariable *api.Variable
	Items            NomadJobObjectItems `mapstructure:"items"`
	Namespace        string              `mapstructure:"namespace"`
	Path             string              `mapstructure:"path"`
}
