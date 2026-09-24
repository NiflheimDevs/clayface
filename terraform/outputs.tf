output "vms" {
  description = "Map of VM name to { hypervisor, role, os_family, mac, ip }. hypervisor is the libvirt host alias from lab.yaml; role is the terraform module that built the VM ('gateway' for the OPNsense edge VM, otherwise the `module:` value from lab.yaml - 'dc', 'client', 'alpine', 'app') and names the Ansible playbook that owns it; os_family is 'windows' or 'linux' and decides how Ansible connects; mac is the pinned NIC MAC for modules that derive one; ip is the optional static address declared in lab.yaml. Consumed by the Ansible dynamic inventory."

  value = merge(
    {
      (local.edge_vm_name) = {
        hypervisor = local.edge_host
        role       = "gateway"
        os_family  = "linux"
        mac        = null
        ip         = null
      }
    },
    {
      for name, p in local.vm_placements :
      name => {
        hypervisor = p["host"]

        # role IS the module name. "Which playbook targets this VM" and
        # "which module built it" were always the same question; keeping two
        # vocabularies only produced the dc/windows conflation this replaces.
        role = p["module"]

        # How Ansible connects. Deliberately orthogonal to role: dc and
        # client are both windows, and a future member server would be a
        # third windows role. Reading it from module_os is also the typo
        # guard for `module:` in lab.yaml.
        os_family = local.module_os[p["module"]]

        # Pinned MAC, dc/client/app — extend the merge below when another
        # module needs a DHCP reservation.
        mac = try(
          merge(
            { for n, m in module.dc_host_a : n => m.mac },
            { for n, m in module.dc_host_b : n => m.mac },
            { for n, m in module.client_host_a : n => m.mac },
            { for n, m in module.client_host_b : n => m.mac },
            { for n, m in module.app_host_a : n => m.mac },
            { for n, m in module.app_host_b : n => m.mac },
          )[name],
          null
        )

        # Static address, optional: the `ip:` key in lab.yaml.
        ip = try(p["ip"], null)
      }
    }
  )
}
