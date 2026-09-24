variable "name" {
  type = string
}

# Bridge device this VM's NIC attaches to. Supplied by the root module out of
# lab.yaml's `networks:` map (locals.tf -> network_bridges), so the bridge
# name lives in exactly one place instead of as a literal in every module.
#
# No default on purpose: a module call that forgets this should fail the plan
# rather than quietly attach the VM to a bridge nobody chose.
variable "bridge" {
  description = "Name of the Linux bridge on the hypervisor that this VM's NIC attaches to."
  type        = string
}

variable "base_image_path" {
  description = "Absolute path of the qcow2 base image on the hypervisor (built per base-image/README.md). Must exist on every host that runs a dc VM."
  type        = string
  default     = "/var/lib/libvirt/images/ws-base.qcow2"
}
