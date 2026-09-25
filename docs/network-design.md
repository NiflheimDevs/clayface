# Clayface — Network Design (zones, the DMZ boundary, the VPN pool)

Design for the lab's L2 segments and the filtered boundary between them: which
zones exist, what may cross between them, why the boundary sits where it does,
and what the boundary deliberately is *not*.

Status: **built.** `lab.yaml`'s `networks:` map is the source of truth for the
segments; terraform attaches VMs to them; `ansible/playbooks/hosts.yml` builds
the bridges and the VXLAN overlays on the hypervisors; and
`ansible/playbooks/opnsense_dmz.yml` builds the boundary itself on the edge VM.
The internal user/server split is **not** built — section 11.

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
The AD identity layer (`docs/ad-identity-design.md`). IDP01. The attacker
machine and the WAN port forward (roadmap Phase 3). The internal user/server
split (section 11). Packer-based image automation, which is what would make the
one manual step in section 8 disappear (roadmap Phase 7).

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
| WAN | `wan` | `vm-wan0` | — (NAT leg) | — | — | — | built |
| VPN pool | `vpn` | — | `10.0.100.0/24` | `server: none` | — | — | **declared, not implemented** |

Guests, by zone:

| VM | Zone | Address | Module / role |
|---|---|---|---|
| `opnsense01` | all three | `10.0.0.1`, `10.0.10.1`, WAN via NAT | `opnsense` / `gateway` |
| `dc01` | `lan` | `10.0.0.10` | `dc` / `dc` |
| `app01` | `dmz` | `10.0.10.10` | `app` / `app` |
| `client01` | `lan` (commented out) | `10.0.0.20` | `client` / `client` |

`wan` is in the map although it has no subnet, because "this leg has no subnet,
no gateway and no overlay" is itself a fact worth declaring rather than
implying. The WAN is a NAT leg: OPNsense does the NAT, and nothing inside the
lab may be reached from it unless a port forward says so.

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

## 8. Automation, and the one manual step

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

- **One manual step.** The DMZ interface address (section 8). Not automatable
  on this OPNsense release.
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
  touched by this wave.

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
