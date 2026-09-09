provider "libvirt" {
  alias = "host_a"
  uri   = "qemu+ssh://${try(local.lab["hosts"]["host_a"]["user"], local.libvirt_uri_host["user"])}@${try(local.lab["hosts"]["host_a"]["address"], local.libvirt_uri_host["address"])}/system"
}

provider "libvirt" {
  alias = "host_b"
  uri   = "qemu+ssh://${try(local.lab["hosts"]["host_b"]["user"], local.libvirt_uri_host["user"])}@${try(local.lab["hosts"]["host_b"]["address"], local.libvirt_uri_host["address"])}/system"
}
