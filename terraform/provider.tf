provider "libvirt" {
  alias = "host_a"
  uri   = "qemu+ssh://${var.libvirt_hosts["host_a"].user}@${var.libvirt_hosts["host_a"].address}/system"
}

provider "libvirt" {
  alias = "host_b"
  uri   = "qemu+ssh://${var.libvirt_hosts["host_b"].user}@${var.libvirt_hosts["host_b"].address}/system"
}
