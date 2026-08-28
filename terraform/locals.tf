locals {
  lab = yamldecode(file("${path.module}/../lab.yaml"))

  vm_placements = local.lab["vm_placements"]

  edge_host = local.lab["edge"]["host"]

  edge_host_attrs = local.lab["hosts"][local.edge_host]

  edge_vm_name = local.lab["edge_vm"]["name"]
}
