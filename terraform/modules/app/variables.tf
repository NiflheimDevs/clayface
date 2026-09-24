variable "name" {
  type = string
}

# The Ubuntu base image is built by hand (see docs/app01-design.md) and already
# carries Docker and the Docker Compose v2 plugin. It is used only as a
# read-only backing file: the domain boots the per-VM overlay this module
# creates, never the base itself.
#
# The file deliberately has no `.qcow2` extension, so its format cannot be
# inferred from the name — every reference to it declares `qcow2` explicitly.
variable "base_image_path" {
  description = "Absolute path of the Ubuntu qcow2 base image on the hypervisor. Must exist on every host that runs an app VM."
  type        = string
  default     = "/var/lib/libvirt/images/ubuntu24.04-base"
}
