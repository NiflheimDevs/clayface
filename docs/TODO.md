# TODO

Deferred fixes. Not blocking current work.

> **Both network items below are now done** (2026-09-25), along with the VPN
> pool declaration. [`network-plan.md`](network-plan.md) is the plan they were
> executed from; [`network-design.md`](network-design.md) is the design they
> produced and is the current authority on the zones and the boundary.

## Move hardcoded network facts into lab.yaml — DONE

Closed 2026-09-25 by [`network-plan.md`](network-plan.md) Part A.

The single-source-of-truth rule says static lab facts live only in
`lab.yaml`, but a few had leaked into playbook vars:

- `ansible/playbooks/start_vms.yml`
  - `lab_dns_ip: 10.0.0.1` — the lab DNS (OPNsense) address. The whole
    `10.0.0.0/24` lab subnet is a fact and existed nowhere in lab.yaml.
  - `lab_dns_port: 53`
  - `vm_bridge_name: vm-br0` — duplicated `network.bridge_name` from
    lab.yaml (which `lab_inventory.py` already distributes as a host var;
    the playbook should use that var instead of redefining it).
- `ansible/playbooks/hosts.yml`
  - Monitor IP `10.0.0.2/24` added to the bridge (task itself is marked
    "temp") — same subnet fact, hardcoded.

What landed: `lab.yaml` gained a `networks:` map (one entry per L2 domain,
each with `bridge` and optionally `subnet` / `gateway` / `dns` / `monitor` /
`vxlan_id`), read by terraform through `locals.tf` and by Ansible through
`lab_inventory.py`, which hands every host the whole map as `lab_networks`.
`start_vms.yml`, `hosts.yml` and `edge.yml` now loop it instead of carrying
literals. The duplicate is gone; "the lab subnet" is no longer a fact that
exists in prose.

Not in scope of this fix (deliberate exceptions, see CLAUDE.md):
- `LAB_DOMAIN` default `clayface` in `inventory/terraform_vms.py` —
  owned by the OPNsense image, mirrored there on purpose.

## Add network segmentation (DMZ vs internal zones) — DONE

Closed 2026-09-25 by [`network-plan.md`](network-plan.md) Part B; the design
it produced is [`network-design.md`](network-design.md). The same wave
declared the remote-access VPN pool (`10.0.100.0/24`) as a fact with no server
behind it — see Part C, and the `VPN_initial_access` correction in
`redclay.yaml`.

redclay.yaml describes conceptual zones (external / DMZ / internal), but
the actual lab was one flat L2 (`vm-br0`) with every VM on it. That made
"compromise DMZ host, pivot to internal" trivial — no boundary existed to
cross, and firewall-rule evasion was not modeled.

What landed: the `networks:` map's second segment (`dmz`, `10.0.10.0/24`,
bridge `vm-dmz0`, VXLAN 101), a third NIC on OPNsense, APP01 moved to
`network: dmz` / `ip: 10.0.10.10`, and
`ansible/playbooks/opnsense_dmz.yml` building the boundary — five firewall
rules from `networks.dmz.allow` in lab.yaml plus two logged denies, a dnsmasq
range so app01 gets a lease at all, and Unbound on the new leg.

**What the boundary is not:** the hypervisor is multi-homed into every zone
(a provisioning fact, outside the attacker model), and DC01 / CLIENT01 / IDP01
are still in one flat internal segment — the internal user/server split
(`10.0.20.0/24`, `10.0.30.0/24`) remains unbuilt.

Work involved:
- second bridge + VXLAN on the hypervisors (hosts.yml pattern repeats)
- third interface on the OPNsense VM in its terraform module
- OPNsense interface/firewall rules for the DMZ↔internal boundary

High realism per unit of effort — do this before building the web-app
attack path, not after.

**Update (history, kept because it is honest):** the web-app attack path was
built first anyway, deliberately — APP01 exists (`docs/app01-design.md`) and,
until 2026-09-25, sat on the flat L2 with everything else. For that whole
period APP01's compromise was *equivalent* to an internal foothold, because
there was no boundary to cross from it. The same sentence is the reason this
item existed.

**Now closed, and the narrative may change accordingly:** APP01 is in the DMZ
behind a default-deny boundary, and a compromise there is a genuinely weaker
position than an internal foothold. The Chapter 5 chain that loots the
`svc-idp-ldap` credential and binds to LDAP on dc01 works through a single
deliberate, logged allowance — which is what makes it a pivot *across* a
boundary rather than a walk down a flat network.
