package main

import (
	"os"

	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"
)

func ExpandVariables(client *api.Client, variablemetadata []*api.VariableMetadata) (variables []api.Variable) {
	logger = zap.L()
	for _, v := range variablemetadata {
		variable_items, _, err := client.Variables().GetVariableItems(v.Path, &api.QueryOptions{})
		if err != nil {
			logger.Error("failed to fetch variable items from Nomad",
				zap.String("variablePath", v.Path),
				zap.String("variableNamespace", v.Namespace),
				zap.Error(err))
			os.Exit(1)
		}

		variable := api.Variable{
			Namespace:   v.Namespace,
			Path:        v.Path,
			CreateTime:  v.CreateTime,
			CreateIndex: v.CreateIndex,
			ModifyTime:  v.ModifyTime,
			ModifyIndex: v.ModifyIndex,
			Items:       variable_items,
			Lock:        v.Lock,
		}
		variables = append(variables, variable)
	}
	return
}

func FetchNomadJobGroupsForController(client *api.Client) []NomadJobGroupObject {
	logger = zap.L()
	variablemetadata, _, err := client.Variables().List(&api.QueryOptions{
		Prefix: NOMAD_VAR_NOMADJOB_PREFIX,
	})
	if err != nil {
		logger.Error("failed to fetch variablemetadata from Nomad",
			zap.Error(err),
		)
		os.Exit(1)
	}
	logger.Info("successfully fetched variables list from Nomad for NomadJobGroups")

	// Epxand variable metadata so that we have the content (`Items`) within each object
	variables := ExpandVariables(client, variablemetadata)

	// Convert api.Variables to NomadJobGroupObjects for further processing
	nomad_job_objects := ConvertVariableToNomadJobGroupStruct(variables)

	// Filter out all jobspecs to check based on controllername
	logger.Info("filtering NomadJobGroups for controller relevance",
		zap.String("controllerNamespace", controller_namespace),
		zap.String("controllerName", controller_name),
	)

	nomad_job_objects_relevant_for_controller := []NomadJobGroupObject{}

	for _, object := range nomad_job_objects {
		if (object.Items.ControllerName == controller_name) && (object.Namespace == controller_namespace) {
			logger.Info("accepting NomadJobGroup as it matches controller name and/or namespace",
				zap.String("variablePath", object.Path),
				zap.String("variableNamespace", object.Namespace),
			)
			nomad_job_objects_relevant_for_controller = append(nomad_job_objects_relevant_for_controller, object)
		} else {
			logger.Warn(
				"skipping NomadJobGroup as it does not match given controller name and/or namespace",
				zap.String("variablePath", object.Path),
				zap.String("variableNamespace", object.Namespace),
			)
		}
	}

	logger.Info("NomadJobGroup filtering complete",
		zap.Int("nomadJobGroupsToProcess", len(nomad_job_objects_relevant_for_controller)),
	)
	return nomad_job_objects_relevant_for_controller
}

func FetchGitRepositoriesForController(client *api.Client) []GitRepositoryObject {
	variablemetadata, _, err := client.Variables().List(&api.QueryOptions{
		Prefix: NOMAD_VAR_GITREPOSITORY_PREFIX,
	})
	if err != nil {
		logger.Error("failed to fetch variablemetadata from Nomad",
			zap.Error(err),
		)
		os.Exit(1)
	}
	logger.Info("successfully fetched variables list from Nomad for GitRepositories")

	// Epxand variable metadata so that we have the content (`Items`) within each object
	variables := ExpandVariables(client, variablemetadata)

	// Convert api.Variables to NomadJobGroupObjects for further processing
	nomad_gitrepo_objects := ConvertVariableToGitRepositoryStruct(variables)

	// Filter out all jobspecs to check based on controllername
	logger.Info("filtering GitRepositories for controller relevance",
		zap.String("controllerNamespace", controller_namespace),
		zap.String("controllerName", controller_name),
	)

	gitrepository_objects_relevant_for_controller := []GitRepositoryObject{}

	for _, object := range nomad_gitrepo_objects {
		if (object.Items.ControllerName == controller_name) && (object.Namespace == controller_namespace) {
			logger.Info("accepting GitRepository as it matches controller name and/or namespace",
				zap.String("variablePath", object.Path),
				zap.String("variableNamespace", object.Namespace),
			)
			gitrepository_objects_relevant_for_controller = append(gitrepository_objects_relevant_for_controller, object)
		} else {
			logger.Warn(
				"skipping GitRepository as it does not match given controller name and/or namespace",
				zap.String("variablePath", object.Path),
				zap.String("variableNamespace", object.Namespace),
			)
		}
	}

	logger.Info("GitRepository filtering complete",
		zap.Int("gitRepositories", len(gitrepository_objects_relevant_for_controller)),
	)
	return gitrepository_objects_relevant_for_controller
}
