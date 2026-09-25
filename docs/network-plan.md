# Network plan — DMZ, segment rename, VPN pool

Plan for `docs/roadmap.md` Phase 2 plus two coupled refactors. Written
2026-09-24. **Nothing here is implemented.** This document is the detail
behind three `docs/TODO.md` items; it is not itself a work item.

## Why

The supervisor-approved target topology (`red-clay/Overview.md`) has four
zones: external, a DMZ for the application tier, an internal user zone, an
internal server/identity zone, plus a remote-access VPN pool. What runs is one
flat L2 — `vm-br0` = `10.0.0.0/24` — with `app01`, `dc01` and `client01` all on
it and no firewall between them.

Four documents record that gap as the project's biggest honesty problem:
`docs/app01-design.md` §3.1 ("APP01 being compromised is currently equivalent
to an internal host being compromised"), `docs/TODO.md:29-54`,
`docs/thesis-plan.md:159-168` and `docs/roadmap.md:14-15`. Until it closes, no
Chapter 5 narrative may treat "the DMZ web app fell" as a distinct, weaker
position than "an internal host fell".

Three changes, deliberately bundled because they touch the same files:

1. **A DMZ segment.** `app01` moves onto its own bridge behind a filtered
   OPNsense boundary. One NIC only — dual-homing would make the boundary
   bypassable at L2 and destroy the point.
2. **A rename and a restructure.** `vm-br0` / `bridge_name` does not say what
   the segment is. It becomes `vm-lan0` / `lan`, and the network facts
   restructure into a `networks:` map. This also closes a duplication the repo
   already tracks (`docs/TODO.md:5-27`): the bridge name lives in `lab.yaml`
   (consumed by Ansible only) *and* as six hardcoded literals in the terraform
   modules, with nothing linking them — `terraform/locals.tf` never reads the
   `network:` block at all.
3. **The VPN pool as a fact.** `10.0.100.0/24` is already described as the
   remote-access pool (`red-clay/Overview.md:85`) but exists nowhere in code.
   It gets declared, with no server behind it.

## Decisions taken

| Decision | Choice |
|---|---|
| `lab.yaml` network shape | `networks:` map, per segment |
| Internal addressing | Stays flat `10.0.0.0/24`. `dc01` / `client01` do not move. Only `app01` moves, to `10.0.10.10` |
| VPN scope | Pool declared as a fact only. No WireGuard, no VPN server |
| OPNsense DMZ config | API-driven from Ansible, with a one-time hand-assignment fallback if the API turns out not to cover interface assignment |
| Names | `vm-lan0` / `vm-dmz0` / `vm-wan0`, keys `lan` / `dmz` / `wan` |

What stays unbuilt, and the docs must keep saying so: the user/server internal
split (`10.0.20.0/24` / `10.0.30.0/24`) and the VPN server itself.

## What is real before and after

| Zone | Subnet | Before | After |
|---|---|---|---|
| LAN (internal) | `10.0.0.0/24` | built, `vm-br0` | built, `vm-lan0`, **unchanged addresses** |
| DMZ | `10.0.10.0/24` | does not exist | built, `vm-dmz0`, filtered at OPNsense |
| Internal split (user / server) | `10.0.20.0/24`, `10.0.30.0/24` | target only | target only |
| VPN pool | `10.0.100.0/24` | nowhere | declared fact, no server |
| WAN | n/a | built, `vm-wan0` | unchanged |

---

## Prerequisites

**The lab is down.** `docs/app01-verification-pending.md:16-30` records: 8
domains all `shut off`, no `vm-br0`, `sshd` inactive on `host_a`, docker
inactive.

This work lands after `docs/roadmap.md` Phases 0 and 1, not interleaved with
them:

1. **Phase 0** — commit the uncommitted working tree in logical chunks.
   Otherwise the rename diff is entangled with the whole app tier and cannot be
   reviewed.
2. **Phase 1** — bring the lab up and verify what is already written. First
   ever `app` module apply, live `ubuntu-base` builder-domain hazard. It must
   not share a debugging session with a network change.
3. This plan.

Because the lab is down there is no "keep it working" constraint, but there is
still a reason to split into three commits: each is offline-verifiable, so a
failure is attributable to one of them.

### Pre-flight recon, before writing any playbook

One read-only session on a running OPNsense, in the idiom
`docs/app01-verification-pending.md` §4 already prescribes for the unverified
`d_nat` rule. This collapses most of the uncertainty below into facts:

```bash
# Does an assignment endpoint exist, or only configuration of an
# already-assigned interface?
curl -sk -u "$KEY:$SECRET" https://10.0.0.1/api/interfaces/ | head -50
# What field names does OPNsense itself write?
ssh root@10.0.0.1 'grep -A30 "<interfaces>" /conf/config.xml'
ssh root@10.0.0.1 'grep -B2 -A40 "<filter>" /conf/config.xml'
```

Then confirm app01's addressing model, which is a hard prerequisite:

```bash
ssh clayface@10.0.0.30 'cat /etc/netplan/*.yaml'
```

If that shows a static `10.0.0.30`, app01 will come up unreachable on the DMZ
bridge and `app.yml` will hang — that is a base-image fix, not a module fix,
and it must be known before the DMZ commit, not during it. If it is
`dhcp4: true` (most likely, since the whole design leans on the dnsmasq
reservation), nothing to do.

---

## Part A — the map and the rename

### A1. Three commits, in this order

**Commit 1 — the map, with no behaviour change.** `lab.yaml` restructured into
`networks:` with `lan.bridge` still `vm-br0`, plus `dmz`, `wan`, `vpn` and the
new `subnet` / `gateway` / `monitor` / `dns` keys. Teach `locals.tf`, the five
modules, `main.tf`, `outputs.tf`, `lab_inventory.py`, `hosts.yml`, `edge.yml`
and `start_vms.yml` to read it.

> **The proof this commit is correct: `terraform plan` reports no changes.**
> The bridge string is byte-identical, so the module bodies emit the same XML.
> If the plan proposes anything, the plumbing is wrong — and that is discovered
> with the old name still in place, where the blast radius is zero.

This commit is what kills the duplication: after it, both later commits are
one-line `lab.yaml` edits instead of refactors. It also deletes
`start_vms.yml`'s two `vm_bridge_name: vm-br0` copies (`:21`, `:92`), which
`docs/TODO.md:14-16` flags, replacing them with a `libvirt_bridge` hostvar from
the terraform `vms` output.

**Commit 2 — the rename, alone.** One line: `lan.bridge: vm-br0 → vm-lan0`.

> **Mandatory pre-step: create `vm-lan0` on the host before `terraform apply`.**
> ```bash
> sudo ip link add vm-lan0 type bridge && sudo ip link set vm-lan0 up
> ```
> See A4 for why. And **do not delete `vm-br0`** until the rename is verified —
> an orphan bridge with no member costs nothing and makes rollback instant.

**Commit 3 — the DMZ.** `app01` to `network: dmz` / `ip: 10.0.10.10`; the
opnsense module gains a third NIC; `opnsense_dmz.yml` is added; `deploy.sh`
gains its step and a reorder; `app_validate.yml`'s SSRF default changes; docs
and vault come in line.

### A2. `lab.yaml` shape

Replace the flat `network:` block:

```yaml
networks:
  lan:
    bridge: vm-lan0
    subnet: 10.0.0.0/24
    gateway: 10.0.0.1
    dns: 10.0.0.1                 # OPNsense Unbound - see the note below
    monitor: 10.0.0.2/24          # control node on this leg
    vxlan_id: 100
  dmz:
    bridge: vm-dmz0
    subnet: 10.0.10.0/24
    gateway: 10.0.10.1
    dns: 10.0.10.1
    monitor: 10.0.10.2/24
    vxlan_id: 101
  wan:
    bridge: vm-wan0               # NAT leg: no subnet, no gateway, no overlay

# Shared by every overlay, declared once.
vxlan_port: 4789
vxlan_parent_interface: wlan0

# Remote-access VPN pool - DECLARED, NOT IMPLEMENTED.
#
# There is no VPN server anywhere in this lab. No WireGuard config exists, no
# OPNsense VPN instance is configured, and nothing answers on this range. This
# key exists so the pool the target topology describes has a home in the
# machine-read facts, and so no document can quietly imply the lab supports VPN
# initial access when it does not.
vpn:
  pool: 10.0.100.0/24
  server: none
```

Each key earns its place by displacing a literal:

| Key | Literal it kills |
|---|---|
| `subnet` | the `10.0.0.0/24` written in prose across four docs (`docs/TODO.md:11-12`) |
| `gateway` | `.1` in every zone — implied everywhere, written nowhere |
| `monitor` | `hosts.yml:69`'s `10.0.0.2/24`, marked `(temp)` |
| `dns` | four copies of `10.0.0.1`: `hosts.yml:10`, `start_vms.yml:22`, `dc.yml:40`, `opnsense.yml:62` |

`deploy.sh:19`'s `no_proxy="clayface,10.0.0.1,..."` stays a literal, with a
comment pointing at `lab.yaml`. Teaching a bash prelude to parse YAML adds a
new failure mode to the one script that must work before anything else does, in
exchange for one duplicated address that is already a documented mirrored
constant. The trade-off goes in the comment.

Only `bridge` is required per segment; the rest are optional, because the WAN
leg has none of them. **A missing `vxlan_id` is the statement "this leg has no
overlay"** — that is what makes `wan` self-describing and what keeps the
future Kali attacker VM pinned to the host that holds OPNsense.

`vm_placements.app01` gains `network: dmz` and `ip: 10.0.10.10`. The key is
optional and defaults to `lan`, so `dc01`, `client01` and the two `alpine`
placements need no edit.

### A3. Does VXLAN need refactoring too? — yes

One overlay per L2 segment, so a DMZ means **two** VXLAN interfaces per host,
not one. Today's shape says exactly one: `vxlan_name`, `vxlan_id`,
`vxlan_port` and `vxlan_parent_interface` are flat top-level scalars, and
`hosts.yml:97-143` is an inline eight-task block for a single interface.

- **Per segment:** `vxlan_id` moves under the segment. The interface name is
  **derived** — `vxlan<id>` — so `vxlan_name` stops being a knob that can
  disagree with the id. The vault already drifted here: it calls the interface
  `vxlan0` (`red-clay/Report/Actions/Bare Metal VM.md:16`) while `lab.yaml:27`
  says `vxlan100`.
- **Global, stays top-level:** `vxlan_port` (4789 is the standard) and
  `vxlan_parent_interface` (`wlan0` is a property of the host NIC, not of any
  segment).
- **Derived in the inventory, not declared:** `vxlan_remote_ip` /
  `vxlan_peer_ips` already come from the `hosts` map
  (`lab_inventory.py:64-68`). They stay derived and become a peer *list* used
  once per segment, since every overlay has the same remote peer.

`hosts.yml` then creates a VXLAN per segment that declares a `vxlan_id`,
attached to that segment's own bridge — a nested loop instead of a copy of the
block. Two rules worth stating in the playbook:

- **Never enslave two bridges to one VXLAN.** That would merge the LAN and DMZ
  L2 domains and silently undo the segmentation.
- When `host_b` is commented out there are no peers, so the whole section is
  skipped — exactly as today.

### A4. Terraform

`terraform/locals.tf` — start reading the map, with a guard in the style of the
existing ones (`edge_host_attrs`, `module_os`):

```hcl
networks = local.lab["networks"]

# Guard: indexing a map with a missing key is an error, so a typo'd
# `network:` in lab.yaml fails the plan instead of silently putting a VM on
# no bridge. Same trick as module_os.
network_bridges = { for k, v in local.networks : k => v["bridge"] }

placement_network = {
  for name, p in local.vm_placements : name => try(p["network"], "lan")
}
```

`terraform/main.tf` — all 10 module blocks change. The 8 placement-driven
blocks add one argument:
`bridge = local.network_bridges[local.placement_network[each.key]]`. The 2
`opnsense` blocks have no `vm_placements` entry, so they take both bridges from
the map explicitly: `bridge = local.network_bridges["lan"]` and
`wan_bridge = local.network_bridges["wan"]`.

Each of the 5 modules gains a `bridge` variable and loses its literal
(`alpine/main.tf:50`, `app/main.tf:98`, `client/main.tf:105`,
`domaincontroller/main.tf:83`, `opnsense/main.tf:51`); `opnsense` additionally
gains `wan_bridge` (`main.tf:64`).

`terraform/outputs.tf` — each `vms` entry gains `network`, plus a `bridges`
list for opnsense. This is what lets `start_vms.yml` precheck the *right*
bridge instead of one hardcoded name.

**Interface ordering hazard.** `libvirt_domain.devices.disks` and `interfaces`
share the device ordering space, so a future edit that adds a disk ahead of the
NICs could move `vtnet2`. The mitigation is to **append** the DMZ NIC, never
insert it, and to **pin its MAC** — which turns NIC renumbering from a silent
misconfiguration of the WAN into a named assertion failure in the playbook
(A5). Paying for the pinned MAC now is cheaper than debugging a wrong-interface
firewall rule later.

**Provider behaviour — the load-bearing assumption.** `dmacvicar/libvirt`
0.9.8 is the v2 Plugin Framework rewrite, and in its source only top-level
`name` and `type` carry `RequiresReplace`. A nested change like
`devices.interfaces[*].source.bridge.bridge` therefore takes the in-place
`Update` path: stop domain, `DomainUndefineFlags` keeping nvram and TPM,
`DomainDefineXML`, restart. No terraform replace, no state surgery, and the
per-VM overlay disk is a separate resource, so guests' filesystems are
untouched — the only guest-visible effect is a reboot. **This must be confirmed
by the first `terraform plan`, not trusted** (see Verification Layer 2).

**`hosts.yml` and `edge.yml` must run before `terraform apply`.** Today the
order is apply-then-hosts, which works only because `vm-br0` and `vm-wan0`
already exist on the host. Once the bridge name changes, apply-with-`vm-lan0`
-while-`vm-lan0`-does-not-exist fails at the Update path: the domain is
stopped, undefined, `DomainDefineXML` succeeds, and then `virsh start` fails
with `Network bridge vm-lan0 not found`. Apply exits non-zero, leaving the
domain undefined and stopped. Recovery is easy and nothing is lost, but it is a
self-inflicted failure with a one-line fix. Both playbooks read only `lab.yaml`
through the inventory and need nothing from terraform, so there is no
dependency reason to keep them after apply; reordering also means a
`hosts.yml` failure surfaces before terraform state is touched.

### A5. Ansible hypervisor side

`ansible/inventory/lab_inventory.py` — replace the `NETWORK_VARS` flattening
(`:34-43`) with the structured map plus the global overlay knobs on the
`hypervisors` group: `lab_networks`, `vxlan_port`, `vxlan_parent_interface`,
and the derived `vxlan_peers` list. Publish `lab_vpn` on `all.vars` alongside
`lab_ad`.

`ansible/playbooks/hosts.yml` — replace the single-bridge tasks (`:40-70`) and
the inline VXLAN block (`:97-143`) with loops:

1. one `ip -o link show` to gather existing links, then a loop over
   `lab_networks | dict2items` creating each missing bridge and bringing it up;
2. a loop adding `monitor` where the segment declares one;
3. a loop running `resolvectl dns <bridge> <dns>` where the segment declares a
   resolver — this generalises the current LAN-only fix (`:84-90`) and is what
   lets the control node resolve lab names on every leg it is attached to;
4. a nested VXLAN section over segments declaring `vxlan_id`, guarded by
   `vxlan_peers | length > 0`.

`ansible/playbooks/edge.yml` — `wan_name` becomes
`lab_networks['wan']['bridge']`.

`ansible/playbooks/start_vms.yml` — delete the two `vm_bridge_name: vm-br0`
literals (`:21`, `:92`) and the bridge-existence prechecks that use them; both
plays precheck `{{ libvirt_bridge }}` instead, from the new `vms` output field.

`terraform_vms.py` — set `libvirt_bridge` per VM from that field.

**One correction that is not a rename.** `CLAUDE.md:176-177` advises persisting
the resolver with `nmcli con mod vm-br0 ipv4.dns 10.0.0.1`. That advice is
already wrong: a bridge created with `ip link add` is not a NetworkManager
connection, so `nmcli con mod` on it fails. `resolvectl dns <bridge>` works on
the unmanaged device, which is why `hosts.yml` uses it. Fix the advice to match
the mechanism, or the next person follows a command that cannot work.

---

## Part B — the DMZ

### B1. app01

`modules/app` stays single-NIC. Its `interfaces` block loses the literal and
takes `var.bridge`; with `network: dmz` that bridge is `vm-dmz0`. Its pinned
MAC, `ip:`, DHCP-reservation chain and other settings are untouched.

Result: `virsh domiflist app01` shows exactly **one** interface, on `vm-dmz0`.

### B2. OPNsense — what is baked vs what is automated

> **Amended 2026-09-25, from the live box.** The paragraphs below were written
> from documentation and end in "attempt it and see". That attempt has now been
> made, and the answer is **no**: there is no REST API for an interface's IPv4
> address on the OPNsense release this lab runs. The original text is kept
> underneath as the record of what was known before the attempt; the finding
> that replaces it is this:
>
> - **The assignment *is* API-settable.** `/api/interfaces/assignment/`
>   (`searchItem`, `addItem`, `delItem`, `reconfigure`) exists and works: it is
>   what creates the `opt1 → vtnet2` row. This is the `assignments` controller
>   the original text could not find in the published docs.
> - **The address is not.** On this release the assignment model carries seven
>   fields — `descr`, `identifier`, `icon`, `optgroup`, `if`, `lock` — and
>   covers assignment **only**. `setItem` accepts `type4`, `ipaddr`, `subnet`
>   and `enable` without complaint, answers `{"result":"saved"}`, and writes
>   **none** of them: the config gains `<if>`, `<descr>` and `<lock>` and
>   nothing else. Silent and inert. `interfaces/assignment/pending` (the
>   pending-action endpoint the newer UI drives) **404s** on this box —
>   `{"errorMessage":"Endpoint not found"}` — which was the first signal that
>   its controller is not the one the current docs describe.
> - The address lives in `config.xml`, written by the **legacy
>   `src/www/interfaces.php` form**. That path needs a GUI session
>   (`guiconfig.inc` → `authgui.inc`) plus a CSRF token (`csrf.inc`,
>   `LegacyCSRF::checkToken()`, from `$_POST[$securityTokenKey]` or the
>   `X-CSRFToken` header). There is **no HTTP Basic or API-key route into
>   legacy pages**, so it is not reachable with the credentials `deploy.sh`
>   already has.
> - **Only 27.1 (master) has the rich model** — `type4`/`ipaddr`/`subnet`/
>   `enable`/`pending_action` on the assignment item, plus a working
>   `pendingAction`. Tags 26.7.1 through 26.7.4 all still carry the thin one.
>   Verified by reading the release tags' sources, not the docs site.
>
> **Therefore:** the address is a **one-time hand step**, exactly as the
> fallback below anticipated, and `opnsense_dmz.yml` **asserts** it rather than
> setting it — it reads the interface back and stops with the UI steps if the
> address is missing. It is a mirrored constant of the same class as the LAN IP
> and the DNS domain, and `docs/opnsense-image.md` records it.
>
> Asserting rather than scraping the legacy form was a deliberate choice. The
> form is reachable only with a session cookie plus CSRF, which would mean a
> **new** `LAB_OPNSENSE_PASS` credential in the user's `deploy.env` and a pile
> of untestable session plumbing, to automate a step that happens once per lab
> build and that fail-closed makes harmless to defer. The offer stands if the
> manual step ever becomes a nuisance — 27.1's model is the cleaner fix.

**The finding as it stood before the attempt, stated with its confidence.**
OPNsense's public API documentation
(`docs.opnsense.org/development/api/`) for the `interfaces` controller shows
`settings` (`get` / `set` / `reconfigure`) and `overview`, plus sub-controllers
for bridge / gif / gre / lagg / loopback / neighbor / vip / vlan / vxlan.
**There is no documented `assignments` controller** — the step that creates the
`opt1 → vtnet2` mapping in the first place. Configuring an *already-assigned*
interface through `settings/set/<ifname>` is plausible and probably works, but
that is inference, not documentation.

Endpoints that **are** documented:

| Purpose | Endpoint | Status |
|---|---|---|
| Search filter rules | `GET /api/firewall/filter/searchRule` | **confirmed; returns the full generated ruleset (36 rows on this box), config rules under `Firewall -> Filter -> rules`** |
| Add / update rule | `POST .../addRule`, `.../setRule/<uuid>` | **confirmed. Returns HTTP 200 with `{"result":"failed","validations":{...}}` on rejection — the status code proves nothing, so callers must assert `result`.** |
| Apply ruleset | `POST /api/firewall/filter/apply` | endpoint confirmed |
| Interface list for rules | `GET /api/firewall/filter/getInterfaceList` | endpoint confirmed |
| Unbound ACL | `/api/unbound/settings/{search_acl,add_acl,set_acl}` | endpoint confirmed, body unverified |
| DHCP range | `/api/dnsmasq/settings/{search_range,add_range,set_range}` + `service/reconfigure` | endpoint confirmed, body unverified |
| Config read-back | `GET /api/core/backup/download/this` | **confirmed — returns `config.xml`, which is how the playbook verifies what it wrote** |

No page documents request-body field names for any of them. Read them off a
live system (`/conf/config.xml`) — same remedy as the `d_nat` rule.

**The DHCP range is a hard prerequisite, not a nicety.** app01 gets its address
from dnsmasq, so unless `opnsense_dmz.yml` enables DHCP on the DMZ interface
with a range, app01 boots with **no lease at all**. That is also the reason the
un-configured state is safe: it fails closed, loudly, at
`wait_for_connection`.

Put this in a **new** `ansible/playbooks/opnsense_dmz.yml` rather than growing
`opnsense.yml` — the existing file is 591 lines of DHCP/DNS pinning with one
purpose, and the DMZ policy has a different lifecycle (it changes when the
policy changes, not when a VM is added). Copy its idiom exactly: one play,
`hosts: localhost`, `connection: local`, `module_defaults` for auth,
`validate_certs: false`, `status_code: 200`,
search → index → add-if-missing → update-if-changed → reconfigure, **never
delete a row**, own objects identified by a `descr` marker. The header comment
must carry the same honest "written from documentation, verify against the live
box" status as `opnsense.yml:437-459`.

**No `opnsense.yml` change is needed for app01's IP move.** That playbook reads
`groups['vms_pinned']`, which carries `libvirt_ip` from the terraform output, so
changing `ip:` to `10.0.10.10` in `lab.yaml` rewrites the dnsmasq row
automatically through the existing update-by-name path. That is the payoff of
the never-delete design and belongs in the commit message.

### B3. The hazard that would silently void the whole change

> **Amended 2026-09-25, from the live box. The premise below is false, and the
> real property is better.** There is no pass-any-per-interface automatic rule.
> OPNsense's generated system ruleset contains **no** rule that passes traffic
> merely because it arrived on an interface; the automatic rules are narrow and
> named — DHCP (68→67, 67→68), `sshlockout`, `virusprot`, IPv6 ICMP, and a
> LAN-only anti-lockout. Everything else falls through to the default deny.
>
> So **a newly assigned interface is fail-closed**, not wide open: with no pass
> rule written for it, the DMZ is denied from the moment it comes up. There is
> no `disable_automation` field to set (confirmed absent, not merely unset) and
> no window to close.
>
> This inverts the section's conclusion in a useful direction. The hazard was
> never "the interface might be open"; it was "the deploy might *look* finished
> while the DMZ is unreachable" — which is why the DHCP range and the address
> are prerequisites (B2), and why the playbook fails loudly on a missing
> address instead of quietly leaving app01 off the network. It also makes the
> one manual step in B2 safe to defer: while the address is unset, the segment
> is unreachable rather than unfiltered. An attacker cannot land in a zone that
> has no route to it.
>
> The ordering constraint below therefore **ceases to be a safety constraint**.
> The playbook still assigns, then writes rules, then applies — but that is
> because a rule cannot name an interface that does not exist yet, not because
> a window has to be closed.
>
> The assertion that survives, and is now the load-bearing one, is different:
> **no rule on the DMZ interface may be an unrestricted pass.** That is checked
> after the ruleset is written, against the interface's own rows, and catches
> the real failure mode — a rule that is wider than it reads (a missing
> destination port, a source of `any`) — rather than an automatic rule that
> does not exist.

**The hazard as originally written.** A newly assigned OPNsense interface comes
with automatic rules — in the common case a pass-any-from-this-interface and a
pass-any-to-this-interface. A freshly assigned DMZ interface is therefore wide
open in both directions before a single rule is written. If those are left on:

- the boundary is fake;
- every check in Part D's Layer 3 fails, looking like a rule problem rather
  than an automatic-rules problem;
- worse, if the play dies partway the lab sits in a state where the DMZ looks
  segmented in the diagram and is not segmented in fact — exactly the class of
  dishonesty this repo's docs exist to avoid.

So the field (the GUI label is along the lines of "Disable this interface's
automatic rules"; the config field is probably `disable_automation`) must be
set **in the same call that brings the interface up**, so there is never a
window where the interface is up and auto-ruled. Section order inside the
playbook: assign → assert device → set IP and disable automation together →
DHCP range → filter rules → apply.

And add an assertion, so the state is provable rather than assumed:

```yaml
- name: Read the interface configuration back
  ansible.builtin.uri:
    url: "{{ opnsense_url }}/api/interfaces/settings/get/opt1"
    method: GET
  register: opt1_cfg

- name: Refuse to continue if the DMZ interface still has its automatic rules
  ansible.builtin.assert:
    that: opt1_cfg.json.interface.disable_automation | default('0') | string == '1'
    fail_msg: >-
      opt1 is assigned but its automatic rules are still enabled. A freshly
      assigned OPNsense interface passes any-to-any in both directions until
      they are disabled, so the DMZ boundary would be a diagram and not a
      filter. Nothing after this point is meaningful. Set it in the UI
      (Interfaces -> [opt1] -> uncheck automatic rules) or fix the field name
      in this playbook, then re-run.
    quiet: true
```

This assertion is the most valuable part of the playbook: it converts "the
boundary exists" from a claim into a checked precondition. **(Kept as the
record of the intent. The assertion that was actually built checks the
unrestricted-pass property described above — same purpose, aimed at the failure
that exists.)**

### B4. The boundary policy

Rules on interface `dmz`, in order:

| # | Action | Proto / port | Source | Destination | Why |
|---|---|---|---|---|---|
| 1 | pass | tcp 389 | DMZ net | `10.0.0.10` (dc01) | **The deliberate allowance.** See below. |
| 1b | pass | tcp 636 | DMZ net | `10.0.0.10` (dc01) | Same rule, LDAPS. A pf rule carries one port, so 389 and 636 are two rules. |
| 2 | pass | tcp/udp 53 | DMZ net | `10.0.10.1` | So `dc01.clayface` resolves and the pivot is expressible as a hostname. Narrow to the resolver, not "any". |
| 3 | pass | udp 67 | DMZ net | `10.0.10.1` | DHCP. Not implied by anything once the interface is filtered, so it must be written explicitly. |
| 4 | block (log) | any | DMZ net | `10.0.0.0/24` | The explicit deny, logged. This is the rule that produces the Ch5 evidence. |
| 5 | block (log) | any | DMZ net | any | Catch-all, last. There is an implicit deny, but an explicit logged one makes an attempted pivot *visible*. |
| 6 | pass | tcp 443 | `10.0.0.0/24` | `10.0.10.10` | Optional: internal users browsing the portal. |

> **Two corrections to the table above, 2026-09-25, from the live box.** Both
> change what the table means for anyone reading it later.
>
> 1. **`389, 636` is not one rule.** OPNsense's `destination_port` field
>    (`PortField`) takes **one port, one range, or an alias — never a comma
>    list**. `addRule` answers HTTP 200 with
>    `{"result":"failed","validations":{"rule.destination_port":"Please
>    specify a valid portnumber, name, alias or range."}}`. Row 1 is therefore
>    **two pf rules**, and the playbook builds it as such — `networks.dmz.allow`
>    gives a `ports:` **list** per entry, and the playbook expands it one rule
>    per port, suffixing `[port N]` onto the description only when an entry
>    names more than one (so the description-keyed ownership index stays
>    unique). **The firewall's rule count is not the length of the `allow`
>    list** — 4 entries on this box become 5 rules.
> 2. **`tcp+udp` is one rule, but it is spelled `tcp/udp`.** `protocol`
>    (`ProtocolField`) does support a combined value via `<AddOptions>`; the
>    API canonicalises it to `TCP/UDP`. Row 2 is unchanged in meaning — worth
>    recording because it is the one place the two fields behave differently.
>
> The `result` assertion this implies is in B2's amended table: `addRule` and
> `setRule` report rejection in the body, not the status code, so a playbook
> that only checks the HTTP status silently loses rules.

Rules 2 and 3 reach the firewall's own address on 53/67, which is normal and
not a hole. **Deliberately absent: `10.0.10.1:443`.** The firewall's management
UI must not be reachable from the zone the attacker lands in. That decision is
what drives the SSRF change in B6.

Note on rule 1's destination: `dc01` is currently commented out of
`vm_placements`, so its address is not derivable from terraform. The target has
to be a declared fact.

**The one deliberate allowance, and why it is not a mistake.** The documented
Chapter 5 chain (`app/internal/seed/seed.sql:112-117`,
`docs/app01-design.md` §8.5) is: loot the `svc-idp-ldap` credential out of
app01's portal database, then bind to LDAP on `dc01`. app01 is not
domain-joined, so that credential is the only pivot. Deny DMZ → dc01:389/636 and
the pivot dies, and so does the chain. The rule stays — narrow, logged, and
commented in the playbook with the reason, so it reads as a deliberate design
decision rather than an oversight. This is the "planted weakness, documented"
pattern the rest of the project uses.

Worth one line in the docs: `3268` / `3269` (global catalog) stay **closed** on
purpose. The pivot needs a bind on `dc01:389`, not a forest-wide search. An
attacker-noticeable difference is a design decision, not an oversight.

**Where the allowed set lives.** Recommended: a small `dmz.allow` list in
`lab.yaml`, read by the playbook, alongside the rest of the policy facts. The
precedent is `app.weaknesses` — a design decision expressed as data, read by
a playbook, asserted by a validator. The reason is that "the DMZ may reach dc01
on 389/636" is the single most consequential decision in this wave, and putting
it in one auditable list makes widening or narrowing the boundary a `lab.yaml`
diff a supervisor can read rather than a line buried in Ansible. The
alternative — hardcoded rules with a long comment — is defensible, but then the
boundary is only discoverable by reading a playbook.

### B5. Control-node reachability

`app.yml`, `app_validate.yml` and `ad_validate.yml` reach app01 from the
control node. With the DMZ on its own bridge, the control node has no path
unless one is created.

**Decision: the control node gets an address on every bridge**, via the
`monitor` key and the loop in A5 — the same thing `hosts.yml:69` already does
for the LAN leg. Consequences to state in the docs rather than leave implicit:

- The hypervisor is a member of every zone. This is a provisioning fact outside
  the attacker model (the attacker never holds the hypervisor, and the model is
  fixed as external-only), so it does not weaken the boundary claim.
- What the boundary governs is **VM-to-VM** traffic, which is what Chapter 5
  measures. Say that in the DMZ section, in those words.
- Note that once the control node is on the DMZ bridge, its SSH to app01 is
  same-subnet L2 and does not traverse OPNsense at all; the boundary still
  applies to everything app01 sends *out* of the DMZ, which is the point.
- The alternative — no address, route `10.0.10.0/24` via `10.0.0.1`, plus a
  firewall rule admitting the control node — is arguably more faithful but buys
  the same bypass expressed as a rule, and adds a routing dependency that makes
  `app.yml` fail for a reason that is not the app. Not chosen. If wanted later,
  it is a `monitor` removal plus one rule.

**Ansible's path to the OPNsense API never crosses the DMZ** — the control node
talks to `10.0.0.1` over the LAN leg. So the DMZ rules cannot lock Ansible out
of the API. That is why the DMZ works at all, and it is worth one sentence in
the playbook header.

### B6. Two things the DMZ breaks, and their fixes

**`app_validate.yml`'s SSRF probe.** Its defaults are
`http://127.0.0.1:8080/healthz` and `http://10.0.0.1/` (`:185-187`), both IP
literals on purpose, and the ON assertion (`:791`) requires **at least one**
target to answer with a body. The targets are fetched by the portal, on app01,
from the DMZ — so `http://10.0.0.1/` is now refused by rule 4, and the ON
assertion can fail for a reason that is not the toggle. That is exactly the
failure class the file's own comment at `:177-183` already rejects other
targets for.

**Fix: drop `http://10.0.0.1/` from the default list; keep
`127.0.0.1:8080/healthz`.**

- `127.0.0.1:8080` is the portal's own loopback listener. It is an IP literal
  (so the hardened check refuses it for the right reason), it answers a body
  (`ok` from `/healthz`), and it is **boundary-independent** — no DMZ rule can
  change its outcome. It is therefore a *better* ON-evidence target than the
  LAN address ever was.
- The LAN-address target moves to the documented side: still reachable via the
  `LAB_APP_SSRF_TARGETS` override, with a comment recording the new expected
  outcome. With the toggle ON it now returns **nothing** — and that absence is
  **evidence of the boundary**, not of the toggle. That reframes the probe
  honestly: the SSRF *toggle* is proven by loopback, the *boundary* is proven
  by the LAN target being dropped.
- Update `app/internal/web/api.go:296`'s comment the same way: from the DMZ,
  `http://10.0.0.1/` is the firewall's LAN address and is deliberately
  unreachable; the reachable management address is `10.0.10.1`, and the ruleset
  deliberately does not open `:443` there either. Cite the reason, so the
  comment teaches the boundary instead of contradicting it.

Worth writing down as a result rather than hiding: the firewall now **bounds**
the SSRF weakness. Addresses the portal could previously reach internally are
denied, so the weakness's blast radius is a network control's doing. That is a
defense-in-depth observation Chapters 4 and 5 should claim deliberately.

Rejected alternative: retarget the default to `http://10.0.10.1/`. It would
work only if the ruleset allowed DMZ → `10.0.10.1:443`, which means opening the
firewall's management UI to the attacker's zone — a real weakening of the
boundary this wave builds, in exchange for a validator target. `127.0.0.1`
costs nothing.

**`ad_gpo.yml:160`** grants WinRM from `-RemoteAddress 10.0.0.0/24`. Still
correct — only Windows VMs are in the LAN. Add a comment saying the DMZ is
excluded on purpose, so a future reader does not "fix" it.

**Also flagged, not fixed:** the WAN port forward in `opnsense.yml` may need a
linked filter pass rule. If it does, add it in the same idiom and record the
finding in the verification-pending file. Do not fix it blind — it is the same
unverified controller.

### B7. Ordering in `deploy.sh`

```
ansible hosts.yml            # every segment bridge + per-segment VXLAN + addresses + resolvectl
ansible edge.yml             # idempotent no-op for vm-wan0
terraform apply              # app01 gains bridge vm-dmz0; opnsense01 gains NIC3
ansible start_vms.yml --tags gateway
ansible opnsense.yml         # existing DHCP/DNS pinning (app01 row -> 10.0.10.10)
ansible opnsense_dmz.yml     # NEW: assign opt1, DHCP range, filter rules, Unbound; assert the address
ansible start_vms.yml --tags members
dc.yml / ad.yml / ad_gpo.yml / client.yml / app.yml / app_validate.yml / ad_validate.yml
```

The two constraints, both satisfied above:

1. `hosts.yml` and `edge.yml` move **above** `terraform apply` (see A4).
2. The DMZ interface must be **assigned, DHCP-serving and filtered before app01
   boots**, or app01 either gets no lease or sits on a segment with no policy.
   Both are before `--tags members`.

**Failure behaviour of the intermediate states, stated plainly because it is
reassuring and non-obvious:**

| State | Result |
|---|---|
| `opnsense_dmz.yml` never runs | app01 gets **no DHCP lease** (dnsmasq is not serving the DMZ), so `app.yml` fails loudly at `wait_for_connection`. Fails closed. |
| Interface assigned, no filter rules yet | app01 has no address at all (the address is the manual step) and, once it does, **denied by default** — there is no pass-any-per-interface automatic rule to leave on. Safe; see B3. |
| Rules written but not applied | The running ruleset is still the previous one, so the DMZ is still denied. The window is between the write and the `apply`, and it closes on the next task. |

`opnsense_dmz.yml` holds a third state, and it is the one `deploy.sh` must not
swallow: on a fresh lab it **stops the deploy on purpose** at the address
assert. Everything else it configures has already been configured by then, the
DMZ is fail-closed in the meantime, and the alternative is a deploy that reports
success and leaves app01 unreachable. Set the address (the play prints the exact
UI steps), re-run the playbook, carry on from the next step. **Do not add
`|| true` to that step** — it converts the one honest failure into a fifteen
minute timeout inside `app.yml`, which is strictly worse to debug.

---

## Part C — the VPN pool

Add the `vpn:` block from A2, publish it from `lab_inventory.py` as `lab_vpn`
on `all.vars`, and record in `docs/TODO.md` that **nothing reads it yet** — the
established pattern in this repo for a fact declared ahead of its consumer
(`ad.weaknesses` is the other one).

Two honesty fixes that go with it:

- `docs/redclay.yaml:659` lists `VPN_initial_access` under `core_supports_well`.
  That is **false** and stays false until something implements a server. Move it
  out of that list into a plainly-marked in-scope-not-built form.
- `docs/thesis-plan.md:497-501`'s accuracy warning should be narrowed rather
  than deleted — see the table below.

Update the places that already assume a pool with no server:
`red-clay/Overview.md` (keep the `planned` styling — still true),
`red-clay/Attacker/Attack Paths.md:12,22-23,32`, and
`docs/redclay.yaml:456-460`, whose `conceptual_zones` lists three zones and no
VPN.

---

## Verification

Precondition: `docs/roadmap.md` Phases 0-1 complete and green. This procedure
is meaningless on a down lab, and Phase 1's own hazards must be behind us.

`docs/roadmap.md:32` sets a standing rule: capture evidence from Phase 1
onward. The artifacts named below are Chapter 5 material, not a smoke test.

### Offline, per commit, before any boot

```bash
terraform -chdir=terraform validate && terraform -chdir=terraform fmt -check -recursive
ansible-playbook --syntax-check ansible/playbooks/{hosts,edge,start_vms,opnsense_dmz,app_validate}.yml
ansible-inventory -i ansible/inventory/lab_inventory.py --host host_a | python3 -m json.tool
ansible-inventory -i ansible/inventory/terraform_vms.py --list | grep -A3 '"app01"'
python3 -c "import yaml;yaml.safe_load(open('lab.yaml'))"
grep -rn 'vm-br0\|bridge_name' --exclude-dir=.git --exclude-dir=vendor .
```

### Layer 0 — preconditions, once, by hand

```bash
sudo systemctl enable --now sshd docker
sudo cp --reflink=auto /var/lib/libvirt/images/ubuntu24.04-base{,-bck}
sudo chown libvirt-qemu:libvirt-qemu /var/lib/libvirt/images/ubuntu24.04-base
virsh -c qemu:///system undefine ubuntu-base --nvram      # AFTER the backup
ssh clayface@10.0.0.30 'cat /etc/netplan/*.yaml'
```

The `undefine` is the highest-risk item in the whole roadmap: the live builder
domain points straight at the read-only base image, and booting it corrupts
every overlay stacked on top, silently.

### Layer 1 — the L2 exists

```bash
ip -br link show | grep -E 'vm-(lan|dmz|wan)0'
ip -br addr show vm-lan0 ; ip -br addr show vm-dmz0
resolvectl dns vm-lan0
bridge link show | grep vxlan
```

Expected: three bridges up; `vm-lan0 10.0.0.2/24`, `vm-dmz0 10.0.10.2/24`,
`vm-wan0` with no address; `resolvectl dns vm-lan0` shows `10.0.0.1`. With
`host_b` commented out, **no VXLANs** — `ip link show vxlan100` failing is
correct and matches today. Re-enable `host_b` and expect both `vxlan100` and
`vxlan101`, each enslaved to its own bridge and never crossed.

### Layer 2 — the terraform claim

```bash
terraform -chdir=terraform plan
```

Expected `~` (update in place) on `app01` and `opnsense01`, **not** `-/+`. If
it says "must be replaced" anywhere, **stop** — the RequiresReplace reading is
wrong, state churn and the overlay-disk question come back, and the strategy
changes. This single plan output is the load-bearing verification of the
riskiest assumption in the plan.

```bash
terraform -chdir=terraform output -json vms | python3 -m json.tool
```

Expected: `app01` carries `network: dmz`, `bridge: vm-dmz0`, `ip: 10.0.10.10`.

### Layer 3 — the boundary is real

**3a. app01 has exactly one NIC.** The most important check in the document.

```bash
ssh clayface@10.0.10.10 'hostname; ip -br addr'
```

Expected: **one** interface carrying an address, `10.0.10.10/24`. A second
interface with a `10.0.0.x` address means app01 is dual-homed, the boundary is
bypassable at L2, and every result below is void.

**3b. The DMZ cannot reach the internal zone, except on the allowed ports.**
Run the probe **on app01** — the control node is L2-adjacent to the DMZ bridge,
so its own reachability proves nothing about the filter.

```bash
ssh clayface@10.0.10.10 'for p in 389 636 3268 3269 88 135 445 5985 3389; do
  timeout 3 bash -c "echo > /dev/tcp/10.0.0.10/$p" 2>/dev/null \
    && echo "$p OPEN" || echo "$p closed"; done'
```

Expected: `389 OPEN`, `636 OPEN`, everything else `closed`.

**3c. The firewall's own LAN address is unreachable from the DMZ.** This is the
check that proves the automatic rules were really disabled — it is how B3's
failure mode gets caught:

```bash
ssh clayface@10.0.10.10 'curl -s -m 3 -o /dev/null -w "%{http_code}\n" http://10.0.0.1/; echo "rc=$?"'
```

Expected `000` / non-zero `rc`. **A `200` here means the DMZ is unfiltered.**

**3d. The rest of the internal subnet is dark.**

```bash
ssh clayface@10.0.10.10 'nmap -Pn -p 88,135,139,445,3389,5985 --open 10.0.0.0/24'
```

Expected: no output.

**3e. The reverse direction.** Once `dc01` is back in `lab.yaml` and booted:

```powershell
Test-NetConnection -ComputerName 10.0.10.10 -Port 443
```

Expected `TcpTestSucceeded : False`, and the deny rule's log entry in OPNsense
— that log line is the Chapter 5 artifact.

### Layer 4 — the boundary is not *too* tight (the documented chain survives)

**The test that matters most for the thesis, and the one most likely to be
skipped.** If it fails, Chapter 5's chain has no second hop.

```bash
ssh clayface@10.0.10.10 'ldapsearch -x -H ldap://dc01.clayface \
  -D "svc-idp-ldap@clayface.local" -w "$LAB_USER_PASS" \
  -b "DC=clayface,DC=local" "(objectClass=user)" sAMAccountName | head -20'
```

Expected: a successful bind and a user list. A refusal here means the DMZ rule
is wrong, **not** that the AD layer is broken — discriminate before debugging
the wrong component. This also confirms the DNS allowance (rule 2), since the
bind uses the hostname.

### Layer 5 — OPNsense's own view

```bash
ssh root@10.0.0.1 'grep -A12 "<interfaces>" /conf/config.xml'
ssh root@10.0.0.1 'grep -B2 -A30 "<filter>" /conf/config.xml'
```

Expected: `opt1` with its IP and device; the `descr`-marked rules on `opt1` and
**no auto-generated pass-any rule for `opt1`**. UI cross-check: Interfaces →
`[opt1]` shows the automatic-rules box unchecked; Firewall → Rules → DMZ lists
only ours.

### Layer 6 — the tier still validates

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/app_validate.yml
ssh clayface@10.0.10.10 'nmap -Pn -p- 10.0.10.10'
```

Expected: green with all eleven toggles `true`, tally reporting no
declared-but-unasserted toggles; nmap showing only 80 and 443 open on app01.
Note that this playbook reaches app01 over the DMZ leg — worth a comment so the
next reader is not surprised.

### Layer 7 — idempotence

Run `opnsense_dmz.yml` twice; second run reports 0 changed. Run `deploy.sh`
end-to-end once more; Ansible reports 0 changed throughout
(`docs/roadmap.md` metric 1).

### The six artifacts to keep

The `terraform plan` showing in-place update; `ip -br addr` on app01 showing one
interface; the 3b port scan; the 3c `curl` returning nothing; the OPNsense deny
log entries; and the successful `ldapsearch` bind. Together those are the "a
boundary exists and it is crossed deliberately" result.

---

## Risk and rollback

| State | What happens | Verdict |
|---|---|---|
| Commit 1 only | Plan is clean; `vm-br0` still exists because `lab.yaml` still says so; the dmz bridge exists and is empty | Safe |
| Commit 2, `vm-lan0` **not** created on the host | Apply fails at the Update: stop, undefine, define OK, then `virsh start` → `Network bridge vm-lan0 not found`. Apply exits non-zero, domain left undefined | Recoverable: create the bridge, re-apply. **No data loss.** The sharpest failure mode, and the reason for the reorder |
| Commit 2, bridge created | Both domains restart onto the new bridge | Safe |
| Commit 3, `opnsense_dmz.yml` not run | app01 boots on `vm-dmz0` with no lease → `app.yml` fails loudly | Fails closed |
| Commit 3, assigned but auto-rules on | app01 reaches everything routable | **Dangerous.** Prevented by B3 |

**Commit 2 rollback:** revert the one line, `terraform apply`, `hosts.yml`
recreates `vm-br0`. Because the update is in-place and the overlay is a separate
`libvirt_volume`, the guests' filesystems are untouched — the only
guest-visible effect is a reboot.

**Commit 3 rollback:** `lab.yaml` back to `network: lan` / `ip: 10.0.0.30`,
revert the module's third NIC, apply, re-run `opnsense.yml` (the update path
rewrites app01's dnsmasq row back — nothing to delete by hand). **The stale DMZ
filter rules remain in OPNsense and must be deleted by hand** — the playbook
never deletes a row, deliberately. Say so in the rollback note rather than
letting someone discover it.

**The teardown worktree.** `.claude/worktrees/chore+lab-teardown/` (branch
`worktree-chore+lab-teardown`, not on main) is a semantic conflict, not a
textual one:

- `ansible/playbooks/teardown.yml:529` reads `lab.network.vxlan_name`,
  `lab.network.bridge_name` and `lab.network.wan_name`. All three keys vanish in
  commit 1. `include_vars` still succeeds (it loads the whole file), so the
  failure is a Jinja undefined-variable error on the `--network` path — loud,
  but not until someone runs it.
- `teardown.sh:72,141` and that worktree's `CLAUDE.md:294` hardcode `vm-br0` /
  `vm-wan0` / `vxlan100` in user-facing prose.
- It carries its own `CLAUDE.md` copy, so the `vm-br0` edits conflict.

Land commits 1-3 on main, then rebase the teardown worktree as one follow-up
where the change is small and mechanical — replace the three key reads with a
loop over `lab.networks | dict2items` deriving `vxlan<id>`, mirroring the shape
the hypervisor playbook now uses. Do not update the worktree in parallel; it
would mean editing the same three reads twice.

**Bridges are runtime-only.** `hosts.yml` creates them with `ip link add` and
the per-leg resolver with `resolvectl`; both are lost on reboot. This plan
doubles the number of runtime-only objects. Worth a `docs/TODO.md` entry —
NetworkManager persistence would make the lab survive a reboot.

---

## Documentation updates

**A. Assertions that are now false — change, do not soften.**

| File | What it says | New text |
|---|---|---|
| `docs/TODO.md:29-54` | "no boundary exists to cross"; the *Update* paragraph calling APP01's compromise equivalent to an internal foothold | Item **done**, with the closure date. Keep the "built out of order" history — it is honest |
| `docs/TODO.md:5-27` | "the whole `10.0.0.0/24` lab subnet is a fact and exists nowhere in lab.yaml" | **Done**: the map declares subnet/gateway/dns/monitor; `start_vms.yml`'s duplicate is deleted. Keep the `LAB_DOMAIN` exception note |
| `docs/app01-design.md:64` | "There is no network segmentation. One flat L2, `vm-br0`" | New reality, plus the **one NIC** property |
| `docs/app01-design.md` §3.1 | the whole "Not a DMZ host" section | Rewritten: app01 *is* a DMZ host behind a filtered boundary — **and** the honest successor statement, that the control node is multi-homed into every zone by construction |
| `docs/app01-design.md` §15, §17 | "No network segmentation, the single largest gap"; "When does segmentation land?" | Struck, answered, pointed at `docs/network-design.md` |
| `docs/app01-design.md` §3.2 | pinned-MAC example hardcodes `ip: 10.0.0.30` | `10.0.10.10`, `network: dmz` |
| `docs/roadmap.md:93-105` | the Phase 2 DMZ item | Done; keep the connectivity-matrix item, now richer (per-zone) |
| `docs/roadmap.md:63-65` | host prep `vm-br0`, `resolvectl dns vm-br0` | `vm-lan0` |
| `docs/thesis-plan.md:159-168` | "APP01 sits on the same flat L2 … say 'containerized application host', not 'DMZ host'" | **The instruction is now obsolete.** app01 *is* a DMZ host. Say so, and say which zone dc01/client01 are in |
| `docs/thesis-plan.md:497-501` | the Overview accuracy warning | Successor below |
| `docs/redclay.yaml:157` | `not_yet: DMZ segmentation (still on the flat vm-br0 L2)` | Remove; add `address: 10.0.10.10` beside app01, since that block is the deliberate "what is actually deployed" record |
| `docs/redclay.yaml:659` | `VPN_initial_access` under `core_supports_well` | **Must change.** Nothing supports it |
| `docs/app01-verification-pending.md:24` | `ip -br link show vm-br0` | `vm-lan0`; and add the new unverified items (interface assignment, the auto-rules field, rule-body field names, the port-forward's linked filter rule) in that file's §4 idiom |

**B. The accuracy warning, handled honestly.** `docs/thesis-plan.md:497-501`
records that `Overview.md` shows three internal subnets plus a VPN pool while
reality is one flat L2. After this wave, reality is: **one flat internal L2
(`lan`, `10.0.0.0/24`), a real DMZ (`dmz`, `10.0.10.0/24`), a declared but
unimplemented VPN pool (`10.0.100.0/24`), and the user/server split still
aspiration.** The right move is a short explicit "what is real and what is not"
table rather than deletion — the diagram is the supervisor-approved target, and
the repo already has the `built`/`planned` class convention. Extend that to
zones rather than breaking it.

**C. Vault prose** — keep `[[wikilinks]]` and the author's voice; no
reformatting.

| File | Edit |
|---|---|
| `red-clay/Overview.md` mermaid | Move APP01 and the DMZ subgraph from `planned` to `built`; VPN and KALI stay `planned`; the internal user/server split stays `planned` |
| `red-clay/Overview.md` addressing table | Add a Status column or sentence. The "currently runs one flat `10.0.0.0/24` on `vm-br0`" note becomes: LAN is `vm-lan0`, the DMZ exists as `vm-dmz0`, `host_b` is commented out |
| `red-clay/Architecture/Infrastructure.md:27` | the edge label `VXLAN 100 over wlan0 / bridge vm-br0` → both overlays and both bridges |
| `red-clay/Report/Actions/OPNSense.md:6,8` | LAN interface is `vm-lan0`; **add the DMZ interface paragraph** — this file is the only record of how the OPNsense image was built, so the third NIC's assignment belongs here |
| `red-clay/Report/Actions/Bare Metal VM.md:11,19,21` | bridge is `vm-lan0`; add the second bridge and second VXLAN commands, since this file is the hand-run version of what `hosts.yml` now does |

**D. `CLAUDE.md`** — the six `vm-br0` lines (`:92`, `:172`, `:174`, `:176`,
`:177`, `:301`); note that `:176-177`'s `nmcli` advice is wrong regardless (A5).
The architecture section becomes two segments with the DMZ's one-NIC property
stated as the security claim. The mirrored-constants list gains the DMZ
gateway. A short new section on the boundary: what it allows, why 389/636
specifically, and that widening it is a design change.

**E. New file: `docs/network-design.md`.** The DMZ deserves its own design
document in the shape of `docs/app01-design.md`: zones and addressing; why the
boundary sits at OPNsense; the one-NIC rule and its rationale; the deliberate
LDAP allowance and its provenance; **"what the boundary is not"** (the
hypervisor is multi-homed into every zone — a provisioning fact, outside the
attacker model); the VPN pool as declared-not-implemented; the segment/VNI
separation rule; and a verification section mirroring the one above.

**F. New file: `docs/opnsense-image.md`.** There is currently none. Record how
the image was built, the LAN and WAN interface assignment, and the DMZ
interface (whether API-assigned or hand-assigned).

---

## Out of scope

The user/server internal split (`10.0.20.0/24`, `10.0.30.0/24`); a real VPN
server; the Kali WAN-side attacker VM (`docs/roadmap.md` Phase 3); IDP01
(Phase 5); redesigning any weakness toggle beyond the SSRF default change;
Packer-based image automation (Phase 7).

## Open judgement calls

Recorded rather than decided, so the choice can be made at implementation time:

1. **`dmz.allow` in `lab.yaml`** versus hardcoded rules in the playbook. B4
   recommends the former, with the `app.weaknesses` precedent; the cost is
   policy living in the facts file.
2. **A new playbook** versus a tagged section of `opnsense.yml`. B2 recommends
   the new file on blast-radius grounds; the fallback is `--tags dmz`, which
   keeps the file but loses the attributable `deploy.sh` step.
3. **Whether `opnsense.yml` also reads `lab.yaml`** for `opnsense_host`,
   closing the fourth copy of `10.0.0.1`. Recommended, in commit 1 where the
   risk is lowest — but it changes a file whose invariants everything depends
   on.
4. **The pinned DMZ NIC MAC** (A4). Recommended, because it converts NIC
   renumbering from a silent WAN misconfiguration into a named assertion
   failure.
5. **`deploy.sh`'s `no_proxy` literal** (A2). Recommended to leave, with a
   comment.
