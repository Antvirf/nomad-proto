job "nomadops" {
  region      = "global"
  datacenters = ["dc1"]
  type        = "service"
  namespace   = "default"

  group "nomadops" {
    network {}

    task "nomadops" {
      driver = "raw_exec"

      config {
        command = "/home/ubuntu/go/bin/nomadops"
      }

      env {
        // Standard Nomad CLI/API ENV vars
        NOMAD_ADDR                   = "https://localhost:4646"
        NOMAD_SKIP_VERIFY            = "true"

        // Custom env vars/configs for the operator
        NOMAD_GITOPS_CONTROLLER_NAME = "nomadops" // configurable in case multiple controllers are desired
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
