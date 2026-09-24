# OPNsense base image — migration plan

Status: **plan only. Nothing in this document has been executed.**

Companion to [`network-plan.md`](network-plan.md) — the DMZ / segment-rename /
VPN-pool plan. The two overlap in three places and §1.1 says which plan owns
what. Read that section before acting on either.

Goal: bring the OPNsense edge VM onto the same footing as every other guest —
a published base image plus a per-VM overlay, with all lab-specific facts
applied by Ansible at deploy time instead of baked into a hand-built disk.

## 1. Why this is needed

`opnsense01` is the only VM in the lab that does not follow the repo pattern.
It was created by hand on top of a volume that already existed, and the
terraform module was written to describe that one VM rather than to create
one.

Observed state (read-only survey of the live host):

| Concern | opnsense01 | every other module |
| --- | --- | --- |
| volume | none — module has no `libvirt_volume` | `libvirt_volume` with `backing_store.path = var.base_image_path` |
| disk path | hardcoded `/var/lib/libvirt/images/opnsense.qcow2` | `"${upper(var.name)}.qcow2"` in pool `default` |
| pinned MAC | none | `sha256(var.name)` → `52:54:00:xx:xx:xx` |
| `mac` output | none — nothing to merge in `terraform/outputs.tf` | exported, consumed by `vms_pinned` |
| CPU model | `qemu64`, `svm` disabled | `cpu { mode = "host-model" }` |
| console | no `<graphics>` / `<video>` / console device | SPICE + qxl |
| lab.yaml entry | `edge_vm:` has only `name:` | `vm_placements:` entry with `host`, `module`, `ip` |
| base image | none exists | `ws-base.qcow2`, `client-base.qcow2`, `ubuntu24.04-base.qcow2` |

Two consequences beyond untidiness:

1. **There is no base image.** The running VM's disk *is* the artifact. The
   installed builder domain `opnsense` boots the *same*
   `/var/lib/libvirt/images/opnsense.qcow2` as the running `opnsense01` — the
   read-only-backing-file hazard CLAUDE.md documents for `client-creator`, but
   worse, because here there is not even an overlay to corrupt. Any change to
   one is a change to the other.

2. **Reproducibility.** The current disk carries hand-made configuration that
   nothing recreates. If that volume is lost, the lab loses DHCP, DNS and the
   `clayface` zone — the thing every other playbook depends on.

The counter-argument that made this tolerable until now: OPNsense is the one
component treated as "always correct", the image owns the DNS domain
(`CLAUDE.md`, the `.clayface` exception), and the DNS domain is a mirrored
constant precisely because the image baked it in. That assumption is what this
plan retires — not by putting the domain in `lab.yaml`, but by making the
image stock so the *playbook* owns it.

### 1.1 Relationship to `docs/network-plan.md`

`docs/network-plan.md` plans the DMZ, the `vm-br0` → `vm-lan0` rename, the
`networks:` map in `lab.yaml`, and the VPN pool declaration. This plan is
**complementary, not competing**: it covers what that plan deliberately leaves
alone — that OPNsense has no base image and no overlay, so its disk *is* the
artifact.

Three places where they touch, and who wins:

- **`10.0.0.1` and `10.0.0.0/24`.** network-plan Part A already moves these
  into `lab.yaml` as `networks.lan.{gateway,dns,subnet}` and deletes four
  copies of the address. This plan adopts that shape (§7 OD-1) rather than
  inventing a second home. If Part A lands first, that half of OD-1 is already
  done.
- **The third NIC (DMZ).** network-plan commit 3 adds a NIC to
  `modules/opnsense`. That interacts with the NIC-order rule in §2.2.
- **`docs/opnsense-image.md`.** network-plan's documentation section (item F)
  already names that file as required and says it does not exist. This plan is
  what produces its content: §2 is the "what is baked" half, §3 is the "how it
  was built" half.

**Sequencing recommendation: this plan first.** network-plan commit 1's stated
proof is "`terraform plan` reports no changes", and that proof is only
meaningful once the module actually manages a volume. Doing the image
migration afterwards rewrites the module twice.

### 1.2 One correction to `docs/network-plan.md`

`network-plan.md` §B2 states that no `assignments` controller is documented and
concludes that interface assignment is unverified inference, with a one-time
hand assignment as the fallback. **That is wrong, and it is checkable.**
OPNsense 26.7 ships `Interfaces/Api/AssignmentController` with
`searchItemAction`, `getItemAction($ifname)`, `setItemAction($ifname)`,
`addItemAction`, `delItemAction` and `reconfigureAction`. Confirmed live:
`GET /api/interfaces/assignment/getItem/lan` answers 200. Interface assignment
is playbook-settable, so the DMZ interface does not need a hand-assignment
fallback — which matters, because a hand-assigned interface becomes a mirrored
constant of the same class as the LAN IP.

This correction came from the same investigation that produced this plan. The
"OPNsense is always presumed correct" question turned up several API paths that
had been written off as unavailable; an earlier conclusion of mine to that
effect was retracted after the source was read. Treat §B2's endpoint table as
worth re-checking the same way before relying on its "unverified" column.

## 2. What the base image must contain

The image must be a **stock OPNsense 26.7 install on ZFS** plus exactly four
build-time facts. Everything else is either already stock or settable over the
REST API.

### 2.1 Must be baked (no API path exists)

1. **Root password.** The installer sets it. Rotatable later via
   `auth/user/set` (`<password type="UpdateOnlyTextField"/>`), but a box needs
   a login to be usable at all.
2. **One API key pair on root.** This is the bootstrap credential. OPNsense's
   `ApiControllerBase` authenticates only through the `Local API`
   (key/secret) authenticator — there is no user-password path into `/api/`.
   Without a key in the image there is no way to run the wizard that would
   create one. This pair is what `deploy.env` holds.
3. **`system.ssh.enabled`** — *only if* deploy-time SSH to the firewall is
   wanted. `openssh_enabled()` returns true only when
   `$config['system']['ssh']['enabled']` is set; the
   `<ssh><group>admins</group>` node present in both the factory sample and
   the live image is inert. No API for it. Decision: default **off**, since
   nothing in the repo uses SSH to the edge.
4. **`system.webgui.protocol`.** Factory default is `https`; the live image
   deviates with `http`. No API. Decision: take the factory `https` and stop
   deviating — `opnsense.yml` already probes https-then-http with
   `validate_certs: false`.

### 2.2 Must be baked for mechanical reasons

5. **ZFS, not UFS.** The live image is ZFS (`core/snapshots/isSupported` →
   `{"supported":true}`, one boot environment `default`). ZFS is what makes
   `SnapshotsController` (`add` / `activate` / `del`) work, i.e. "reset the
   firewall to a known state" as a recovery tool. Keep it.
6. **The same NIC count and order as the terraform module declares.** The
   factory config carries `<if>mismatch0</if>` (lan) and `<if>mismatch1</if>`
   (wan) — placeholders the installer resolves to concrete device names. The
   image will end up with `vtnet0`/`vtnet1` hardcoded. If the builder domain's
   NICs differ in count or order from `terraform/modules/opnsense/main.tf`,
   **LAN and WAN swap silently**: the lab LAN lands on the WAN bridge. Build
   the image with exactly two virtio NICs, bridge first. (The vault's build
   record agrees on the order: `red-clay/Report/Actions/OPNSense.md` — "1 act
   as a LAN and 2nd one act as WAN".)

   **Coordination with `docs/network-plan.md`:** that plan's commit 3 adds a
   third NIC (DMZ). Adding a NIC *after* the base is published is safe — the
   installer's resolution pins only the first two, and a third device comes up
   as `vtnet2` regardless. What is not safe is reordering. Two options: build
   the base with two NICs and let the DMZ append a third, or build it with
   three now so the DMZ becomes a pure config change. Recommend two now and
   let the DMZ be network-plan's change — but record in
   `docs/opnsense-image.md` that the third NIC was never in the image.

### 2.3 Must NOT be baked

Everything below is applied by `core/initial_setup/configure` plus the module
APIs at deploy time. Baking any of it re-creates the problem this plan exists
to solve.

- hostname, domain, language, timezone, DNS servers, `dnsallowoverride`
- LAN IP / CIDR, WAN mode and its fields (ipaddr, gateway, mtu, mss,
  blockpriv, blockbogons)
- interface assignment beyond the installer's own resolution
- the DHCP range and `configure_dhcp`
- dnsmasq `regdhcp`, `regdhcpstatic`, `dhcp.domain`, `dhcp.fqdn`, `dhcp.local`
- the Unbound domain forward for `clayface`
- firewall rules, NAT, aliases, tunables, offloading settings
- `trigger_initial_wizard` — leave the factory value present; the wizard
  clears it, the same way the Windows unattend clears OOBE.

Note the factory defaults already cover several things the live image has:
`timezone Etc/UTC`, `dnsallowoverride 1`, dnsmasq enabled on port 53053 with
`interface=lan`, unbound enabled, `nat outbound mode automatic`, the two
default "allow LAN to any" rules, the `admins` group with `page-all`, console
menu on, `powerd` hadp, bogons monthly. Those need no work.

### 2.4 The lab facts that must MOVE, not disappear

These are currently baked and must be re-homed:

| Fact | Current home | New home |
| --- | --- | --- |
| LAN address `10.0.0.1/24` | image | `lab.yaml` (see open decision OD-1) |
| DNS domain `clayface` | image | `lab.yaml` (OD-1) |
| DHCP range `10.0.0.100-200` | image | `lab.yaml` or the wizard defaults |
| 3 DHCP/DNS reservation rows | image | `opnsense.yml`, from `vms_pinned` |
| Unbound forward `clayface → 127.0.0.1:53053` | image | `opnsense.yml` (already does this) |
| A stale Unbound forward `lab → 127.0.0.1:53053` | image | delete. Provenance: `red-clay/Report/Actions/OPNSense.md` records the original domain as `lab`, later changed to `clayface`; the forward row outlived the rename. Delete it explicitly in the pin phase — nothing else will |

## 3. Base image build procedure

Follow the shape of `base-image/README.md` (the Windows precedent): manual,
documented, with a pre-generalize copy kept as the escape hatch.

1. Create a builder domain with two virtio NICs, the first on the lab bridge,
   booting `vm/images/OPNsense-26.7-dvd-amd64.iso`, installing to a fresh
   qcow2 with **ZFS**.
2. Set the root password at install time.
3. First boot: complete the GUI wizard far enough to have a usable box, then
   create one API key pair on root (System → Access → Users).
4. Do **not** set hostname, domain, LAN address, DHCP range, or any lab
   firewall rule. If the installer forces a LAN address, leave the stock
   `192.168.1.1/24` and let the playbook move it.
5. Optionally set `system.ssh.enabled` and confirm `webgui.protocol`.
6. Shut down, publish as `opnsense-base.qcow2`, keep
   `opnsense-base-bck.qcow2` (pre-publish copy) as the escape hatch.
7. **Undefine the builder domain**: `virsh -c qemu:///system undefine opnsense
   --nvram`. Once published, the base is a read-only backing file forever —
   booting it directly corrupts every overlay above it.

### 3.1 Migration options for the existing volume

- **Option A (recommended): build fresh.** Clean stock image; the live disk's
  baked facts become the playbook's job, which is the whole point. Costs one
  wizard run and a re-bootstrap of the firewall config; the live VM is
  discarded.
- **Option B (fast, does not meet the goal): promote the current disk.** Copy
  `/var/lib/libvirt/images/opnsense.qcow2` to `opnsense-base.qcow2` and build
  overlays on it. Gets the terraform shape right in an afternoon, but bakes
  every lab fact and leaves the playbook unable to own the domain or LAN
  address. Acceptable only as a stopgap.
- Either way, **copy, do not move.** The running VM keeps working until the
  cutover.

## 4. Changes, by area

Ordered so each stage leaves the lab working.

### Stage 1 — terraform module (`terraform/modules/opnsense/`)

- Add a `base_image_path` variable, matching the other modules.
- Add `resource "libvirt_volume" "opnsense"` — name
  `"${upper(var.name)}.qcow2"`, pool `default`, `target.format.type = "qcow2"`,
  `backing_store.path = var.base_image_path`. **`capacity` must be ≥ the base's
  virtual size** — check with `virsh vol-info --pool default
  opnsense-base.qcow2` (the file is root-only, so `qemu-img info` needs sudo).
  Undersizing makes the overlay smaller than its own backing file, which qemu
  refuses to open.
- Point the domain's disk `source.file.file` at the new volume instead of the
  hardcoded path.
- Add `cpu { mode = "host-model" }`.
- Add SPICE graphics + qxl video + console, so the firewall has a console like
  everything else (today it has none — a broken firewall needs SSH or the
  serial console on the host).
- Add a `mac` output derived from `sha256(var.name)`, same construction as
  `modules/domaincontroller`.
- Add `base_image_path` defaults to the `module "opnsense_host_a"` block in
  `terraform/main.tf` (and `_host_b`).
- **Add the gateway module to `module_os` in `terraform/locals.tf`** as
  `gateway = "linux"`. Today `gateway` is absent from that map and the
  `os_family` is hardcoded in `outputs.tf`; adding it makes the typo guard
  cover the gateway too.

### Stage 2 — terraform outputs and lab.yaml

- `terraform/outputs.tf`: merge the opnsense module's MAC into the edge entry
  the way the dc/client/app macs are merged, instead of hardcoding
  `mac = null, ip = null`. The existing comment says to do this when another
  module needs a DHCP reservation — the gateway does now.
- `lab.yaml`: add `ip:` to the `edge_vm` block. This is what puts the gateway
  into `vms_pinned`, which is what makes `opnsense.yml` write a reservation
  for the firewall itself. (Whether that is even desirable is OD-1.)
- If OD-1 resolves to "move the domain and LAN address into lab.yaml", that
  edit happens here, and `opnsense.yml` / `terraform_vms.py` / `dc.yml` stop
  mirroring constants and read `lab.yaml` like everything else. This is the
  same change `docs/TODO.md` already asks for under "Move hardcoded network
  facts into lab.yaml".

### Stage 3 — the bootstrap problem

This is the part with no precedent in the repo and the most ways to get it
wrong.

Stock LAN is `192.168.1.1/24`. The Ansible control node **is** `host_a`, which
sits on the home LAN at `192.168.1.161/24`. So `192.168.1.1` resolves on-link
via `wlan0` and ARPs to the home router, **not** to the VM. Any API call to
the freshly-booted firewall would hit whatever is at that address on the home
network.

Mitigation, in order:

1. Give the builder image a LAN address that does not collide, **or**
2. Have the bootstrap play run `ip route add 192.168.1.1/32 dev vm-br0` before
   the first API call, then remove it after the wizard moves the LAN to
   `10.0.0.1`.
3. The wizard calls `service reload delay` and moves the LAN interface, so the
   play must wait for `10.0.0.1` to answer before continuing. The existing
   scheme probe in `opnsense.yml` is the right pattern; it needs to also
   handle "the address changed".

Split `ansible/playbooks/opnsense.yml` into two phases, or add a first block
guarded by a variable:

- **bootstrap (once per fresh VM):** resolve the API credential, call
  `core/initial_setup/configure` with hostname / domain / timezone / DNS / LAN
  IP+CIRD / WAN mode / root password, then wait for the new address.
- **pin (idempotent, every deploy):** what the playbook does today — dnsmasq
  `regdhcp` / `regdhcpstatic`, one reservation per `vms_pinned` member, the
  Unbound `clayface` forward, reconfigure, verify.

**Ordering hazard:** `configure` is a bootstrap, not a patch.
`flush_network_lan()` deletes and recreates every LAN DHCPv4 range;
`flush_network_wan()` deletes and recreates the gateway rows. Re-running it on
a live lab destroys the reservation rows the pin phase wrote. So the bootstrap
must run before the pin phase and must not re-run — guard it on a fact the box
reports (e.g. skip when the LAN address is already correct), not on a
local marker.

### Stage 4 — deploy.sh

Insert the bootstrap step after `start_vms.yml --tags gateway` and before the
existing `opnsense.yml` step. The current ordering comment in `deploy.sh`
explains why pinning precedes the member boot, and it stays true — the
bootstrap simply goes one step earlier.

### Stage 5 — a validate playbook

`app_validate.yml` and `ad_validate.yml` establish the pattern: a read-only
playbook that asserts the built thing is what the design doc says, including
negative cases. The edge deserves the same treatment, and it is the honest
answer to "OPNsense is always presumed correct":

- the LAN address and DNS domain match `lab.yaml`
- every `vms_pinned` name resolves through the firewall
- the firewall has no rules beyond the default LAN-allow pair, unless a
  documented exception exists
- the stale `lab` Unbound forward is absent
- WAN has an address and a default route (see §5)

### Stage 6 — the image document

Write `docs/opnsense-image.md`, the file `docs/network-plan.md` already marks
as required and missing: what the image contains (§2), how it was built (§3),
the LAN/WAN interface assignment, and — if the DMZ has not landed yet — a note
that the third NIC was never in the image.

## 5. Adjacent breakage worth fixing in the same pass

- **`vm-wan0` has no uplink.** The bridge exists and carries one vnet, and
  nothing attaches a physical interface. So OPNsense's WAN has no address and
  no default route, and the lab has no upstream internet.
  `playbooks/edge.yml` only creates and brings up the bridge. The live WAN
  interface is `dhcp`, so this is a wiring gap, not a config one. Decide
  whether the WAN should NAT to the home LAN (`wlan0`) or stay isolated — for
  a red-team lab isolated may be correct, but then the interface should be
  static and the intent documented, because right now it reads as broken.
- **`edge_vm` has no `ip:`**, so the gateway can never be in `vms_pinned` and
  the firewall never gets a reservation. Fine while its address is baked;
  must be settled once it is not.
- **`ansible/inventory/terraform_vms.py` sets `ansible_host` to
  `<name>.<LAB_DOMAIN>` for every VM**, so the gateway's `ansible_host` is
  `opnsense01.clayface`, which does not resolve. Harmless today because every
  playbook reaches the gateway by its literal address, but it is a trap for
  the first playbook that targets `vms_gateway` directly.
- **`deploy.env` still holds the example's placeholder API key and secret**
  (byte-identical to `deploy.env.example`). Whichever image ships, the
  credential in `deploy.env` must match the key baked into it. Per the
  standing note, this is not a finding — only a functional coupling that the
  new build makes explicit.
- **`docs/TODO.md`'s two network items are `docs/network-plan.md`'s job, not
  this plan's.** This plan settles only the half that plan deliberately keeps
  as an exception (the DNS domain) and supplies the image documentation its
  item F asks for. Do not duplicate that work here.

## 6. Verification

Each stage, in order:

1. **Image:** boot the published base once in a throwaway domain, confirm it
   comes up at stock `192.168.1.1`, that `/api/` answers with the baked key,
   and that `core/snapshots/isSupported` is true. Undefine the throwaway.
2. **Module:** `terraform plan` on a scratch VM name; confirm a volume is
   created with a backing store and that `virsh dumpxml` shows the overlay, the
   `host-model` CPU, the pinned MAC and a console.
3. **Overlay sizing:** `virsh vol-info` on the new overlay vs the base.
4. **Bootstrap:** run the bootstrap against a fresh overlay; confirm the LAN
   moves to `10.0.0.1`, the domain is `clayface`, and the address survives a
   reboot.
5. **Pin:** run the existing `opnsense.yml`; confirm every `vms_pinned` name
   resolves via `nslookup <name>.clayface 10.0.0.1` and that a second run
   changes nothing.
6. **End-to-end:** full `./deploy.sh` from `--destroy` on a clean state, with
   `dc01` and `client01` re-enabled in `lab.yaml` (they are commented out in
   the current working tree). This is the only test that proves the edge is
   reproducible, because it is the only one that starts with nothing.

## 7. Open decisions

- **OD-1 — `10.0.0.1` and `clayface`.** Split, because the two have different
  answers.
  - **The address and subnet: adopt network-plan Part A2's `networks:` map.**
    Do not invent a second home. `networks.lan.gateway` / `.dns` / `.subnet`
    already displace the four copies of `10.0.0.1`; network-plan's open call 3
    asks whether `opnsense.yml` should read it too. Recommend yes.
  - **The DNS domain `clayface`: still a mirrored constant, for now.** Its
    stated justification in `CLAUDE.md` is "the image baked it in", which this
    plan removes — but the domain stays special for a second, independent
    reason: `dc.yml` derives the AD forest name from it (`LAB_DOMAIN` +
    `.local`), so it is one knob moving two things. Moving it into `lab.yaml`
    is a separate, smaller change that should follow the image, not ride along
    with it. `TODO.md`'s exception note should be rewritten at that point, not
    deleted now.
- **OD-2 — bootstrap vs bake the LAN address.** Baking `10.0.0.1/24` into the
  image avoids the `192.168.1.1` collision entirely and costs one violation of
  "no lab facts baked". The collision is solvable in a few lines either way.
  Recommend: keep the image stock and solve it in the playbook, since the
  whole point is that the image stops being special.
- **OD-3 — WAN model.** network-plan A2 declares the WAN leg as "NAT leg: no
  subnet, no gateway, no overlay", which settles the declaration but not the
  wiring — §5 records that `vm-wan0` currently has no uplink at all. Decide
  whether the WAN NATs to the home LAN (`wlan0`) or is left genuinely
  isolated, and make the live interface mode match the answer. This is a
  prerequisite for reaching the internet from inside the lab and for
  `docs/roadmap.md`'s Phase 3 WAN verification.
- **OD-4 — keep the current volume as a fallback base?** Yes as a file; no as
  the published base.
- **OD-5 — SSH to the firewall.** Default off. Turn on only if a playbook
  needs it, and then it is a baked fact with no API.

## 8. Rollback

Before the cutover, the existing `opnsense.qcow2` is the whole lab's DNS.
Keep a copy off the pool (`./backups/`, gitignored) and record its
`virsh vol-info` output. Rollback is: restore the file, revert the terraform
module to the hardcoded path, `terraform apply`, re-run `hosts.yml`.

After the cutover there is no in-place rollback of the image — the recovery
tool is the ZFS boot environment the base ships with, which is a reason to
keep ZFS in §2.2.

## 9. Sequencing summary

Stage numbers below match the §4 headings; §3 is the prerequisite.

| Stage | Change | Lab still works after? |
| --- | --- | --- |
| 0 | decisions OD-1…OD-5 | yes |
| §3 | build + publish the stock base image, keep the pre-publish copy | yes — nothing references it yet |
| 1 | module: overlay volume, `base_image_path`, MAC, `cpu.mode`, console, `module_os.gateway` | yes — verified on a scratch VM name |
| 2 | outputs + `lab.yaml` (`ip:` on the gateway) | yes |
| 3 | bootstrap phase in `opnsense.yml`, and the `192.168.1.1` route | yes, if guarded as in §3 |
| 4 | the `deploy.sh` step | yes |
| — | **cutover: `opnsense01` onto the overlay, undefine the builder, drop the old disk** | **the cutover — see §8** |
| 5 | validate playbook | yes |
| 6 | `docs/opnsense-image.md` | yes |
