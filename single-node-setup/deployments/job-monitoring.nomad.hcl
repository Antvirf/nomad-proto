# Exactly one `job` block allowd per file
job "monitoring" {
  # Network/datacenter topology
  region      = "global"
  datacenters = ["dc1"]

  # Namespace - similar to Kubernetes; jobs must be unique per namespace
  # Must be created with the CLI, e.g. `nomad namespace apply -description "My testing namespace" testing-namespace`
  # https://developer.hashicorp.com/nomad/tutorials/manage-clusters/namespaces
  namespace = "default"

  # Which Nomad Scheduler type to use
  # https://developer.hashicorp.com/nomad/docs/schedulers
  # Service is a long-lived thing that should never go down, similar to `Deployment` or `StatefulSet` in k8s
  # System are tasks that should run on EVERY node, similar to `DaemonSet` in k8s
  # Batch for one-off jobs, equivalent to Kubernetes `Job`, or perhaps `Pod` that exists on completion
  # System Batch is similar to Batch, but runs once on EVERY node to perhaps make configuration changes etc.
  type = "service"


  # Group defines a series of tasks that should be co-located on the SAME NODE
  # So roughly similar to a `Pod` in terms of colocation on the same node
  # Tasks in the same Nomad Group also share their networking
  # An 'Allocation' in Nomad terms is the declaration that "this group (=this set of tasks) will run on nodes [X,Y]"
  # Scheduling is the process of defining an `Allocation`
  group "prometheus" {

    # Define networking options for the group - everything in the group shares the same network
    network {
      port "promport" {} # Dynamic port
    }

    # Service blocks help route traffic to your application
    # Either via Consul, or via Nomad's native service discovery
    service {
      name     = "prometheus"
      port     = "promport"
      provider = "consul"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.prometheus.rule=Host(`localhost`) && PathPrefix(`/prometheus/`)"
      ]
    }

    # Tasks are the primary 'unit of work' in Nomad
    # Each task can use exactly one driver (docker, java, exec, etc.)
    # Somewhat similar to each `container` in a `Pod` in Kubernetes
    task "promcontainer" {
      driver = "docker"

      # Driver specific config for docker, roughly equivalent to args/configs passed to the Docker Daemon when doing `docker run`
      config {
        image        = "prom/prometheus:v2.53.0" # latest at time of writing
        network_mode = "host"

        # Args passed to the container itself - equivalent to `CMD` of the Docker container
        args = [
          "--web.listen-address=:${NOMAD_PORT_promport}",
          "--web.route-prefix=/prometheus/",
          "--web.external-url=/prometheus/",
          "--config.file=/etc/prometheus/prometheus.yml",
        ]

        # Docker volumes - these are needed to make use of any local files that the container would want to access
        # Usually this will include the templates and config files
        volumes = [
          "local/prometheus.yml:/etc/prometheus/prometheus.yml",
        ]
      }

      # Templating system that allows you to create files for the task, e.g. configs
      template {
        change_mode = "restart" # What to do to this task if the configuration file changes - default is `restart`
        data        = file("./templates-prometheus.yml")
        destination = "local/prometheus.yml"
      }
    }
  }

  group "loki" {

    network {
      # Needs to be static since the Docker Driver needs to know where to look
      port "lokiport" {
        to     = 3100
        static = 3100
      }
    }

    service {
      name     = "loki"
      port     = "lokiport"
      provider = "consul"
    }

    task "loki" {
      driver = "docker"

      config {
        network_mode = "host"
        image        = "grafana/loki:3.0.0" # latest at time of writing
        volumes = [
          "local/loki.yml:/etc/loki/config/loki.yml",
        ]
      }

      template {
        change_mode = "restart" # What to do to this task if the configuration file changes - default is `restart`
        data        = file("./templates-loki.yml")
        destination = "local/loki.yml"
      }
    }
  }

  group "grafana" {

    network {
      # Dynamic port configuration - note that our config file MUST be aware of this port
      port "http" {}
    }

    service {
      name     = "grafana"
      port     = "http"
      provider = "consul"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.grafana.rule=Host(`localhost`) && PathPrefix(`/grafana/`)"
      ]
    }

    task "dashboard" {
      driver = "docker"

      config {
        image        = "grafana/grafana:10.4.4" # latest at time of writing
        network_mode = "host"
        volumes = [
          "local/grafana.ini:/etc/grafana/grafana.ini",
          "local/grafana-datasources.yml:/etc/grafana/provisioning/datasources/nomad-datasources.yml",
        ]
      }

      template {
        change_mode = "restart"
        data        = file("./templates-grafana.ini")
        destination = "local/grafana.ini"
      }

      # Data source templates to declaratively add source to Grafana
      # Docs on Prometheus: https://grafana.com/docs/grafana/latest/datasources/prometheus/
      # Docs on Loki: https://grafana.com/docs/grafana/latest/datasources/loki/
      # Docs on Tempo: https://grafana.com/docs/grafana/latest/datasources/tempo/configure-tempo-data-source/
      template {
        change_mode = "noop"
        data        = file("./templates-grafana-datasources.yml")
        destination = "local/grafana-datasources.yml"
      }
    }
  }
}
