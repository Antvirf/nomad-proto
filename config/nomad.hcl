datacenter = "dc1"
data_dir   = "/tmp/nomad"
bind_addr  = "0.0.0.0"

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled = true
}

tls {
  http = true
  rpc  = true

  ca_file   = "nomad-agent-ca.pem"
  cert_file = "global-server-nomad.pem"
  key_file  = "global-server-nomad-key.pem"

  verify_server_hostname = false
  verify_https_client    = false
}

consul {
  # See docs at https://developer.hashicorp.com/nomad/docs/configuration/consul
  address = "127.0.0.1:8500"

  # The below options are only needed if ACLs are enabled
  # In Consul dev mode ACLs are disabled, so configuring these will break stuff between Nomad and Consul
  # auth    = "admin:password"
  # token   = "abcd1234"
}

# monitoring
telemetry {
  collection_interval        = "1s"
  disable_hostname           = true
  prometheus_metrics         = true
  publish_allocation_metrics = true
  publish_node_metrics       = true
}

# Plugins - configuring Docker driver
plugin "docker" {
  config {
    extra_labels = ["job_name", "job_id", "task_group_name", "task_name", "namespace", "node_name", "node_id"]
    logging {
      type = "loki"
      config {
        loki-url = "http://loki.service.consul:3100/loki/api/v1/push"
        labels   = "com.hashicorp.nomad.alloc_id,com.hashicorp.nomad.job_id,com.hashicorp.nomad.job_name,com.hashicorp.nomad.namespace,com.hashicorp.nomad.node_id,com.hashicorp.nomad.node_name,com.hashicorp.nomad.task_group_name,com.hashicorp.nomad.task_name"
      }
    }
  }
}


