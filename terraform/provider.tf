provider "libvirt" {
  alias = "host_a"
  uri   = "qemu+ssh://${local.lab["hosts"]["host_a"]["user"]}@${local.lab["hosts"]["host_a"]["address"]}/system"
}

provider "libvirt" {
  alias = "host_b"
  uri   = "qemu+ssh://${local.lab["hosts"]["host_b"]["user"]}@${local.lab["hosts"]["host_b"]["address"]}/system"
}
