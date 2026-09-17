output "name" {
  description = "Name of the libvirt domain created by this module"
  value       = libvirt_domain.client.name
}

output "mac" {
  description = "The pinned MAC address of the VM's NIC. Consumed by the root module's `vms` output so Ansible (and OPNsense static mappings) can tie name -> MAC -> DHCP lease."
  value       = local.vm_mac
}
