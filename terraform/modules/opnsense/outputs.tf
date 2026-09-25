# The pinned MAC of the DMZ NIC (vtnet2). Exported so the Ansible side can
# assert that the interface it is about to configure is the one terraform
# attached: NIC order is device order, so an edit that ever renumbered them
# would silently repoint the firewall rules at a different interface. With the
# MAC checked, that becomes a named failure instead.
output "dmz_mac" {
  description = "Pinned MAC address of the edge VM's DMZ NIC (the third one, vtnet2 in the guest)."
  value       = local.dmz_mac
}
