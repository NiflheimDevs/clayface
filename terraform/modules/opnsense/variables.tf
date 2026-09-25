variable "name" {
  type = string
}

# The edge VM is the only guest with more than one leg, so it takes one bridge
# per segment. All of them come from lab.yaml's `networks:` map by way of the
# root module (locals.tf -> network_bridges); the edge VM has no
# `vm_placements` entry, so main.tf names the segments explicitly rather than
# deriving them.
#
# No defaults on purpose: a module call that forgets one should fail the plan
# rather than quietly attach a NIC to a bridge nobody chose.
variable "bridge" {
  description = "Name of the Linux bridge on the hypervisor for the edge VM's LAN (inside) NIC."
  type        = string
}

variable "wan_bridge" {
  description = "Name of the Linux bridge on the hypervisor for the edge VM's WAN (outside, NAT) NIC."
  type        = string
}

variable "dmz_bridge" {
  description = "Name of the Linux bridge on the hypervisor for the edge VM's DMZ NIC. Appended third, so the guest sees it as vtnet2."
  type        = string
}
