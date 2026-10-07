locals {
  lab = yamldecode(file("${path.module}/../lab.yaml"))

  # lab.yaml keeps `vm_placements:` (null when empty); try() covers the key
  # being absent, the ternary covers it being null.
  vm_placements = try(local.lab["vm_placements"] != null ? local.lab["vm_placements"] : {}, {})

  # The network segments, one entry per L2 domain (`networks:` in lab.yaml).
  networks = local.lab["networks"]

  # Segment name -> bridge device name. This is what main.tf hands to each
  # module's `bridge` variable; before it existed the bridge name was a
  # literal repeated in all five modules, with nothing linking the copies to
  # lab.yaml.
  network_bridges = { for k, v in local.networks : k => v["bridge"] }

  # Segment name -> subnet and gateway. These facts are consumed by firewall
  # and Ansible configuration; the guest itself receives its WAN address from
  # OPNsense DHCP.
  #
  # `try` rather than a direct index because `bridge` is still the only key
  # every segment must carry — a segment with no subnet is a legitimate thing
  # to declare, and the WAN leg was exactly that until crusader moved onto it.
  # A missing key yields "" here; the module is what refuses to build a guest
  # from it (a `subnet:` with no "/" fails the plan at the prefix split).
  network_subnets  = { for k, v in local.networks : k => try(v["subnet"], "") }
  network_gateways = { for k, v in local.networks : k => try(v["gateway"], "") }

  # VM name -> segment name, defaulting to `lan`, so only a VM that is
  # deliberately elsewhere carries a `network:` key. Indexing network_bridges
  # with the result is also the guard: a typo'd `network:` in lab.yaml names a
  # key that does not exist, which is an error, so the plan fails loudly
  # instead of attaching the VM to no bridge. Same trick as module_os.
  placement_network = {
    for name, p in local.vm_placements : name => try(p["network"], "lan")
  }

  # The edge VM is the only guest with more than one leg: `lan` inside, `wan`
  # outside, `dmz` for the segment the attacker lands in. Not a lab.yaml fact
  # because it is a property of what OPNsense IS here, not a placement choice
  # someone makes. Order is NIC order — see the append-only note in the
  # opnsense module.
  gateway_networks = ["lan", "wan", "dmz"]

  edge_host = local.lab["edge"]["host"]

  # The edge VM's uplink bridge, from `edge.uplink_bridge` in lab.yaml. It is
  # the device libvirt's stock `default` NAT network creates, so this is the
  # one bridge the lab attaches to that it does not build itself. Indexed
  # rather than try()'d for the same reason `edge.host` is: a lab.yaml without
  # it should fail the plan, not silently build a firewall with no way off the
  # lab.
  edge_uplink_bridge = local.lab["edge"]["uplink_bridge"]

  # This is the guard for `edge.host`, not dead code. Indexing a map with a
  # key that does not exist is an error, so a typo'd edge.host fails the plan
  # instead of silently dropping the gateway VM. Kept for that reason even
  # though nothing dereferences it.
  edge_host_attrs = local.lab["hosts"][local.edge_host]

  # Every module under terraform/modules/ that a vm_placement may name, and
  # the guest OS family it produces.
  module_os = {
    alpine   = "linux"
    app      = "linux"
    crusader = "linux"
    dc       = "windows"
    client   = "windows"
  }

  # Provider configurations are always evaluated by terraform, even when no
  # resource uses them. A hypervisor commented out of lab.yaml therefore
  # still needs a syntactically valid (never actually used) URI, built from
  # whichever host IS defined.
  libvirt_uri_host = try(local.lab["hosts"]["host_a"], local.lab["hosts"]["host_b"])

  edge_vm_name = local.lab["edge_vm"]["name"]
}
