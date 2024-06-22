job "monitoring" {
  datacenters = ["dc1"]
  type        = "service"

  group "prometheus" {
    network {
      port "promport" {
        static = 9090
        to     = 9090
      } # Free to use any port here, as our references to it from Grafana use dynamic ports
    }
    service {
      name     = "prometheus"
      port     = "promport"
      provider = "consul"
      tags = [
        "traefik.enable=true",
      ]
    }

    task "promcontainer" {
      driver = "docker"
      config {
        image        = "prom/prometheus:v2.53.0" # latest at time of writing
        network_mode = "host"
        volumes = [
          "local/prometheus.yml:/etc/prometheus/prometheus.yml",
        ]
      }
      template {
        change_mode = "noop"
        destination = "local/prometheus.yml"
        data        = <<EOH
---
global:
  scrape_interval:     5s
  evaluation_interval: 5s

scrape_configs:
  - job_name: 'nomad_metrics'
    consul_sd_configs:
    - server: 'consul.service.consul:8500'
      services: ['nomad-client', 'nomad']
    scrape_interval: 5s
    scheme: https
    tls_config:
      insecure_skip_verify: true
    relabel_configs:
    - source_labels: ['__meta_consul_tags']
      regex: '(.*)http(.*)'
      action: keep
    metrics_path: /v1/metrics
    params:
      format: ['prometheus']
EOH
      }
    }
  }

  group "loki" {
    network {
      port "lokiport" {} # use a dynamic port to avoid clashes
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
        // args = [
        //   "--config.file=/etc/loki/config/loki.yml",
        // ]
        volumes = [
          "local/loki.yml:/etc/loki/config/loki.yml",
        ]
      }

      template {
        change_mode = "noop"
        destination = "local/loki.yml"
        data        = <<EOH
---
global:
  scrape_interval:     5s
  evaluation_interval: 5s
EOH
      }
    }
  }

  group "grafana" {
    network {
      port "http" {
        static = 3000
        to     = 3000
      }
    }

    service {
      name     = "grafana"
      port     = "http"
      provider = "consul"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.grafana.rule=Host(`localhost`)"
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
        change_mode = "noop"
        destination = "local/grafana.ini"
        data        = <<EOH
[auth.anonymous]
enabled = true
org_role = Admin
[security]
admin_user=admin
admin_password=admin
EOH
      }
      # Data source templates
      # Docs on Prometheus: https://grafana.com/docs/grafana/latest/datasources/prometheus/
      # Docs on Loki: https://grafana.com/docs/grafana/latest/datasources/loki/
      # Docs on Tempo: https://grafana.com/docs/grafana/latest/datasources/tempo/configure-tempo-data-source/
      template {
        change_mode = "noop"
        destination = "local/grafana-datasources.yml"
        data        = <<EOT
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    # Resolving Prometheus using Nomad's own features, without consul
    url: http://{{ range service "prometheus" }}{{ .Address }}{{ .Port }}{{ end }}
    jsonData:
      httpMethod: POST
      manageAlerts: true

  - name: Loki
    type: loki
    access: proxy
    # Resolving Loki's IP via Consul, but dynamic port via Nomad
    url: http://loki.service.consul:{{ range service "loki" }}{{ .Port }}{{ end }}
    jsonData:
      timeout: 60
      maxLines: 100
EOT
      }
    }
  }
}
