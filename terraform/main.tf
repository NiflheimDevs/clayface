module "opnsense_host_a" {
  source = "./modules/opnsense"

  providers = {
    libvirt = libvirt.host_a
  }

  for_each = local.edge_host == "host_a" ? { (local.edge_vm_name) = true } : {}

  name = each.key
}

module "opnsense_host_b" {
  source = "./modules/opnsense"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = local.edge_host == "host_b" ? { (local.edge_vm_name) = true } : {}

  name = each.key
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

  name = each.key
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

  name = each.key
}
