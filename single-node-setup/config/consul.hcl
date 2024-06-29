datacenter = "dc1"
data_dir   = "/tmp/nomad/consul/"
log_level  = "INFO"

client_addr    = "0.0.0.0"
bind_addr      = "0.0.0.0"
advertise_addr = "{{ GetPrivateIP }}"

server           = true
bootstrap_expect = 1

connect {
  enabled = true
}

ui_config {
  enabled = true
}
ui = true

ports {
  dns = 8600
}
