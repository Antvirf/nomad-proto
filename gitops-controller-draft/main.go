package main

import (
	"os"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Externally configurable variables - setup done in `init()` below
	SYNC_INTERVAL_CRON   string
	ONE_OFF              string
	controller_name      string
	controller_namespace string

	// Internally configurable vars
	NOMAD_VAR_NOMADJOB_PREFIX      = "nomadops/v1/nomadjobgroup/"
	NOMAD_VAR_GITREPOSITORY_PREFIX = "nomadops/v1/gitrepository/"

	// Derived internal vars
	controller_git_clone_base_path string
	logger                         = zap.L()
)

func init() {
	// Set up logger
	zap_encoder_config := zap.NewProductionEncoderConfig()
	zap_encoder_config.EncodeTime = zapcore.ISO8601TimeEncoder
	zap_config := zap.NewProductionConfig()
	zap_config.DisableCaller = true
	zap_config.EncoderConfig = zap_encoder_config
	zap.ReplaceGlobals(zap.Must(zap_config.Build()))

	// Set up env vars
	SYNC_INTERVAL_CRON = GetEnv("NOMAD_GITOPS_CONTROLLER_SYNC_CRON_EXPRESSION", "*/15 * * * * *")
	ONE_OFF = GetEnv("NOMAD_GITOPS_ONE_OFF", "false")
	controller_name = GetEnv("NOMAD_GITOPS_CONTROLLER_NAME", "nomadops")
	controller_namespace = GetEnv("NOMAD_GITOPS_CONTROLLER_NAMESPACE", "default")

	// Set up derived internal vars
	controller_git_clone_base_path = "/local/tmp/nomad/" + controller_name

	// Set up controller working directory
	err := os.MkdirAll(controller_git_clone_base_path, os.ModePerm)
	if err != nil {
		logger.Fatal("failed to create controller base path", zap.Error(err))
		panic(err)
	}
}

func main() {
	// Set up Nomad ClientConfig for all controllers to use
	// This uses the same env vars as the Nomad CLI, so set `env` block in the Nomad job spec accordingly
	clientConfig := api.DefaultConfig()

	// Run the controllers - usually with cron, unless ONE_OFF is set

	if strings.ToLower(ONE_OFF) == "true" {
		ControllerGitRepository(clientConfig)
		ControllerNomadJobGroup(clientConfig)
	} else {
		c := cron.New(cron.WithSeconds())
		c.AddFunc(SYNC_INTERVAL_CRON, func() {
			logger.Info("starting reconciliation loop")
			ControllerGitRepository(clientConfig)
			ControllerNomadJobGroup(clientConfig)
		})
		c.Start()
		select {} // Keeps program running forever
	}
}
