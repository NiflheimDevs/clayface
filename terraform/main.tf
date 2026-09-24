# One module block per (module, host) pair, because a provider cannot be
# selected dynamically: adding a hypervisor means adding a provider block in
# provider.tf and a block per module here.
#
# Every block that builds a VM from `vm_placements` passes the same bridge
# expression: the placement's `network:` key (defaulting to `lan`) resolved
# through lab.yaml's `networks:` map. It is written out rather than shared
# because terraform has no way to factor a module argument without inventing
# a wrapper module; the expression is the same three locals everywhere, and
# a typo'd `network:` in lab.yaml fails the plan on every one of them.

module "opnsense_host_a" {
  source = "./modules/opnsense"

  providers = {
    libvirt = libvirt.host_a
  }

  for_each = local.edge_host == "host_a" ? { (local.edge_vm_name) = true } : {}

  name = each.key

  # The edge VM is not in `vm_placements` — `edge.host` names its hypervisor
  # instead — so its segments are named explicitly rather than derived.
  bridge     = local.network_bridges["lan"]
  wan_bridge = local.network_bridges["wan"]
}

module "opnsense_host_b" {
  source = "./modules/opnsense"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = local.edge_host == "host_b" ? { (local.edge_vm_name) = true } : {}

  name = each.key

  bridge     = local.network_bridges["lan"]
  wan_bridge = local.network_bridges["wan"]
}

module "alpine_host_a" {
  source = "./modules/alpine"

  providers = {
    libvirt = libvirt.host_a
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_a" && p["module"] == "alpine"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "alpine_host_b" {
  source = "./modules/alpine"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_b" && p["module"] == "alpine"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "dc_host_a" {
  source = "./modules/domaincontroller"
  providers = {
    libvirt = libvirt.host_a
  }
  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_a" && p["module"] == "dc"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "dc_host_b" {
  source = "./modules/domaincontroller"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_b" && p["module"] == "dc"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "client_host_a" {
  source = "./modules/client"

  providers = {
    libvirt = libvirt.host_a
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_a" && p["module"] == "client"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "client_host_b" {
  source = "./modules/client"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_b" && p["module"] == "client"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "app_host_a" {
  source = "./modules/app"

  providers = {
    libvirt = libvirt.host_a
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_a" && p["module"] == "app"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "app_host_b" {
  source = "./modules/app"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_b" && p["module"] == "app"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}
