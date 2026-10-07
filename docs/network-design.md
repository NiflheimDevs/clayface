# Clayface — Network Design (zones, the DMZ boundary, the VPN pool)

Design for the lab's L2 segments and the filtered boundary between them: which
zones exist, what may cross between them, why the boundary sits where it does,
and what the boundary deliberately is *not*.

Status: **built.** `lab.yaml`'s `networks:` map is the source of truth for the
segments; terraform attaches VMs to them; `ansible/playbooks/hosts.yml` builds
the bridges and the VXLAN overlays on the hypervisors; and
`ansible/playbooks/opnsense_dmz.yml` builds the boundary itself on the edge VM.
The attacker leg is addressed and carries a VM (section 8.1). The internal
user/server split is **not** built — section 11.

Companion documents: `docs/network-plan.md` (the plan this was executed from,
and the record of what was discovered while executing it),
`docs/opnsense-image.md` (how the edge image was built, including the one
interface fact that is not automatable), `docs/app01-design.md` (the host that
sits in the DMZ), `docs/ad-identity-design.md` (the identity layer the one
deliberate allowance reaches).

---

## 1. Scope and non-goals

**In scope.** The segment map and its addressing; the bridges and VXLAN
overlays that carry it; the OPNsense interfaces; the filter policy between the
zones; the DHCP and DNS the DMZ needs to function at all; and the VPN pool as a
declared fact.

**Out of scope.** The application tier's own design (`docs/app01-design.md`).
The AD identity layer (`docs/ad-identity-design.md`). IDP01. The internal
user/server split (section 11). Packer-based image automation, which is what
would make the manual steps in section 8 disappear (roadmap Phase 7).

The attacker machine is in scope only as far as its segment goes: the `wan`
leg's addressing, the firewall's own address on it, and the fact that the VM is
configured without Ansible. What runs *on* the box — the tooling, the
engagement workflow — is not, and neither is the WAN port forward, which is
written but still unverified (section 11).

**Explicitly rejected.**

- **VLANs instead of separate bridges.** A VLAN would have been one bridge with
  tagged sub-interfaces, which is closer to how a real corporate network does
  it. It was rejected because the hypervisor-side change is not smaller (the
  same `hosts.yml` loop, with tag plumbing added), it puts the tag in the path
  of every existing VM, and it makes the segment invisible in `virsh
  domiflist` — the command this lab actually debugs with. Bridges are the
  honest primitive for a lab with four VMs per host.
- **A second OPNsense VM as a stand-alone DMZ firewall.** Two firewalls would
  be a more defensible boundary and a much larger change: a second image, a
  second automation surface, and a routing story between them. The boundary
  this document describes is the one-VM version, and section 6 states what that
  costs.
- **Per-zone control-node access being removed.** The control node is
  multi-homed into every zone by construction (section 6). Making it
  single-homed would mean Ansible could no longer reach the DMZ hosts it
  provisions, which is a provisioning requirement, not an attacker-model
  statement.

---

## 2. Zones and addressing

Static facts live in `lab.yaml` under `networks:` and nowhere else. This table
is a rendering of that map, not a second copy of it.

| Zone | Segment key | Bridge | Subnet | Gateway / DNS | Control node | VXLAN VNI | Status |
|---|---|---|---|---|---|---|---|
| Internal LAN | `lan` | `vm-lan0` | `10.0.0.0/24` | `10.0.0.1` | `10.0.0.2/24` | 100 | built |
| DMZ | `dmz` | `vm-dmz0` | `10.0.10.0/24` | `10.0.10.1` | `10.0.10.2/24` | 101 | built |
| Attacker (WAN) | `wan` | `vm-wan0` | `10.1.0.0/24` | `10.1.0.1` | — | — | built |
| VPN pool | `vpn` | — | `10.0.100.0/24` | `server: none` | — | — | **declared, not implemented** |

Guests, by zone:

| VM | Zone | Address | Module / role |
|---|---|---|---|
| `opnsense01` | all three | `10.0.0.1`, `10.1.0.1`, `10.0.10.1` | `opnsense` / `gateway` |
| `dc01` | `lan` | `10.0.0.10` | `dc` / `dc` |
| `app01` | `dmz` | `10.0.10.10` | `app` / `app` |
| `client01` | `lan` (commented out) | `10.0.0.20` | `client` / `client` |
| `crusader` | `wan` | Dynamic OPNsense DHCP lease, registered as `crusader.clayface` | `crusader` / `crusader` |
| *uplink* | *(none — the bridge libvirt's `default` network creates)* | `192.168.122.0/24` (that network's own DHCP) | `opnsense` / `gateway`, 4th NIC |

`wan` carries only `bridge`, `subnet` and `gateway`. The keys it omits are
statements rather than gaps:

- **no `vxlan_id`** — no overlay, so the leg is host-local by definition. That
  is what pins the attacker VM to the host holding OPNsense.
- **no `monitor`** — nothing on the control node manages that VM. It is
  configured by cloud-init at first boot, so Ansible never dials in, and the
  control node stays off this leg.
- **no `dns`** — OPNsense's WAN interface is not a resolver. A `dns:` here
  would point the bridge at an address that never answers.

The leg was addressed in this wave because a guest now lives on it (section
8.1). Before that it declared nothing but a bridge, and that was the standing
proof that `bridge` is the only key a segment must carry. The WAN is still the
NAT leg in the sense that matters: nothing inside the lab may be reached from
it unless a port forward says so.

### 2.1 Why the map is a map

Before this wave, "the lab subnet" was `10.0.0.0/24` written in prose across
four documents and hardcoded in two playbooks. The `networks:` map replaced
that with one declaration, read by terraform through `locals.tf` and by Ansible
through `lab_inventory.py` (which hands every host the whole map as
`lab_networks`).

The practical payoff is that **adding a zone is a data edit**. `dmz` cost one
entry in `lab.yaml`: no new terraform module, no new playbook, no edit to
`hosts.yml`. The `hosts.yml` loop builds a bridge, a monitor address, a
resolver binding and a VXLAN for every entry in the map, so the second segment
arrived through the same code path as the first.

The same is true of the VXLAN: `vxlan_id` under a segment is what makes it an
overlay. Declaring one is the statement "this leg is carried between hosts",
and a segment without one is host-local by definition. With `host_b` currently
commented out there is no peer, so no overlay is actually built — but VNI 101
is declared and will follow automatically when the second host returns.

---

## 3. Why the boundary sits at OPNsense

The boundary is a set of pf filter rules on the edge VM, evaluated between the
`lan` and `dmz` interfaces. Three properties make that the right place:

- **It is a separate guest.** The rules are not on `app01`, so compromising
  `app01` does not touch them. This is the opposite of a host firewall, where
  the attacker's first act after initial access is to disable the control that
  is supposed to contain them.
- **It is the router.** Every packet between the zones is already passing
  through it, so the boundary costs no new hop and no new failure mode.
- **It is the thing that is already configured declaratively.** The edge VM's
  DHCP and DNS pinning was already API-driven; adding its filter rules there
  gives the lab one automation surface for the edge rather than two.

The boundary is **default-deny**. OPNsense's generated system ruleset contains
no pass-any-per-interface rule — the automatic rules are narrow and named
(DHCP 68→67 and 67→68, `sshlockout`, `virusprot`, IPv6 ICMP, and a LAN-only
anti-lockout). Everything not matched by a rule on the interface falls through
to the deny. An interface with no pass rules written for it is therefore
unreachable, not open.

**This is worth stating because the opposite is a common assumption**, and
because it is what makes section 8's manual step safe: while the DMZ address is
unset, the segment is unreachable rather than unfiltered. An attacker cannot
land in a zone that has no route to it.

---

## 4. The one-NIC rule

**Each guest holds exactly one NIC, on exactly one zone.** `app01` has one
interface on `vm-dmz0` and no other. The edge VM is the only exception, and it
is the exception because being multi-homed *is* its role.

The rule exists because a second NIC is the standard way a segmentation design
is quietly defeated. A DMZ host with an interface on the internal LAN has the
boundary drawn around it rather than across it: the attacker who lands there
walks out over the interface the firewall rules were never asked about. The
control node is the honest counterexample — it *is* multi-homed — and section 6
is where that is stated rather than hidden.

The rule is enforced structurally rather than by policy: `lab.yaml` gives a VM
one `network:` key, terraform's `placement_network` local resolves it to one
bridge, and the client/dc/app modules each attach one NIC. A second NIC would
take a second module change, which is the right amount of friction.

---

## 5. The boundary policy

The policy is **data**: `networks.dmz.allow` in `lab.yaml`, read by
`opnsense_dmz.yml`. It lives there rather than inside the playbook because "what
may the zone the attacker lands in reach?" is the single most consequential
decision in this design. As a `lab.yaml` list it is a diff a supervisor can
read; buried in Ansible it would only be discoverable by reading a playbook.
The precedent is `app.weaknesses` — a design decision expressed as data, read by
a playbook, asserted by a validator.

Rules, in the order the firewall actually holds them — this is the sequence
`opnsense_dmz.yml` reports at the end of a run, and the numbers below are the
live `sequence` values, not a reading order someone chose:

| # | Action | Proto / port | On | Source | Destination | Why |
|---|---|---|---|---|---|---|
| 1 | pass | tcp 389 | `opt1` | DMZ net | `10.0.0.10` (dc01) | **The one deliberate allowance.** Section 5.1. |
| 2 | pass | tcp 636 | `opt1` | DMZ net | `10.0.0.10` (dc01) | Same allowance, LDAPS. |
| 3 | pass | tcp/udp 53 | `opt1` | DMZ net | `10.0.10.1` | So `dc01.clayface` resolves and the pivot is expressible as a hostname. |
| 4 | pass | udp 67 | `opt1` | DMZ net | `10.0.10.1` | DHCP. Must be written explicitly — nothing else provides it. |
| 5 | pass | tcp 443 | `lan` | `10.0.0.0/24` | `10.0.10.10` | Internal users browsing the portal. The only row not on the DMZ leg. |
| 6 | block (log) | any | `opt1` | DMZ net | `10.0.0.0/24` | The explicit deny. This is the rule the Chapter 5 evidence comes from. |
| 7 | block (log) | any | `opt1` | DMZ net | any | Catch-all, last. An attempt to pivot is *visible*, not merely dropped. |

The `On` column is not decoration: row 5 is a rule on the **LAN** interface
about DMZ traffic, and reading the table without it invites the conclusion that
the DMZ leg carries a pass rule into `10.0.10.10`, which it does not. The two
blocks come after it so that the DMZ leg's own list ends with its catch-all —
which is what makes the denies the last thing the DMZ leg evaluates.

**The table is a one-way ratchet, and that is deliberate.** `allow` is applied
by creating and updating rules, never by deleting them, so *adding* an entry
converges on the next run but *removing* one does not: the rule stays on the
firewall and simply stops being one lab.yaml asks for. The playbook fails the
run when it finds such a rule rather than deleting it — it cannot tell "lab.yaml
no longer wants this" from "something else put this here", and guessing wrong
deletes a rule someone depends on. So a shrinking boundary is a two-step
change: edit `lab.yaml`, re-run, delete the named rule in the UI, re-run.

Two facts about the table that a reader would otherwise mis-derive:

- **Rows 1 and 2 are two pf rules, not one.** OPNsense's `destination_port`
  field takes one port, one range, or an alias — never a comma list. In
  `lab.yaml` the entry is one list (`ports: [389, 636]`), because "which ports"
  is the fact; the playbook expands it to one rule per port, suffixing `[port
  N]` onto the description so the description-keyed ownership index stays
  unique. **The firewall rule count is not the length of the `allow` list.**
- **`protocol` does accept a combined value**, spelled `tcp/udp` and
  canonicalised by the API to `TCP/UDP`. So row 3 is one rule where rows 1–2
  are two. The two fields behave differently, and this is the only place it
  matters.

Every rule carries `log`. The passes are logged as well as the denies, so the
Chapter 5 evidence includes *what was used* and not only what was refused.

**Deliberately absent: `10.0.10.1:443`.** The firewall's own management UI must
not be reachable from the zone the attacker lands in. That decision is what
drove the SSRF-probe change in `app_validate.yml` (section 9), and it has a
second consequence recorded in `app/internal/web/api.go`: from the DMZ, the
address `10.0.0.1` is the firewall's *LAN* address, on a leg the portal has no
route to at all — so it is not "blocked", it is simply not on the network. The
reachable management address is `10.0.10.1`, and `:443` there is closed.

**Also deliberately absent: `3268` / `3269`** (the global catalog). The pivot
needs a bind on `dc01:389`, not a forest-wide search. An attacker-noticeable
difference is a design decision, not an oversight.

### 5.1 The one deliberate allowance

The documented Chapter 5 chain (`app/internal/seed/seed.sql`,
`docs/app01-design.md` §8.5) is: loot the `svc-idp-ldap` credential out of
app01's portal database, then bind to LDAP on `dc01`.

`app01` is **not domain-joined**. It carries a service account as a credential
and nothing more, so that credential is the *only* pivot from the DMZ into the
domain. Deny DMZ → `dc01` on 389/636 and the chain dies with it.

So the rule stays, and it stays narrow: a bind on 389/636, not a forest search,
not SMB, not a shell. It is logged. And it is commented here and in the
playbook with its provenance, so a reader finds a decision with a reason rather
than a hole with a story attached.

This is the "planted weakness, documented" pattern the rest of the project
uses. The difference from the application weaknesses is that this one is a
*network* control's deliberate exception, which is why it is in the same list
as the rest of the policy rather than in a comment: narrowing it or widening it
should be a reviewable diff.

---

## 6. What the boundary is not

Stated at length because a diagram of two zones with a firewall between them
invites conclusions the design does not support.

- **It is not a hypervisor boundary.** The hypervisor holds a bridge for every
  segment and has an address on each. A guest that escaped to the hypervisor
  would be outside this model entirely. That is a provisioning fact — the host
  must be able to build and attach the segments — and not a claim about
  containment.
- **It is not a control-node boundary.** The control node is deliberately
  multi-homed into every zone (`10.0.0.2` and `10.0.10.2`), because Ansible has
  to reach the guests it provisions and the DMZ guests are no exception. A
  compromise of the control node is therefore equivalent to a compromise of
  every zone at once. This is the single biggest gap in the boundary, it is
  structural rather than accidental, and it belongs in the attacker model as an
  explicit exclusion rather than as an omission.
- **It does not constrain the LAN.** Rules 5 and 6 filter traffic *leaving* the
  DMZ. `dc01` and `client01` reach each other, and reach `app01` on 443, with
  nothing between them. The internal user/server split is unbuilt (section 11).
- **It is one firewall, not two.** A stand-alone DMZ firewall would put the
  boundary on a device the attacker's segment does not route through in order
  to reach its own gateway. Rejected in section 1 for cost, and the cost is
  real: rules 3 and 4 necessarily expose the firewall's own address to the DMZ
  on 53 and 67, because it is that segment's resolver and DHCP server.
- **It does not stop the application weakness.** SQL injection in the portal is
  still SQL injection. What the boundary does is bound the *blast radius* of
  the SSRF weakness specifically — addresses the portal could previously reach
  internally are now denied. That is defense in depth and worth claiming in
  Chapter 4, but it is a network control containing an application flaw, not a
  fix for it.

---

## 7. DHCP and DNS on the DMZ leg

The DMZ leg needs both, and neither is optional:

- **DHCP.** `app01` gets its address from dnsmasq on OPNsense, so unless
  `opnsense_dmz.yml` gives the DMZ leg an interface list entry and a pool,
  `app01` boots with **no lease at all**. The pool is `10.0.10.100`–`10.0.10.200`
  (`dhcp_start` / `dhcp_end` in `lab.yaml`); `app01` itself is not in it, since
  it holds a reservation.
- **DNS.** Unbound must be told to answer on the DMZ leg, or `dc01.clayface`
  does not resolve from inside the DMZ and the Chapter 5 pivot cannot be
  expressed as a hostname. That is a small thing that changes the narrative: a
  chain that ends in an IP address reads as an accident, and one that ends in a
  hostname reads as a target.

Both are configured by the same playbook, from the same map, and both are
asserted at the end of it.

`subnet_mask` is deliberately **not** in the map for the DMZ. dnsmasq takes it
from the interface address, and a second copy could only ever disagree with
`subnet`.

---

## 8. Automation, and the manual steps

`ansible/playbooks/opnsense_dmz.yml` is one play, `hosts: localhost`, against
the OPNsense REST API. It is idempotent: a second run reports `changed=0`. It

1. finds the edge VM's DMZ NIC **by the MAC terraform pinned to it** and makes
   sure that device is assigned as an OPNsense interface;
2. writes the filter rules from `networks.dmz.allow` plus the two logged
   denies, and applies them;
3. enables dnsmasq and Unbound on the new leg;
4. verifies — the interface is addressed as `lab.yaml` says, no rule on it is
   an unrestricted pass, every rule `lab.yaml` asks for is configured, and the
   leg answers DNS for the lab domain.

Three design points worth recording:

- **The interface is found by MAC, not by name.** NIC order is device order and
  the guest names interfaces by position (`vtnet0`, `vtnet1`, `vtnet2`), so an
  edit that renumbered the NICs would silently repoint the firewall rules at a
  different interface. Terraform pins a MAC derived from the VM name for the
  DMZ NIC specifically to turn that into a named assertion failure.
- **The playbook never deletes a rule it does not own.** Ownership is by
  `description`; rules it owns are updated in place, rules it does not are left
  alone. This is the same idiom as `opnsense.yml`, and it is what lets an
  operator add a rule by hand without the next run removing it.
- **`addRule` and `setRule` report rejection in the response body, not the
  status code.** A rejected rule comes back as HTTP 200 with
  `{"result":"failed","validations":{...}}`. Asserting on the status code would
  mean silently losing rules, so every create and update asserts on `result`.

**The one manual step: the interface's IPv4 address.** This playbook *asserts*
it and stops with instructions; it cannot set it. OPNsense 26.7 has no REST path
to an interface's IPv4 configuration — the `interfaces/assignment` controller
is backed by a seven-field model that covers assignment only, and a `setItem`
carrying address fields answers `{"result":"saved"}` while writing none of
them. The address lives in `config.xml`, written by the legacy `interfaces.php`
form, which authenticates by GUI session rather than by API key. Only 27.1's
model makes the address API-settable.

Interfaces → Assignments → the DMZ row:

```
IPv4 Configuration Type : Static IPv4
IPv4 address            : 10.0.10.1/24
Block private networks  : OFF   (the DMZ is RFC1918 — ticking this drops
                                 the whole segment)
Block bogon networks    : OFF
Enabled                 : ON
```

Then Save, Apply, and re-run the playbook.

**Why asserting is better than automating here.** The step happens once per lab
build. Automating it would mean a new `LAB_OPNSENSE_PASS` credential in
`deploy.env` and a pile of session-and-CSRF plumbing against a legacy page, to
replace a UI form with a script that cannot be tested until the form has been
filled in once anyway. Section 3 is what makes deferring it safe: an unaddressed
DMZ interface is default-denied, so the manual step is a gap in *reachability*,
not in *filtering*. And the playbook fails loudly rather than letting `app.yml`
time out at `wait_for_connection` fifteen minutes later.

`deploy.sh` runs the playbook between `opnsense.yml` and
`start_vms.yml --tags members`, and it must not be `|| true`-swallowed: on a
fresh lab it stops the deploy once, deliberately, at the address assert. Fix
the address, re-run the playbook, carry on from the next step.

### 8.1 The attacker leg, and the second address

`crusader` sits on `wan` at `10.1.0.100`, and the firewall's own address on that
leg is `10.1.0.1`. Three things about it are worth recording, because each one
is a place the design could have gone the other way.

**The address is a manual step, and the second one.** OPNsense 26.7 has no REST
path to an interface's IPv4 configuration — the same wall section 8 describes —
so `10.1.0.1/24` is typed into the UI once and asserted thereafter, by section
4e of `ansible/playbooks/opnsense.yml`. That assert reads the **running**
interface list, not `config.xml`, for the reason recorded in `opnsense_dmz.yml`:
an interface can carry its address in saved config and have none in the kernel,
and every API read that says "configured" then describes a leg with no route on
it. It is skipped when the leg declares no subnet, so a lab with no attacker VM
is not asked to address a segment nothing uses.

The UI step, complete because the defaults are wrong for this subnet:

```
Interfaces -> Assignments -> WAN
  IPv4 Configuration Type : Static IPv4
  IPv4 address            : 10.1.0.1/24
  Block private networks  : OFF   <-- ON by default on WAN, and 10.1.0.0/24
                                      is RFC1918, so leaving it on drops the
                                      whole attacker leg
  Block bogon networks    : OFF
  Enabled                 : ON
Save, then Apply.
```

**The guest configures itself, because nothing else can reach it.** This is the
one VM in the lab that Ansible does not own. It sits on the far side of the
filter on a leg with no DHCP and no DNS, so there is no address for Ansible to
dial in to in order to set one — a circularity the DMZ hosts do not have, since
they at least get a lease and a resolver on their own leg. `crusader` breaks it
from inside: the terraform module attaches a NoCloud seed
(`libvirt_cloudinit_disk`), and cloud-init writes the address, the route and the
hostname at first boot with nothing reaching in.

That choice is why this leg declares no `monitor`. Giving the control node an
address here would have made the leg Ansible-manageable and would have added a
third leg to the multi-homing that section 6 already calls the largest gap in
the model — a real cost, paid for a VM that needs nothing from Ansible.

**Config changes reach the guest the same way: a second NoCloud seed.**
`ansible/playbooks/crusader.yml` is the one playbook allowed to touch the VM,
and it still never dials in. It renders `base-image/clayface.ovpn` (appending
the CA from `LAB_VPN_CA_FILE` when it is set — without it the client drops
`remote-cert-tls server` and does not verify the server's identity, an
accepted risk the playbook warns about at run time — and writing an auth
file from `LAB_VPN_USER` / `LAB_VPN_PASS`, the latter defaulting to the
former, so `openvpn-client@clayface` never blocks on an interactive
prompt), builds a second `cidata`-labelled ISO on the hypervisor, attaches it
to the running domain as a virtio disk — the terraform seed holds `vdb`, the
push seed takes `vdc` — and reboots the guest. The seed's instance-id is
hashed from the rendered user-data (the kali module's own idiom), so an
unchanged config is a no-op and a changed one re-runs the per-instance
modules on the next boot. The user-data deliberately carries no `network:` key:
the NIC belongs to the keyfile the first seed wrote, and a second seed
reaching the network stage would reopen a question that cost the day once
already. Terraform's next apply drops the attached device from the domain XML;
re-running the playbook re-attaches it. The tunnel itself is only verifiable
from the guest console — `systemctl status openvpn-client@clayface` and
`ip addr show tun0` — which the playbook says rather than pretends otherwise.

Two consequences follow from the module pinning no MAC, and they are one
decision rather than two. `vms_pinned` in `terraform_vms.py` requires **both** a
MAC and an `ip:`, so `opnsense.yml` never writes a DHCP reservation for this VM
— correct, since a reservation on a leg with no pool points at nothing. And
`ip:` in `lab.yaml` keeps one meaning (this VM's static address) while the
mechanism that realizes it differs per module: reservation for a MAC-pinning
module, cloud-init for this one.

**What the seed may not do is let cloud-init configure the NIC.** With no
network config from any source cloud-init renders a *fallback* config — DHCP for
the first candidate NIC — and on this base that is rendered by `eni` (ifupdown)
into `/etc/network/interfaces.d/50-cloud-init` and handed to dhcpcd. This leg
serves no DHCP, dhcpcd falls back to an IPv4 link-local address, and
NetworkManager finds eth0 already carrying an address it did not put there and
reports the device `unmanaged` from then on. So the seed's `network-config` file
carries `config: disabled`, and the address is written as a NetworkManager
keyfile from user-data instead. The disable cannot live in user-data: cloud-init
resolves the network config from a merged config object it assembles before it
has read user-data, so `network: {config: disabled}` is missing on exactly the
first boot and present on every boot after.

The renderer is also why the keyfile is a keyfile. The seed's `network-config`
channel is rendered by the first renderer cloud-init finds in a priority list
compiled into it, and on this base that is `eni`, which has no `match:` concept —
it names its stanza after the config *key*, so an `ethernets: primary:` entry
became `iface primary`, an interface that does not exist. Steering that list
means `system_info`, which cloud-init 24.2 deprecated in user-data and ignores
there. A second `autoconnect-priority=100` on the keyfile is what beats the
generic DHCP profile NM auto-creates for any ethernet device (`Wired connection
1`, from the hand-built install, which binds no interface name).

**It depends on cloud-init being in the base image, and fails silently if it is
not.** `KALI-base.qcow2` carries cloud-init and NetworkManager; a base built
from Kali's installer carries neither, and the `kali-cloud` images do. If
cloud-init is absent the seed is attached, nothing reads it, and the guest comes
up on a link-local address with no error logged anywhere — so the check is
`ip -br addr` on the guest (or the GDM banner's hostname), never the exit code of
a plan or a playbook. The interface is matched in the keyfile by name (`eth0`);
a base that names its NIC `enp1s0`/`ens3` would not match, which is the same
class of silent failure and is why the guest is what gets checked.

**Nothing about this touches the hypervisor.** The `hosts.yml` loop reads every
segment, but a segment declaring no `monitor`, no `dns` and no `vxlan_id` gives
it nothing to do — no new bridge (`vm-wan0` already existed), no address, no
overlay. The attacker leg is a data edit plus one UI field, which is the
property section 2.1 claims for the map and this is the first segment to test it
from the other direction.

---

### 8.2 The uplink, and who gets a network

The lab had no route off it at all. `vm-wan0` was a bridge with no ports — the
attacker leg and OPNsense's own WAN interface talking to each other and to
nothing else — and a guest on it could not resolve a package mirror, let alone
reach one. This section adds the route and fixes the posture that comes with it.

**A fourth NIC, on libvirt's stock NAT network.** The module appends the NIC
last so the guest sees it as `vtnet3`; LAN, WAN and DMZ keep their numbering.
Its bridge is named by `edge.uplink_bridge` in `lab.yaml`, and it points at
`virbr0` — the device libvirt's `default` network creates. `default` ships with
the libvirt package: a NAT network on `192.168.122.0/24` with its own DHCP,
whose forwarding libvirt masquerades to the host's uplink. That is the route.
It is deliberately **not** a `networks:` entry — those are bridges `hosts.yml`
creates with `ip link add` and owns, whereas this device is created by libvirt
and libvirt owns it; `ip link add` on it would fight libvirt for the device, and
the gateway and VM inventories would treat a host-NAT leg as a lab segment. It
is a property of the edge, like `edge.host`.

**What it costs: the route off the lab is host state, not lab state.** Nothing
in this repo creates `default`. It comes with the package, and a host where it
has been undefined or never started has no `virbr0` — which surfaces as
`virsh start` failing with "Network bridge virbr0 not found", the same error
`start_vms.yml` prechecks the segment bridges for. It checks this one too: the
gateway's `bridges` output carries the uplink bridge alongside the three
segment bridges. `virsh -c qemu:///system net-list` on the edge host is the
check. A lab-owned network was the alternative and was rejected as more
machinery than the dependency is worth: it would mean a network definition, a
playbook section and a template in this repo, all to own a device the package
already provides.

One consequence is a trap rather than a detail: because that subnet is RFC1918,
the address OPNsense's uplink interface receives is private, so that
interface's **Block private networks must be OFF** or the leg is dropped
silently — the same trap the attacker leg's WAN interface carries, and two of
the six UI steps below.

The lab now holds two masquerades on the egress path, which is worth knowing
when reading a packet capture: OPNsense translates `10.1.0.0/24` to
`192.168.122.x` (the manual Hybrid rule below), and libvirt translates that to
the host's LAN address.

**It is a separate leg rather than a re-purposing of the WAN.** OPNsense's
interface *named* `wan` is `vtnet1` at `10.1.0.1/24`, and the port-forward rule,
the section 4e assert and crusader's `gateway:` all point at it. Naming the
uplink `wan` instead would have made OPNsense's WAN the real uplink and moved
the attacker leg to an OPT name — conceptually tidier, and invasive in exactly
three places. The fourth NIC costs one interface assignment and breaks nothing.

**The posture: the attacker has a network, the corporate side does not.** Only
`10.1.0.0/24` egresses. `dc01`, `client01` and `app01` stay sealed. This was a
decision rather than a default, and the alternatives were real:

- *All egress* is the more faithful corporate network — real enterprises have
  internet — and the argument for it is that it closes an open problem. Phase 6
  of the roadmap has to show how an attacker on a LAN host obtains tooling, and
  egress would make that step a download. It was rejected because Phase 6 also
  measures reproducibility — clean `deploy.sh` runs converging and a second
  Ansible run reporting zero changes — and a domain controller that reaches
  Windows Update changes its own patch level under the measurement. The blast
  radius is the second reason: deliberately unpatched, intentionally
  misconfigured Windows VMs with a route out is what GOAD's own README warns
  against.
- *Filtered egress for everything* — allow updates and tooling, deny the rest,
  log it — is what a corporate network actually looks like and would be the
  better answer if the internal tier needed anything. Nothing in the Chapter 5
  chain needs it. `app.yml`'s only internet dependency is a `docker pull` for
  the database and proxy images, and that runs in play 1 **on the control
  node**; `app01` only ever `docker load`s a tarball.

The decision is one line of prose here and, in the running lab, the *absence*
of three rules. Section 11 records it as a deviation, and the roadmap's Phase D
entry that asked for this decision can now be answered.

**The non-obvious part: nothing is translated automatically.** OPNsense's
automatic outbound NAT creates rules for each interface *except the WAN*, and
the attacker leg **is** the interface OPNsense calls WAN. So under automatic
mode the leg that needs translating is the one excluded by definition, and the
interfaces that are included are translated towards `10.1.0.1` — the attacker
leg's own address, which routes nowhere. The mode has to move to **Hybrid** with
one manual rule. This is the single most likely way to end up with a configured
uplink that carries no traffic, which is why the assert below exists.

**Six manual steps, and only the first outcome is asserted.** The interface is
assigned, given DHCP, given a gateway and a default route, NAT'd, filtered, and
given a resolver — all in the UI, because 26.7 has no REST path to an
interface's IPv4 configuration, which is the same wall sections 8 and 8.1 hit.
`opnsense.yml` section 4f asserts the *first* outcome that is visible through
the API — the interface is assigned, `up` and addressed — and prints the whole
list when it is not. It deliberately does not try to assert the NAT rule, the
route or the filter rule: those are not readable from the control node in any
way that would distinguish "configured" from "configured and not working", and
a check that guessed would be worse than none. The only real test is from the
guest: `ip -br addr`, then `ping 192.168.122.1`, then a name lookup.

**Redefining this domain deserved care, and the plan is clean.** `opnsense01`
has no overlay and no base image — `opnsense.qcow2` is the live artifact,
written in place — and its LAN and WAN NICs carry libvirt-assigned MACs that
the module does not pin, while OPNsense binds an interface assignment partly by
MAC. A domain *replace* could therefore have churned those MACs and silently
broken the assignment. Terraform does not do that: the plan for this change is
`3 to add, 1 to change, 0 to destroy`, and the interface diff is a **pure
addition** — the existing three keep their MACs, so the assignment is untouched.
Copy the image aside anyway before applying; it has no snapshot behind it and
the copy costs a minute. The seeded `kali-base` domain is one `virsh start` from
corrupting a read-only backing file — the same class of hazard with the opposite
shape: a writable image with nothing behind it to roll back to.

---

## 9. What the boundary breaks, and the fixes

Segmentation is a change to a running lab, and it invalidated two things that
had been written against a flat network. Both are fixed in this wave:

- **`app_validate.yml`'s SSRF probe** defaulted to `http://10.0.0.1/` as one of
  its targets, fetched by the portal *from the DMZ*. That address is now
  unreachable, which would have made the ON assertion fail for a reason that is
  not the toggle. The default target list is now loopback only
  (`http://127.0.0.1:8080/healthz`), which is boundary-independent and
  therefore better ON-evidence than the LAN address ever was. The LAN address
  moves to the documented side, via `LAB_APP_SSRF_TARGETS`: with the toggle ON
  it now returns nothing, and that absence is evidence of the **boundary**. The
  SSRF *toggle* is proven by loopback; the *boundary* is proven by the LAN
  target being dropped.
- **`ad_gpo.yml`'s WinRM rule** grants `-RemoteAddress 10.0.0.0/24`, which
  excludes the DMZ. That is correct — only Windows VMs are in the LAN, and a
  host in the attacker's segment should not be able to open WinRM on its peers
  whatever the policy store says. A comment now says so, so a future reader does
  not "fix" it by widening it to the whole lab.

---

## 10. Verification

The boundary is verified in four layers, and each layer answers a different
question. All are runnable from the control node.

**Layer 1 — the L2 exists.** `ip -br link show vm-lan0`, `vm-dmz0`, `vm-wan0`
on the hypervisor; `virsh domiflist app01` shows one interface, on `vm-dmz0`.

**Layer 2 — the guests are where the map says.** `terraform output -json vms`
carries `network` and `bridge` per VM; `opnsense01` additionally carries
`bridges` and `dmz_mac`. `dc01` resolves to `10.0.0.10` and `app01` to
`10.0.10.10` from the control node.

**Layer 3 — the boundary is real.** Every check the playbook asserts, plus the
live ones:

```
# unrestricted pass on the DMZ interface: must be empty
GET /api/firewall/filter/searchRule   # filter rows by interface == opt1

# a rule lab.yaml asks for is missing: must be empty
# (description-keyed comparison against networks.dmz.allow)

# a rule lab.yaml does NOT ask for is present: must be empty
# (same read, the other direction — shape checks alone would pass a
#  port-scoped rule this lab never wrote down)

# the DMZ is not routable to the LAN:  dc01:445 must fail from app01,
#                                      dc01:88  must fail from app01
# the DMZ is routable to LDAP:         dc01:389 must succeed from app01
# the firewall's UI is not reachable:  10.0.10.1:443 must fail from app01
```

Port 389 appears exactly once above, and on the succeed side. It is the one
deliberate allowance; listing it as a must-fail too would describe a boundary
that denies the project's own attack path.

**Layer 4 — the boundary is not *too* tight.** The documented chain survives:
from `app01`, a bind to `dc01.clayface:389` with the `svc-idp-ldap` credential
recovered from the portal database succeeds. If this fails, the boundary has
been drawn through the middle of the project's own attack path.

### 10.1 What was actually measured, 2026-09-25

Layers 1-4 above are no longer a plan. Observed:

```
opnsense_dmz.yml    ok=56 changed=0 failed=0   (idempotent over two runs)
app_validate.yml    ok=44 changed=0 failed=0   (11/11 toggles ON, 0 drift)
                    connected as https://app01.clayface:443
```

And the boundary probed from **inside** `app01` — the measurement the control
node cannot make, because it holds an address on every segment and its own
traffic therefore leaves by the LAN leg:

```
OPEN   10.0.0.10:389     the one deliberate allowance
OPEN   10.0.0.10:636     its LDAPS half
CLOSED 10.0.0.10:445     SMB
CLOSED 10.0.0.10:88      Kerberos
CLOSED 10.0.0.10:135     RPC
CLOSED 10.0.10.1:443     the firewall's own UI, from the DMZ
CLOSED 10.0.10.1:22      ssh on the firewall, from the DMZ
OPEN   10.0.10.1:53      the DNS allowance
CLOSED 8.8.8.8:53        nothing off-segment
dc01.clayface -> 10.0.0.10
```

The `app_validate.yml` line is itself a Layer-3 result, not a separate claim:
`app01.clayface` is a name the control node can only resolve by asking
`10.0.10.1`, so the run exercises the DNS allowance on the DMZ leg end to end.

**One caveat on the last line of the probe.** `8.8.8.8:53` being closed does
*not* isolate the deny-any catch-all as the cause. The WAN interface has no IPv4
address, so the lab has no route off-segment whether or not a deny rule exists.
The catch-all is present, logged, and read back by the playbook; this probe does
not independently prove it is what closed that path. The DMZ→LAN results have no
such confound — those addresses are on-link and reachable from the control node,
so `CLOSED` there is the filter and nothing else.

The playbook's final report prints the boundary it built and the state of each
assertion, and it is the fastest read of "what is actually there".

---

## 11. Documented deviations and known limitations

- **Manual steps, all of them interface configuration.** The DMZ interface
  address (section 8) and the WAN's (section 8.1) are one field each. The
  uplink is six steps, because an interface that has to route needs an
  assignment, an address, a gateway, a NAT rule, a filter rule and a resolver —
  and none of them is automatable on this OPNsense release, for the same
  reason: 26.7 has no REST path to an interface's IPv4 configuration
  (section 8.2).
- **Only the attacker leg has internet, by decision** (section 8.2). The
  corporate VMs are sealed. This is a posture the thesis has to defend rather
  than a limitation, and the roadmap's Phase D entry owes it a paragraph.
- **The uplink is asserted only as far as the API can see, and is measured from
  the guest as far as routing.** Section 4f proves the interface is assigned, up
  and addressed; it cannot prove the NAT rule, the default route or the filter
  rule work. `ping 8.8.8.8` from `crusader` does — 0% loss, which exercises all
  three at once — so "OPNsense has an uplink" and "crusader can reach the
  internet" are no longer two separate claims. The deployment playbook now
  adds the attacker leg to Unbound DNS → General → Listen Interfaces and the
  guest resolves through `10.1.0.1` (`dig google.com` returns an address, no
  timeout). The same playbook installs quick blocks for OPNsense management
  ports, so the attacker can route out without reaching the router login.
- **The uplink depends on host state that this repo does not create.**
  `virbr0` belongs to libvirt's stock `default` network, which comes with the
  libvirt package. Nothing here defines it, autostarts it or repairs it, so a
  host where it has been undefined shows up as `virsh start` failing with
  "Network bridge virbr0 not found" — prechecked by `start_vms.yml`, which is
  told the bridge name through the gateway's `bridges` output. Section 8.2 has
  the reasoning for depending on it rather than building a network of our own.
- **The attacker VM's address depends on cloud-init in its base image, and on a
  seed that stops cloud-init from touching the NIC.** `KALI-base.qcow2` has
  cloud-init and NetworkManager; a base built from Kali's installer has neither.
  When cloud-init is missing the NoCloud seed is attached and nothing reads it,
  so the guest comes up link-local with no error logged anywhere — the check is
  `ip -br addr` on the guest, not the exit code of anything (section 8.1). The
  seed's `network-config` file carries `config: disabled` because the fallback
  config cloud-init would otherwise generate is DHCP rendered by `eni`, which on
  a leg with no DHCP server leaves eth0 link-local and `unmanaged`. The address
  itself is a NetworkManager keyfile in user-data with
  `autoconnect-priority=100`, which is what beats the generic DHCP profile NM
  auto-creates. The seed rides a virtio disk, not a SATA cdrom: as a cdrom,
  cloud-init's systemd generator was cut off partway through the datasource
  probe on roughly every other fresh-overlay boot, and the guest booted with no
  hostname and no address and nothing logged.
- **The attacker leg is verified from the guest, not from the control node.**
  `crusader` holds `10.1.0.100`, pings `10.1.0.1`, reaches `8.8.8.8` and
  resolves names through `10.1.0.1`. What has *not* run is
  section 4e's assert as a task, so a fresh deploy still learns the leg's state
  by booting the guest rather than by failing the playbook.
- **The internal split is unbuilt.** `10.0.20.0/24` (users) and `10.0.30.0/24`
  (servers) exist in the target topology and nowhere else. `dc01`, `client01`
  and IDP01 share one flat internal segment, so a foothold on the DC reaches
  the workstation without crossing anything. The DMZ boundary is real; the
  internal tiering is not.
- **The control node is multi-homed into every zone** (section 6). Structural,
  and the largest gap in the model.
- **The VPN pool is declared and unimplemented.** `lab.yaml` says
  `server: none` and nothing answers on `10.0.100.0/24`. It is declared so the
  address range has a home in the machine-read facts, and so no document can
  quietly imply the lab supports VPN initial access when it does not.
  `docs/redclay.yaml`'s technique coverage had `VPN_initial_access` under
  `core_supports_well`; it was removed rather than annotated.
- **VXLAN is a single-peer static mesh.** Fine for two hosts, breaks at three;
  multi-host needs full-mesh FDB entries or a spine. VNI 101 is declared and
  will follow VNI 100 when `host_b` returns, but neither is built while the lab
  is single-host.
- **The WAN port forward is unverified.** It was written from documentation,
  not against a live system, and it may need a linked filter pass rule. Not
  touched by this wave. The attacker VM now exists, so it is finally the thing
  that could verify it: `LAB_WAN_EXPOSE_APP=443` plus a probe from `crusader`
  is the test, and until that runs the port forward is still a claim.

---

## 12. Open questions

- **Should the boundary log be collected?** The two deny rules log to OPNsense's
  own filter log, which nothing currently reads. That log is the natural first
  data source for the detection half of the project (Wazuh/SIEM), and it is
  already being produced — but collecting it is a Phase-6 question, not this
  one.
- **Should the DMZ get its own resolver instead of using the firewall's?** The
  DMZ reaches `10.0.10.1` on 53 by necessity, which is the firewall's own
  address. An internal resolver reachable from the DMZ would remove that
  exposure and add a service to build. Not worth it at this size.
- **Does the internal split earn its cost?** Two more zones is two more
  boundaries, two more policies and two more places for the lab to be wrong,
  in exchange for a segmentation story that matches the proposal's diagram. The
  DMZ boundary is the one the attacker model actually crosses; the internal
  split is where the remaining fidelity lives, and the trade is a judgement
  call for the roadmap rather than for this document.
- **Should `client01` be re-enabled?** It is commented out of `lab.yaml` and its
  `10.0.0.20` address is still declared. Until it comes back, the LAN has one
  Windows host and the lateral-movement story has one fewer step.
