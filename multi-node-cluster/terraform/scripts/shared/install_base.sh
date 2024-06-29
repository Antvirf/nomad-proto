DEBIAN_FRONTEND=noninteractive

echo "CONFIGS: Installing shared stuff"
apt-get update
apt-get upgrade -y

apt-get update
sleep 5

apt-get install -y wget unzip curl
