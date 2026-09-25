# Clayface — the OPNsense edge image

How `opnsense.qcow2` — the disk `opnsense01` boots — was built, what is baked
into it, what is applied at deploy time, and the one fact about it that is
neither.

Status: **hand-built, in place.** There is no base image and no overlay: the
running VM's disk *is* the artifact. `terraform/modules/opnsense/main.tf` points
at `/var/lib/libvirt/images/opnsense.qcow2` directly, and the installed builder
domain (`opnsense`) boots the same file — so a change to one is a change to the
other, and losing that file loses the lab's DHCP, DNS and `clayface` zone
together. [`docs/opnsense-image-plan.md`](opnsense-image-plan.md) plans the
migration to the base-image-plus-overlay pattern every other guest uses; nothing
in it has been executed.

---

## 1. Why this document exists

`base-image/` documents the Windows images, `docs/app01-design.md` §5 documents
the Ubuntu one, and this file is the third. It was written late, and the gap it
fills is the reason the DMZ wave needed a manual step (section 4): facts about
the edge image that live nowhere but the disk cannot be checked, and a design
that assumes one of them is automatable will be wrong.

---

## 2. How it was built

Hand-built on `host_a`, from the OPNsense installer ISO in `vm/images/`. There
is no unattend, no sysprep and no automation: a stock 26.7 install, then
configuration through the GUI.

Facts decided at build time, in the order they mattered:

- **ZFS, not UFS.** This is what makes the snapshot/rollback API work
  (`core/snapshots/isSupported` answers `{"supported":true}`, one boot
  environment `default`), which is the edge VM's escape hatch.
- **Three vCPU.** Two is not enough — the dashboard stalls under load. Recorded
  in the build notes as a plain observation and worth keeping: it is the kind of
  thing that gets "optimised" back down and then reads as a broken firewall.
- **A root password, set by the installer.** Lab-only and deliberately weak,
  consistent with every other credential in this lab.
- **One API key pair on `root`.** This is the bootstrap credential, and it is
  load-bearing: `ApiControllerBase` authenticates **only** through the Local API
  (key/secret) authenticator, so there is no user-password route into `/api/` at
  all. Without a key baked in, there is no way to make the API calls that would
  create one. The pair lives in the gitignored `deploy.env` as
  `OPNSENSE_API_KEY` / `OPNSENSE_API_SECRET`, and it is created in the UI at
  System → Access → Users.
- **The WebGUI listens on HTTP, not HTTPS.** This deviates from the factory
  default of `https` and there is no API for it. It is why every playbook that
  talks to the box probes https first and falls back to http with
  `validate_certs: false`, rather than assuming a scheme. If a stock image is
  ever built, take the factory `https` and stop deviating.
- **SSH is off.** Nothing in the repo reaches the edge over SSH; the API is the
  only management path.

The DNS domain `clayface` is configured here too — dnsmasq's domain plus an
Unbound forward — which is why it is a documented mirrored constant in the
repo rather than a `lab.yaml` key: `lab.yaml` cannot change what the image baked
in. See the exceptions list in `CLAUDE.md`.

---

## 3. Interfaces

NIC order in the terraform module is device order, and the guest names its
interfaces by it. Order matters more here than anywhere else in the lab,
because the firewall rules name interfaces.

| Guest name | Position | Bridge | Role | Address | Assigned by |
|---|---|---|---|---|---|
| `vtnet0` | 1st | `vm-lan0` | LAN (inside) | `10.0.0.1/24` | baked |
| `vtnet1` | 2nd | `vm-wan0` | WAN (outside, NAT) | DHCP from the home LAN | baked |
| `vtnet2` | 3rd | `vm-dmz0` | DMZ | `10.0.10.1/24` | **hand, once** (section 4) |

**The DMZ NIC is appended third, never inserted.** Disks and interfaces share
the ordering space, so inserting a NIC ahead of an existing one renumbers
everything after it and silently repoints the firewall rules at a different
interface. `terraform/modules/opnsense/main.tf` carries that rule as a comment
at the interface list, and pins the DMZ NIC's MAC — derived from the VM name
with a `-dmz` suffix, so it cannot collide with `opnsense01`'s own LAN MAC —
purely so that a future renumbering becomes a named assertion failure in
`opnsense_dmz.yml` rather than a boundary that quietly stopped existing.

The LAN and WAN assignments are baked. Their addresses are baked too: `10.0.0.1`
appears in the repo as a mirrored constant (`opnsense_host` in `opnsense.yml`
and `opnsense_dmz.yml`, `lab_dns_forwarder` in `dc.yml`, the `no_proxy` list in
`deploy.sh`), and the DMZ's `10.0.10.1` was added to that class by this wave.
`start_vms.yml` is deliberately *not* on that list: its `lab_dns_ip` reads
`networks.lan.dns` from lab.yaml, so it follows a lab.yaml edit rather than
having to be changed alongside one.

---

## 4. The one fact that is neither baked nor automatable

**The DMZ interface's IPv4 address.** It is set by hand, once per lab build, in
the GUI.

OPNsense 26.7 has no REST path to an interface's IPv4 configuration. The
`interfaces/assignment` controller — the one that owns assignment, and that
*does* work over the API — is backed by a seven-field model (`descr`,
`identifier`, `icon`, `optgroup`, `if`, `lock`) covering assignment only. A
`setItem` carrying `type4` / `ipaddr` / `subnet` / `enable` answers
`{"result":"saved"}` and writes **none** of them: the config gains `<if>`,
`<descr>` and `<lock>`, and nothing else. `interfaces/assignment/pending` 404s
on this release. The address lives in `config.xml`, written by the legacy
`src/www/interfaces.php` form, which authenticates by **GUI session plus CSRF**
(`guiconfig.inc` → `authgui.inc`; `LegacyCSRF::checkToken()`) — there is no HTTP
Basic or API-key route into legacy pages. Only 27.1 (master) has the model that
makes the address settable; every 26.7.x point release still has the thin one.

The steps, once per lab build — Interfaces → Assignments → the DMZ row:

```
IPv4 Configuration Type : Static IPv4
IPv4 address            : 10.0.10.1/24
Block private networks  : OFF   (the DMZ is RFC1918 — ticking this drops
                                 the whole segment)
Block bogon networks    : OFF
Enabled                 : ON
```

Save, Apply. `ansible/playbooks/opnsense_dmz.yml` does everything else and
**asserts** this, reading the address back and stopping with these instructions
if it is missing. It does not scrape the legacy form: that would need a new
`LAB_OPNSENSE_PASS` credential in `deploy.env` and untestable session machinery,
to replace a one-time form with a script.

Deferring it is safe because the DMZ is **fail-closed**: OPNsense's generated
system ruleset has no pass-any-per-interface rule, so an interface with no pass
rule written for it is denied. While the address is unset the segment is
unreachable rather than unfiltered. `docs/network-design.md` §3 and §8 have the
long version.

**If the image migration in `docs/opnsense-image-plan.md` is taken up**, this
address is a candidate for baking alongside the LAN IP and the DNS domain —
which is exactly the mirrored-constant outcome that plan argues against, and it
is the honest resolution given the API has no path to it.

---

## 5. What the image does *not* carry

Everything else is applied at deploy time by playbooks, and the split is worth
stating so a future change lands on the right side of it:

| Applied by | What |
|---|---|
| `hosts.yml` / `edge.yml` | the bridges themselves, and the VXLAN overlays |
| `opnsense.yml` | DHCP reservations and DNS records for every pinned VM, one dnsmasq row each, driven by the pinned MAC chain |
| `opnsense_dmz.yml` | the DMZ interface assignment, its dnsmasq range, Unbound on the new leg, and the whole filter policy |
| **the image** | the OS, the API key, the LAN/WAN interface assignment and addressing, the `clayface` DNS domain, the DMZ interface's *address* (section 4) |

No server roles, no policy, no reservations. The image is "a stock OPNsense that
knows which NIC is which and what the lab's domain is called" — and even that
last part is scheduled to move to the playbooks if the migration plan lands.
