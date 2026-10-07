# The pinned MAC of the DMZ NIC (vtnet2). Exported so the Ansible side can
# assert that the interface it is about to configure is the one terraform
# attached: NIC order is device order, so an edit that ever renumbered them
# would silently repoint the firewall rules at a different interface. With the
# MAC checked, that becomes a named failure instead.
output "dmz_mac" {
  description = "Pinned MAC address of the edge VM's DMZ NIC (the third one, vtnet2 in the guest)."
  value       = local.dmz_mac
}

# The pinned MAC of the uplink NIC (vtnet4's device, vtnet3). Exported for the
# DMZ's reason — and one more, because this is the NIC OPNsense has to be told
# about by hand: the identifier the box allocates it (opt1, opt2, ...) is not
# predictable, and `opt1` is already the DMZ. The MAC is the only stable name
# this interface has before anyone has configured it.
output "uplink_mac" {
  description = "Pinned MAC address of the edge VM's uplink NIC (the fourth one, vtnet3 in the guest)."
  value       = local.uplink_mac
}
