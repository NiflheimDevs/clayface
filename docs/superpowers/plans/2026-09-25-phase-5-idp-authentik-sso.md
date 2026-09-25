# Phase 5 — IDP01: Authentik as the lab's identity provider

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build IDP01 — an Authentik identity provider on the LAN that binds
LDAP to `dc01` as `svc-idp-ldap` — and move the portal's authentication onto it
over OIDC, so the modern identity leg in `red-clay/Architecture/Identity.md` is
built rather than planned.

**Architecture:** IDP01 is a fourth role, not a special case: a terraform module
under `terraform/modules/idp/`, a `vm_placements` entry in `lab.yaml`, a
`vms_idp` inventory group, and its own playbook. The portal becomes a
confidential OIDC relying party: it validates the ID token itself, and it maps
AD groups to portal roles **on its own side** from a `lab.yaml` map, so nothing
in Authentik holds a role that a lab.yaml change could contradict.

**Tech Stack:** Terraform (dmacvicar/libvirt v0.9.8) + Ansible; Authentik
`ghcr.io/goauthentik/server:2026.8.3` with PostgreSQL 16 and Redis, on the
Ubuntu base image the `app` role already uses; Go 1.27 with
`github.com/coreos/go-oidc/v3` and `golang.org/x/oauth2` in the portal.

**Spec:** `docs/roadmap.md` Phase 5, `docs/ad-identity-design.md` (the IDP bind
account, section 14's deviations register), `red-clay/Architecture/Identity.md`
(the figure this phase makes true), `docs/network-design.md` (the DMZ boundary),
and `ansible/playbooks/app.yml` + `terraform/modules/app/` (the pattern this
phase follows for a new service role).

## Global Constraints

- **`ad.idp.read_ous` limits what the IdP can see, and the search bases must
  respect it.** `svc-idp-ldap` holds an explicit read ACE on `OU=Users` and
  `OU=Groups` only, and `ad_validate.yml` asserts exactly that. The LDAP
  source's user base is therefore `OU=Users,OU=Clayface,DC=clayface,DC=local`
  and its group base is `OU=Groups,OU=Clayface,DC=clayface,DC=local` — never
  the domain root, and never the anchor OU `OU=Clayface`, which the bind
  account deliberately cannot read. A source pointed at a wider base does not
  fail loudly; it silently syncs nothing.
- **The DMZ boundary gains exactly one allowance, and it is a `lab.yaml` entry.**
  The portal (in the DMZ) reaches IDP01 (on the LAN) over tcp 9000, which the
  default-deny filter blocks today. `networks.dmz.allow` is where an allowance
  is expressed — it is data, and `opnsense_dmz.yml` builds the rules from it.
  Do not add a rule by hand in the UI, and do not widen the existing
  DMZ→`dc01` 389/636 entry, which is the Chapter 5 pivot.
- **IDP01 sits on the LAN** (`10.0.0.30`), matching
  `docs/ad-identity-design.md` and the roadmap. The IdP holds the directory
  bind credential; it does not belong in the zone the attacker lands in.
- **The bind credential is `svc-idp-ldap` and nothing else.** No Domain Admin
  account, no `svc-app-portal`, no domain-wide read. If the source needs a
  wider base to work, the answer is a wider ACE in `lab.yaml`, which Phase 4
  already asserts — not a different account.
- **The portal's OIDC client secret is generated, not typed.** It lives in the
  gitignored `deploy.env` (`LAB_OIDC_CLIENT_SECRET`) and reaches both the
  blueprint and the portal's environment file from there. `lab.yaml` names the
  client id; it never holds a secret.
- **`LAB_APP_SSH_PASS` is reused for IDP01.** Both Linux roles run the same
  Ubuntu base image, which is where the `clayface` account and its password are
  baked. The variable's name is app-specific and the credential is not; the
  rename is deliberately out of scope here because it touches `deploy.env` and
  `app.yml` for no functional gain. Noted so a reader does not think it is an
  oversight.
- **Authored files take the newest available versions.** `go.mod` gets
  `go-oidc/v3` and `golang.org/x/oauth2` at their current releases, and every
  dependency is vendored so the portal's image build does not need the network.
- **The lab's default posture stays vulnerable.** `app.weaknesses` remains all
  `true`; `app.oidc.enabled: true` adds a second way in and does not remove the
  first. The local password form keeps working — that contrast is the point of
  the phase, not a regression.

## Review Focus

The spec describes the integration. It does not describe the hostnames, the
certificate reality, or what happens when Authentik is mid-migration. These are
the five ways this work is most likely to bite, most likely first. Each one gets
a check in the task that owns it.

1. **The OIDC `iss` claim does not match the configured issuer.** Authentik
   derives the issuer from the request's `Host` header, so a portal configured
   with `http://idp01.clayface:9000/...` while discovery was answered over
   `http://10.0.0.30:9000/...` produces a token whose `iss` the portal rejects —
   and the failure reads as "invalid token", not as "you used the wrong
   hostname". Task 5 asserts discovery's `issuer` equals the configured string
   before the portal is pointed at it.
2. **Authentik's first boot is slow and its health endpoint answers late.**
   PostgreSQL initialises, migrations run, then the server listens. A
   `wait_for` with a two-minute budget fails on a healthy first boot. Task 4
   waits five minutes and dumps `docker compose logs` in the failure path
   rather than leaving a bare timeout.
3. **The LDAP source syncs nothing and reports success.** A base DN the bind
   account cannot read, or a bind DN in the wrong form, produces an enabled
   source with zero users and no error on the Authentik side. Task 5 asserts
   two specific AD users and one group with its members are present in
   Authentik, by name — the only assertion that distinguishes "synced" from
   "silent".
4. **The portal's OIDC config is partially set.** `OIDC_ENABLED=true` with an
   empty `OIDC_CLIENT_SECRET` currently has no way to fail: `Validate()` would
   not know to require it. Task 6 makes `Validate()` reject a enabled-but-
   incomplete block, naming the missing environment variable, and unit-tests
   it.
5. **The portal's redirect URI is not what Authentik was told to accept.**
   Authentik matches `redirect_uris` in strict mode by default, and the
   mismatch surfaces as an error page in the browser after a successful
   password entry. Task 5 writes the URI from `lab.yaml` and Task 7 asserts the
   portal's own `/portal/sso/login` returns a `redirect_uri` byte-identical to
   it.

---

## File Structure

| File | Responsibility in this phase |
| --- | --- |
| `terraform/modules/idp/{main,variables,outputs}.tf` | New: builds IDP01 as a libvirt VM |
| `terraform/locals.tf` | `module_os` gains `idp = "linux"` |
| `terraform/main.tf` | `module.idp_host_a` / `idp_host_b` blocks |
| `terraform/outputs.tf` | the `mac` merge gains the IdP modules |
| `lab.yaml` | the `idp01` placement, `app.oidc.*`, and the DMZ allowance |
| `ansible/inventory/terraform_vms.py` | the `vms_idp` group, its dispatch branch, and the SSH connection |
| `ansible/playbooks/idp.yml` | New: the Authentik stack and the blueprint |
| `ansible/playbooks/idp_validate.yml` | New: the assertions |
| `app/internal/config/config.go` | the `OIDC` config block and its validation |
| `app/internal/web/sso.go` | New: the authorization-code flow and the role mapping |
| `app/internal/web/sso_test.go` | New: unit tests over the seam |
| `app/internal/web/router.go` | the SSO routes; a nil-database guard in `audit` |
| `app/go.mod` | `go-oidc/v3`, `x/oauth2`, and a `vendor/` tree |
| `ansible/playbooks/app.yml` | renders the `OIDC_*` environment |
| `ansible/playbooks/app_validate.yml` | asserts the SSO route |
| `deploy.sh`, `deploy.env.example` | the new steps and the new variables |
| `docs/idp01-design.md` | New: what IDP01 is and why |
| `red-clay/Architecture/Identity.md` | the figure stops calling the modern leg planned |

Tasks 1–3 are the infrastructure, Tasks 4–5 the IdP itself, Tasks 6–7 the
relying party, Task 8 the wiring and documentation. Tasks 1–3 must land before
Task 4 (there is no VM to configure without them), and Task 6 before Task 7
(there is no route to assert), but Tasks 1–3 and 6 are independent of each
other and can proceed in parallel.

---

### Task 1: The IDP terraform module

**Files:**
- Create: `terraform/modules/idp/main.tf`
- Create: `terraform/modules/idp/variables.tf`
- Create: `terraform/modules/idp/outputs.tf`

**Interfaces:**
- Produces: a module with variables `name` (string), `bridge` (string, no
  default), `base_image_path` (string), `memory_mib` (number, default 6144),
  `vcpu` (number, default 3), `disk_gib` (number, default 60); outputs `name`
  and `mac`. Identical in shape to `terraform/modules/app`, which is what makes
  `terraform/main.tf` a copy of the app blocks.

- [ ] **Step 1: Read the module being copied**

Read `terraform/modules/app/main.tf`, `variables.tf` and `outputs.tf` in full.
The three files below are that module with the sizes changed and the comments
re-pointed; the MAC derivation, the volume naming and the domain shape are the
same on purpose, because a second libvirt VM shape in one repo is a thing to
maintain rather than a thing to have.

- [ ] **Step 2: Write `terraform/modules/idp/variables.tf`**

```hcl
variable "name" {
  description = "VM name, and the name the overlay volume is derived from."
  type        = string
}

# No default on purpose, exactly as the app module has none: a bridge that
# silently defaulted would attach the IdP to the wrong segment, and the only
# symptom is a VM nobody can reach. main.tf passes it from lab.yaml.
variable "bridge" {
  description = "Bridge device this VM's NIC is attached to, from lab.yaml `networks:`."
  type        = string
}

# The same Ubuntu base image the `app` role uses. There is no separate IdP base
# to build: the image is a stock Ubuntu with a `clayface` account, and the
# Authentik stack arrives as containers.
variable "base_image_path" {
  description = "Base qcow2 the overlay is stacked on."
  type        = string
  default     = "/var/lib/libvirt/images/ubuntu24.04-base"
}

# Larger than the app module's 4096: this VM runs four containers, two of them
# a JVM-heavy Django app and a PostgreSQL. Authentik's own published guidance is
# 2 vCPU / 4 GiB for server+worker alone, and the database and cache are inside
# this VM too.
variable "memory_mib" {
  description = "Guest memory in MiB."
  type        = number
  default     = 6144
}

variable "vcpu" {
  description = "Guest vCPU count."
  type        = number
  default     = 3
}

# The images are ~1.5 GiB before extraction and PostgreSQL's data directory
# grows with every sync; 60 GiB leaves room without another resize.
variable "disk_gib" {
  description = "Overlay volume capacity in GiB. Must be >= the base image's virtual size."
  type        = number
  default     = 60
}
```

- [ ] **Step 3: Write `terraform/modules/idp/main.tf`**

```hcl
# A single-NIC Linux VM for the identity provider. Deliberately the same shape
# as modules/app: the pinned MAC, the volume naming and the q35 domain are the
# lab's one Linux VM shape, and the two modules differ only in size.

locals {
  # The MAC is derived from the name so it is stable across `terraform apply`
  # runs and known before the VM exists - which is what lets
  # playbooks/opnsense.yml turn it into a DHCP reservation. Same derivation as
  # modules/app and modules/dc (see the CLAUDE.md note on sysprepped Windows
  # guests booting as WIN-XXXXXXXXXXX).
  mac_seed = sha256(var.name)
  vm_mac   = format(
    "52:54:00:%s:%s:%s",
    substr(local.mac_seed, 0, 2),
    substr(local.mac_seed, 2, 2),
    substr(local.mac_seed, 4, 2),
  )
}

resource "libvirt_volume" "idp" {
  name = "${upper(var.name)}.qcow2"
  pool = "default"

  # An overlay on the base image, never a copy: the base is a read-only backing
  # file and every VM built from it shares its blocks. Booting the base itself
  # corrupts every overlay stacked on it.
  capacity = var.disk_gib
  format   = "qcow2"

  backing_store = {
    path = var.base_image_path
    format = "qcow2"
  }
}

resource "libvirt_domain" "idp" {
  name     = var.name
  memory   = var.memory_mib
  vcpu     = var.vcpu
  type     = "kvm"
  firmware = "bios"

  # The lab's Linux VMs are q35 with host-model CPU passthrough; the Windows
  # client is the only UEFI guest.
  os {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
  }

  features {
    acpi = true
    apic = true
  }

  cpu {
    mode = "host-model"
  }

  disk {
    volume_id = libvirt_volume.idp.id
  }

  network_interface {
    bridge = var.bridge
    mac    = local.vm_mac
  }

  # No graphics device: the VM is installed by cloud-init-free first boot and
  # driven over SSH, and an unused SPICE/VNC listener is attack surface the lab
  # does not need. `virsh console` still works.
  console {
    type        = "pty"
    target_port = "0"
    target_type = "serial"
  }

  graphics {
    type = "none"
  }
}
```

- [ ] **Step 4: Write `terraform/modules/idp/outputs.tf`**

```hcl
output "name" {
  description = "VM name, so callers can iterate modules by their own keys."
  value       = libvirt_domain.idp.name
}

output "mac" {
  description = <<-EOT
    The pinned NIC MAC. terraform knows it before the guest exists and the
    guest cannot change it, which is what lets playbooks/opnsense.yml create a
    DHCP reservation + DNS record for a name the guest has not chosen yet.
  EOT
  value       = local.vm_mac
}
```

- [ ] **Step 5: Validate the module parses**

Run: `terraform -chdir=terraform validate`

Expected: `Success! The configuration is valid.` — the module is not yet
referenced, so this only proves it parses; the next task proves it builds.

- [ ] **Step 6: Commit**

```bash
git add terraform/modules/idp
git commit -m "feat(lab): add the IdP terraform module

Same single-NIC Linux VM shape as modules/app, larger: the Authentik
stack runs four containers including PostgreSQL in this guest. The MAC
derivation is unchanged, so the VM can be pinned by DHCP the same way."
```

---

### Task 2: Place IDP01 and reach it from the inventory

**Files:**
- Modify: `lab.yaml` (the `idp01` entry under `vm_placements`; an `ad.idp.spn`
  reference is Phase 4's, not touched here)
- Modify: `terraform/locals.tf:43-48` (`module_os`)
- Modify: `terraform/main.tf` (append the two IdP blocks)
- Modify: `terraform/outputs.tf:54-64` (the `mac` merge)
- Modify: `ansible/inventory/terraform_vms.py` (group, dispatch, connection)

**Interfaces:**
- Consumes: the module from Task 1.
- Produces: `idp01` in the `vms`, `vms_idp`, `vms_linux` and `vms_pinned`
  inventory groups, with `ansible_host: idp01.clayface`,
  `libvirt_ip: 10.0.0.30` and SSH connection variables.

- [ ] **Step 1: Add the placement to `lab.yaml`**

Under `vm_placements`, after the `app01` entry:

```yaml
  # The identity provider (roadmap Phase 5). On the LAN, not the DMZ: it holds
  # the directory bind credential, and the DMZ is where the attacker lands.
  #
  # `ip:` is required, as it is for every VM that must be reachable by name
  # before it can be configured - opnsense.yml turns this plus the module's
  # pinned MAC into a DHCP reservation and a DNS A record.
  #
  # 10.0.0.30 must be OUTSIDE lab.yaml's LAN DHCP pool; if it is not, pick
  # another address rather than shrinking the pool, so an existing lease is
  # never invalidated.
  idp01:
    module: idp
    host: host_a
    ip: 10.0.0.30
```

- [ ] **Step 2: Register the module OS**

In `terraform/locals.tf`, add to `module_os`:

```hcl
  module_os = {
    alpine = "linux"
    app    = "linux"
    dc     = "windows"
    client = "windows"
    # The IdP is a Linux guest. Indexing this map is the typo guard for
    # `module:` in lab.yaml, so a new module must be added here before any
    # placement can name it.
    idp    = "linux"
  }
```

- [ ] **Step 3: Add the module blocks to `terraform/main.tf`**

Append, matching the `app_host_a` / `app_host_b` blocks exactly:

```hcl
module "idp_host_a" {
  source = "./modules/idp"

  providers = {
    libvirt = libvirt.host_a
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_a" && p["module"] == "idp"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}

module "idp_host_b" {
  source = "./modules/idp"

  providers = {
    libvirt = libvirt.host_b
  }

  for_each = {
    for name, p in local.vm_placements :
    name => p if p["host"] == "host_b" && p["module"] == "idp"
  }

  name   = each.key
  bridge = local.network_bridges[local.placement_network[each.key]]
}
```

- [ ] **Step 4: Extend the MAC merge in `terraform/outputs.tf`**

In the `mac` expression (lines 54–64), add two lines:

```hcl
        mac = try(
          merge(
            { for n, m in module.dc_host_a : n => m.mac },
            { for n, m in module.dc_host_b : n => m.mac },
            { for n, m in module.client_host_a : n => m.mac },
            { for n, m in module.client_host_b : n => m.mac },
            { for n, m in module.app_host_a : n => m.mac },
            { for n, m in module.app_host_b : n => m.mac },
            { for n, m in module.idp_host_a : n => m.mac },
            { for n, m in module.idp_host_b : n => m.mac },
          )[name],
          null
        )
```

Also update the output's `description` to name `idp` among the roles.

- [ ] **Step 5: Check the plan before applying**

Run: `./deploy.sh --plan`

Expected: one new resource group — `module.idp_host_a["idp01"]` with a volume and
a domain — and no change to anything else. A plan that wants to recreate an
existing VM means a shared resource name collided; the volume is named
`${upper(var.name)}.qcow2`, so a collision would be with another `IDP01`.

- [ ] **Step 6: Teach the inventory about the new group**

In `ansible/inventory/terraform_vms.py`:

The docstring's group list gains a line after `vms_app`:

```
    vms_idp       : identity providers, role "idp"
```

The group-init dict (lines 115–129) gains:

```python
        "vms_idp": {"hosts": []},
```

The role dispatch (lines 156–161) gains a branch:

```python
        elif role == "idp":
            inventory["vms_idp"]["hosts"].append(name)
```

The connection block (lines 207–219) changes its condition from `role == "app"`
to cover both Linux roles:

```python
        # The two Linux roles that carry credentials share the branch. They run
        # the same base image, so the account and password are the same one;
        # os_family cannot choose this, because "linux" also describes the
        # alpine guests, which carry no credentials at all (see the module
        # docstring).
        elif role in ("app", "idp"):
            hostvars.update({
                "ansible_connection": "ssh",
                "ansible_user": "clayface",
                "ansible_password": LAB_APP_SSH_PASS,
                "ansible_ssh_common_args": "-o StrictHostKeyChecking=accept-new",
            })
```

The `LAB_APP_SSH_PASS` docstring line changes to say it covers both roles:

```
    LAB_APP_SSH_PASS     Password of the `clayface` local account baked into
                         the Ubuntu base image; used for both the SSH login
                         and sudo on role "app" and role "idp" VMs
```

- [ ] **Step 7: Apply and prove the inventory sees it**

```bash
./deploy.sh --skip-ansible      # terraform init + apply
terraform -chdir=terraform output -json vms | python3 -c '
import json,sys
v = json.load(sys.stdin)
print(json.dumps(v.get("idp01"), indent=2))'
ansible-inventory -i ansible/inventory/lab_inventory.py \
                  -i ansible/inventory/terraform_vms.py --graph vms_idp
ansible-inventory -i ansible/inventory/lab_inventory.py \
                  -i ansible/inventory/terraform_vms.py --host idp01
```

Expected: the `vms` output shows `role: idp`, `os_family: linux`,
`network: lan`, `bridge: vm-lan0`, a `mac`, and `ip: 10.0.0.30`; the graph
lists `idp01` under `vms_idp`; the host dump shows the SSH connection variables.

- [ ] **Step 8: Pin the address and prove the name resolves**

`opnsense.yml` creates the reservation from `vms_pinned`, which `idp01` now
belongs to. Run it, then confirm the bootstrap path:

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/opnsense.yml
# Start it and wait for the guest to take its lease
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/start_vms.yml
getent hosts idp01.clayface
```

Expected: `10.0.0.30  idp01.clayface`. If instead it resolves to a different
address, the guest booted before the reservation existed and holds a pool
lease — reboot it, or bootstrap with `-e ansible_host=<current ip>` and re-run
`opnsense.yml`.

- [ ] **Step 9: Commit**

```bash
git add lab.yaml terraform/locals.tf terraform/main.tf terraform/outputs.tf \
        ansible/inventory/terraform_vms.py
git commit -m "feat(lab): place IDP01 on the LAN and teach the inventory about it

idp01 at 10.0.0.30, module idp, role idp, in its own vms_idp group. It
shares the Linux credential branch with the app role because both run the
same base image; os_family cannot choose that branch, since the alpine
guests are also Linux and carry no credentials."
```

---

### Task 3: The one firewall allowance the portal needs

The portal is in the DMZ; IDP01 is on the LAN. OPNsense's default-deny filter
blocks that path today, and the OIDC token exchange is the portal's own
server-side request — it cannot be avoided by a browser redirect.

**Files:**
- Modify: `lab.yaml` (`networks.dmz.allow`)

**Interfaces:**
- Consumes: `networks.dmz.allow`, already read by `opnsense_dmz.yml`.
- Produces: a rule permitting `app01` → `10.0.0.30` tcp 9000.

- [ ] **Step 1: Read the existing entry's shape**

```bash
cd /home/kiasoh/crapijat/College/project
python3 -c "
import yaml, json
lab = yaml.safe_load(open('lab.yaml'))
print(json.dumps(lab['networks']['dmz'].get('allow'), indent=2))"
```

The entry that permits the Chapter 5 LDAP pivot is the template. **Mirror its
key set exactly** — do not invent keys, and do not assume the schema from this
plan. The steps below give the values; the shape comes from what this command
prints.

- [ ] **Step 2: Write the failing check**

Before adding the allowance, prove the path is closed. From `app01`:

```bash
ansible -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
        vms_app -b -m ansible.builtin.shell \
        -a 'timeout 5 bash -c "cat < /dev/null > /dev/tcp/10.0.0.30/9000" && echo OPEN || echo CLOSED'
```

Expected: `CLOSED`. (Port 9000 is not listening yet either — Task 4 starts the
stack — so this step establishes the baseline and is re-run in Step 5 for the
firewall's own answer. Note which of the two reasons applies when you read it.)

- [ ] **Step 3: Add the allowance to `lab.yaml`**

Add to `networks.dmz.allow`, alongside the existing entry:

```yaml
      # The portal is a confidential OIDC client: it exchanges the
      # authorization code for tokens in a server-side request from APP01 to
      # the IdP, and a browser redirect cannot carry that. Port 9000 is
      # plain HTTP by design (see docs/idp01-design.md on why the lab runs
      # OIDC in the clear), so this is one port, one destination, and no
      # wider than the exchange needs.
      #
      # IDP01 -> dc01 does NOT need an entry here: both are on the LAN, which
      # is a single L2 domain, so the bind never reaches the firewall.
      - description: OIDC token exchange, portal to the identity provider
        destination: 10.0.0.30
        protocol: tcp
        ports: [9000]
```

- [ ] **Step 4: Apply the filter rules**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/opnsense_dmz.yml
```

Expected: the play read-backs confirm the DMZ interface address (its one
deliberate halt on a fresh lab) and report the new rule among the applied set.
If the play's rule count assertion is a fixed number, it must be updated in the
same change — a count is a fact about `lab.yaml`, and `lab.yaml` just changed.

- [ ] **Step 5: Prove the ruleset, not the service, is what answers**

Run the Step 2 command again.

Expected: still `CLOSED` for a *different* reason — nothing listens on 9000 yet.
The definitive check is Step 6 of Task 4, which runs the same probe against a
listening socket. Record both outputs: the pair is what distinguishes a firewall
denial from an absent service, and that distinction is exactly the kind of thing
the thesis's evidence chapter should not have to guess at.

- [ ] **Step 6: Commit**

```bash
git add lab.yaml
git commit -m "feat(lab): allow the portal to reach the identity provider

One DMZ allowance for the OIDC token exchange, expressed as a
networks.dmz.allow entry so opnsense_dmz.yml builds the rule. The LDAP
bind from IDP01 to dc01 needs no rule: both are on the LAN, which is one
L2 domain."
```

---

### Task 4: Build IDP01 — the Authentik stack

Follows `ansible/playbooks/app.yml` exactly: build or pull on the control node,
`docker save` to a tarball, copy it to the guest, `docker load`, `compose up`,
wait for the health endpoint. The guest has no route to an image registry worth
relying on, and a lab that cannot be rebuilt offline is a lab with one working
day.

**Files:**
- Create: `ansible/idp/docker-compose.yml`
- Create: `ansible/idp/.env.j2`
- Create: `ansible/playbooks/idp.yml`

**Interfaces:**
- Consumes: the `vms_idp` group from Task 2; `LAB_IDP_DB_PASS`,
  `LAB_IDP_SECRET_KEY`, `LAB_IDP_ADMIN_PASS` from `deploy.env`.
- Produces: `http://idp01.clayface:9000/` serving Authentik, and the vars
  `idp_issuer`, `idp_client_id`, `idp_client_secret` on the control node, which
  Tasks 5 and 7 consume.

- [ ] **Step 1: Write the compose file**

`ansible/idp/docker-compose.yml`:

```yaml
# The Authentik stack for the lab's identity provider.
#
# Sized for a lab, not for production: one worker, default concurrency, and no
# external object storage. The version is pinned rather than tracked, for the
# same reason the portal's base images are pinned in its compose file - a lab
# whose rebuild silently changes product version is not reproducible.
#
# Ports: 9000 (HTTP) is the one the lab uses. 9443 (HTTPS) is published for
# completeness but nothing points at it - everything in this lab speaks plain
# HTTP over the segment, which docs/idp01-design.md records as a deviation.
services:
  postgresql:
    image: docker.io/library/postgres:16-alpine
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -d $${POSTGRES_DB} -U $${POSTGRES_USER}"]
      start_period: 20s
      interval: 30s
      retries: 5
      timeout: 5s
    volumes:
      - database:/var/lib/postgresql/data
    environment:
      POSTGRES_PASSWORD: ${PG_PASS:?database password required}
      POSTGRES_USER: ${PG_USER:-authentik}
      POSTGRES_DB: ${PG_DB:-authentik}

  redis:
    image: docker.io/library/redis:7-alpine
    restart: unless-stopped
    command: --save 60 1 --loglevel warning
    healthcheck:
      test: ["CMD-SHELL", "redis-cli ping | grep PONG"]
      start_period: 20s
      interval: 30s
      retries: 5
      timeout: 3s
    volumes:
      - redis:/data

  server:
    image: ghcr.io/goauthentik/server:2026.8.3
    restart: unless-stopped
    command: server
    environment:
      AUTHENTIK_SECRET_KEY: ${AUTHENTIK_SECRET_KEY:?secret key required}
      AUTHENTIK_REDIS__HOST: redis
      AUTHENTIK_POSTGRESQL__HOST: postgresql
      AUTHENTIK_POSTGRESQL__USER: ${PG_USER:-authentik}
      AUTHENTIK_POSTGRESQL__NAME: ${PG_DB:-authentik}
      AUTHENTIK_POSTGRESQL__PASSWORD: ${PG_PASS}
      AUTHENTIK_BOOTSTRAP_PASSWORD: ${AUTHENTIK_BOOTSTRAP_PASSWORD:?bootstrap password required}
      AUTHENTIK_BOOTSTRAP_TOKEN: ${AUTHENTIK_BOOTSTRAP_TOKEN:?bootstrap token required}
      AUTHENTIK_ERROR_REPORTING__ENABLED: "false"
    volumes:
      - ./media:/media
      - ./custom-templates:/templates
      # The blueprints are the write path: Ansible ships YAML here and
      # Authentik instantiates whatever it finds on start. Read-only, because
      # the running instance must never rewrite the artifact Ansible owns.
      - ./blueprints:/blueprints/custom:ro
    ports:
      - "0.0.0.0:9000:9000"
      - "0.0.0.0:9443:9443"
    depends_on:
      postgresql:
        condition: service_healthy
      redis:
        condition: service_healthy

  worker:
    image: ghcr.io/goauthentik/server:2026.8.3
    restart: unless-stopped
    command: worker
    environment:
      AUTHENTIK_SECRET_KEY: ${AUTHENTIK_SECRET_KEY:?secret key required}
      AUTHENTIK_REDIS__HOST: redis
      AUTHENTIK_POSTGRESQL__HOST: postgresql
      AUTHENTIK_POSTGRESQL__USER: ${PG_USER:-authentik}
      AUTHENTIK_POSTGRESQL__NAME: ${PG_DB:-authentik}
      AUTHENTIK_POSTGRESQL__PASSWORD: ${PG_PASS}
      AUTHENTIK_ERROR_REPORTING__ENABLED: "false"
    user: root
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./media:/media
      - ./certs:/certs
      - ./custom-templates:/templates
      - ./blueprints:/blueprints/custom:ro
    depends_on:
      postgresql:
        condition: service_healthy
      redis:
        condition: service_healthy

volumes:
  database:
    driver: local
  redis:
    driver: local
```

- [ ] **Step 2: Write the environment template**

`ansible/idp/.env.j2`:

```jinja
# Rendered by ansible/playbooks/idp.yml from deploy.env. Every secret here is
# generated once and lives in the gitignored deploy.env, never in lab.yaml and
# never in this template.
PG_PASS={{ idp_db_pass }}
PG_USER=authentik
PG_DB=authentik

AUTHENTIK_SECRET_KEY={{ idp_secret_key }}
AUTHENTIK_BOOTSTRAP_PASSWORD={{ idp_admin_pass }}
AUTHENTIK_BOOTSTRAP_TOKEN={{ idp_bootstrap_token }}

# Authentik derives the OIDC issuer from the request's Host header, so the
# portal's configured issuer and this VM's name must agree. Binding the lab's
# domain here is what keeps discovery's `issuer` stable whichever address the
# caller used.
AUTHENTIK_WEB__PATH=/
```

- [ ] **Step 3: Write the playbook**

`ansible/playbooks/idp.yml`, following `app.yml`'s two-play shape:

```yaml
---
# IDP01: the lab's identity provider.
#
# Two plays, the same split app.yml uses. Play 1 runs on the control node
# because that is the only machine with a working image registry and the only
# one whose docker can build; play 2 runs on the guest, which has neither and
# gets everything through a saved tarball.
#
# Run:
#   ansible-playbook -i inventory/lab_inventory.py \
#                    -i inventory/terraform_vms.py playbooks/idp.yml
#
# What it does (idempotent):
#   1. pulls the three images on the control node and saves them to a tarball
#   2. creates the guest's directory tree
#   3. copies the tarball, loads it on the guest (guarded by image presence)
#   4. ships the compose file and the rendered .env verbatim
#   5. ships the blueprint directory
#   6. docker compose up --detach, then waits for the readiness endpoint
#
# The images are PULLED, not built: there is nothing of ours in them. The
# tarball exists so the guest never needs a registry, which is the same reason
# app.yml saves the portal image.

- name: Prepare the identity provider images
  hosts: localhost
  connection: local
  gather_facts: false

  vars:
    idp_dir: "{{ playbook_dir }}/../idp"
    idp_control_tarball: "/tmp/clayface-idp-images.tar"
    idp_images:
      - "ghcr.io/goauthentik/server:2026.8.3"
      - "docker.io/library/postgres:16-alpine"
      - "docker.io/library/redis:7-alpine"

  tasks:
    - name: Assert the docker CLI is available on the control node
      ansible.builtin.command: docker version --format '{{ "{{" }}.Server.Version{{ "}}" }}'
      changed_when: false
      register: idp_docker_version

    - name: Report the control node's docker version
      ansible.builtin.debug:
        msg: "control node docker server {{ idp_docker_version.stdout }}"

    - name: Pull the images the identity provider needs
      ansible.builtin.command: "docker pull {{ item }}"
      loop: "{{ idp_images }}"
      register: idp_pull
      changed_when: "'Downloaded newer image' in idp_pull.stdout or 'Pull complete' in idp_pull.stdout"

    - name: Save the images to a tarball for the guest
      ansible.builtin.command: >-
        docker save --output {{ idp_control_tarball }} {{ idp_images | join(' ') }}
      changed_when: true

    - name: Report the tarball size
      ansible.builtin.stat:
        path: "{{ idp_control_tarball }}"
      register: idp_tarball_stat

    - name: Report it
      ansible.builtin.debug:
        msg: "image tarball is {{ (idp_tarball_stat.stat.size / 1048576) | round(1) }} MiB"

- name: Build the identity provider
  hosts: vms_idp
  become: false
  gather_facts: true

  vars:
    idp_dir: "{{ playbook_dir }}/../idp"
    idp_app_dir: /opt/clayface-idp
    idp_control_tarball: "/tmp/clayface-idp-images.tar"
    idp_images:
      - "ghcr.io/goauthentik/server:2026.8.3"
      - "docker.io/library/postgres:16-alpine"
      - "docker.io/library/redis:7-alpine"

  tasks:
    - name: Assert docker and compose are present on the guest
      ansible.builtin.command: "{{ item }}"
      loop:
        - docker --version
        - docker compose version
      changed_when: false

    - name: Create the identity provider's directory tree
      ansible.builtin.file:
        path: "{{ item }}"
        state: directory
        owner: clayface
        group: clayface
        mode: "0755"
      loop:
        - "{{ idp_app_dir }}"
        - "{{ idp_app_dir }}/blueprints"
        - "{{ idp_app_dir }}/media"
        - "{{ idp_app_dir }}/certs"
        - "{{ idp_app_dir }}/custom-templates"
      become: true

    - name: Copy the image tarball to the guest
      ansible.builtin.copy:
        src: "{{ idp_control_tarball }}"
        dest: "{{ idp_app_dir }}/images.tar"
        owner: clayface
        group: clayface
        mode: "0644"

    - name: Check which images the guest already has
      ansible.builtin.command: "docker image inspect {{ item }}"
      loop: "{{ idp_images }}"
      register: idp_present
      changed_when: false
      failed_when: false

    - name: Load the images the guest is missing
      ansible.builtin.command: "docker load --input {{ idp_app_dir }}/images.tar"
      when: idp_present.results | selectattr('rc', 'ne', 0) | list | length > 0
      changed_when: true

    - name: Ship the compose file
      ansible.builtin.copy:
        src: "{{ idp_dir }}/docker-compose.yml"
        dest: "{{ idp_app_dir }}/docker-compose.yml"
        owner: clayface
        group: clayface
        mode: "0644"

    - name: Ship the blueprints
      ansible.builtin.copy:
        src: "{{ idp_dir }}/blueprints/"
        dest: "{{ idp_app_dir }}/blueprints/"
        owner: clayface
        group: clayface
        mode: "0644"

    - name: Render the environment file
      ansible.builtin.copy:
        content: "{{ lookup('template', idp_dir + '/.env.j2') }}"
        dest: "{{ idp_app_dir }}/.env"
        owner: clayface
        group: clayface
        # The file holds the database password, the secret key and the
        # bootstrap token. 0600 on a single-tenant lab VM is the cheap half of
        # a control that production would express as a secret store.
        mode: "0600"
      no_log: true

    - name: Start the identity provider
      ansible.builtin.command:
        cmd: docker compose up --detach
        chdir: "{{ idp_app_dir }}"
      register: idp_compose_up
      changed_when: "'Started' in idp_compose_up.stderr or 'Created' in idp_compose_up.stderr or 'Recreated' in idp_compose_up.stderr"

    - name: Wait for the readiness endpoint
      ansible.builtin.uri:
        url: "http://{{ inventory_hostname }}.{{ lookup('env', 'LAB_DOMAIN') | default('clayface', true) }}:9000/-/health/ready/"
        status_code: 200
        use_proxy: false
      delegate_to: localhost
      register: idp_ready
      retries: 60
      delay: 5
      until: idp_ready.status == 200
      # A first boot initialises PostgreSQL, runs the migrations and only then
      # binds the port. Five minutes is generous on purpose: a two-minute
      # budget fails on a healthy lab VM, and a timeout here is the least
      # informative failure this playbook can produce.
      rescue:
        - name: Show the container logs after a failed readiness check
          ansible.builtin.command:
            cmd: docker compose logs --tail=80
            chdir: "{{ idp_app_dir }}"
          changed_when: false
          register: idp_logs

        - name: Print them
          ansible.builtin.debug:
            msg: "{{ idp_logs.stdout_lines }}"

        - name: Fail with the reason named
          ansible.builtin.fail:
            msg: >-
              Authentik did not answer /-/health/ready/ within five minutes.
              The container logs are above. Check the postgresql healthcheck
              first: a ${PG_PASS} mismatch and a slow first boot look the same
              from here.

    - name: Report the containers
      ansible.builtin.command:
        cmd: docker compose ps --format '{{ "{{" }}.Service{{ "}}" }} {{ "{{" }}.State{{ "}}" }}'
        chdir: "{{ idp_app_dir }}"
      changed_when: false
      register: idp_ps

    - name: Report
      ansible.builtin.debug:
        msg: "{{ idp_ps.stdout_lines }}"
```

- [ ] **Step 4: Add the credentials to `deploy.env.example`**

```bash
# --- Identity provider (IDP01, roadmap Phase 5) ---
#
# Generate each of these once and keep them in the gitignored deploy.env:
#   LAB_IDP_DB_PASS:        openssl rand -base64 24
#   LAB_IDP_SECRET_KEY:     openssl rand -base64 48
#   LAB_IDP_BOOTSTRAP_TOKEN: openssl rand -hex 32
#   LAB_IDP_ADMIN_PASS:     the akadmin console password (lab-only)
#   LAB_OIDC_CLIENT_SECRET: openssl rand -base64 32
LAB_IDP_DB_PASS=
LAB_IDP_SECRET_KEY=
LAB_IDP_BOOTSTRAP_TOKEN=
LAB_IDP_ADMIN_PASS=
LAB_OIDC_CLIENT_ID=portal
LAB_OIDC_CLIENT_SECRET=
```

Then add the same names with generated values to the real `deploy.env`, and make
`deploy.sh` load them (Task 8 wires the exports; for now, export them in the
shell that runs this playbook).

- [ ] **Step 5: Run it**

```bash
set -a; . ./deploy.env; set +a
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/idp.yml
```

Expected: the tarball is written, the guest loads three images, four containers
report `running`, and the readiness check passes. The first run takes several
minutes; that is the migrations, not a hang.

- [ ] **Step 6: Prove the firewall allowance now carries traffic**

```bash
ansible -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
        vms_app -b -m ansible.builtin.shell \
        -a 'timeout 5 bash -c "cat < /dev/null > /dev/tcp/10.0.0.30/9000" && echo OPEN || echo CLOSED'
curl -sS -o /dev/null -w '%{http_code}\n' http://idp01.clayface:9000/-/health/live/
```

Expected: `OPEN` from APP01 (Task 3's allowance working) and `204` or `200` from
the control node. This is the pair Task 3 step 5 deferred.

- [ ] **Step 7: Prove idempotence**

Run the playbook again.

Expected: the pull, load and copy tasks report `ok`; `docker compose up
--detach` reports no `Recreated`; the readiness check passes on the first
attempt.

- [ ] **Step 8: Commit**

```bash
git add ansible/idp ansible/playbooks/idp.yml deploy.env.example
git commit -m "feat(lab): build IDP01 running Authentik

The stack arrives the way the portal's does - pulled and saved on the
control node, loaded on the guest, so the guest never needs a registry.
Blueprint directory is bind-mounted read-only: Ansible owns that YAML."
```

---

### Task 5: The blueprint — bind the directory and publish the application

Everything in Authentik is created through a blueprint: a YAML file in the
bind-mounted directory with the instantiate label, which Authentik applies on
start. That keeps the IdP's configuration in the repository, reviewable, and
rebuildable — the same reason the firewall rules are built from `lab.yaml`
rather than clicked.

**Files:**
- Create: `ansible/idp/blueprints/portal.yaml`
- Create: `ansible/playbooks/idp_validate.yml`
- Modify: `lab.yaml` (`app.oidc.*`)

**Interfaces:**
- Consumes: `idp_bootstrap_token` (Task 4), `lab_ad.idp.bind_account`,
  `ad.idp.read_ous`, `lab_ad.groups`.
- Produces: `lab_app.oidc` — `{enabled, issuer, client_id, redirect_url,
  role_claim, role_map, default_role}` — which Task 6 reads as the portal's
  configuration and Task 7 asserts.

- [ ] **Step 1: Declare the relying party's configuration in `lab.yaml`**

Under `app:`, alongside `service_account`:

```yaml
  # How the portal authenticates users (roadmap Phase 5). The directory-bind
  # half lives under `ad.idp` - that is the identity provider reading AD, and
  # this is the portal reading the identity provider. Two directions, two
  # owners.
  oidc:
    # A toggle, like every other posture in this lab: the local-password path
    # stays available so Chapter 3's baseline and Chapter 5's modern leg can
    # both be demonstrated without editing code.
    enabled: true

    # The issuer, in the form BOTH callers use. The portal fetches discovery
    # from this URL server-side, and the authorization endpoint the browser is
    # sent to comes out of that same document - so the browser never needs a
    # name it cannot resolve.
    #
    # Authentik derives its own issuer from the request's Host header. This
    # string and the name the portal uses to reach the IdP must therefore be
    # the same name, or the token's `iss` is rejected. idp_validate.yml
    # asserts the two agree before anything depends on it.
    issuer: http://idp01.clayface:9000/application/o/portal/

    # Public identifier. The secret is NOT here: it is LAB_OIDC_CLIENT_SECRET
    # in the gitignored deploy.env, because lab.yaml is committed.
    client_id: portal

    # Must match the provider's registered redirect URI byte for byte;
    # Authentik matches in strict mode. This is the portal's own address as
    # the control node and a LAN browser reach it.
    redirect_url: http://app01.clayface:8080/portal/sso/callback

    # Which ID token claim carries the groups, and how an AD group becomes a
    # portal role. The mapping deliberately lives HERE, in the relying party,
    # rather than as Authentik group attributes: the portal is what enforces
    # the role, so the portal is what should hold the rule, and a lab.yaml
    # change then cannot leave the IdP asserting a role nothing honours.
    role_claim: groups
    role_map:
      GG-IT-Admins: admin
      GG-Employees: user
    default_role: user
```

- [ ] **Step 2: Write the blueprint**

`ansible/idp/blueprints/portal.yaml`:

```yaml
# The lab's identity provider configuration, as one blueprint.
#
# Authentik instantiates any blueprint in the bind-mounted directory whose
# metadata carries the instantiate label, on every start. Instantiation is
# idempotent by the entries' own `id` fields, so a re-run updates rather than
# duplicates.
#
# !Env reads a container environment variable. The bind password and the OIDC
# client secret come from the rendered .env, which comes from the gitignored
# deploy.env - so this file, which IS committed, holds no credential.
version: 1
metadata:
  name: Clayface lab - portal SSO
  labels:
    blueprints.goauthentik.io/instantiate: "true"

entries:
  # --- The directory bind -------------------------------------------------
  #
  # Base DNs are OU=Users and OU=Groups UNDER the anchor OU, because
  # `ad.idp.read_ous` gives this account an explicit read ACE on those two OUs
  # and nothing else. Pointing a search base at the domain root would not
  # error: the source would report enabled, sync nothing, and look healthy.
  #
  # A useful consequence of that scoping, worth stating in the thesis: the
  # provider cannot see OU=ServiceAccounts, so the service accounts - including
  # the Kerberoastable one W1 plants - are not exposed through SSO at all. The
  # weak account is reachable by an attacker through Kerberos, not through the
  # product that legitimately holds a directory credential.
  - id: clayface-ldap-source
    model: authentik.sources.ldap.ldapsource
    identifiers:
      slug: clayface-ad
    attrs:
      name: Clayface Active Directory
      enabled: true
      server_uri: ldap://10.0.0.10:389
      bind_cn: >-
        CN={{ lab_ad.idp.bind_account }},OU=ServiceAccounts,OU={{ lab_ad.ou }},{{ ad_base_dn }}
      bind_password: !Env [LAB_USER_PASS]
      base_dn: "{{ ad_base_dn }}"
      additional_user_dn: "OU=Users,OU={{ lab_ad.ou }}"
      additional_group_dn: "OU=Groups,OU={{ lab_ad.ou }}"
      # objectSid is the only attribute that is stable across a rename, and a
      # lab rebuilds often enough that name-based matching produces duplicates.
      object_uniqueness_field: objectSid
      group_membership_field: member
      sync_users: true
      sync_groups: true
      sync_parent_group: false
      # Service accounts are unreachable anyway (see above); this filter keeps
      # the intent explicit for anyone who later widens the read ACE.
      user_object_filter: "(objectClass=user)"
      group_object_filter: "(objectClass=group)"

  # --- The signing keypair ------------------------------------------------
  #
  # Referenced by name from the default install rather than generated here: a
  # generated keypair means embedding private key material in a committed file.
  # The token signature is validated against the JWKS Authentik publishes at
  # the issuer, so a self-signed certificate is sufficient - nothing validates
  # it against a CA.
  - id: clayface-signing-key
    model: authentik.crypto.certificatekeypair
    state: present
    identifiers:
      name: authentik Self-signed Certificate
    attrs:
      name: authentik Self-signed Certificate

  # --- The groups claim ---------------------------------------------------
  #
  # Authentik's default `profile` scope carries no groups, so the portal would
  # receive a token with no way to decide a role. This mapping publishes the
  # user's group names under the claim `lab.yaml` names in `role_claim`.
  #
  # `ak_groups` are Authentik's own groups, which the LDAP source populates
  # from AD; the names are the AD names, which is what makes the relying
  # party's role_map readable against lab.yaml's `ad.groups`.
  - id: clayface-groups-scope
    model: authentik.providers.oauth2.scopemapping
    identifiers:
      name: Clayface groups
    attrs:
      name: Clayface groups
      scope_name: groups
      description: The user's directory group names, so the relying party can map them to roles.
      expression: |
        return {"groups": [g.name for g in request.user.ak_groups.all()]}

  # --- The OIDC provider --------------------------------------------------
  #
  # Confidential client: the portal exchanges the code server-side, so the
  # secret never reaches a browser.
  #
  # The flows are the default install's; `!Find` locates them by slug rather
  # than by primary key, which is not knowable in a committed file. The
  # explicit-consent authorization flow is used deliberately: consent is part
  # of what the thesis's identity chapter describes, and it is one more
  # observable step in the browser.
  - id: clayface-portal-provider
    model: authentik.providers.oauth2.oauth2provider
    identifiers:
      name: Clayface portal
    attrs:
      name: Clayface portal
      authorization_flow: !Find [authentik.flows.flow, [slug, default-provider-authorization-explicit-consent]]
      invalidation_flow: !Find [authentik.flows.flow, [slug, default-invalidation-flow]]
      client_type: confidential
      client_id: !Env [LAB_OIDC_CLIENT_ID]
      client_secret: !Env [LAB_OIDC_CLIENT_SECRET]
      signing_key: !KeyOf clayface-signing-key
      redirect_uris:
        - matching_mode: strict
          url: "{{ lab_app.oidc.redirect_url }}"
      property_mappings:
        - !Find [authentik.providers.oauth2.scopemapping, [scope_name, openid]]
        - !Find [authentik.providers.oauth2.scopemapping, [scope_name, email]]
        - !Find [authentik.providers.oauth2.scopemapping, [scope_name, profile]]
        - !KeyOf clayface-groups-scope
      # The subject identifier is the AD account name, which makes a token's
      # `sub` readable in the portal's audit trail and makes an SSO login
      # traceable back to a directory object.
      sub_mode: user_username
      include_claims_in_id_token: true
      access_token_validity: hours=1
      refresh_token_validity: days=1

  # --- The application ----------------------------------------------------
  #
  # What a user sees on the Authentik dashboard, and what the portal's
  # redirect lands on. The slug fixes the issuer path, so changing it changes
  # lab.yaml's `app.oidc.issuer` in the same commit.
  - id: clayface-portal-application
    model: authentik.core.application
    identifiers:
      slug: portal
    attrs:
      name: Clayface Portal
      slug: portal
      provider: !KeyOf clayface-portal-provider
      meta_description: The customer portal, authenticated through this provider.
```

- [ ] **Step 3: Add the OIDC client to `deploy.env` and re-run the playbook**

```bash
# in deploy.env, if not already set
grep -q '^LAB_OIDC_CLIENT_SECRET=' deploy.env || \
  printf 'LAB_OIDC_CLIENT_SECRET=%s\n' "$(openssl rand -base64 32)" >> deploy.env
set -a; . ./deploy.env; set +a
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/idp.yml
```

Expected: the blueprint is copied and the containers restart onto it. A
blueprint syntax error is reported in the `worker` container's log, not at
`docker compose up` — read them if nothing appears:

```bash
ansible -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
        vms_idp -b -m ansible.builtin.shell \
        -a 'cd /opt/clayface-idp && docker compose logs --tail=40 worker | grep -i blueprint'
```

- [ ] **Step 4: Write the failing assertion — the LDAP sync**

`ansible/playbooks/idp_validate.yml`:

```yaml
---
# IDP01's assertions: the blueprint was applied, the directory bind works, and
# the issuer the portal will be configured with is the issuer Authentik emits.
#
# Read-only. The write path is the blueprint, and a validator that configured
# anything would make a passing run the only evidence that the artifact is
# right.
#
# Run:
#   ansible-playbook -i inventory/lab_inventory.py \
#                    -i inventory/terraform_vms.py playbooks/idp_validate.yml

- name: Validate the identity provider
  hosts: localhost
  connection: local
  gather_facts: false

  vars:
    idp_base: "{{ lab_app.oidc.issuer | regex_replace('/application/o/.*$', '') }}"
    idp_api: "{{ idp_base }}/api/v3"
    idp_token: "{{ lookup('env', 'LAB_IDP_BOOTSTRAP_TOKEN') }}"
    # The users this asserts were synced. Named rather than counted: a count
    # passes for a source that synced the wrong OU.
    idp_expect_users:
      - bob.sing
      - abed.nad
    idp_expect_group: GG-Developers
    idp_expect_group_members:
      - bob.sing

  tasks:
    - name: Fail if the bootstrap token is not in the environment
      ansible.builtin.assert:
        that: idp_token | length > 0
        fail_msg: >-
          LAB_IDP_BOOTSTRAP_TOKEN is not set. It is generated once into
          deploy.env (see deploy.env.example) and read by both idp.yml and
          this playbook; `set -a; . ./deploy.env; set +a` first.

    # --- The issuer contract, BEFORE anything depends on it ---
    #
    # Authentik derives the issuer from the request's Host header. If the
    # portal is configured with a name that answers with a different issuer,
    # every token is rejected with a message about the token rather than about
    # the hostname. This is the assertion that turns that into a sentence.
    - name: Fetch the discovery document
      ansible.builtin.uri:
        url: "{{ lab_app.oidc.issuer }}.well-known/openid-configuration"
        return_content: true
        use_proxy: false
      register: idp_discovery

    - name: Assert the issuer Authentik emits is the one lab.yaml configures
      ansible.builtin.assert:
        that: idp_discovery.json.issuer == lab_app.oidc.issuer
        fail_msg: >-
          lab.yaml `app.oidc.issuer` is {{ lab_app.oidc.issuer }} but the
          discovery document says {{ idp_discovery.json.issuer }}. The value
          must be the exact string in the token's `iss` claim, and Authentik
          builds it from the Host header of the request that reached it. Fix
          lab.yaml (usually to the same hostname the portal uses) rather than
          relaxing the portal's validation.

    - name: Assert the provider holds the client id the portal will send
      ansible.builtin.uri:
        url: "{{ idp_api }}/providers/oauth2/?name=Clayface%20portal"
        headers:
          Authorization: "Bearer {{ idp_token }}"
        return_content: true
        use_proxy: false
      register: idp_provider

    - name: Assert it
      ansible.builtin.assert:
        that:
          - idp_provider.json.pagination.count == 1
          - idp_provider.json.results[0].client_id == lab_app.oidc.client_id
        fail_msg: >-
          expected exactly one OAuth2 provider named 'Clayface portal' with
          client_id {{ lab_app.oidc.client_id }}; found
          {{ idp_provider.json.pagination.count }} provider(s).

    # --- The source is enabled AND synced ---
    #
    # These are two separate facts. An enabled source whose base DN the bind
    # account cannot read syncs zero objects and reports no error anywhere, so
    # the second half is the one that proves the AD scoping and the bind
    # credential are both right.
    - name: Fetch the LDAP source
      ansible.builtin.uri:
        url: "{{ idp_api }}/sources/ldap/?search=clayface"
        headers:
          Authorization: "Bearer {{ idp_token }}"
        return_content: true
        use_proxy: false
      register: idp_source

    - name: Assert the source is enabled
      ansible.builtin.assert:
        that:
          - idp_source.json.pagination.count == 1
          - idp_source.json.results[0].enabled | bool
        fail_msg: "the LDAP source is missing or disabled"

    - name: Assert the directory users were synced
      ansible.builtin.uri:
        url: "{{ idp_api }}/core/users/?search={{ item }}"
        headers:
          Authorization: "Bearer {{ idp_token }}"
        return_content: true
        use_proxy: false
      register: idp_user
      loop: "{{ idp_expect_users }}"
      loop_control:
        label: "{{ item }}"

    - name: Assert each expected user is present in Authentik
      ansible.builtin.assert:
        that:
          - item.json.pagination.count == 1
          - item.json.results[0].username == item.item
        fail_msg: >-
          {{ item.item }} is not in Authentik. If EVERY user is missing, the
          bind or the base DN is wrong; if only some are, check that the
          account is in OU=Users,OU={{ lab_ad.ou }} - the source cannot read
          outside it.
      loop: "{{ idp_user.results }}"
      loop_control:
        label: "{{ item.item }}"

    - name: Fetch the expected group
      ansible.builtin.uri:
        url: "{{ idp_api }}/core/groups/?search={{ idp_expect_group }}"
        headers:
          Authorization: "Bearer {{ idp_token }}"
        return_content: true
        use_proxy: false
      register: idp_group

    - name: Assert the group synced with its members
      ansible.builtin.assert:
        that:
          - idp_group.json.pagination.count == 1
          - idp_group.json.results[0].users_obj | map(attribute='username') | sort
            == idp_expect_group_members | sort
        fail_msg: >-
          {{ idp_expect_group }} in Authentik has members
          {{ idp_group.json.results[0].users_obj | default([]) | map(attribute='username') | list }},
          expected {{ idp_expect_group_members }}. Group MEMBERSHIP is a
          separate sync from group creation, and a group with no members is
          what a `member` attribute the bind account cannot read looks like.

    # --- The redirect URI, byte for byte ---
    - name: Assert the provider accepts the redirect URI lab.yaml declares
      ansible.builtin.assert:
        that: >-
          lab_app.oidc.redirect_url in
          (idp_provider.json.results[0].redirect_uris | map(attribute='url') | list)
        fail_msg: >-
          lab.yaml declares {{ lab_app.oidc.redirect_url }} but the provider
          holds {{ idp_provider.json.results[0].redirect_uris | map(attribute='url') | list }}.
          Authentik matches redirect URIs in strict mode, so a mismatch
          surfaces in the browser only after a successful password entry.
```

- [ ] **Step 5: Run it and watch the sync assertion fail first**

The blueprint exists but the source has not synced yet — Authentik syncs on a
schedule, not on instantiation.

```bash
set -a; . ./deploy.env; set +a
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/idp_validate.yml
```

Expected: FAIL on "Assert each expected user is present in Authentik" with
`bob.sing is not in Authentik.`

- [ ] **Step 6: Trigger the sync, then make the playbook do it**

The API can start a sync on demand. Add to `idp_validate.yml`, before the user
assertion:

```yaml
    # Authentik syncs the LDAP source on a schedule, so a fresh build has an
    # empty user list for a while. Triggering the sync here rather than
    # sleeping keeps the validator deterministic: it asserts the source CAN
    # sync, which is the property under test, instead of that it happened to
    # have synced by the time the playbook ran.
    #
    # A POST, which is why this is the one write in this playbook - and it is
    # a write to the IdP's own schedule, not to its configuration.
    - name: Trigger an LDAP sync
      ansible.builtin.uri:
        url: "{{ idp_api }}/sources/ldap/{{ idp_source.json.results[0].slug }}/sync/"
        method: POST
        headers:
          Authorization: "Bearer {{ idp_token }}"
        status_code: [200, 204]
        use_proxy: false
      changed_when: false

    - name: Wait for the sync to finish
      ansible.builtin.uri:
        url: "{{ idp_api }}/sources/ldap/{{ idp_source.json.results[0].slug }}/"
        headers:
          Authorization: "Bearer {{ idp_token }}"
        return_content: true
        use_proxy: false
      register: idp_sync_state
      until: >-
        idp_sync_state.json.sync_status is defined and
        idp_sync_state.json.sync_status.status != 'running'
      retries: 60
      delay: 2
      failed_when: >-
        idp_sync_state.json.sync_status is defined and
        idp_sync_state.json.sync_status.status == 'error'

    - name: Fail with the sync error if there was one
      ansible.builtin.fail:
        msg: >-
          the LDAP sync reported an error:
          {{ idp_sync_state.json.sync_status.logs | default('no logs') }}.
          A bind failure and an unreadable base DN both land here, which is
          why the log is printed rather than summarised.
      when: >-
        idp_sync_state.json.sync_status is defined and
        idp_sync_state.json.sync_status.status == 'error'
```

- [ ] **Step 7: Run it and see it pass**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/idp_validate.yml
```

Expected: every assertion passes. If the group-members assertion fails while the
users pass, read the failure message — it names the missing members and the
likely cause, which is the `member` attribute rather than the group object.

- [ ] **Step 8: Prove the misconfiguration is caught**

Point the source at a base the bind account cannot read, which is the silent
failure this validator exists for:

```bash
ansible -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
        localhost -c local -m ansible.builtin.shell -a 'true'
# edit ansible/idp/blueprints/portal.yaml: additional_user_dn -> "OU={{ lab_ad.ou }}"
# then re-run idp.yml and idp_validate.yml
```

Expected: `idp.yml` succeeds — the source is enabled and the blueprint is valid,
which is exactly why the silent case is dangerous — and `idp_validate.yml`
fails on the missing users, naming the base DN in its message. Revert the edit
and re-run both to confirm green.

- [ ] **Step 9: Commit**

```bash
git add ansible/idp/blueprints ansible/playbooks/idp_validate.yml lab.yaml
git commit -m "feat(lab): bind the identity provider to AD and publish the portal

One blueprint creates the LDAP source, the groups scope, the confidential
OIDC provider and the application. The search bases are the two OUs
svc-idp-ldap can actually read; a wider base would sync nothing and look
healthy, so the validator asserts named users and a group's membership."
```

---

### Task 6: The portal as an OIDC relying party

**Files:**
- Modify: `app/go.mod`, add `app/vendor/`
- Modify: `app/internal/config/config.go`
- Create: `app/internal/config/oidc_test.go`
- Create: `app/internal/web/sso.go`
- Create: `app/internal/web/sso_test.go`
- Modify: `app/internal/web/router.go`

**Interfaces:**
- Consumes: `lab_app.oidc.*` (Task 5) rendered into `OIDC_*` environment
  variables by Task 7.
- Produces: `config.OIDC` (`Enabled`, `Issuer`, `ClientID`, `ClientSecret`,
  `RedirectURL`, `RoleClaim`, `RoleMap`, `DefaultRole`); the routes
  `GET /portal/sso/login` and `GET /portal/sso/callback`; and
  `(*web.Server).ssoRolePriority(groups []string) string`.

- [ ] **Step 1: Add the dependencies**

```bash
cd /home/kiasoh/crapijat/College/project/app
go get github.com/coreos/go-oidc/v3/oidc@latest
go get golang.org/x/oauth2@latest
go mod tidy
go mod vendor
git status --short go.mod go.sum vendor | head -20
```

Expected: `go.mod` gains both as direct requirements and `vendor/` gains their
trees. The vendor directory is committed: the portal's container image is built
on the control node, and a build that needs the module proxy is a build that
fails on a lab network without one. Go's toolchain uses `vendor/` automatically
when it is present, so no `-mod=vendor` flag is needed in the Dockerfile.

- [ ] **Step 2: Write the failing config test**

`app/internal/config/oidc_test.go`:

```go
package config

import (
	"strings"
	"testing"
)

func TestParseOIDCRoleMap(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    map[string]string
		wantErr string
	}{
		{
			name: "two mappings",
			in:   "GG-IT-Admins=admin,GG-Employees=user",
			want: map[string]string{"GG-IT-Admins": "admin", "GG-Employees": "user"},
		},
		{
			name: "whitespace around entries is tolerated",
			in:   " GG-IT-Admins = admin , GG-Employees = user ",
			want: map[string]string{"GG-IT-Admins": "admin", "GG-Employees": "user"},
		},
		{
			name: "empty string is an empty map, not an error",
			in:   "",
			want: map[string]string{},
		},
		{
			name:    "a pair with no separator is an error",
			in:      "GG-IT-Admins",
			wantErr: "GG-IT-Admins",
		},
		{
			name:    "an empty group is an error",
			in:      "=admin",
			wantErr: "empty group name",
		},
		{
			name:    "an empty role is an error",
			in:      "GG-IT-Admins=",
			wantErr: "empty role",
		},
		{
			name:    "a duplicated group is an error rather than a silent override",
			in:      "GG-IT-Admins=admin,GG-IT-Admins=user",
			wantErr: "duplicate",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseOIDCRoleMap(tc.in)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error mentioning %q, got none (map=%v)", tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q does not mention %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestValidateRejectsIncompleteOIDC(t *testing.T) {
	base := func() *Config {
		return &Config{
			DBPassword:    "x",
			SessionSecret: "y",
			Weak:          Weaknesses{ForgeableSessionToken: false, HardcodedDBCredentials: false},
			OIDC:          OIDC{Enabled: true, Issuer: "http://idp01.clayface:9000/application/o/portal/"},
		}
	}

	cases := []struct {
		name    string
		mutate  func(*Config)
		wantEnv string
	}{
		{"no issuer", func(c *Config) { c.OIDC.Issuer = "" }, "OIDC_ISSUER"},
		{"no client id", func(c *Config) { c.OIDC.ClientID = "" }, "OIDC_CLIENT_ID"},
		{"no client secret", func(c *Config) { c.OIDC.ClientSecret = "" }, "OIDC_CLIENT_SECRET"},
		{"no redirect url", func(c *Config) { c.OIDC.RedirectURL = "" }, "OIDC_REDIRECT_URL"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := base()
			c.OIDC.ClientID, c.OIDC.ClientSecret, c.OIDC.RedirectURL = "portal", "s", "http://x/cb"
			tc.mutate(c)

			err := c.Validate()
			if err == nil {
				t.Fatalf("expected Validate to reject the config")
			}
			if !strings.Contains(err.Error(), tc.wantEnv) {
				t.Fatalf("error %q does not name %s", err.Error(), tc.wantEnv)
			}
		})
	}

	t.Run("disabled OIDC needs none of them", func(t *testing.T) {
		c := base()
		c.OIDC = OIDC{Enabled: false}
		if err := c.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
```

- [ ] **Step 3: Run it and see it fail**

Run: `cd app && go test ./internal/config/ -run 'OIDC' -v`

Expected: compile failure — `undefined: parseOIDCRoleMap`, `unknown field OIDC`,
`unknown field Enabled`.

- [ ] **Step 4: Add the config block**

In `app/internal/config/config.go`, add the type beside `Weaknesses`:

```go
// OIDC is the portal's relying-party configuration. It is a separate struct
// from Weaknesses because it is not a weakness: it is how the portal works in
// the Phase 5 posture, and every field in it is a correct setting.
type OIDC struct {
	Enabled      bool
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	// RoleClaim is the ID token claim carrying the group names. Named rather
	// than hardcoded because the claim's name is the identity provider's
	// choice, and a provider that renames it should be a lab.yaml edit.
	RoleClaim string
	// RoleMap maps a directory group to a portal role. It lives on the
	// relying party rather than as an IdP-side group attribute: the portal is
	// what enforces a role, so the portal is what should hold the rule.
	RoleMap map[string]string
	// DefaultRole is what a user in no mapped group gets. A user with no role
	// still authenticates - refusing them would make the mapping a second
	// authorization system, and the portal already has one.
	DefaultRole string
}
```

Add the fields to `Config`:

```go
	OIDC OIDC
```

Add the environment constants beside the existing `Env*` names:

```go
	EnvOIDCEnabled      = "OIDC_ENABLED"
	EnvOIDCIssuer       = "OIDC_ISSUER"
	EnvOIDCClientID     = "OIDC_CLIENT_ID"
	EnvOIDCClientSecret = "OIDC_CLIENT_SECRET"
	EnvOIDCRedirectURL  = "OIDC_REDIRECT_URL"
	EnvOIDCRoleClaim    = "OIDC_ROLE_CLAIM"
	EnvOIDCRoleMap      = "OIDC_ROLE_MAP"
	EnvOIDCDefaultRole  = "OIDC_DEFAULT_ROLE"
```

In `Load()`, after the existing fields are read:

```go
	oidcRoleMap, err := parseOIDCRoleMap(envOr(EnvOIDCRoleMap, ""))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", EnvOIDCRoleMap, err)
	}
	cfg.OIDC = OIDC{
		// Off unless it is asked for: with the variable unset the portal has
		// exactly the posture it had before the identity provider existed, so
		// the Phase 2 and 3 behaviour stays reproducible.
		Enabled:      envBoolOr(EnvOIDCEnabled, false),
		Issuer:       strings.TrimRight(envOr(EnvOIDCIssuer, ""), "/"),
		ClientID:     envOr(EnvOIDCClientID, ""),
		ClientSecret: envOr(EnvOIDCClientSecret, ""),
		RedirectURL:  envOr(EnvOIDCRedirectURL, ""),
		RoleClaim:    envOr(EnvOIDCRoleClaim, "groups"),
		RoleMap:      oidcRoleMap,
		DefaultRole:  envOr(EnvOIDCDefaultRole, "user"),
	}
```

Note `envBoolOr(EnvOIDCEnabled, false)`: the helper's documented behaviour is
"unset or unparsable means enabled", which is right for the weakness toggles and
wrong here. This call site is the exception, and it is worth a comment saying so
rather than changing the helper.

Add `parseOIDCRoleMap`:

```go
// parseOIDCRoleMap reads the keyword form `group=role,group=role`.
//
// Every malformed input is an error rather than a skip. A silently dropped
// entry means a user in that group authenticates with the default role, which
// looks like a working login and is an authorization bug - the kind that is
// found by accident. An empty string is an empty map, because a lab that maps
// no groups is a legitimate configuration and not a typo.
func parseOIDCRoleMap(raw string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		group, role, found := strings.Cut(pair, "=")
		if !found {
			return nil, fmt.Errorf("entry %q is not in group=role form", pair)
		}
		group, role = strings.TrimSpace(group), strings.TrimSpace(role)
		if group == "" {
			return nil, fmt.Errorf("entry %q has an empty group name", pair)
		}
		if role == "" {
			return nil, fmt.Errorf("entry %q has an empty role", pair)
		}
		if _, dup := out[group]; dup {
			return nil, fmt.Errorf("duplicate entry for group %q", group)
		}
		out[group] = role
	}
	return out, nil
}
```

Add `strings` to the imports if it is not already there, and to `Validate()`:

```go
	if c.OIDC.Enabled {
		// A half-configured relying party is the failure mode with no natural
		// symptom: the login route appears, the redirect happens, and the
		// exchange fails with an error from the library about a token. Naming
		// the environment variable is the whole point of this check.
		for _, req := range []struct {
			env string
			val string
		}{
			{EnvOIDCIssuer, c.OIDC.Issuer},
			{EnvOIDCClientID, c.OIDC.ClientID},
			{EnvOIDCClientSecret, c.OIDC.ClientSecret},
			{EnvOIDCRedirectURL, c.OIDC.RedirectURL},
		} {
			if req.val == "" {
				return fmt.Errorf("%s is required when %s is true", req.env, EnvOIDCEnabled)
			}
		}
	}
```

- [ ] **Step 5: Run the config tests**

Run: `cd app && go test ./internal/config/ -v`

Expected: PASS.

- [ ] **Step 6: Write the failing SSO tests**

`app/internal/web/sso_test.go`:

```go
package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"clayface/app/internal/config"
)

// fakeOIDC is the seam. It replaces the two network calls - the code exchange
// and the signature verification - and nothing else, so the handlers under
// test run their real state and nonce checks, their real role mapping and
// their real cookie handling.
type fakeOIDC struct {
	claims    map[string]any
	verifyErr error
	authURL   string
}

func (f *fakeOIDC) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	// Echo the state so the test can assert one was generated, and so a test
	// that forgets to set the cookie can still locate it.
	return f.authURL + "?state=" + state
}

func (f *fakeOIDC) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "fake"}, nil
}

func (f *fakeOIDC) Verify(_ context.Context, rawIDToken string) (*oidc.IDToken, error) {
	if f.verifyErr != nil {
		return nil, f.verifyErr
	}
	raw, err := json.Marshal(f.claims)
	if err != nil {
		return nil, err
	}
	return &oidc.IDToken{
		Issuer:  "http://idp01.clayface:9000/application/o/portal/",
		Subject: "bob.sing",
		Claims:  raw,
	}, nil
}

func oidcTestServer(t *testing.T, f *fakeOIDC) *Server {
	t.Helper()
	cfg := &config.Config{
		SessionSecret: "test-secret",
		Weak: config.Weaknesses{
			ForgeableSessionToken:   false,
			HardcodedDBCredentials:  false,
			VerboseErrorsDebugEndpoint: false,
		},
		OIDC: config.OIDC{
			Enabled:     true,
			Issuer:      "http://idp01.clayface:9000/application/o/portal/",
			ClientID:    "portal",
			ClientSecret: "secret",
			RedirectURL: "http://app01.clayface:8080/portal/sso/callback",
			RoleClaim:   "groups",
			RoleMap: map[string]string{
				"GG-IT-Admins":  "admin",
				"GG-Employees":  "user",
			},
			DefaultRole: "user",
		},
	}
	s := &Server{cfg: cfg, oidc: f}
	return s
}

func TestSSORolePriority(t *testing.T) {
	s := oidcTestServer(t, &fakeOIDC{})

	cases := []struct {
		name   string
		groups []string
		want   string
	}{
		{"no groups gets the default", nil, "user"},
		{"one mapped group", []string{"GG-Employees"}, "user"},
		{"both groups gets the higher one", []string{"GG-Employees", "GG-IT-Admins"}, "admin"},
		{"the higher one first is the same answer", []string{"GG-IT-Admins", "GG-Employees"}, "admin"},
		{"an unmapped group is ignored", []string{"GG-Developers"}, "user"},
		{"case is not guessed at", []string{"gg-it-admins"}, "user"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := s.ssoRolePriority(tc.groups); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSSOLoginSetsStateCookieAndRedirects(t *testing.T) {
	s := oidcTestServer(t, &fakeOIDC{authURL: "http://idp01.clayface:9000/application/o/authorize/"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/portal/sso/login", nil)
	s.handleSSOLogin(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("got status %d, want 302", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" {
		t.Fatal("no Location header")
	}

	var state string
	for _, c := range rec.Result().Cookies() {
		if c.Name == ssoStateCookieName {
			state = c.Value
		}
	}
	if state == "" {
		t.Fatalf("no %s cookie was set", ssoStateCookieName)
	}
	// The state in the cookie is the one in the URL. If these ever diverge,
	// every callback fails - which is a loud failure, and this is what makes
	// it a test failure first.
	if !strings.Contains(loc, "state="+state) {
		t.Fatalf("Location %q does not carry the cookie's state %q", loc, state)
	}
}

func TestSSOCallbackRejectsStateMismatch(t *testing.T) {
	s := oidcTestServer(t, &fakeOIDC{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/portal/sso/callback?code=x&state=attacker", nil)
	req.AddCookie(&http.Cookie{Name: ssoStateCookieName, Value: "expected"})
	s.handleSSOCallback(rec, req)

	if rec.Code == http.StatusFound {
		t.Fatalf("a state mismatch must not complete a login")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName && c.Value != "" {
			t.Fatal("a session cookie was set despite the state mismatch")
		}
	}
}

func TestSSOCallbackCreatesASessionFromTheClaims(t *testing.T) {
	s := oidcTestServer(t, &fakeOIDC{claims: map[string]any{
		"sub":    "bob.sing",
		"groups": []string{"GG-Employees", "GG-Developers"},
	}})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/portal/sso/callback?code=x&state=ok", nil)
	req.AddCookie(&http.Cookie{Name: ssoStateCookieName, Value: "ok"})
	s.handleSSOCallback(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want 303", rec.Code)
	}
	var session string
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			session = c.Value
		}
	}
	if session == "" {
		t.Fatal("no session cookie was set")
	}

	got, err := decodeSession(s.cfg, session)
	if err != nil {
		t.Fatalf("the session cookie does not decode: %v", err)
	}
	if got.Username != "bob.sing" {
		t.Fatalf("session username is %q, want bob.sing", got.Username)
	}
	if got.Role != "user" {
		t.Fatalf("session role is %q, want user (bob.sing is in GG-Employees)", got.Role)
	}
}

func TestSSOCallbackSurfacesAVerificationFailure(t *testing.T) {
	s := oidcTestServer(t, &fakeOIDC{verifyErr: errors.New("oidc: failed to verify signature")})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/portal/sso/callback?code=x&state=ok", nil)
	req.AddCookie(&http.Cookie{Name: ssoStateCookieName, Value: "ok"})
	s.handleSSOCallback(rec, req)

	if rec.Code == http.StatusSeeOther {
		t.Fatal("a failed verification must not complete a login")
	}
	if !strings.Contains(rec.Body.String(), "signature") {
		t.Fatalf("the body does not name the failure: %q", rec.Body.String())
	}
}
```

Add `strings` to the test imports.

- [ ] **Step 7: Run them and see them fail**

Run: `cd app && go test ./internal/web/ -run 'SSO' -v`

Expected: compile failure — `unknown field oidc in Server`, `undefined:
s.handleSSOLogin`, `undefined: ssoStateCookieName`, `undefined: s.ssoRolePriority`.

- [ ] **Step 8: Write `sso.go`**

```go
package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// ssoStateCookieName carries the CSRF state between the login redirect and the
// callback. Scoped to the callback path and short-lived: it exists for one
// round trip and there is no reason for it to be sent anywhere else.
const ssoStateCookieName = "clayface_sso_state"

// ssoClient is the seam between the handlers and the network.
//
// The flow has exactly two calls that cannot be made locally - exchanging the
// authorization code and verifying the token's signature - and both are here.
// Everything else in this file (state, nonce, claim extraction, role mapping,
// session creation) is local logic, and the tests exercise it through a fake
// implementation of this interface rather than through a live identity
// provider.
type ssoClient interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

// liveSSO is the production implementation: the library's own config and
// verifier, wired together.
type liveSSO struct {
	cfg      *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func (l *liveSSO) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return l.cfg.AuthCodeURL(state, opts...)
}

func (l *liveSSO) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return l.cfg.Exchange(ctx, code, opts...)
}

func (l *liveSSO) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return l.verifier.Verify(ctx, rawIDToken)
}

// newSSO builds the live client. Discovery happens here, at startup, so a
// misconfigured issuer fails the process rather than the first login: a portal
// that starts and then cannot authenticate anyone is harder to diagnose than
// one that does not start and says why.
func newSSO(ctx context.Context, cfg *config.Config) (ssoClient, error) {
	provider, err := oidc.NewProvider(ctx, cfg.OIDC.Issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery against %s: %w", cfg.OIDC.Issuer, err)
	}
	return &liveSSO{
		cfg: &oauth2.Config{
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			RedirectURL:  cfg.OIDC.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email", cfg.OIDC.RoleClaim},
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID}),
	}, nil
}

// handleSSOLogin starts the authorization-code flow.
func (s *Server) handleSSOLogin(w http.ResponseWriter, r *http.Request) {
	if s.oidc == nil {
		http.NotFound(w, r)
		return
	}

	state, err := randomToken()
	if err != nil {
		s.errorDetail(w, r, "could not start the login flow", err)
		return
	}
	nonce, err := randomToken()
	if err != nil {
		s.errorDetail(w, r, "could not start the login flow", err)
		return
	}

	// The nonce rides in the state cookie alongside the state rather than in
	// its own cookie: they have the same lifetime, the same scope, and one
	// cookie is one thing to get wrong instead of two. `state.nonce`.
	http.SetCookie(w, &http.Cookie{
		Name:     ssoStateCookieName,
		Value:    state + "." + nonce,
		Path:     "/portal/sso/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
		// Not Secure: the portal is served over HTTP in this lab (the
		// certificate is self-signed, docs/idp01-design.md records why). A
		// Secure cookie here would be silently dropped and every login would
		// fail with a state mismatch.
	})

	url := s.oidc.AuthCodeURL(state, oidc.Nonce(nonce))
	http.Redirect(w, r, url, http.StatusFound)
}

// handleSSOCallback completes the flow.
func (s *Server) handleSSOCallback(w http.ResponseWriter, r *http.Request) {
	if s.oidc == nil {
		http.NotFound(w, r)
		return
	}

	cookie, err := r.Cookie(ssoStateCookieName)
	if err != nil {
		http.Error(w, "no login in progress", http.StatusBadRequest)
		return
	}
	// Cleared on every path, including the failures: a state cookie that
	// outlives its round trip is a replay window.
	http.SetCookie(w, &http.Cookie{
		Name: ssoStateCookieName, Value: "", Path: "/portal/sso/", MaxAge: -1,
	})

	state, nonce, found := strings.Cut(cookie.Value, ".")
	if !found {
		http.Error(w, "malformed login state", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state {
		http.Error(w, "login state mismatch", http.StatusBadRequest)
		return
	}

	token, err := s.oidc.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		s.errorDetail(w, r, "code exchange failed", err)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		s.errorDetail(w, r, "the token response carried no id_token", nil)
		return
	}

	idToken, err := s.oidc.Verify(r.Context(), rawIDToken)
	if err != nil {
		s.errorDetail(w, r, "id_token verification failed", err)
		return
	}
	if idToken.Nonce != "" && idToken.Nonce != nonce {
		s.errorDetail(w, r, "id_token nonce mismatch", nil)
		return
	}

	var claims struct {
		Sub      string          `json:"sub"`
		Username string          `json:"preferred_username"`
		Name     string          `json:"name"`
		Email    string          `json:"email"`
		Groups   json.RawMessage `json:"-"`
	}
	if err := idToken.Claims(&claims); err != nil {
		s.errorDetail(w, r, "could not read the id_token claims", err)
		return
	}

	groups, err := extractGroups(idToken, s.cfg.OIDC.RoleClaim)
	if err != nil {
		s.errorDetail(w, r, "could not read the groups claim", err)
		return
	}

	username := claims.Username
	if username == "" {
		// Authentik is configured with sub_mode user_username, so `sub` is the
		// directory account name. Falling back to it keeps the session
		// identifiable with a provider that omits preferred_username.
		username = idToken.Subject
	}
	if username == "" {
		s.errorDetail(w, r, "the id_token identifies no user", nil)
		return
	}

	role := s.ssoRolePriority(groups)

	setSessionCookie(w, s.cfg, newSession(username, role))
	s.audit(r, username, "sso_login", fmt.Sprintf("role=%s groups=%s", role, strings.Join(groups, "|")))
	http.Redirect(w, r, "/portal/profile", http.StatusSeeOther)
}

// ssoRolePriority maps the directory groups to exactly one portal role.
//
// The priority order is fixed and explicit. A user in both GG-IT-Admins and
// GG-Employees gets admin: taking the union would be a different authorization
// model, and taking whichever the provider happened to list first would make
// the answer depend on claim ordering, which no one controls.
//
// Group names are compared exactly. Case-insensitive matching would be
// friendlier and would also mean a lab.yaml typo silently grants a role, so
// the strict comparison is the deliberate choice.
func (s *Server) ssoRolePriority(groups []string) string {
	rank := map[string]int{"admin": 2, "user": 1}
	best, bestRank := "", 0
	for _, g := range groups {
		role, ok := s.cfg.OIDC.RoleMap[g]
		if !ok {
			continue
		}
		if r := rank[role]; r > bestRank {
			best, bestRank = role, r
		}
	}
	if best == "" {
		return s.cfg.OIDC.DefaultRole
	}
	return best
}

// extractGroups reads the configured claim, which may be a list of names or a
// single name. Authentik's scope mapping returns a list; a provider that
// returns a space-separated string is common enough that accepting both is
// cheaper than the debugging session.
func extractGroups(idToken *oidc.IDToken, claim string) ([]string, error) {
	var raw map[string]json.RawMessage
	if err := idToken.Claims(&raw); err != nil {
		return nil, err
	}
	value, ok := raw[claim]
	if !ok {
		return nil, nil
	}

	var list []string
	if err := json.Unmarshal(value, &list); err == nil {
		return list, nil
	}
	var one string
	if err := json.Unmarshal(value, &one); err == nil {
		return strings.Fields(one), nil
	}
	return nil, fmt.Errorf("claim %q is neither a list nor a string", claim)
}

// randomToken returns a URL-safe 256-bit random string.
func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
```

- [ ] **Step 9: Wire the routes and the startup**

In `app/internal/web/router.go`, add the field to `Server`:

```go
type Server struct {
	cfg *config.Config
	db  *sql.DB
	// oidc is nil when OIDC is disabled, and both handlers answer 404 in that
	// case. Nil rather than a no-op implementation, because "the route exists
	// and does nothing" is worse than "the route does not exist".
	oidc ssoClient
}
```

In `NewServer`, after the mux is built:

```go
	if cfg.OIDC.Enabled {
		client, err := newSSO(context.Background(), cfg)
		if err != nil {
			return nil, err
		}
		s.oidc = client
		mux.HandleFunc("GET /portal/sso/login", s.handleSSOLogin)
		mux.HandleFunc("GET /portal/sso/callback", s.handleSSOCallback)
	}
```

Add the nil-database guard to `audit`:

```go
func (s *Server) audit(r *http.Request, username, action, detail string) {
	// The SSO handlers can run without a database: their unit tests exercise
	// the whole flow with no store. A login that cannot be recorded is not a
	// reason to fail the login, and a nil check here is cheaper than making
	// every caller carry the distinction.
	if s.db == nil {
		return
	}
	...
}
```

Add a login link to the login page so the flow is reachable from a browser:
in `portal.go`'s login template, add, guarded on the config:

```
{{ if .SSOEnabled }}<p><a href="/portal/sso/login">Sign in with Clayface SSO</a></p>{{ end }}
```

and put `SSOEnabled: s.cfg.OIDC.Enabled` into the page's data.

- [ ] **Step 10: Run the tests**

Run: `cd app && go test ./... -v 2>&1 | tail -40`

Expected: PASS. Then run the full build, which is what proves `vendor/` is
complete:

Run: `cd app && go build ./... && go vet ./...`

Expected: no output.

- [ ] **Step 11: Commit**

```bash
git add app/go.mod app/go.sum app/vendor app/internal/config \
        app/internal/web
git commit -m "feat(portal): authenticate through the identity provider over OIDC

Authorization code flow with state and nonce, the ID token verified
locally against the provider's JWKS, and AD groups mapped to portal roles
on the relying party's side from lab.yaml. The local password path stays
available, so the Phase 2 baseline and the Phase 5 posture can both be
demonstrated. Dependencies are vendored: the image build must not need a
module proxy."
```

---

### Task 7: Give the portal its OIDC environment, and assert the route

**Files:**
- Modify: `ansible/playbooks/app.yml` (the `.env` render and the vars)
- Modify: `ansible/playbooks/app_validate.yml`
- Modify: `ansible/inventory/terraform_vms.py` (only if `lab_app` needs exposing)

**Interfaces:**
- Consumes: `lab_app.oidc.*` (Task 5) and the routes from Task 6.
- Produces: a running portal whose `/portal/sso/login` redirects to the IdP's
  authorization endpoint.

- [ ] **Step 1: Extend the environment render in `app.yml`**

In the `app_env_content` var, after the existing weakness loop:

```jinja
{% if lab_app.oidc.enabled | default(false) %}
OIDC_ENABLED=true
OIDC_ISSUER={{ lab_app.oidc.issuer }}
OIDC_CLIENT_ID={{ lab_app.oidc.client_id }}
OIDC_CLIENT_SECRET={{ app_oidc_client_secret }}
OIDC_REDIRECT_URL={{ lab_app.oidc.redirect_url }}
OIDC_ROLE_CLAIM={{ lab_app.oidc.role_claim }}
OIDC_ROLE_MAP={% for group, role in lab_app.oidc.role_map.items() %}{{ group }}={{ role }}{% if not loop.last %},{% endif %}{% endfor %}

OIDC_DEFAULT_ROLE={{ lab_app.oidc.default_role }}
{% endif %}
```

with, in the vars block:

```yaml
    # The client secret is deliberately NOT read from lab.yaml: that file is
    # committed. It comes from deploy.env through the same lookup the other
    # credentials use.
    app_oidc_client_secret: "{{ lookup('env', 'LAB_OIDC_CLIENT_SECRET') }}"
```

and a guard, in the `pre_tasks`:

```yaml
    - name: Fail if OIDC is enabled but its client secret is missing
      ansible.builtin.assert:
        that: app_oidc_client_secret | length > 0
        fail_msg: >-
          lab.yaml has `app.oidc.enabled: true` but LAB_OIDC_CLIENT_SECRET is
          not in the environment. It must be the same value the blueprint gave
          the provider, which is the one in deploy.env. Export it:
          `set -a; . ./deploy.env; set +a`.
      when: lab_app.oidc.enabled | default(false)
```

Note the render is `no_log`-adjacent: the `.env` task should keep whatever
`no_log` posture it already has, and the `OIDC_CLIENT_SECRET` line makes that
non-optional — add `no_log: true` to the render task if it is not already set.

- [ ] **Step 2: Make `lab_app` available to the playbooks**

`lab.yaml`'s `app:` block is not currently surfaced to Ansible, which is why
`app.yml` uses the `LAB_*` environment variables instead. Now that OIDC is
declared in `lab.yaml` and its values are static facts, read it the way `ad.yml`
reads `lab_ad`. Confirm how `lab_ad` reaches Ansible:

```bash
grep -rn "lab_ad\b" ansible/ --include=*.yml --include=*.cfg | head
```

Mirror exactly what that does for `lab_app` — if `lab_ad` comes from
`group_vars/all.yml` or a `vars_files` entry, extend the same file; do not
invent a second mechanism.

- [ ] **Step 3: Write the failing assertion**

Add to `ansible/playbooks/app_validate.yml`:

```yaml
    # --- The SSO route, from the perspective of a browser ---
    #
    # The full OIDC flow needs a browser and a human, so it is a documented
    # manual step (docs/idp01-design.md). What is automatable is the first leg:
    # the portal must send the browser to the identity provider's authorization
    # endpoint, with the client id and the redirect URI it was configured with.
    # A portal whose issuer is wrong fails here, before the browser does.
    - name: Request the SSO login route
      ansible.builtin.uri:
        url: "{{ lab_app.oidc.redirect_url | regex_replace('/portal/sso/callback$', '') }}/portal/sso/login"
        follow_redirects: none
        status_code: [302, 303]
        use_proxy: false
      register: app_sso_login
      when: lab_app.oidc.enabled | default(false)

    - name: Assert it redirects to the identity provider
      ansible.builtin.assert:
        that:
          - app_sso_login.location is search('idp01')
          - app_sso_login.location is search('client_id=' + lab_app.oidc.client_id)
          - >-
            (app_sso_login.location | urldecode) is
            search('redirect_uri=' + (lab_app.oidc.redirect_url | urldecode))
        fail_msg: >-
          /portal/sso/login redirected to {{ app_sso_login.location }}. It must
          point at the identity provider with client_id
          {{ lab_app.oidc.client_id }} and redirect_uri
          {{ lab_app.oidc.redirect_url }}. A redirect_uri that differs by even
          one character is rejected by the provider only AFTER a successful
          password entry, which is why it is asserted here.
      when: lab_app.oidc.enabled | default(false)
```

- [ ] **Step 4: Run it with OIDC off and prove the assertion is skipped, then on**

```bash
cd /home/kiasoh/crapijat/College/project
sed -i 's/^    enabled: true$/    enabled: false/' lab.yaml   # the app.oidc toggle only
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/app.yml
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/app_validate.yml
```

Expected: the SSO assertion is skipped, and the rest of `app_validate.yml`
passes — the portal is fully functional on its local-password path with OIDC
off. That is the Phase 2/3 posture reproducing, and it is the check that the
new code did not make SSO mandatory.

Then turn it back on and prove the route:

```bash
python3 - <<'PY'
import re
p = 'lab.yaml'
s = open(p).read()
s = re.sub(r'^(  oidc:\n(?:.*\n)*?    )enabled: false$', r'\1enabled: true', s, flags=re.M)
open(p, 'w').write(s)
PY
grep -A2 '^  oidc:' lab.yaml
set -a; . ./deploy.env; set +a
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py ansible/playbooks/app.yml
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py ansible/playbooks/app_validate.yml
```

Expected: both pass, and the redirect assertion reports the authorization URL.

- [ ] **Step 5: Prove the live flow by hand, once**

The end-to-end flow needs a browser. From the control node (which is dual-homed
on both segment bridges, so it reaches the DMZ directly):

```bash
python3 -m webbrowser "http://app01.clayface:8080/portal/sso/login"
```

Expected, in order: the portal redirects to `idp01.clayface:9000`; the Authentik
login form appears; sign in as `bob.sing`; the consent screen appears; the
browser returns to `/portal/profile` and shows `bob.sing`. Capture the browser
URLs and a screenshot of the profile page as evidence — the standing rule is
that evidence is captured when the flow is demonstrated, not reconstructed
afterwards.

Then sign out and sign in as `abed.nad` and confirm the profile page shows an
admin role, which is the role mapping working end to end.

- [ ] **Step 6: Commit**

```bash
git add ansible/playbooks/app.yml ansible/playbooks/app_validate.yml lab.yaml
git commit -m "feat(lab): point the portal at the identity provider

The OIDC environment is rendered from lab.yaml with the client secret
from deploy.env, and app_validate asserts the first leg of the flow - the
authorization redirect, its client id and its redirect URI. That last one
is the mismatch a browser only reports after a successful password entry."
```

---

### Task 8: Wire the deploy, document the design, and update the figure

**Files:**
- Modify: `deploy.sh`
- Modify: `docs/roadmap.md` (Phase 5)
- Create: `docs/idp01-design.md`
- Modify: `red-clay/Architecture/Identity.md`
- Modify: `CLAUDE.md` (the architecture section and the docs list)
- Modify: `docs/planted-weaknesses.md` (the deviations bullet about plain LDAP)

**Interfaces:**
- Consumes: everything above.
- Produces: `./deploy.sh` building the whole lab in one run.

- [ ] **Step 1: Add the steps to `deploy.sh`**

`idp.yml` must run after `ad.yml` (the bind account has to exist) and before
`app.yml` (the portal's environment names the issuer, and the provider must be
there to answer discovery). `idp_validate.yml` goes immediately after `idp.yml`
so a broken blueprint fails before the portal is pointed at it.

Insert between the `lab_data.yml` step (Phase 4) and the `app.yml` step:

```bash
step "Ansible: build the identity provider (idp.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/idp.yml"

step "Ansible: validate the identity provider (idp_validate.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/idp_validate.yml"
```

Match the surrounding style exactly — the `step` helper, the line
continuations, and whether the existing steps pass `--ask-become-pass`.

Add the new environment variables to the `no_proxy` export if it is a list of
lab addresses: `10.0.0.30` is a new address the control node talks to, and the
ambient proxy will answer 503 for it exactly as it does for `10.0.0.1` and
`10.0.10.1`.

- [ ] **Step 2: Prove the whole thing deploys**

```bash
set -a; . ./deploy.env; set +a
./deploy.sh 2>&1 | tee /tmp/phase5-deploy.log
```

Expected: green through `ad_validate.yml`. The first run of a fresh lab still
stops once at `opnsense_dmz.yml` for the DMZ interface address; that is
unchanged and deliberate.

- [ ] **Step 3: Write `docs/idp01-design.md`**

Cover, in the style of the other design docs: what IDP01 is (an Authentik
deployment on the Ubuntu base image, one NIC on the LAN, four containers); why
Authentik rather than Keycloak or ADFS (open source, a blueprint mechanism that
keeps configuration in the repository, an LDAP source that binds an existing
directory without a forest trust, and a working OIDC provider); the LDAP bind
and why the base DNs are the two OUs `svc-idp-ldap` can read; the OIDC
integration and where the group-to-role mapping lives; the plain-HTTP deviation
and that the client secret crosses the LAN in the clear, matching section 14's
register; and the DMZ allowance with a pointer to the `lab.yaml` entry that
expresses it.

State honestly what the lab does **not** do: no TLS anywhere on the IdP leg, no
MFA, no session revocation beyond the token lifetimes, and no federation between
AD FS and Authentik — the AD side is a bind, which is a directory read, not a
trust.

- [ ] **Step 4: Update the figure in `Red-clay/Architecture/Identity.md`**

The Mermaid diagram currently marks IDP01 and PORTAL as `planned` and draws
`IDP01 -.->|"federation (planned)"| DC01`. Change:

- `IDP01` and `PORTAL` move to `classDef built`.
- `PORTAL -->|"OIDC authorization code"| IDP01` becomes a solid edge.
- `IDP01 -.->|"federation (planned)"| DC01` becomes
  `IDP01 -->|"LDAP bind (svc-idp-ldap)"| DC01`, solid, and worded as a bind
  rather than as federation, because that is what it is.
- `APP01SVC -.->|"LDAP / Kerberos when useful"| DC01` keeps its dashed style
  but the label becomes the accurate one: the portal's own account is used by
  the Chapter 5 chain, not by the portal's steady state, which now delegates
  authentication to IDP01.

Replace the closing note — "**IDP01 ↔ AD federation** is planned, not built" —
with what is now true, and keep the vault's prose style: no reformatting beyond
these edits.

- [ ] **Step 5: Correct the deviations bullet in `docs/planted-weaknesses.md`**

The bullet currently attributes the plain-LDAP credential exposure to the IdP.
IDP01 is on the LAN, so its bind never crosses the DMZ boundary; the credential
that does cross is the one the Chapter 5 chain recovers and uses from APP01:

```markdown
- **LDAP is plain on 389**, not LDAPS. The DMZ boundary's one allowance is
  DMZ -> `dc01` tcp 389/636, so the credential the Chapter 5 chain recovers
  crosses that boundary in the clear when it is used to bind, and the
  firewall's log records it. IDP01, which binds as `svc-idp-ldap` from the
  LAN, does not cross the boundary at all. Documented, not incidental.
```

- [ ] **Step 6: Update `CLAUDE.md`**

Three edits, all factual:

- The docs list gains `docs/idp01-design.md` and `docs/planted-weaknesses.md`
  (the latter is Phase 4's, and if it landed there already, only the IdP entry
  is new).
- The architecture section's "current state" paragraph gains IDP01 among the
  built VMs, with its role, address and what `idp.yml` deploys.
- The flat-network note stays and is still accurate: `dc01`, `client01` and
  `IDP01` are one segment. Do not write as though the user/server split exists.

- [ ] **Step 7: Strike the completed Phase 5 items in `docs/roadmap.md`**

Annotate the Phase 5 section the way the earlier phases are annotated: what
landed, when (2026-09-25), and where the design lives. Any item this plan
deliberately did not do — TLS, MFA, AD FS federation — goes in as explicit
future work rather than being silently dropped.

- [ ] **Step 8: Commit**

```bash
git add deploy.sh docs/idp01-design.md docs/roadmap.md docs/planted-weaknesses.md \
        CLAUDE.md red-clay/Architecture/Identity.md
git commit -m "docs(lab): record Phase 5 and make the identity figure true

IDP01 is built and the portal authenticates through it, so the modern leg
of the identity diagram stops being dashed. The design doc states the
deviations - plain HTTP, no MFA, no AD FS federation, a bind rather than a
trust - rather than leaving them to be inferred."
```

---

## Self-Review

**Spec coverage.** The roadmap's Phase 5 items map as: decide the IdP product →
Authentik, chosen by the user and recorded in `docs/idp01-design.md` (Task 8);
provision IDP01 with a terraform module, a `lab.yaml` placement including `ip:`,
and a deploy playbook → Tasks 1–4; bind the IdP to AD → Task 5; integrate
portal → OIDC → IDP01 → AD → Tasks 6–7; update the `Identity.md` figure → Task 8.
The CLAUDE.md data-flow rule is honoured throughout: the VM placement, the OIDC
client configuration and the DMZ allowance are all `lab.yaml` facts, and nothing
is duplicated into a playbook.

Three spec-adjacent decisions are made explicitly rather than by default, and
each is argued where it lands: IDP01 on the LAN rather than in the DMZ (Global
Constraints), the group-to-role mapping on the relying party rather than in
Authentik (Task 5 Step 1's comment), and one new firewall allowance rather than
a widened existing one (Task 3).

**Placeholder scan.** No `TBD`, no `TODO`, no "similar to Task N". Two steps
depend on reading a file this plan does not quote — Task 3 Step 1 (the shape of
`networks.dmz.allow` entries) and Task 7 Step 2 (how `lab_ad` reaches Ansible) —
and both say so and give the command that answers it, rather than guessing at a
schema.

**Type consistency.** `lab_app.oidc` has the same seven keys everywhere it
appears (Task 5 declares them; Task 6 names them as `OIDC_*` constants; Task 7
renders them; Task 8 asserts them). `ssoClient`'s three methods are implemented
by `liveSSO` and by the test's `fakeOIDC` with the same signatures.
`ssoStateCookieName` and `ssoRolePriority` are introduced in Task 6 and used
there and in the tests. The `idp` role string is spelled the same in `lab.yaml`,
`locals.tf`'s `module_os`, the module block filters, the inventory dispatch and
`hosts: vms_idp`.

**Review Focus.** (1) issuer/`iss` mismatch — Task 5 Step 4's discovery
assertion, which runs before the portal is pointed anywhere; (2) slow first boot
— Task 4 Step 3's five-minute budget and its log-dumping rescue; (3) a source
that syncs nothing — Task 5 Steps 4–8, which assert named users and a group's
membership and then prove the silent case is caught by deliberately widening a
base DN; (4) a half-configured relying party — Task 6 Step 2's table test, with
one case per environment variable; (5) a redirect-URI mismatch — Task 5 Step 4's
URI assertion on the provider and Task 7 Step 3's on the portal, checking the
same string from both ends.

## Execution Handoff

Nine tasks, with Tasks 1–3 (infrastructure) and Tasks 6 (the portal's Go
changes) independent of each other and everything else sequenced behind them.
The high-risk surfaces are the blueprint's `!Find`/`!Env` references, which fail
in a container log rather than at apply time, and the OIDC issuer contract,
which fails as a token error rather than as a hostname error — both are called
out with their own checks.

**Recommended: Subagent-driven** — the two failure surfaces above are exactly
the kind a fresh reviewer catches cheaply, because the failure message does not
name its cause and the implementer of a task is the worst-placed person to
notice that.
