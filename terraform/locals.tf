locals {
  lab = yamldecode(file("${path.module}/../lab.yaml"))

  # lab.yaml keeps `vm_placements:` (null when empty); try() covers the key
  # being absent, the ternary covers it being null.
  vm_placements = try(local.lab["vm_placements"] != null ? local.lab["vm_placements"] : {}, {})

  edge_host = local.lab["edge"]["host"]

  edge_host_attrs = local.lab["hosts"][local.edge_host]

  # Provider configurations are always evaluated by terraform, even when no
  # resource uses them. A hypervisor commented out of lab.yaml therefore
  # still needs a syntactically valid (never actually used) URI, built from
  # whichever host IS defined.
  libvirt_uri_host = try(local.lab["hosts"]["host_a"], local.lab["hosts"]["host_b"])

  edge_vm_name = local.lab["edge_vm"]["name"]
}
