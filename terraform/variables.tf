variable "libvirt_hosts" {
  type = map(object({
    address = string
    user    = string
  }))
}

