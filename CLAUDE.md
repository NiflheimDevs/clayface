# CLAUDE.md

Guidance for Claude Code (and any other coding agent) working in this repository.

## What this project is

Bachelor final project: a **full red team simulation lab** — an IaC-provisioned
virtualized corporate network with intentional, documented misconfigurations,
built for practicing and demonstrating the ethical hacking lifecycle (initial
access, privilege escalation, lateral movement, data exposure).

Project name: **Clayface** — this is now canonical. Older docs (the knowledge
base used to, vault notes sometimes) say "RedClay" / "red clay" / "red-clay";
same project.

Docs hierarchy:

- `docs/redteam_corporate.md` — the proposal: source of truth for scope/goals.
- `docs/redclay.yaml` — project knowledge base: goals, target topology,
  identity architecture, decisions, agent guidance. It describes **intent**
  and is not read by any tooling. Machine-consumed facts live in `lab.yaml`
  (see the data flow rule below); if the knowledge base contradicts
  `lab.yaml` or the code, `lab.yaml`/code wins.
- `red-clay/` — Obsidian vault: personal notes, learning courses, and a work
  journal. **They can be outdated or aspirational** — if a doc contradicts
  the actual code/config in `terraform/` or `ansible/`, trust the code, and
  if the discrepancy is significant, stop and ask the user before acting
  on it. (`red-clay/Playground/Knowledge Base.md` predates
  `docs/redclay.yaml` and largely overlaps it — treat the yaml as newer.)

## Repo layout

```
docs/
  redteam_corporate.md      # Project proposal — source of truth for scope/goals
  redclay.yaml              # Project knowledge base — goals, topology, decisions (intent only)
lab.yaml                    # SINGLE SOURCE OF TRUTH: hosts, network, VM placements
deploy.sh                   # End-to-end provisioner: terraform init/apply + ansible playbooks
base-image/                 # Windows base image build files (unattend-dc.xml + guide)
terraform/                  # Terraform root module (libvirt provider, multi-host)
  locals.tf                 #   Loads lab.yaml via yamldecode()
  modules/alpine/           #   Alpine Linux VM module
  modules/opnsense/         #   OPNsense edge/gateway VM module
  modules/domaincontroller/ #   Windows Server VM module (qcow2 backing on base_image_path)
ansible/
  ansible.cfg               # Default inventory: lab_inventory.py
  requirements.yml          # Collections: ansible.windows, microsoft.ad
  inventory/lab_inventory.py     # Dynamic inventory from lab.yaml (hypervisors, edge)
  inventory/terraform_vms.py     # Dynamic inventory from terraform `vms` output
  playbooks/                # hosts.yml (hypervisors), edge.yml, start_vms.yml,
                            # dc.yml (Windows DC: hostname, admin password, promote)
vm/images/                  # ISOs (tiny11, OPNsense) — gitignored, large binaries
red-clay/                   # Obsidian vault: notes, courses, journal, diagrams
```

**Data flow rule:** static facts (host IPs, users, network settings, VM
placement) live ONLY in `lab.yaml`. Terraform reads it via `locals.tf`,
Ansible via `inventory/lab_inventory.py`. Derived state (which VMs exist)
belongs to terraform and flows to Ansible through the `vms` output
(name → hypervisor + role) into `inventory/terraform_vms.py`. Never
duplicate a fact in two places — extend `lab.yaml` or a terraform output.

Exceptions to the rule — values owned elsewhere, mirrored here:
- **DNS domain** (`.clayface`): owned by the OPNsense image. Ansible
  assumes it via `LAB_DOMAIN` env (default `clayface`) in
  `terraform_vms.py`. Deliberately NOT in lab.yaml — lab.yaml can't change
  what the image baked in.
- **AD domain** (`clayface.local`, NetBIOS `CLAYFACE`): DERIVED from the
  DNS domain in `playbooks/dc.yml` (`LAB_DOMAIN` + `.local`, uppercased for
  NetBIOS), not declared anywhere. It is a child of the DNS domain, so a
  separate knob could only ever disagree with it. Setting `LAB_DOMAIN`
  moves the inventory addresses and the forest name together.
- **OPNsense LAN IP** (`10.0.0.1`): mirrored constant, baked into the
  OPNsense image. Appears as `opnsense_host` in `opnsense.yml`, `lab_dns_ip`
  in `start_vms.yml`, and `lab_dns_forwarder` in `dc.yml`.
- **Gateway VM host**: the top-level `edge.host` scalar in lab.yaml, not
  listed as a placement. Exactly one edge host holds BY CONSTRUCTION (single
  scalar) — no count check. A guard local in `terraform/locals.tf` fails the
  plan if `edge.host` doesn't name a host in `hosts`.

## Architecture (current state)

- **Two hypervisor hosts** (`host_a` = 192.168.1.161, `host_b` = 192.168.1.134),
  both reached over SSH (`qemu+ssh://...`) via the `dmacvicar/libvirt` provider
  (v0.9.8 — the provider is mid-rewrite and known-flaky; pinned deliberately).
- Hosts are joined by a **VXLAN overlay** (vxlan100, port 4789, over `wlan0`)
  and a bridge (`vm-br0`) — configured by the Ansible hypervisor playbook.
- **OPNsense** VM acts as the edge (DHCP/DNS) — DNS resolves VM names so
  Ansible reaches VMs as `<name>.clayface` (domain owned by the OPNsense
  image, mirrored in `terraform_vms.py`).
- **Terraform → Ansible link**: terraform's `vms` output (VM name → hypervisor
  + role) feeds `ansible/inventory/terraform_vms.py`, which groups VMs into
  `terraform_vms`, `vms_gateway` (started first), `vms_dc` (Windows DCs, WinRM
  connection vars; also members of `vms_linux` so start_vms.yml starts them),
  and `vms_linux`. Gateway detection is by `role` in the output — not by VM
  name prefix; roles are `gateway` (edge VM), `dc` (module `dc` in
  vm_placements), `linux` (everything else).
- **Gateway VM** (`opnsense01`): exactly one, always on the host named by
  `edge.host` in lab.yaml. Per-host module blocks instantiate it only on
  that host; because `edge.host` is a single scalar, zero-or-two-edge states
  are impossible by construction (a typo'd name fails the plan via the
  locals.tf guard).
- Adding a VM = one entry in `lab.yaml` under `vm_placements`. Adding a host
  = an entry under `hosts` + one provider block + per-module blocks in
  `terraform/main.tf` (providers can't be selected dynamically).

The end goal (per the Overview and proposal): Proxmox/KVM + Terraform + Packer
+ Ansible building an AD environment with planted weaknesses (Kerberoastable
accounts, bad ACLs, NTLM relay targets), plus detection via Wazuh/SIEM.
Much of this is **not built yet** — the lab currently has Alpine + OPNsense
VMs, plus a Windows `dc01` VM (from the Windows base image) that
`playbooks/dc.yml` promotes to the forest root. Don't assume components
exist; check first.

## Windows domain controller (dc01, role "dc")

- The base image (`ws-base.qcow2`, module variable `base_image_path`) is
  built per `base-image/README.md`: sysprep
  `/generalize /oobe /shutdown /unattend:` with
  `base-image/unattend-dc.xml` — so first boot answers OOBE itself,
  enables WinRM (all firewall profiles), and activates Administrator with
  the lab password. No console step needed; if a base was built WITHOUT
  the unattend, its first boot stops at the OOBE password screen and must
  be completed once on the SPICE console before dc.yml can connect.
- **The image carries NO server roles, and must not.** AD DS does not
  support sysprep `/generalize` (Microsoft's sysprep server-role support
  table lists it as unsupported), so `dc.yml` installs the role during
  promotion instead. Building the image with the role installed is what
  broke the first attempt — see the README's troubleshooting section for
  the AppX / `AppXSvc` / patch-level failure modes on build 26100.
- Connection is **WinRM** (HTTP, port 5985), not SSH. Control node needs
  `pywinrm` (`pip install pywinrm`) and the collections from
  `ansible/requirements.yml` (`ansible-galaxy collection install -r
  ansible/requirements.yml`).
- **A WinRM failure is almost always the control node, not the image.** The
  unattend's specialize pass already configures the guest correctly
  (`winrm quickconfig -q`, both `WINRM-HTTP-In-TCP*` firewall rules,
  `LocalAccountTokenFilterPolicy=1` — all three verified on a built VM).
  Before rebuilding an image over a WinRM error, check these three:
  1. **Transport must be NTLM** — set in `inventory/terraform_vms.py`.
     Ansible defaults to `plaintext` (HTTP Basic), which a stock guest
     rejects with HTTP 401 because it enables only `Negotiate`. NTLM also
     works before the domain exists, which Kerberos does not.
  2. **Bypass the ambient HTTP proxy.** pywinrm builds a plain
     `requests.Session()`, which honours `HTTP_PROXY`; a local proxy cannot
     resolve `*.clayface` and answers with HTTP **503**, which looks exactly
     like a broken guest. `deploy.sh` exports `no_proxy` covering
     `clayface`. Running ansible by hand needs the same variable (or
     `env -u HTTP_PROXY`).
  3. **The control node must resolve `<name>.clayface`.** That name is served
     by OPNsense on the `vm-br0` leg; the machine's other resolvers have
     never heard of the lab domain. `playbooks/hosts.yml` sets this up while
     it configures the bridge (`resolvectl dns vm-br0 10.0.0.1`). That is
     runtime-only and lost on reboot, so on a fresh boot run hosts.yml — or
     make it persistent with `sudo nmcli con mod vm-br0 ipv4.dns 10.0.0.1
     ipv4.ignore-auto-dns yes` followed by `sudo nmcli con up vm-br0`
     (bounces the lab interface).

  Symptom → cause: `Code 503` = proxy; `credentials were rejected` on
  `plaintext` = transport; `NameResolutionError` = control-node DNS.
- `playbooks/dc.yml` (idempotent): renames the host to its inventory name
  `dc01` from lab.yaml, enforces the Administrator password, points the
  NIC's DNS client at `127.0.0.1`, promotes (the `microsoft.ad.domain`
  module installs the AD DS role AND creates the forest), reboots, sets a
  DNS forwarder to OPNsense, then verifies the domain's own A record and
  the `_ldap._tcp.dc._msdcs` SRV record resolve locally with `-DnsOnly`.
  The DNS client must be set BEFORE promotion: netlogon registers the DC's
  records by dynamic update into the configured resolver, so leaving it on
  OPNsense means the AD zone comes up without its own records.
- **DHCP / name resolution (all VMs, not just the DC)**: VMs get DHCP from
  OPNsense (Dnsmasq backend); no IPs are hardcoded. The DC VMs have a
  **pinned MAC** (derived deterministically from the VM name in the dc
  module; check `terraform -chdir=terraform output -json vms`). The
  binding name → MAC → stable IP + DNS registration is automated by
  `playbooks/opnsense.yml` via the OPNsense REST API — one-time prep is
  creating an API key in the OPNsense UI (System → Access → Users), then
  run with `OPNSENSE_API_KEY=... OPNSENSE_API_SECRET=...` (also settable:
  `LAB_DNS_URL`, `LAB_DOMAIN`, `LAB_DC_IP_BASE`, `LAB_DHCP_DNS_SERVER`).
  It runs in deploy.sh before dc.yml; without it (or before the first
  run), bootstrap with the VM's current IP:
  `ansible-playbook ... dc.yml -e ansible_host=<ip>`.
- **DNS topology — dc01 is the lab resolver.** OPNsense (Unbound on
  `10.0.0.1`) keeps the `clayface` zone via its dnsmasq domain forward, so
  `<name>.clayface` resolves from boot #0 and the Ansible control node can
  always reach the DC by name. `clayface.local` is authoritative on `dc01`
  itself, which forwards everything else to OPNsense. Handing `dc01` out to
  DHCP clients is a **separate, opt-in step**: `opnsense.yml` sets dnsmasq
  option 6 only when `LAB_DHCP_DNS_SERVER` is set, and refuses to do so if
  that address does not answer DNS. Enable it only after `dc.yml` reports
  its DNS verification passing — pointing the lab at a resolver that is not
  yet a resolver removes name resolution lab-wide (IP connectivity survives,
  so recovery is `dc.yml -e ansible_host=<ip>` plus deleting the option row).
- Passwords: Administrator + DSRM default to `Admin@123`, overridable with
  `LAB_WIN_ADMIN_PASS` (read by the inventory script, dc.yml, and baked in
  the unattend file — change all three together). Lab-only and
  deliberately weak; DSRM sharing the same password is a known shortcut
  to revisit.

## How to operate

Full deploy:

```bash
./deploy.sh                # terraform init+apply, then ansible playbooks (asks for become pass)
./deploy.sh --plan         # terraform plan only
./deploy.sh --skip-ansible # terraform only
./deploy.sh --destroy      # teardown
```

Manual ansible runs follow the same pattern as deploy.sh: combine the
lab.yaml host inventory with the terraform VM one (`-i
inventory/lab_inventory.py -i inventory/terraform_vms.py`).

## Known pain points / open issues

- **Hypervisor addresses come from the home LAN's DHCP and can drift.**
  `lab.yaml` `hosts.*.address` is what terraform's `qemu+ssh://` URI and the
  `hypervisors` inventory target, and the NIC is `ipv4.method auto` — so a
  lease change silently breaks `terraform plan` and `playbooks/hosts.yml`
  with `No route to host` until the new address is written back. Pin the
  lease (DHCP reservation or a static address) before relying on it; the
  `vm-br0` leg (`10.0.0.2/24`, static in NetworkManager) is not a substitute,
  because hosts.yml is what creates that bridge.
- **Image management is manual** — ISOs/qcow2 base images are copied to hosts
  by hand into `vm/images/`; provisioning does not fetch them.
- **VXLAN is a single-peer static mesh** (`vxlan_remote_ip`) — fine for two
  hosts, breaks at three. Multi-host needs full-mesh FDB entries (see
  `vxlan_peer_ips` in lab_inventory.py) or a spine.
- libvirt provider 0.9.x has known bugs (no TTY/monitor on VMs was hit before).

## Conventions & cautions

- `terraform.tfstate*`, `.terraform/`, and `terraform.tfvars` are gitignored —
  **never commit state or tfvars** (they may contain lab credentials).
- Ansible playbooks run with `--ask-become-pass`; they modify host networking
  (bridges, VXLAN). Be careful with changes to the hypervisor/edge playbooks —
  a mistake can drop network access to the hosts.
- This is an isolated lab for authorized security education. Offensive
  tooling/techniques discussed in docs are for use **inside this lab only**.
- The user is a student learning IaC as a side objective — prefer clear,
  explainable changes over clever ones, and surface trade-offs.
- The Obsidian vault uses `[[wikilinks]]` and is user-authored prose. When
  editing vault notes, keep that style; don't reformat or restructure them
  unless asked.
