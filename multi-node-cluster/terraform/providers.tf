terraform {
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "0.60.1"
    }
    ssh = {
      source  = "loafoe/ssh"
      version = "2.7.0"
    }
  }
}

provider "proxmox" {
  # Authenticates via env vars
  # PROXMOX_VE_USERNAME
  # PROXMOX_VE_PASSWORD

  insecure = true
  endpoint = "https://pve.local.aviitala.cloud:8006/"
}
