variable "name" {
  type = string
}

variable "base_image_path" {
  description = "Absolute path of the qcow2 base image on the hypervisor (built per base-image/README.md). Must exist on every host that runs a dc VM."
  type        = string
  default     = "/var/lib/libvirt/images/ws-base.qcow2"
}

output "mac" {
  description = "The pinned MAC address of the VM's NIC. Consumed by the root module's `vms` output so Ansible (and OPNsense static mappings) can tie name -> MAC -> DHCP lease."
  value       = local.dc_mac
}
