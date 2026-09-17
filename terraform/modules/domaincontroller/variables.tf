variable "name" {
  type = string
}

variable "base_image_path" {
  description = "Absolute path of the qcow2 base image on the hypervisor (built per base-image/README.md). Must exist on every host that runs a dc VM."
  type        = string
  default     = "/var/lib/libvirt/images/ws-base.qcow2"
}
