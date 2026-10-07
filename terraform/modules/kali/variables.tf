variable "name" {
  type = string
}

# Bridge device this VM's NIC attaches to, supplied by the root module out of
# lab.yaml's `networks:` map (locals.tf -> network_bridges). For crusader that
# is the WAN bridge: the attacker box sits outside the boundary rather than on
# the segment it is attacking, which is the point of the placement.
#
# No default on purpose: a module call that forgets this should fail the plan
# rather than quietly attach the VM to a bridge nobody chose.
variable "bridge" {
  description = "Name of the Linux bridge on the hypervisor that this VM's NIC attaches to."
  type        = string
}

# Crusader uses an ordinary OPNsense DHCP lease on the WAN leg. Its MAC is not
# pinned, so the lease is intentionally not a Terraform-managed DNS reservation.
# The Kali base image is built by hand and is used only as a read-only backing
# file: the domain boots the per-VM overlay this module creates, never the base
# itself. Booting a base directly corrupts every overlay stacked on it, which is
# why the builder domain must be undefined once the base is published.
#
# The base image must boot with a normal DHCP-capable network manager and an
# Ansible-reachable SSH account. OPNsense supplies the lease and DNS binding;
# this module deliberately attaches no cloud-init seed.
variable "base_image_path" {
  description = "Absolute path of the Kali qcow2 base image on the hypervisor. Must exist on every host that runs a crusader VM."
  type        = string
  default     = "/var/lib/libvirt/images/KALI-base.qcow2"
}
