package main

import "github.com/hashicorp/nomad/api"

// Structs

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

// Interfaces

type ControllerObject interface {
	GetPath() string
	GetNamespace() string
	GetControllerName() string
}

// Functions

func (obj NomadJobGroupObject) GetPath() string           { return obj.Path }
func (obj GitRepositoryObject) GetPath() string           { return obj.Path }
func (obj NomadJobGroupObject) GetNamespace() string      { return obj.Namespace }
func (obj GitRepositoryObject) GetNamespace() string      { return obj.Namespace }
func (obj NomadJobGroupObject) GetControllerName() string { return obj.Items.ControllerName }
func (obj GitRepositoryObject) GetControllerName() string { return obj.Items.ControllerName }

func (nomad_job_group_object NomadJobGroupObject) ConvertToNomadVariable() *api.Variable {
	return &api.Variable{
		Path:      nomad_job_group_object.Path,
		Namespace: nomad_job_group_object.Namespace,
		Items: api.VariableItems{
			"spec":                              nomad_job_group_object.Items.Spec,
			"status":                            nomad_job_group_object.Items.Status,
			"controller_name":                   nomad_job_group_object.Items.ControllerName,
			"git_repository_name":               nomad_job_group_object.Items.GitRepositoryName,
			"nomad_job_relative_path":           nomad_job_group_object.Items.NomadJobRelativePath,
			"nomad_job_regex_path_filter":       nomad_job_group_object.Items.NomadJobRegexPathFilter,
			"nomad_job_group_relative_path":     nomad_job_group_object.Items.NomadJobGroupRelativePath,
			"nomad_job_group_regex_path_filter": nomad_job_group_object.Items.NomadJobGroupRegexPathFilter,
		},
	}
}
