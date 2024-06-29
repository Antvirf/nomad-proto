#!/usr/bin/env bash
set -eo pipefail

${BASE_PACKAGES_SNIPPET}

${DNSMASQ_CONFIG_SNIPPET}

${CONSUL_INSTALL_SNIPPET}

# Add the server config
cat <<EOF > "/etc/consul.d/consul.hcl"
datacenter = "dc1"
server = true
ui = true
bootstrap_expect = 1
data_dir = "/var/lib/consul"
retry_join = []
client_addr = "0.0.0.0"
bind_addr = "{{ GetPrivateIP }}"
leave_on_terminate = true
enable_syslog = true
connect {
  enabled = true
}
EOF

systemctl restart consul

echo "Finished with consul.sh"
