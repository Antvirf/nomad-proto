job "monitoring" {
  datacenters = ["dc1"]
  type        = "service"

  group "prometheus" {
    network {
      mode = "bridge"
      port "promport" {
        static = 9090
        to     = 9090
      }
    }
    service {
      name         = "prometheus"
      port         = "promport"
      provider     = "consul"
      address_mode = "alloc" # otherwise, Consul resolves this service to 127.0.0.1 instead
    }

    task "promcontainer" {
      driver = "docker"
      config {
        image = "prom/prometheus:v2.53.0" # latest at time of writing
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
      mode = "bridge"
      port "lokiport" {
        static = 3100
        to     = 3100
      }
    }
    service {
      name         = "loki"
      port         = "lokiport"
      provider     = "consul"
      address_mode = "alloc" # otherwise, Consul resolves this service to 127.0.0.1 instead
    }

    task "loki" {
      driver = "docker"
      config {
        image = "grafana/loki:3.0.0" # latest at time of writing
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
      mode = "bridge"
      port "http" {
        static = 3000
        to     = 3000
      }
    }

    service {
      name         = "grafana"
      port         = "http"
      provider     = "consul"
      address_mode = "alloc" # otherwise, Consul resolves this service to 127.0.0.1 instead
    }

    task "dashboard" {
      driver = "docker"
      config {
        image = "grafana/grafana:10.4.4" # latest at time of writing
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
        data        = <<EOH
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus.service.consul:9090
    jsonData:
      httpMethod: POST
      manageAlerts: true

  - name: Loki
    type: loki
    access: proxy
    url: http://loki.service.consul:3100
    jsonData:
      timeout: 60
      maxLines: 100
EOH
      }
    }
  }
}