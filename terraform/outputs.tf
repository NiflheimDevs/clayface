output "vms" {
  description = "Map of VM name to { hypervisor, role }. hypervisor is the libvirt host alias from lab.yaml; role is 'gateway' for the OPNsense edge VM and 'linux' for the rest. Consumed by the Ansible dynamic inventory."

  value = merge(
    {
      (local.edge_vm_name) = {
        hypervisor = local.edge_host
        role       = "gateway"
      }
    },
    {
      for name, p in local.vm_placements :
      name => {
        hypervisor = p["host"]
        role       = "linux"
      }
    }
  )
}
