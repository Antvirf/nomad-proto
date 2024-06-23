# Based on https://developer.hashicorp.com/nomad/tutorials/load-balancing/load-balancing-traefik#load-balancing-traefik

job "traefik" {
  region      = "global"
  datacenters = ["dc1"]
  type        = "service"
  namespace   = "default"

  update {
    max_parallel = 0 # force update instead of rolling deployment
  }

  group "traefik" {
    count = 1

    network {
      port "http" {
        static = 8080
      }

      port "api" {
        static = 8081
      }
    }

    service {
      name = "traefik"

      check {
        name     = "alive"
        type     = "tcp"
        port     = "http"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "traefik" {
      driver = "docker"

      config {
        image        = "traefik:v2.2"
        network_mode = "host"

        volumes = [
          "local/traefik.toml:/etc/traefik/traefik.toml",
        ]
      }

      template {
        change_mode = "restart"
        data        = file("./templates-traefik.toml")
        destination = "local/traefik.toml"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
