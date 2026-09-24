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

  # VM name -> segment name, defaulting to `lan`, so only a VM that is
  # deliberately elsewhere carries a `network:` key. Indexing network_bridges
  # with the result is also the guard: a typo'd `network:` in lab.yaml names a
  # key that does not exist, which is an error, so the plan fails loudly
  # instead of attaching the VM to no bridge. Same trick as module_os.
  placement_network = {
    for name, p in local.vm_placements : name => try(p["network"], "lan")
  }

  # The edge VM is the only guest with more than one leg: `lan` inside and
  # `wan` outside. Not a lab.yaml fact because it is a property of what
  # OPNsense IS here, not a placement choice someone makes.
  gateway_networks = ["lan", "wan"]

  edge_host = local.lab["edge"]["host"]

  # This is the guard for `edge.host`, not dead code. Indexing a map with a
  # key that does not exist is an error, so a typo'd edge.host fails the plan
  # instead of silently dropping the gateway VM. Kept for that reason even
  # though nothing dereferences it.
  edge_host_attrs = local.lab["hosts"][local.edge_host]

  # Every module under terraform/modules/ that a vm_placement may name, and
  # the guest OS family it produces.
  module_os = {
    alpine = "linux"
    app    = "linux"
    dc     = "windows"
    client = "windows"
  }

  # Provider configurations are always evaluated by terraform, even when no
  # resource uses them. A hypervisor commented out of lab.yaml therefore
  # still needs a syntactically valid (never actually used) URI, built from
  # whichever host IS defined.
  libvirt_uri_host = try(local.lab["hosts"]["host_a"], local.lab["hosts"]["host_b"])

  edge_vm_name = local.lab["edge_vm"]["name"]
}
