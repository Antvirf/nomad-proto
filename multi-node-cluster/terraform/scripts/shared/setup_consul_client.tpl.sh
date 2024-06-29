# Consul client config
mkdir -p /etc/consul.d
cat <<EOF >'/etc/consul.d/consul.hcl'
datacenter = "dc1"
data_dir = "/var/lib/consul"
retry_join = ${CONSUL_SERVER_IPS}
client_addr = "0.0.0.0"
bind_addr = "{{ GetPrivateIP }}"
leave_on_terminate = true
enable_syslog = true
disable_update_check = true
enable_debug = true
EOF

systemctl restart consul
