 # Multi-node Consul + Nomad deployment for Proxmox VE with Terraform

A configurable setup of Consul and Nomad VMs based on Ubuntu 22.04, deployed to a Proxmox VE. Inspired and based heavily on [this excellent resource](https://github.com/groovemonkey/tutorialinux-hashistack/tree/master).

Current setup is not ready for production, several more things to do:

- Set up monitoring similar to the local example
- Set up test workloads and ensure/fix Consul Connect
- Encrypt traffic between Consul nodes
- Encrypt traffic between Nomad nodes
- Consul ACLs
- Nomad ACLs
