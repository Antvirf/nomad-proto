data "local_file" "ssh_public_key" {
  filename = "${path.root}/../../../../.ssh/id_rsa.pub"
}


resource "proxmox_virtual_environment_download_file" "ubuntu_cloud_image" {
  content_type = "iso"
  datastore_id = "local"
  node_name    = "pve"
  url          = "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img"
}

