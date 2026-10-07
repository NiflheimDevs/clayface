output "vms" {
  description = "Map of VM name to { hypervisor, role, os_family, network, bridge, mac, ip }. hypervisor is the libvirt host alias from lab.yaml; role is the terraform module that built the VM ('gateway' for the OPNsense edge VM, otherwise the `module:` value from lab.yaml - 'dc', 'client', 'alpine', 'app') and names the Ansible playbook that owns it; os_family is 'windows' or 'linux' and decides how Ansible connects; network is the lab.yaml segment the VM sits on; bridge is that segment's bridge device, and the gateway additionally carries `bridges`, the list of every leg it is attached to, and `dmz_mac`, the pinned MAC of its DMZ NIC, and `uplink_mac`, the pinned MAC of the NIC that carries it off the lab; mac is the pinned NIC MAC for modules that derive one; ip is the optional static address declared in lab.yaml. Consumed by the Ansible dynamic inventory."

  value = merge(
    {
      (local.edge_vm_name) = {
        hypervisor = local.edge_host
        role       = "gateway"
        os_family  = "linux"
        # The edge VM has no `vm_placements` entry, so it has no `network:`
        # either. Its inside leg is the LAN segment; `bridges` names every leg
        # it holds, in NIC order, which is what lets start_vms.yml precheck
        # all of them.
        network = "lan"
        bridge  = local.network_bridges["lan"]
        # Every leg this VM attaches to, in NIC order, which is what lets
        # start_vms.yml precheck all of them before `virsh start` fails with
        # "Network bridge <name> not found". The three segment legs come from
        # `gateway_networks`; the fourth entry is the uplink, which is not a
        # `networks` entry but IS a bridge the domain needs — and the one the
        # lab does not build, so it is the one most worth checking for.
        bridges = concat(
          [for k in local.gateway_networks : local.network_bridges[k]],
          [local.edge_uplink_bridge],
        )
        mac = null
        ip  = null
        # The DMZ NIC's pinned MAC, so the DMZ playbook can assert that the
        # interface it is configuring is the device terraform attached.
        # See modules/opnsense/outputs.tf.
        dmz_mac = local.edge_host == "host_a" ? try(
          module.opnsense_host_a[local.edge_vm_name].dmz_mac, null
          ) : try(
          module.opnsense_host_b[local.edge_vm_name].dmz_mac, null
        )
        # The uplink NIC's pinned MAC (the fourth, vtnet3), for the same
        # reason and one more: the firewall allocates this interface's name by
        # assignment order, so the MAC is the only stable handle opnsense.yml
        # has on it. See modules/opnsense/outputs.tf.
        uplink_mac = local.edge_host == "host_a" ? try(
          module.opnsense_host_a[local.edge_vm_name].uplink_mac, null
          ) : try(
          module.opnsense_host_b[local.edge_vm_name].uplink_mac, null
        )
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

        # Which segment the VM sits on, and the bridge that segment resolves
        # to. Both are derived from lab.yaml by way of locals.tf, so the
        # answer the playbooks precheck is the same one terraform attached
        # the NIC to.
        network = local.placement_network[name]
        bridge  = local.network_bridges[local.placement_network[name]]

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

output "bridges" {
  description = "Map of lab.yaml segment name to bridge device name. Purely informational: it makes the lab's L2 topology readable from `terraform output` without opening lab.yaml, and it is the single place a consumer can learn every bridge that must exist on a host."

  value = local.network_bridges
}
