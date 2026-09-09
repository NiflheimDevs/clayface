# TODO

Deferred fixes. Not blocking current work.

## Move hardcoded network facts into lab.yaml

The single-source-of-truth rule says static lab facts live only in
`lab.yaml`, but a few have leaked into playbook vars:

- `ansible/playbooks/start_vms.yml`
  - `lab_dns_ip: 10.0.0.1` — the lab DNS (OPNsense) address. The whole
    `10.0.0.0/24` lab subnet is a fact and exists nowhere in lab.yaml.
  - `lab_dns_port: 53`
  - `vm_bridge_name: vm-br0` — duplicates `network.bridge_name` from
    lab.yaml (which `lab_inventory.py` already distributes as a host var;
    the playbook should use that var instead of redefining it).
- `ansible/playbooks/hosts.yml`
  - Monitor IP `10.0.0.2/24` added to the bridge (task itself is marked
    "temp") — same subnet fact, hardcoded.

Fix: add the lab subnet / DNS server address to `network` in lab.yaml,
have the inventories pass them through as host vars, and drop the
playbook-local copies.

Not in scope of this fix (deliberate exceptions, see CLAUDE.md):
- `LAB_DOMAIN` default `clayface` in `inventory/terraform_vms.py` —
  owned by the OPNsense image, mirrored there on purpose.

## Add network segmentation (DMZ vs internal zones)

redclay.yaml describes conceptual zones (external / DMZ / internal), but
the actual lab is one flat L2 (`vm-br0`) with every VM on it. That makes
"compromise DMZ host, pivot to internal" trivial — no boundary exists to
cross, and firewall-rule evasion is not modeled.

Fix: a second bridge (or VLAN) on OPNsense. APP01 (public web) in the DMZ
segment; DC01 / CLIENT01 / IDP01 in the internal segment. OPNsense routes
and filters between them.

Work involved:
- second bridge + VXLAN on the hypervisors (hosts.yml pattern repeats)
- second interface on the APP01 VM in its terraform module
- OPNsense interface/firewall rules for the DMZ↔internal boundary

High realism per unit of effort — do this before building the web-app
attack path, not after.
