# TODO

Deferred fixes. Not blocking current work.

> **Planned, not started:** the two network items below (and the VPN pool) have
> a full implementation plan in [`network-plan.md`](network-plan.md) — the
> `networks:` map, the `vm-br0` → `vm-lan0` rename, the DMZ segment and its
> OPNsense boundary, and the declared VPN pool. Nothing in it is implemented.

## Move hardcoded network facts into lab.yaml

Fix planned in [`network-plan.md`](network-plan.md) Part A.

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

Fix planned in [`network-plan.md`](network-plan.md) Part B. The plan also
declares the remote-access VPN pool (`10.0.100.0/24`) as a fact with no server
behind it — see Part C, and the `VPN_initial_access` correction it requires in
`redclay.yaml`.

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

**Update:** the web-app attack path was built first anyway, deliberately —
APP01 now exists (`docs/app01-design.md`) and sits on the flat L2 with
everything else. The consequence is real and should be stated rather than
hidden: APP01's compromise is currently *equivalent* to an internal
foothold, because there is no boundary to cross from it. Any Ch5 narrative
that treats "DMZ host compromised" as a distinct, weaker position than
"internal host compromised" is wrong until this item is done.
