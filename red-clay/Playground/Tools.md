
### Packer

A tool by HashiCorp that automates building **machine images**. Instead of manually installing an OS and configuring it, you write a template that says "take this ISO, boot it, run these scripts, export the resulting disk image." You build the image once and use it as the base for spinning up VMs.

For your project: you'd use Packer to build a base Windows Server image with WinRM enabled and VirtIO drivers installed, so you can automate it with Ansible later.

### Terraform

Also HashiCorp. Declarative infrastructure provisioning — you describe what you want (this many VMs, this network, these resources) and Terraform figures out how to create/modify/destroy it. It manages state and idempotency.

The **libvirt provider** is a Terraform plugin that talks to libvirt (the KVM management layer) to provision VMs on KVM hosts.

### Ansible

Configuration management tool. Where Terraform creates the infrastructure, Ansible configures what's running on it. You write **playbooks** (YAML files) that describe the desired state of a machine — install this software, create this AD user, set this registry key. Ansible connects over SSH (Linux) or WinRM (Windows) and executes the steps.

### Proxmox

A free, open-source hypervisor platform that runs on bare metal and provides a management layer over KVM and LXC. It gives you a web UI, a REST API, clustering support, and Terraform provider support. It essentially wraps libvirt/KVM in a manageable platform. Much easier for your use case than managing raw libvirt across multiple hosts.
