module "alpine_host_a" {
  source = "./modules/alpine"
  providers = {
    libvirt = libvirt.host_a
  }

  name = "alpine01"
}

module "opnsense_host_a" {
  source = "./modules/opnsense"
  providers = {
    libvirt = libvirt.host_a
  }

  name = "opnsense01"
}

module "alpine_host_b" {
  source = "./modules/alpine"
  providers = {
    libvirt = libvirt.host_b
  }

  name = "alpine02"
}


