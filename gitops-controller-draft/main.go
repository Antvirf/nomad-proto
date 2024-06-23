package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}

// Externally configurable variables
var (
	SYNC_INTERVAL_CRON   = getEnv("NOMAD_GITOPS_CONTROLLER_SYNC_CRON_EXPRESSION", "*/15 * * * * *")
	ONE_OFF              = getEnv("NOMAD_GITOPS_ONE_OFF", "false")
	controller_name      = getEnv("NOMAD_GITOPS_CONTROLLER_NAME", "nomadops")
	controller_namespace = getEnv("NOMAD_GITOPS_CONTROLLER_NAMESPACE", "default")

	// Internally configurable vars
	NOMAD_VAR_NOMADJOB_PREFIX      = "nomadops/v1/nomadjob/"
	NOMAD_VAR_GITREPOSITORY_PREFIX = "nomadops/v1/gitrepository/"

	// Derived internal vars
	controller_git_clone_base_path = "/local/tmp/nomad/" + controller_name
	logger                         = zap.L()
)

func init() {
	err := os.MkdirAll(controller_git_clone_base_path, os.ModePerm)
	if err != nil {
		fmt.Println("failed to create controller base path")
		panic(err)
	}
}

func main() {
	logger := zap.L()
	// Set up Nomad ClientConfig for all controllers to use
	// This uses the same env vars as the Nomad CLI, so set `env` block in the Nomad job spec accordingly
	clientConfig := api.DefaultConfig()

	// Run the controllers - usually with cron, unless ONE_OFF is set

	if strings.ToLower(ONE_OFF) == "true" {
		ControllerGitRepository(clientConfig)
		ControllerNomadJob(clientConfig)
	} else {
		c := cron.New(cron.WithSeconds())
		c.AddFunc(SYNC_INTERVAL_CRON, func() {
			logger.Info("starting reconciliation loop")
			ControllerGitRepository(clientConfig)
			ControllerNomadJob(clientConfig)
		})
		c.Start()
		select {} // Keeps program running forever
	}
}
