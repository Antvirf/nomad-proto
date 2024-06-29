#!/bin/bash
set -e
set -x

# Docker on ubuntu
# Uninstall conflicting
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done

## Installation
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

## Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

## Permissions
sudo usermod -aG docker $USER

## Docker networking - add bridge IP as DNS server
sudo cp config-system/daemon.json /etc/docker/daemon.json
sudo systemctl restart docker

## Loki driver for Docker, so we can use Loki to get logs from all our Docker containers
## https://grafana.com/docs/loki/latest/send-data/docker-driver/
docker plugin install grafana/loki-docker-driver:2.9.2 --alias loki --grant-all-permissions

# Full docs @ https://developer.hashicorp.com/nomad/docs/install
## Nomad/Consul pre-reqs
sudo apt-get update
sudo apt-get install -y software-properties-common
sudo apt-get install -y wget gpg coreutils dnsmasq unzip tree redis-tools jq curl tmux bash-completion

wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor --batch --yes -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list

sudo apt-get update

## Nomad & Consul themselves
sudo apt-get install nomad
sudo apt-get install consul

## Post installation steps for Nomad/Consul as per https://developer.hashicorp.com/nomad/docs/install#post-installation-steps
## Install CNI plugin so we can use bridge network mode
export ARCH_CNI=$( [ $(uname -m) = aarch64 ] && echo arm64 || echo amd64)
curl -L -o cni-plugins.tgz "https://github.com/containernetworking/plugins/releases/download/v1.5.0/cni-plugins-linux-${ARCH_CNI}"-v1.5.0.tgz && \
  sudo mkdir -p /opt/cni/bin && \
  sudo tar -C /opt/cni/bin -xzf cni-plugins.tgz
sudo rm cni-plugins.tgz

## Install consul-cnu plugin so we can use transparent_proxy
curl -L -o consul-cni.zip "https://releases.hashicorp.com/consul-cni/1.5.0/consul-cni_1.5.0_linux_${ARCH_CNI}".zip && \
  sudo unzip -o consul-cni.zip -d /opt/cni/bin -x LICENSE.txt
sudo rm consul-cni.zip

## Configure bridge network routing
echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-arptables && \
  echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-ip6tables && \
  echo 1 | sudo tee /proc/sys/net/bridge/bridge-nf-call-iptables

sudo cp config-system/bridge.conf /etc/sysctl.d/bridge.conf


# Set up DNSMASQ/UFW as per the Nomad walkthrough at https://github.com/schmichael/django-nomadrepo/blob/main/terraform/shared/scripts/setup.sh
sudo ufw disable || echo "ufw not installed"
sudo systemctl disable systemd-resolved.service
sudo cp config-system/dnsmasq.conf /etc/dnsmasq.d/default
sudo chown root:root /etc/dnsmasq.d/default
sudo systemctl restart dnsmasq

echo "Finished"