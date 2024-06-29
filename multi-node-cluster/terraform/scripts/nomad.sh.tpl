#!/usr/bin/env bash
set -eo pipefail

${BASE_PACKAGES_SNIPPET}

# Add nomad Apt repository
curl -fsSL https://apt.releases.hashicorp.com/gpg | apt-key add -
apt-add-repository -y "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
apt-get update

# Add docker (to enable nomad docker driver)
apt-get -y install docker.io

# Install docker loki driver
docker plugin install grafana/loki-docker-driver:2.9.2 --alias loki --grant-all-permissions


${DNSMASQ_CONFIG_SNIPPET}

${CONSUL_INSTALL_SNIPPET}

${CONSUL_CLIENT_CONFIG_SNIPPET}

${NOMAD_INSTALL_SNIPPET}

${CONSUL_TPL_INSTALL_SNIPPET}

echo "Done with nomad.sh"
