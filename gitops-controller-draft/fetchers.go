package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"
)

func ExpandVariables(client *api.Client, variablemetadata []*api.VariableMetadata) (variables []api.Variable) {
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

func FetchNomadJobGroupsForController(client *api.Client) (controller_relevant_nomad_job_objects []NomadJobGroupObject) {
	variablemetadata, _, err := client.Variables().List(&api.QueryOptions{
		Prefix: NOMAD_VAR_NOMADJOB_PREFIX,
	})
	if err != nil {
		logger.Error("failed to fetch variablemetadata from Nomad",
			zap.Error(err),
		)
		panic(err)
	}
	logger.Info("successfully fetched variables list from Nomad for NomadJobGroups")

	variables := ExpandVariables(client, variablemetadata)
	nomad_job_objects := ConvertVariableToNomadJobGroupStruct(variables)
	controller_relevant_nomad_job_objects = FilterObjectForController(nomad_job_objects)
	return
}

func FetchGitRepositoriesForController(client *api.Client) (controller_relevant_gitrepo_objects []GitRepositoryObject) {
	variablemetadata, _, err := client.Variables().List(&api.QueryOptions{
		Prefix: NOMAD_VAR_GITREPOSITORY_PREFIX,
	})
	if err != nil {
		logger.Error("failed to fetch variablemetadata from Nomad",
			zap.Error(err),
		)
		panic(err)
	}
	logger.Info("successfully fetched variables list from Nomad for GitRepositories")

	variables := ExpandVariables(client, variablemetadata)
	nomad_gitrepo_objects := ConvertVariableToGitRepositoryStruct(variables)
	controller_relevant_gitrepo_objects = FilterObjectForController(nomad_gitrepo_objects)
	return
}

func FilterObjectForController[T ControllerObject](objects []T) (filtered_objects []T) {
	if len(objects) == 0 {
		return
	} // do not process an empty list further

	object_type := GetObjectNameFromVariablePath(objects[0].GetPath())

	logger.Debug(fmt.Sprintf("filtering %s for controller relevance", object_type),
		zap.String("controllerNamespace", controller_namespace),
		zap.String("controllerName", controller_name),
	)

	for _, object := range objects {
		if (object.GetControllerName() == controller_name) && (object.GetNamespace() == controller_namespace) {
			logger.Info(fmt.Sprintf("accepting %s as it matches controller name and/or namespace", object_type),
				zap.String("variablePath", object.GetPath()),
				zap.String("variableNamespace", object.GetNamespace()),
			)
			filtered_objects = append(filtered_objects, object)
		} else {
			logger.Warn(
				fmt.Sprintf("skipping %s as it does not match given controller name and/or namespace", object_type),
				zap.String("variablePath", object.GetPath()),
				zap.String("variableNamespace", object.GetNamespace()),
			)
		}
	}
	logger.Info(fmt.Sprintf("%s filtering complete", object_type),
		zap.Int(fmt.Sprintf("%sToProcess", object_type), len(filtered_objects)),
	)
	return
}
