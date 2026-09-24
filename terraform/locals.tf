locals {
  lab = yamldecode(file("${path.module}/../lab.yaml"))

  # lab.yaml keeps `vm_placements:` (null when empty); try() covers the key
  # being absent, the ternary covers it being null.
  vm_placements = try(local.lab["vm_placements"] != null ? local.lab["vm_placements"] : {}, {})

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
