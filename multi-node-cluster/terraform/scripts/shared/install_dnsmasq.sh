#!/bin/bash
apt-get install -y dnsmasq

echo "Setting up dnsmasq"

cat <<EOF > "/etc/dnsmasq.conf"
listen-address=127.0.0.1
port=53
no-negcache
EOF

mkdir -p /etc/dnsmasq.d


cat <<EOF > "/etc/dnsmasq.d/10-consul"
# Enable forward lookup of the 'consul' domain:
server=/consul/127.0.0.1#8600
EOF

# Get rid of systemd
systemctl disable --now systemd-resolved

# Set up nameservers - overwrite existing
echo "nameserver 127.0.0.1" > /etc/resolv.conf
echo "nameserver 8.8.8.8" >> /etc/resolv.conf

systemctl restart dnsmasq
systemctl enable dnsmasq
