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
terraform/                  # Terraform root module (libvirt provider, multi-host)
  locals.tf                 #   Loads lab.yaml via yamldecode()
  modules/alpine/           #   Alpine Linux VM module
  modules/opnsense/         #   OPNsense edge/gateway VM module
ansible/
  ansible.cfg               # Default inventory: lab_inventory.py
  inventory/lab_inventory.py     # Dynamic inventory from lab.yaml (hypervisors, edge)
  inventory/terraform_vms.py     # Dynamic inventory from terraform `vms` output
  playbooks/                # hosts.yml (hypervisors), edge.yml, start_vms.yml
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
  `terraform_vms`, `vms_gateway` (started first), and `vms_linux`. Gateway
  detection is by `role` in the output — not by VM name prefix.
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
VMs only. Don't assume components exist; check first.

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
