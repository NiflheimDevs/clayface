output "vms" {
  description = "Map of VM name (libvirt domain) to the hypervisor alias it runs on. Consumed by the Ansible dynamic inventory."

  value = {
    (module.alpine_host_a.name)   = { hypervisor = "host_a" }
    (module.opnsense_host_a.name) = { hypervisor = "host_a" }
    (module.alpine_host_b.name)   = { hypervisor = "host_b" }
  }
}
