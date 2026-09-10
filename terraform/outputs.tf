output "vms" {
  description = "Map of VM name to { hypervisor, role, mac }. hypervisor is the libvirt host alias from lab.yaml; role is 'gateway' for the OPNsense edge VM, 'dc' for Windows domain controllers and 'linux' for the rest; mac is the pinned NIC MAC (dc VMs) — used for DHCP static mappings and name->IP bootstrap in Ansible. Consumed by the Ansible dynamic inventory."

  value = merge(
    {
      (local.edge_vm_name) = {
        hypervisor = local.edge_host
        role       = "gateway"
        mac        = null
      }
    },
    {
      for name, p in local.vm_placements :
      name => {
        hypervisor = p["host"]
        # "dc" VMs are Windows domain controllers: the Ansible inventory
        # uses this role to group them (vms_dc) and set WinRM connection
        # variables. Everything else is "linux" (SSH).
        role = p["module"] == "dc" ? "dc" : "linux"
        # Pinned MAC of the dc module instances (null for other modules —
        # extend their modules the same way when they need reservations).
        mac = try(
          merge(
            { for n, m in module.dc_host_a : n => m.mac },
            { for n, m in module.dc_host_b : n => m.mac },
          )[name],
          null
        )
      }
    }
  )
}
