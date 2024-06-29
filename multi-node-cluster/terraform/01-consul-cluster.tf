## Render consul installation script from templates
data "template_file" "consul_server_userdata" {
  template = file("${path.module}/scripts/consul.sh.tpl")
  vars = {
    BASE_PACKAGES_SNIPPET  = file("${path.module}/scripts/shared/install_base.sh")
    DNSMASQ_CONFIG_SNIPPET = file("${path.module}/scripts/shared/install_dnsmasq.sh")
    CONSUL_INSTALL_SNIPPET = file("${path.module}/scripts/shared/install_consul.sh")
  }
}


resource "proxmox_virtual_environment_file" "consul_vm_cloud_config" {
  count        = var.consul_node_count
  content_type = "snippets"
  datastore_id = "snippets"
  node_name    = "pve"

  source_raw {
    data = <<-EOF
    #cloud-config
    hostname: consul-vm-${count.index + 1}
    users:
      - default
      - name: ubuntu
        groups:
          - sudo
        shell: /bin/bash
        ssh_authorized_keys:
          - ${trimspace(data.local_file.ssh_public_key.content)}
        sudo: ALL=(ALL) NOPASSWD:ALL

    write_files:
    - encoding: b64
      content: ${base64encode(data.template_file.consul_server_userdata.rendered)}
      owner: root:root
      path: /etc/setup-consul.sh
      permissions: '0755'

    runcmd:
        - apt update
        - apt install -y qemu-guest-agent net-tools
        - systemctl enable qemu-guest-agent
        - systemctl start qemu-guest-agent
        - sudo bash /etc/setup-consul.sh
        - echo "done" > /tmp/cloud-config.done
    EOF

    file_name = "consul-vm-${count.index + 1}-cloud-config.yaml"
  }
}

resource "proxmox_virtual_environment_vm" "consul_vm" {
  count     = var.consul_node_count
  name      = "consul-ubuntu-${count.index + 1}"
  node_name = "pve"

  agent {
    enabled = true
  }

  cpu {
    cores = 2
  }

  memory {
    dedicated = 4096
  }

  disk {
    datastore_id = "local-lvm"
    file_id      = proxmox_virtual_environment_download_file.ubuntu_cloud_image.id
    interface    = "virtio0"
    iothread     = true
    discard      = "on"
    size         = 32
  }

  initialization {
    ip_config {
      ipv4 {
        address = "dhcp"
      }
    }

    user_data_file_id = proxmox_virtual_environment_file.consul_vm_cloud_config[count.index].id
  }

  network_device {
    bridge = "vmbr0"
  }
}


output "consul_node_ip_addresses" {
  value = proxmox_virtual_environment_vm.consul_vm[*].ipv4_addresses[1][0]
}

## Consul join - launch 3 min after creating each vm
## The delay only works if starting entire clusters from scratch correctly; adding new nodes will not execute the delay again
resource "time_sleep" "wait_for_consul" {
  depends_on = [resource.proxmox_virtual_environment_vm.consul_vm]
  lifecycle {
    replace_triggered_by = [
      resource.proxmox_virtual_environment_vm.consul_vm
    ]
  }
  create_duration = "180s"
}
resource "ssh_resource" "consul_join_cluster" {
  depends_on  = [time_sleep.wait_for_consul]
  count       = var.consul_node_count
  host        = resource.proxmox_virtual_environment_vm.consul_vm[count.index].ipv4_addresses[1][0]
  user        = "ubuntu"
  agent       = true
  timeout     = "30s"
  retry_delay = "5s"
  commands = [
    # Connect to each separate VM, and try to join to node #1
    "consul join ${resource.proxmox_virtual_environment_vm.consul_vm[0].ipv4_addresses[1][0]}"
  ]
}
