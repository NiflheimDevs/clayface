# Phase 4 — Wire the AD Weaknesses and Plant the Exfiltration Subject

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the seven `ad.weaknesses` toggles in `lab.yaml` from declarations
that nothing reads into real configuration deltas — each `true` a genuine
misconfiguration with an actively-removing `false` branch — and give the
Chapter 5 impact analysis something to steal.

**Architecture:** Every weakness is one configuration delta applied by the
playbook that owns the object, guarded by a flag, and asserted in both postures
by `ad_validate.yml`. No new module, no new playbook for the weaknesses
themselves: W1/W3/W5/W6/W7 land in `ad.yml`, W2 in `client.yml`, W4 in
`ad_gpo.yml`. The exfiltration subject is a new `lab_data.yml` playbook driven
by a `data.shares` map in `lab.yaml`.

**Tech Stack:** Terraform (dmacvicar/libvirt v0.9.8) + Ansible
(`microsoft.ad`, `ansible.windows`) over WinRM/NTLM; PowerShell through
`ansible.windows.win_powershell`; the AD module and the LAPS module for the
directory objects and their ACLs.

**Spec:** `docs/ad-identity-design.md` (section 13 is the authoritative
normal→vulnerable table for W1–W7; section 14 is the deviations register),
`ignoreme/scenario-design.md` section 4 (the five mechanisms and the
additive-removal rule), `docs/roadmap.md` Phase 4, and
`docs/app01-design.md` (the APP01 toggle table, which is the worked pattern for
the Chapter 4 documentation this plan produces).

## Global Constraints

- **A toggle off is a real configuration, not a skipped task.** Grant, Drift and
  Artifact mechanisms all use additive Ansible/PowerShell idioms, so the
  flag-off branch must actively remove what the flag-on branch added
  (`state: absent`, `RemoveAccessRuleSpecific`, `Set-GPPermission -PermissionLevel
  None`, `setspn -D`, `Remove-ADGroupMember`). This is the rule from
  `ignoreme/scenario-design.md` section 4 and every task below obeys it.
- **Never remove what the baseline owns.** Only GG-IT-Admins is a local
  administrator by design, only Domain Admins may edit GPOs by design, and only
  the built-in Administrator is Ansible's transport. A flag-off branch removes
  exactly the principal it added, matched by SID where an ACL is involved.
- **`ip:` stays required for every Windows VM.** Nothing in this phase removes
  it; W2 and W7 are asserted on `client01`, whose `ip:` is already in
  `lab.yaml`.
- **Toggles are data in `lab.yaml` only.** No playbook may hardcode a weakness
  as enabled, and `ad.weaknesses` stays a flat map of booleans. The one value a
  weakness needs (the account W7 adds) lives next to the object it belongs to
  (`ad.workstation.weak_logon_member`), not inside the toggle map.
- **`client01` and `dc01` must be enabled in `lab.yaml` before Task 3.** Both
  are currently commented out or partially built; `client.yml` and its
  assertions cannot run against a VM that terraform does not create.
- **No credentials in `lab.yaml`.** Passwords come from `deploy.env`
  (`LAB_USER_PASS`, `LAB_SVC_APP_PASS`, `LAB_WIN_ADMIN_PASS`).
- **The firewall is not widened.** `networks.dmz.allow` permits DMZ→`10.0.0.10`
  tcp 389/636 only. Kerberos (88) and the DRSUAPI/RPC traffic DCSync needs do
  not cross that boundary, which is why the escalation half of the chain runs
  from a LAN foothold. This plan states that model; it does not change a rule.
- **Read is the idempotency key.** Every task that mutates the directory reads
  the current state first and reports honestly whether anything moved, and
  wraps the write in `Set-Acl` only when something did — the idiom `ad.yml`
  already uses.

## Review Focus

The spec describes each weakness and its fix. It does not describe what happens
on the second run, on a partially-built lab, or with a misconfigured input.
These are the five ways this software is most likely to bite, most likely
first. Each one gets a test in the task that owns the code.

1. **A flag flipped off while the lab is already vulnerable.** The whole point of
   the additive-removal rule, and the failure is silent: Ansible reports `ok`,
   the ACE stays, and the lab is still weak. Every task's flag-off run asserts
   the artefact is *gone*, not merely that the run succeeded.
2. **`client01` commented out of `lab.yaml` (its state today).** Then W2's task
   never runs, `ad_validate.yml`'s workstation play matches no host, and the
   validator reports green for a workstation that does not exist. Both the
   playbook and the validator must fail loudly rather than pass vacuously.
3. **A misconfigured `weak_logon_member` or a missing account.** W7 adds an
   account name read from `lab.yaml` to a group; a typo must fail with the name
   in the message rather than adding nothing and reporting success.
4. **`LAB_USER_PASS` or `LAB_SVC_APP_PASS` unset or mismatched.** W1's cracking
   step and W7's whole premise depend on the credential the portal plants being
   the same string as the account's real password. A mismatch makes the chain
   silently useless while every assertion still passes.
5. **On → off → on convergence.** A toggle flipped twice must land in the same
   state as one flipped once. This is the same failure as (1) from the other
   direction and is what the "flip it back on" step in each task proves.

---

## File Structure

| File | Responsibility in this phase |
| --- | --- |
| `lab.yaml` | The seven toggles stop being all-false; `ad.workstation.weak_logon_member`, `ad.idp.spn`, and a new `data.shares` map are added |
| `ansible/playbooks/ad.yml` | Owns W1 (SPN), W3 (LAPS read), W5 (delegation), W6 (DCSync rights), W7 (group membership) |
| `ansible/playbooks/client.yml` | Owns W2 (local Administrators membership) |
| `ansible/playbooks/ad_gpo.yml` | Owns W4 (GPO editing permission) |
| `ansible/playbooks/ad_validate.yml` | Owns the assertions for all seven, in both postures |
| `ansible/playbooks/lab_data.yml` | New: plants the shares and documents the impact analysis exfiltrates |
| `deploy.sh` | Runs `lab_data.yml` after `client.yml` |
| `docs/planted-weaknesses.md` | New: the Chapter 4 planted-weakness section |

Task order follows the chain's narrative — W1 and W6 are the escalation, W7 is
the bridge from the DMZ into the LAN — but nothing depends on that order.
Tasks 1–8 are independent of each other.

---

### Task 1: Make the weakness flags a first-class contract

The toggles exist in `lab.yaml` and nothing reads them. Before wiring any of
them, give every playbook that will read them the same var block and the same
guard `app.yml` uses for `app.weaknesses`, so a typo'd key fails loudly instead
of reading as `false`.

**Files:**
- Modify: `lab.yaml` (the `ad.weaknesses` comment; the map stays all-false)
- Modify: `ansible/playbooks/ad.yml:14-25` (vars block) and its `pre_tasks`
- Modify: `ansible/playbooks/client.yml` (vars block, `pre_tasks`)
- Modify: `ansible/playbooks/ad_gpo.yml:19-28` (vars block) and add a `pre_tasks` block
- Modify: `ansible/playbooks/ad_validate.yml:14-20` (vars block)

**Interfaces:**
- Produces: the var `lab_ad_weak` in all four playbooks — a map of the seven
  toggle names to booleans, and the list `lab_ad_weak_keys` naming every toggle
  the playbooks implement. Every later task reads
  `lab_ad_weak.<toggle_name> | bool`.

- [ ] **Step 1: Write the failing check — assert the guard fires**

`lab.yaml` already has the map. Temporarily add a key the playbooks do not
implement, which is the input class the guard exists for:

```bash
cd /home/kiasoh/crapijat/College/project
cp lab.yaml /tmp/lab.yaml.bak
python3 - <<'PY'
import re
p = 'lab.yaml'
s = open(p).read()
s = s.replace("    excessive_group_membership: false\n",
              "    excessive_group_membership: false\n    kerberoastable_svc_app: false\n")
open(p, 'w').write(s)
PY
grep -n "kerberoastable_svc_app" lab.yaml
```

- [ ] **Step 2: Run it and see it fail**

Run: `ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml --list-tasks`

Expected: the playbook parses and lists tasks, and says nothing at all about
`kerberoastable_svc_app` — the undeclared toggle is invisible. That silence is
the failing test.

- [ ] **Step 3: Add the var block and the guard to `ad.yml`**

Add to the `vars:` block in `ansible/playbooks/ad.yml`, after `ad_anchor_dn`:

```yaml
    # The AD weakness toggles, from lab.yaml `ad.weaknesses`. `default({})` is
    # what makes a missing key read as "baseline"; the pre_task below is what
    # stops a typo'd key from doing the same thing silently.
    lab_ad_weak: "{{ (lab_ad | default({}, true)).weaknesses | default({}, true) }}"

    # Every toggle this playbook family implements. Duplicated from lab.yaml
    # on purpose, and it is the same pattern app.yml uses for
    # `app_weakness_keys`: the list is the playbook's declaration of what it
    # knows how to build, and the warn task below is what turns a drift between
    # the two into a message instead of a silently-ignored toggle.
    lab_ad_weak_keys:
      - kerberoastable_idp_bind
      - excessive_local_admin
      - excessive_laps_read
      - weak_gpo_permission
      - unconstrained_delegation
      - excessive_idp_directory_permissions
      - excessive_group_membership
```

Add a `pre_tasks:` block to that play, immediately before `tasks:` (the play
currently has none — `ad.yml` starts its work in `tasks:`):

```yaml
  pre_tasks:
    - name: Fail if lab.yaml has no usable `ad.weaknesses` map
      ansible.builtin.assert:
        that: lab_ad_weak | length > 0
        fail_msg: >-
          lab.yaml has no `ad.weaknesses` map (or it is empty). The seven
          toggles are static lab facts and belong in lab.yaml next to the
          users and groups they govern, not in this playbook. See
          docs/ad-identity-design.md section 13 for the list.

    - name: Note any AD toggle in lab.yaml this playbook does not implement
      ansible.builtin.set_fact:
        lab_ad_weak_unknown: "{{ lab_ad_weak.keys() | reject('in', lab_ad_weak_keys) | list }}"

    - name: Warn about unknown AD toggles
      ansible.builtin.debug:
        msg: >-
          lab.yaml `ad.weaknesses` declares {{ lab_ad_weak_unknown | join(', ') }},
          which ad.yml does not implement and ad_validate.yml does not assert.
          Either the toggle list in the playbooks is stale, or the key is a
          typo — and a typo'd key reads as `false` everywhere, so the lab would
          quietly be its baseline while lab.yaml claims otherwise.
      when: lab_ad_weak_unknown | length > 0
```

- [ ] **Step 4: Add the same var block and guard to `client.yml`, `ad_gpo.yml`, `ad_validate.yml`**

Same `lab_ad_weak` var and same `lab_ad_weak_keys` list in each of the three
plays' `vars:` blocks. `ad_gpo.yml` has no `pre_tasks:` block today — add one
with the same three tasks. `client.yml` and `ad_validate.yml` follow their own
existing `pre_tasks` (or `tasks`) structure; the guard goes first in whichever
block runs first for the play that touches the directory.

- [ ] **Step 5: Run the check again**

Run: `ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml 2>&1 | head -40`

Expected: the run reaches the warn task and prints
`lab.yaml 'ad.weaknesses' declares kerberoastable_svc_app, ...`. The typo is now
visible instead of silent.

- [ ] **Step 6: Remove the bogus key and prove the warning is gone**

```bash
cp /tmp/lab.yaml.bak lab.yaml
grep -c "kerberoastable_svc_app" lab.yaml
```

Expected: `0`.

- [ ] **Step 7: Commit**

```bash
git add lab.yaml ansible/playbooks/ad.yml ansible/playbooks/client.yml \
        ansible/playbooks/ad_gpo.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): give the AD weakness toggles a declared contract

Nothing read lab.yaml's ad.weaknesses, so a typo'd toggle was
indistinguishable from a toggle that was deliberately off. Each playbook
that will read one now declares the list it implements and warns about
keys it does not know, the same shape app.yml uses for app.weaknesses."
```

---

### Task 2: W1 — the Kerberoastable IDP bind account

A service account with a Service Principal Name can be Kerberoasted: any
authenticated principal in the forest may request a TGS for it, the ticket is
encrypted with the account's own key, and it can be taken away and cracked
offline. No SPN means no ticket to request and nothing to crack.

The password half of W1 is already in place and is **not** a delta: the account
is created with `LAB_USER_PASS` (the one shared human password, a documented
deviation in `docs/ad-identity-design.md` section 14) and the domain account
policy sets `MaxPasswordAge 0`, so it never rotates. That is what makes the
captured ticket worth cracking, and it belongs in the Chapter 4 write-up as
context rather than as something this task creates.

**Files:**
- Modify: `lab.yaml` (`ad.idp.spn`)
- Modify: `ansible/playbooks/ad.yml` (new task after "Create the lab user accounts")
- Modify: `ansible/playbooks/ad_validate.yml` (new assertion in the `vms_dc` play)

**Interfaces:**
- Consumes: `lab_ad_weak.kerberoastable_idp_bind` from Task 1.
- Produces: `lab_ad.idp.spn` — the SPN string, read by the playbook and by the
  validator so the two can never name different SPNs.

- [ ] **Step 1: Declare the SPN in `lab.yaml`**

Under `ad.idp`, after `read_ous`:

```yaml
    # The SPN weakness W1 gives this account (docs/ad-identity-design.md
    # section 13). Declared here rather than inline in ad.yml because the
    # validator asserts the same string, and two copies could disagree.
    #
    # The host names IDP01 because IDP01 is the service the account exists
    # for. IDP01 does not exist yet (roadmap Phase 5 adds it at
    # 10.0.0.30); nothing resolves an SPN by name, so the string is
    # forward-consistent rather than broken.
    spn: HTTP/idp01.clayface
```

- [ ] **Step 2: Write the failing assertion**

Add to `ansible/playbooks/ad_validate.yml` in the `vms_dc` play, after
"Assert the domain account policy matches the design":

```yaml
    # --- DELIBERATE WEAKNESS: W1, kerberoastable_idp_bind ---
    #
    # Asserted in both postures from the toggle: with the flag on the account
    # must carry the SPN, with it off the SPN must be gone. The second half is
    # the negative assertion that matters — a kerberoastable account nobody
    # declared is drift, and drift is what the whole validator exists to catch.
    #
    # The lab.yaml `spn` string is read here rather than repeated, so a rename
    # in one place cannot leave the assertion checking a stale SPN.
    - name: Assert the Kerberoastable SPN matches the declared weakness
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory

          $account = '{{ lab_ad.idp.bind_account }}'
          $spn     = '{{ lab_ad.idp.spn }}'
          $want    = [bool]${{ lab_ad_weak.kerberoastable_idp_bind | bool | ternary('true', 'false') }}

          $has     = @((Get-ADUser $account -Properties ServicePrincipalNames).ServicePrincipalNames)
          $present = $has -contains $spn
          if ($present -ne $want) {
              throw "$account SPN '$spn' is $present, the declared weakness W1 says $want"
          }

          # krbtgt's own SPNs are part of how the forest works; every other
          # user object holding one is a Kerberoastable target this lab did not
          # declare. Computer accounts are excluded by using Get-ADUser — a
          # machine SPN is normal and is not what Kerberoasting attacks.
          $others = @(Get-ADUser -LDAPFilter '(servicePrincipalName=*)' `
                        -Properties ServicePrincipalNames |
                        Where-Object { $_.SamAccountName -ne 'krbtgt' -and
                                       $_.SamAccountName -ne $account } |
                        Select-Object -ExpandProperty SamAccountName)
          if ($others.Count) {
              throw "undeclared Kerberoastable accounts: $($others -join ', ')"
          }

          "OK: $account SPN present=$present, no other Kerberoastable user accounts"
      changed_when: false
```

- [ ] **Step 3: Run it with the toggle on and see it fail**

```bash
cd /home/kiasoh/crapijat/College/project
# flip W1 on (it is the input class this test exists for)
sed -i 's/^    kerberoastable_idp_bind: false$/    kerberoastable_idp_bind: true/' lab.yaml
grep -n "kerberoastable_idp_bind" lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/ad_validate.yml
```

Expected: FAIL on "Assert the Kerberoastable SPN matches the declared weakness"
with `svc-idp-ldap SPN 'HTTP/idp01.clayface' is False, the declared weakness W1
says True`. `ad.yml` has not been wired yet — that is the point.

- [ ] **Step 4: Wire the toggle into `ad.yml`**

Add after "Create the lab user accounts" in `ansible/playbooks/ad.yml`:

```yaml
    # --- DELIBERATE WEAKNESS: W1, kerberoastable_idp_bind ---
    #
    # A service account holding a Service Principal Name is Kerberoastable: any
    # authenticated principal in the forest may request a TGS for it, the
    # ticket is encrypted with the account's own key, and it can be taken away
    # and cracked offline. No SPN means no ticket to request, which is why the
    # flag-off branch REMOVES the SPN rather than merely not adding it.
    #
    # `setspn -S` refuses to write an SPN that is already held by another
    # account, so it cannot create a duplicate; `setspn -D` removes one and is
    # only reached when the SPN is present, which is what keeps both branches
    # idempotent. The read is the idempotency key.
    #
    # The password half of W1 is not a delta and is not created here: the
    # account is created with LAB_USER_PASS (the single shared human password,
    # section 14) and the domain policy sets MaxPasswordAge 0, so it never
    # rotates. That is what makes the captured ticket worth cracking, and it is
    # why W1 needs no second change to be exploitable.
    - name: Set the SPN that makes the IDP bind account Kerberoastable
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'

          $account = '{{ lab_ad.idp.bind_account }}'
          $spn     = '{{ lab_ad.idp.spn }}'
          $want    = [bool]${{ lab_ad_weak.kerberoastable_idp_bind | bool | ternary('true', 'false') }}

          $cur = @(setspn -L $account 2>$null |
                     Select-Object -Skip 1 |
                     ForEach-Object { $_.Trim() } |
                     Where-Object { $_ })
          $has = $cur -contains $spn
          if ($has -eq $want) { 'OK'; return }

          if ($want) { setspn -S $spn $account | Out-Null }
          else       { setspn -D $spn $account | Out-Null }
          if ($LASTEXITCODE -ne 0) { throw "setspn failed ($LASTEXITCODE) for $spn on $account" }
          "CHANGED: $account SPN '$spn' -> $want"
      register: idp_spn
      changed_when: "'CHANGED' in (idp_spn.output | join(' '))"
```

- [ ] **Step 5: Run `ad.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `ad.yml` reports `CHANGED: svc-idp-ldap SPN 'HTTP/idp01.clayface' ->
True`; the validator reports `OK: svc-idp-ldap SPN present=True, no other
Kerberoastable user accounts` and passes.

- [ ] **Step 6: Run it a second time and prove idempotence**

Run: the same two commands again.

Expected: `ad.yml` reports the SPN task `ok` (no `CHANGED`) — the read found the
SPN already present.

- [ ] **Step 7: Flip the toggle off and prove the SPN is actively removed**

```bash
sed -i 's/^    kerberoastable_idp_bind: true$/    kerberoastable_idp_bind: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: svc-idp-ldap SPN 'HTTP/idp01.clayface' -> False`, then the
validator passes with `present=False`. A skipped task would have left the SPN
behind and this is the step that catches it.

- [ ] **Step 8: Commit**

```bash
git add lab.yaml ansible/playbooks/ad.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W1, the Kerberoastable IDP bind account

svc-idp-ldap gains HTTP/idp01.clayface while the toggle is on, and the
SPN is actively removed when it is off. The account's weak, non-rotating
password is not created here: it is the shared lab password under a
domain policy with MaxPasswordAge 0, which section 14 already documents."
```

---

### Task 3: W7 — the service account in the all-employees group

W7 is what turns the credential the Chapter 5 chain recovers from the portal
database into a foothold. `svc-app-portal` is in no group today, so the
credential the attacker steals can bind LDAP but cannot log on anywhere. Put it
in `GG-Employees` — the group that carries `SeInteractiveLogonRight` on the
workstations — and the recovered credential becomes an interactive logon on
`client01`, which is the step that moves the attacker across the DMZ boundary
into the LAN.

This is also the ordinary real-world shape of the finding: someone added a
service account to the all-employees group so it could reach a share.

**Files:**
- Modify: `lab.yaml` (`ad.workstation.weak_logon_member`)
- Modify: `ansible/playbooks/ad.yml` (new task after the SPN task)
- Modify: `ansible/playbooks/ad_validate.yml` ("Assert each logon group holds exactly its designed members")

**Interfaces:**
- Consumes: `lab_ad_weak.excessive_group_membership` from Task 1.
- Produces: `lab_ad.workstation.weak_logon_member` — the account W7 adds,
  read by both the playbook and the validator.

- [ ] **Step 1: Declare the account in `lab.yaml`**

Under `ad.workstation`:

```yaml
  workstation:
    logon_group: GG-Employees
    admin_group: GG-IT-Admins
    # The account weakness W7 adds to the logon group. It is the portal's
    # service account: the credential the Chapter 5 chain recovers from the
    # portal database, and W7 is what makes that credential useful for
    # something beyond a bind — it is the step that moves the attacker from
    # the DMZ into the LAN, because the DMZ boundary allows tcp 389/636 to
    # dc01 and nothing else.
    #
    # Named here rather than inline in ad.yml because ad_validate.yml asserts
    # the same account, and a typo'd name in one place would be a silent
    # no-op in the other.
    weak_logon_member: svc-app-portal
```

- [ ] **Step 2: Write the failing assertion**

In `ansible/playbooks/ad_validate.yml`, modify "Assert each logon group holds
exactly its designed members" so the expected set is flag-aware. Replace the
`$want = ...` line and add the W7 block immediately after it:

```powershell
          $want = @({% for entry in lab_ad.users | dict2items | selectattr('value.groups', 'contains', item) %}'{{ entry.key }}'{% if not loop.last %}, {% endif %}{% endfor %})
          {% if lab_ad_weak.excessive_group_membership | bool and item == lab_ad.workstation.logon_group %}
          # --- DELIBERATE WEAKNESS: W7, excessive_group_membership ---
          #
          # While the flag is on, this account is a member of the logon group
          # on purpose and it appears in no lab.yaml `groups` list, so it is
          # appended here rather than derived from the users map. The
          # `item == ...` guard is what keeps it out of the role groups'
          # expectations.
          $want += '{{ lab_ad.workstation.weak_logon_member }}'
          {% endif %}
```

The existing `$missing`/`$extra` comparison then does the work in both
directions: with the flag on the account must be present, with it off it must
not be.

- [ ] **Step 3: Run it with the toggle on and see it fail**

```bash
sed -i 's/^    excessive_group_membership: false$/    excessive_group_membership: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: FAIL on "Assert each logon group holds exactly its designed members"
for `GG-Employees`, with `missing=[svc-app-portal]`.

- [ ] **Step 4: Wire the toggle into `ad.yml`**

Add after the SPN task. The account must exist before it can be a member — it is
created by "Create the lab user accounts" earlier in the same play, so ordering
is already correct.

```yaml
    # --- DELIBERATE WEAKNESS: W7, excessive_group_membership ---
    #
    # The portal's service account is a member of the all-employees group,
    # which is the group that carries SeInteractiveLogonRight on the
    # workstations (client.yml). The consequence is the chain's hinge: the
    # credential the portal database hands over becomes an interactive logon
    # on a domain workstation instead of a credential that can only bind
    # LDAP. Without W7 the recovered credential cannot log on anywhere, and
    # the DMZ boundary — tcp 389/636 to dc01 and nothing else — stops the
    # attacker there.
    #
    # The finding, as an administrator would have made it: the account was
    # added to the group so it could reach a share, and nobody revisited it.
    #
    # Explicit membership rather than microsoft.ad.user's `groups.add`, because
    # the flag-off branch has to REMOVE: `add` is additive and a skipped task
    # would leave the account in the group forever. The existence check runs
    # first so a typo'd name fails with the name in the message instead of
    # adding nothing and reporting success.
    - name: Set the service account's membership of the logon group
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory

          $group  = '{{ lab_ad.workstation.logon_group }}'
          $member = '{{ lab_ad.workstation.weak_logon_member }}'
          $want   = [bool]${{ lab_ad_weak.excessive_group_membership | bool | ternary('true', 'false') }}

          if (-not (Get-ADUser $member -ErrorAction SilentlyContinue)) {
              throw "lab.yaml names '$member' as the W7 logon-group member, but no such account exists. Check ad.workstation.weak_logon_member and ad.users."
          }

          $have = @(Get-ADGroupMember $group | Select-Object -ExpandProperty SamAccountName)
          $is   = $have -contains $member
          if ($is -eq $want) { 'OK'; return }

          if ($want) { Add-ADGroupMember -Identity $group -Members $member }
          else       { Remove-ADGroupMember -Identity $group -Members $member -Confirm:$false }
          "CHANGED: $group member '$member' -> $want"
      register: ad_weak_logon_member
      changed_when: "'CHANGED' in (ad_weak_logon_member.output | join(' '))"
```

- [ ] **Step 5: Run `ad.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: GG-Employees member 'svc-app-portal' -> True`, then the
validator passes and reports `GG-Employees holds 5 members, all expected`.

- [ ] **Step 6: Flip the toggle off and prove the membership is removed**

```bash
sed -i 's/^    excessive_group_membership: true$/    excessive_group_membership: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: GG-Employees member 'svc-app-portal' -> False`, then the
validator passes with 4 members.

- [ ] **Step 7: Prove the typo guard fires**

```bash
sed -i 's/^    weak_logon_member: svc-app-portal$/    weak_logon_member: svc-app-portal1/' lab.yaml
sed -i 's/^    excessive_group_membership: false$/    excessive_group_membership: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
```

Expected: FAIL with `lab.yaml names 'svc-app-portal1' as the W7 logon-group
member, but no such account exists.`

```bash
sed -i 's/^    weak_logon_member: svc-app-portal1$/    weak_logon_member: svc-app-portal/' lab.yaml
sed -i 's/^    excessive_group_membership: true$/    excessive_group_membership: false/' lab.yaml
```

- [ ] **Step 8: Commit**

```bash
git add lab.yaml ansible/playbooks/ad.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W7, the service account in the all-employees group

svc-app-portal joins GG-Employees while the toggle is on and is removed
when it is off. This is the chain's bridge: the credential the portal
database hands over becomes an interactive logon on client01, which is
how the attacker crosses the DMZ boundary instead of stopping at it."
```

---

### Task 4: W6 — the DCSync replication rights

The escalation half of the chain. W1 hands over `svc-idp-ldap`'s credential; W6
turns that credential into the `krbtgt` hash, because a principal holding
`Replicating Directory Changes` **and** `Replicating Directory Changes All` on
the domain root can ask a DC to replicate the password hashes over DRSUAPI.

This is a real-world finding with a real-world cause: directory-sync products
(Azure AD Connect, a backup agent, a monitoring tool) genuinely need these
rights, and the grant is documented as a step in their install guides. Nobody
narrows it afterwards.

**Files:**
- Modify: `ansible/playbooks/ad.yml` (new task after "Grant the IDP bind account read-only access to its scoped OUs")
- Modify: `ansible/playbooks/ad_validate.yml` ("Assert the IDP delegation is scoped")

**Interfaces:**
- Consumes: `lab_ad_weak.excessive_idp_directory_permissions` from Task 1.
- Produces: nothing consumed by later tasks; the ACE is the deliverable.

- [ ] **Step 1: Write the failing assertion**

The existing "Assert the IDP delegation is scoped" task already asserts zero
explicit ACEs for the bind account on the domain root, with the throw message
"this is what weakness W6 looks like". That assertion now has to be flag-aware:
the domain root is the one DN where the expected count depends on the toggle.

Replace the header comment above the task (the `# Two halves...` block) with:

```yaml
    # Two halves, and the second is the one that matters.
    #
    # Scoping is asserted on the ACL, not by a search: every authenticated
    # user in a default AD can read the whole directory, so "can it read the
    # domain root?" answers YES for any account and proves nothing. What the
    # lab actually controls is which OUs the bind account holds an EXPLICIT
    # ACE on, and that is what is checked — exactly one explicit GenericRead
    # Allow on each read OU, and zero explicit ACEs anywhere else in the
    # anchor OU or the child OUs.
    #
    # The DOMAIN ROOT is the exception, and it is the W6 assertion: while
    # `excessive_idp_directory_permissions` is on the account holds exactly
    # two explicit extended-right ACEs there (the DCSync pair), and while it
    # is off it holds none. Asserting the count rather than merely allowing
    # "up to two" is what makes a half-restored baseline fail.
    #
    # Then the functional half: bind as the account and actually read.
```

Replace the `$scope`/`$allowed`/loop block inside the script with:

```powershell
          $scope   = @({% for dn in ad_scope_dns %}'{{ dn }}'{% if not loop.last %}, {% endif %}{% endfor %})
          $allowed = @({% for dn in ad_read_ous_dn %}'{{ dn }}'{% if not loop.last %}, {% endif %}{% endfor %})
          $root    = '{{ ad_base_dn }}'
          $rootWant = if ([bool]${{ lab_ad_weak.excessive_idp_directory_permissions | bool | ternary('true', 'false') }}) { 2 } else { 0 }

          foreach ($dn in $scope) {
              $explicit = @((Get-Acl -Path "AD:\$dn").Access |
                              Where-Object { Test-IsBindAce $_ })
              if ($allowed -contains $dn) {
                  if ($explicit.Count -ne 1 -or
                      $explicit[0].AccessControlType -ne 'Allow' -or
                      $explicit[0].ActiveDirectoryRights -notmatch 'GenericRead') {
                      throw "$dn should carry exactly one explicit GenericRead Allow ACE for the bind account, found $($explicit.Count)"
                  }
              } elseif ($dn -eq $root) {
                  if ($explicit.Count -ne $rootWant) {
                      throw "$dn carries $($explicit.Count) explicit ACE(s) for the bind account; the declared weakness W6 says $rootWant"
                  }
              } elseif ($explicit.Count -ne 0) {
                  throw "$dn carries $($explicit.Count) explicit ACE(s) for the bind account — the delegation is too broad, and it is not the domain-root grant W6 declares"
              }
          }
```

- [ ] **Step 2: Run it with the toggle on and see it fail**

```bash
sed -i 's/^    excessive_idp_directory_permissions: false$/    excessive_idp_directory_permissions: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: FAIL on "Assert the IDP delegation is scoped" with `DC=clayface,DC=local
carries 0 explicit ACE(s) for the bind account; the declared weakness W6 says 2`.

- [ ] **Step 3: Wire the toggle into `ad.yml`**

Add after "Grant the IDP bind account read-only access to its scoped OUs":

```yaml
    # --- DELIBERATE WEAKNESS: W6, excessive_idp_directory_permissions ---
    #
    # The two extended rights that together are DCSync, granted on the DOMAIN
    # ROOT: Replicating Directory Changes and Replicating Directory Changes
    # All. A principal holding both can ask a domain controller to replicate
    # the password hashes over DRSUAPI — including krbtgt's, which is a
    # Domain Admin by way of a forged golden ticket. This is the escalation
    # half of the Chapter 5 chain.
    #
    # The rights are matched by their schema GUIDs rather than by name because
    # that is what the ACE stores and what a reader of the ACL has to look
    # for. Both are needed: the first alone replicates non-secret attributes.
    #
    # Why a real organization has this: directory-sync products (Azure AD
    # Connect, backup agents, monitoring) genuinely need it, the vendor's
    # install guide says to grant it, and nobody narrows it afterwards.
    #
    # The flag-off branch builds the SAME rule object and removes it with
    # RemoveAccessRuleSpecific, which requires an exact match — including
    # inheritance flags — so it cannot strip an unrelated delegation.
    # RemoveAccessRule is the looser variant and deliberately not used.
    #
    # This is deliberately NOT scoped to the read OUs the way the ACE above
    # is. Replication rights are a property of the naming context, not of a
    # subtree, so a grant that looked scoped would not replicate anything and
    # would be a weakness in appearance only.
    - name: Grant the IDP bind account the DCSync replication rights
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory
          Add-Type -AssemblyName System.DirectoryServices

          $who  = '{{ lab_ad.idp.bind_account }}'
          $sid  = (New-Object System.Security.Principal.NTAccount(
                    '{{ lab_ad_netbios }}', $who)
                 ).Translate([System.Security.Principal.SecurityIdentifier])
          $want = [bool]${{ lab_ad_weak.excessive_idp_directory_permissions | bool | ternary('true', 'false') }}

          # Replicating Directory Changes (1131f6aa-...) and Replicating
          # Directory Changes All (1131f6ad-...).
          $guids = @(
              [Guid]'1131f6aa-9c07-11d1-f79f-00c04fc2dcd2',
              [Guid]'1131f6ad-9c07-11d1-f79f-00c04fc2dcd2'
          )

          $acl  = Get-Acl -Path 'AD:\{{ ad_base_dn }}'
          $mine = @($acl.Access | Where-Object {
              if ($_.IsInherited) { return $false }
              try {
                  $_.IdentityReference.Translate(
                      [System.Security.Principal.SecurityIdentifier]).Value -eq $sid.Value
              } catch { return $false }
          })
          $changed = @()

          foreach ($g in $guids) {
              $has = @($mine | Where-Object {
                  $_.AccessControlType -eq 'Allow' -and
                  $_.ActiveDirectoryRights -match 'ExtendedRight' -and
                  $_.ObjectType -eq $g
              }).Count -gt 0
              if ($has -eq $want) { continue }

              $rule = New-Object System.DirectoryServices.ActiveDirectoryAccessRule(
                  $sid,
                  [System.DirectoryServices.ActiveDirectoryRights]::ExtendedRight,
                  [System.Security.AccessControl.AccessControlType]::Allow,
                  $g)
              if ($want) { $acl.AddAccessRule($rule) }
              else       { $acl.RemoveAccessRuleSpecific($rule) | Out-Null }
              $changed += "$g=$want"
          }

          if ($changed.Count -eq 0) { 'OK' }
          else {
              Set-Acl -Path 'AD:\{{ ad_base_dn }}' -AclObject $acl
              "CHANGED: $($changed -join ', ')"
          }
      register: ad_dcsync
      changed_when: "'CHANGED' in (ad_dcsync.output | join(' '))"
```

- [ ] **Step 4: Run `ad.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: 1131f6aa-...=True, 1131f6ad-...=True`, then the validator
passes with `OK: explicit ACEs only on the 2 scoped OUs, no group membership`.

- [ ] **Step 5: Re-run `ad.yml` and prove idempotence**

Run: the same two commands again.

Expected: the DCSync task reports `ok`. Note the reason it must: the ACE's
`ObjectType` is what the `$has` check compares, and the read is the idempotency
key.

- [ ] **Step 6: Flip the toggle off and prove both ACEs are removed**

```bash
sed -i 's/^    excessive_idp_directory_permissions: true$/    excessive_idp_directory_permissions: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: 1131f6aa-...=False, 1131f6ad-...=False`, then the validator
passes with the domain root carrying 0 explicit ACEs. If the validator still
throws, `RemoveAccessRuleSpecific` did not match the ACE it was given — read the
live ACL with `(Get-Acl 'AD:\DC=clayface,DC=local').Access | Where-Object {
-not $_.IsInherited }` and compare the inheritance flags before changing
anything.

- [ ] **Step 7: Commit**

```bash
git add lab.yaml ansible/playbooks/ad.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W6, the DCSync replication rights

svc-idp-ldap holds Replicating Directory Changes and Replicating
Directory Changes All on the domain root while the toggle is on, and both
explicit ACEs are removed when it is off. The scoping assertion now
expects a count per posture at the domain root instead of always zero."
```

---

### Task 5: W2 — every employee is a local administrator

**Files:**
- Modify: `ansible/playbooks/client.yml:226-231`
- Modify: `ansible/playbooks/ad_validate.yml` ("Assert the local Administrators group is the designed set")

**Interfaces:**
- Consumes: `lab_ad_weak.excessive_local_admin` from Task 1.

- [ ] **Step 1: Write the failing assertion**

In `ansible/playbooks/ad_validate.yml`, in the `vms_client` play's "Assert the
local Administrators group is the designed set" task, add after the
built-in-Administrator check and before `$members -join ', '`:

```powershell
          # --- DELIBERATE WEAKNESS: W2, excessive_local_admin ---
          #
          # The logon group's presence here IS the weakness, so the assertion is
          # symmetric: present while the flag is on, absent while it is off.
          # Asserting only the on-direction would let a half-restored baseline
          # pass, which is the failure this whole phase has to avoid.
          $weak     = '{{ lab_ad_netbios }}\{{ lab_ad.workstation.logon_group }}'
          $wantWeak = [bool]${{ lab_ad_weak.excessive_local_admin | bool | ternary('true', 'false') }}
          $hasWeak  = $members -contains $weak
          if ($hasWeak -ne $wantWeak) {
              throw "$weak local administrator = $hasWeak, the declared weakness W2 says $wantWeak"
          }
```

- [ ] **Step 2: Run it with the toggle on and see it fail**

```bash
sed -i 's/^    excessive_local_admin: false$/    excessive_local_admin: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: FAIL with `CLAYFACE\GG-Employees local administrator = False, the
declared weakness W2 says True`.

If it instead reports `skipping: no hosts matched` for the workstation play,
`client01` is still commented out of `lab.yaml` — uncomment it, run
`./deploy.sh --skip-ansible` (or `terraform apply`), then `client.yml`, and
re-run. A validator that passes because its host set is empty is the failure
mode this step exists to surface.

- [ ] **Step 3: Wire the toggle into `client.yml`**

In `ansible/playbooks/client.yml`, after "Grant the workstation admin group
local Administrator rights" (lines 226–231), add:

```yaml
    # --- DELIBERATE WEAKNESS: W2, excessive_local_admin ---
    #
    # With the flag on, every employee is a local administrator on this
    # workstation — the whole logon group, which nests both role groups. That
    # is the ordinary shape of this finding in the field: someone needed to
    # install something once, and adding the group was easier than adding the
    # person.
    #
    # The flag-off branch is `state: absent`, NOT a skipped task.
    # win_group_membership is additive, so a `when:` guard would leave the
    # group in local Administrators forever and the baseline would never come
    # back. The task above carries the designed set unconditionally, so
    # GG-IT-Admins and the built-in Administrator survive in both postures;
    # this task only ever touches the logon group.
    #
    # This is the local-privilege branch of the chain rather than its
    # escalation: local admin is what makes an LSASS dump and the planted
    # shares on this host reachable. Domain Admin comes from W1 and W6.
    - name: Grant the logon group local Administrator rights
      ansible.windows.win_group_membership:
        name: Administrators
        members:
          - "{{ lab_ad_netbios }}\\{{ lab_ad.workstation.logon_group }}"
        state: "{{ 'present' if lab_ad_weak.excessive_local_admin | bool else 'absent' }}"
```

- [ ] **Step 4: Run `client.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/client.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `client.yml` reports the task `changed`; the validator passes and
prints the local Administrators membership including
`CLAYFACE\GG-Employees`.

- [ ] **Step 5: Flip the toggle off and prove the group is removed**

```bash
sed -i 's/^    excessive_local_admin: true$/    excessive_local_admin: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/client.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: the task reports `changed` (it removed the group); the validator
passes with `CLAYFACE\GG-Employees` gone and `CLAYFACE\GG-IT-Admins` and the
built-in Administrator still present.

- [ ] **Step 6: Prove a rebuilt workstation converges**

A workstation rebuilt from the base image has neither the designed set nor the
weakness, so both branches must be reachable from a clean host. Simulate the
starting state by removing only the weakness's principal, then re-run with the
flag on:

```bash
# on the DC, for the record; the automation is what actually has to converge
ansible -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
        vms_client -m ansible.windows.win_shell \
        -a 'Remove-LocalGroupMember -Group Administrators -Member CLAYFACE\GG-Employees -ErrorAction SilentlyContinue; (Get-LocalGroupMember Administrators).Name'
sed -i 's/^    excessive_local_admin: false$/    excessive_local_admin: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/client.yml
```

Expected: the group is back after `client.yml` runs. Leave the toggle on.

- [ ] **Step 7: Commit**

```bash
git add lab.yaml ansible/playbooks/client.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W2, excessive local admin on the workstation

GG-Employees joins the local Administrators group while the toggle is on
and is removed when it is off, because win_group_membership is additive
and a skipped task would never restore the baseline."
```

---

### Task 6: W3 — the logon group can read LAPS passwords

LAPS in this lab has two independent gates: the read-password right on
`OU=Workstations`, and membership of the decryption principal named in the LAPS
policy (`ADPasswordEncryptionPrincipal = GG-IT-Admins`). W3 opens the first one,
and that is enough — a principal holding the read right reads the plaintext
through the LAPS cmdlet. The asymmetry is the finding: the decryption principal
is configured once in a GPO and rarely revisited, while the read ACE on the OU
is what an administrator hands out.

**Files:**
- Modify: `ansible/playbooks/ad.yml` (new task after "Let the workstation admins read LAPS passwords on OU=Workstations")
- Modify: `ansible/playbooks/ad_validate.yml` (new assertion in the `vms_dc` play)

**Interfaces:**
- Consumes: `lab_ad_weak.excessive_laps_read` from Task 1.

- [ ] **Step 1: Write the failing assertion**

Add to `ansible/playbooks/ad_validate.yml` in the `vms_dc` play, after "Assert
the LAPS schema is extended":

```yaml
    # --- DELIBERATE WEAKNESS: W3, excessive_laps_read ---
    #
    # Asserted on the ACL rather than by reading a password: the read ACE on
    # OU=Workstations is the grant W3 makes, and the other gate (the
    # decryption principal in the LAPS GPO) is unchanged by this weakness.
    # Find-LapsADExtendedRights is the cmdlet-shaped answer, but it reports
    # trustees as display names and several of the lab's principals do not
    # resolve consistently that way; the SID comparison is exact.
    - name: Assert the LAPS read delegation matches the declared weakness
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory

          $ou   = '{{ ad_ou_dn.Workstations }}'
          $sid  = (New-Object System.Security.Principal.NTAccount(
                     '{{ lab_ad_netbios }}', '{{ lab_ad.workstation.logon_group }}')
                  ).Translate([System.Security.Principal.SecurityIdentifier])
          $want = [bool]${{ lab_ad_weak.excessive_laps_read | bool | ternary('true', 'false') }}

          $has = @((Get-Acl -Path "AD:\$ou").Access | Where-Object {
              if ($_.IsInherited) { return $false }
              try {
                  $_.IdentityReference.Translate(
                      [System.Security.Principal.SecurityIdentifier]).Value -eq $sid.Value
              } catch { return $false }
          }).Count -gt 0

          if ($has -ne $want) {
              throw "{{ lab_ad.workstation.logon_group }} holds $($has) explicit ACE(s) on $ou; the declared weakness W3 says $want"
          }

          # The baseline ACE must survive in both postures: removing it would
          # take LAPS away from the people whose job it is.
          $admin = (New-Object System.Security.Principal.NTAccount(
                      '{{ lab_ad_netbios }}', '{{ lab_ad.workstation.admin_group }}')
                   ).Translate([System.Security.Principal.SecurityIdentifier])
          $adminHas = @((Get-Acl -Path "AD:\$ou").Access | Where-Object {
              if ($_.IsInherited) { return $false }
              try {
                  $_.IdentityReference.Translate(
                      [System.Security.Principal.SecurityIdentifier]).Value -eq $admin.Value
              } catch { return $false }
          }).Count -gt 0
          if (-not $adminHas) {
              throw "{{ lab_ad.workstation.admin_group }} lost its LAPS read ACE on $ou — the baseline was removed"
          }

          "OK: LAPS read for {{ lab_ad.workstation.logon_group }} present=$has, baseline ACE intact"
      changed_when: false
```

- [ ] **Step 2: Run it with the toggle on and see it fail**

```bash
sed -i 's/^    excessive_laps_read: false$/    excessive_laps_read: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: FAIL with `GG-Employees holds False explicit ACE(s) on
OU=Workstations,OU=Clayface,DC=clayface,DC=local; the declared weakness W3 says
True`.

- [ ] **Step 3: Wire the toggle into `ad.yml`**

Add after "Let the workstation admins read LAPS passwords on OU=Workstations":

```yaml
    # --- DELIBERATE WEAKNESS: W3, excessive_laps_read ---
    #
    # Every employee can read the local administrator password LAPS manages on
    # every workstation. LAPS has two gates and this opens the first: the read
    # right on OU=Workstations. That is enough, because a principal holding the
    # read right reads the plaintext through the LAPS cmdlet — the second gate
    # (the decryption principal in the LAPS GPO) only matters to someone
    # reading the encrypted attribute directly.
    #
    # Why a real organization has this: the read right is what an administrator
    # hands out when helpdesk asks to "look up a workstation password", and it
    # is on an OU rather than on a person, so nobody sees it again.
    #
    # There is no Remove-LapsADReadPasswordPermission cmdlet, so the flag-off
    # branch goes through the ACL directly — the same Get-Acl/Set-Acl idiom the
    # IDP delegation task below uses. Only NON-INHERITED ACEs for this exact
    # SID are removed, which is what leaves the GG-IT-Admins ACE the baseline
    # grants on the same OU untouched. RemoveAccessRuleSpecific demands an
    # exact match; the ACE objects are read from the live ACL, so they are.
    - name: Let the logon group read LAPS passwords on OU=Workstations
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory
          Add-Type -AssemblyName System.DirectoryServices

          $ou   = '{{ ad_ou_dn.Workstations }}'
          $who  = '{{ lab_ad_netbios }}\{{ lab_ad.workstation.logon_group }}'
          $sid  = (New-Object System.Security.Principal.NTAccount(
                     '{{ lab_ad_netbios }}', '{{ lab_ad.workstation.logon_group }}')
                  ).Translate([System.Security.Principal.SecurityIdentifier])
          $want = [bool]${{ lab_ad_weak.excessive_laps_read | bool | ternary('true', 'false') }}

          $acl  = Get-Acl -Path "AD:\$ou"
          $mine = @($acl.Access | Where-Object {
              if ($_.IsInherited) { return $false }
              try {
                  $_.IdentityReference.Translate(
                      [System.Security.Principal.SecurityIdentifier]).Value -eq $sid.Value
              } catch { return $false }
          })

          if ($want) {
              if ($mine.Count -gt 0) { 'OK'; return }
              Set-LapsADReadPasswordPermission -Identity $ou -AllowedPrincipals $who -Confirm:$false
              'CHANGED: granted'
          } else {
              if ($mine.Count -eq 0) { 'OK'; return }
              foreach ($ace in $mine) { $acl.RemoveAccessRuleSpecific($ace) | Out-Null }
              Set-Acl -Path "AD:\$ou" -AclObject $acl
              "CHANGED: removed $($mine.Count) ACE(s)"
          }
      register: laps_read_weak
      changed_when: "'CHANGED' in (laps_read_weak.output | join(' '))"
```

- [ ] **Step 4: Run `ad.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: granted`, then the validator passes with
`OK: LAPS read for GG-Employees present=True, baseline ACE intact`.

- [ ] **Step 5: Re-run `ad.yml`, then flip off and prove removal**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml   # expect: ok, not CHANGED
sed -i 's/^    excessive_laps_read: true$/    excessive_laps_read: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml   # expect: CHANGED: removed 1 ACE(s)
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml  # expect: present=False, baseline ACE intact
```

The `baseline ACE intact` line is the one that matters here: a removal that
matched too loosely would take `GG-IT-Admins`'s ACE with it, and the assertion
above is what fails when that happens.

- [ ] **Step 6: Commit**

```bash
git add lab.yaml ansible/playbooks/ad.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W3, excessive LAPS read

GG-Employees can read the LAPS-managed local administrator password on
OU=Workstations while the toggle is on. Removal goes through the ACL
because the LAPS module has no revoke cmdlet, and the assertion also
proves the GG-IT-Admins baseline ACE survived."
```

---

### Task 7: W5 — unconstrained delegation on the IDP bind account

**Files:**
- Modify: `ansible/playbooks/ad.yml` (new task after the DCSync task)
- Modify: `ansible/playbooks/ad_validate.yml` (new assertion in the `vms_dc` play)

**Interfaces:**
- Consumes: `lab_ad_weak.unconstrained_delegation` from Task 1.

- [ ] **Step 1: Write the failing assertion**

Add to `ansible/playbooks/ad_validate.yml` in the `vms_dc` play:

```yaml
    # --- DELIBERATE WEAKNESS: W5, unconstrained_delegation ---
    #
    # Asserted on userAccountControl bit 524288 (TRUSTED_FOR_DELEGATION). The
    # negative half searches USER objects only: every domain controller is
    # trusted for unconstrained delegation as part of being a DC, so a filter
    # that included computer accounts would assert against the forest's normal
    # shape and fail on a healthy lab.
    #
    # Note that this is NOT the load-bearing step of the Chapter 5 chain — that
    # is W1 then W6, and it is reachable from a LAN foothold because the DMZ
    # boundary permits tcp 389/636 to dc01 and nothing else. W5 is wired,
    # asserted and offered as an adjacent path; docs/planted-weaknesses.md
    # states that scoping decision rather than implying the chain needs it.
    - name: Assert the delegation trust matches the declared weakness
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory

          $who  = '{{ lab_ad.idp.bind_account }}'
          $want = [bool]${{ lab_ad_weak.unconstrained_delegation | bool | ternary('true', 'false') }}

          $uac = (Get-ADUser $who -Properties userAccountControl).userAccountControl
          $has = [bool]($uac -band 524288)
          if ($has -ne $want) {
              throw "$who TRUSTED_FOR_DELEGATION is $has, the declared weakness W5 says $want"
          }

          $others = @(Get-ADUser -LDAPFilter '(userAccountControl:1.2.840.113556.1.4.803:=524288)' `
                       -Properties userAccountControl |
                       Where-Object { $_.SamAccountName -ne $who -and
                                      $_.SamAccountName -ne 'krbtgt' } |
                       Select-Object -ExpandProperty SamAccountName)
          if ($others.Count) {
              throw "undeclared delegated user accounts: $($others -join ', ')"
          }
          "OK: $who TRUSTED_FOR_DELEGATION=$has, no other delegated user accounts"
      changed_when: false
```

- [ ] **Step 2: Run it with the toggle on and see it fail**

```bash
sed -i 's/^    unconstrained_delegation: false$/    unconstrained_delegation: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: FAIL with `svc-idp-ldap TRUSTED_FOR_DELEGATION is False, the declared
weakness W5 says True`.

- [ ] **Step 3: Wire the toggle into `ad.yml`**

Add after the DCSync task:

```yaml
    # --- DELIBERATE WEAKNESS: W5, unconstrained_delegation ---
    #
    # TrustedForDelegation is userAccountControl bit 524288: an account trusted
    # for UNCONSTRAINED delegation that authenticates to a host hands that host
    # a forwardable TGT, and anything running there can capture it and
    # impersonate the account anywhere in the forest. The classic exploitation
    # pairs it with a coercion primitive against the DC, which is why the
    # Chapter 4 write-up names the constrained-delegation and RBCD
    # replacements as the fix rather than a narrower ACL.
    #
    # Why a real organization has this: it used to be the documented way to
    # make a multi-hop Kerberos scenario work, an installer set it, and every
    # later review reads the account as "expected".
    #
    # The unconstrained form was chosen over RBCD on client01 because it is
    # one attribute on an account this lab already owns and because it does
    # not require a second machine account. Section 13 lists both as the
    # weakness's shapes.
    #
    # Set-ADAccountControl is idempotent by nature, but the read is still the
    # guard so the task reports honestly whether anything moved.
    - name: Set the delegation trust on the IDP bind account
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module ActiveDirectory

          $who  = '{{ lab_ad.idp.bind_account }}'
          $want = [bool]${{ lab_ad_weak.unconstrained_delegation | bool | ternary('true', 'false') }}

          $uac = (Get-ADUser $who -Properties userAccountControl).userAccountControl
          $has = [bool]($uac -band 524288)
          if ($has -eq $want) { 'OK'; return }

          Set-ADAccountControl -Identity $who -TrustedForDelegation $want
          "CHANGED: $who TRUSTED_FOR_DELEGATION -> $want"
      register: ad_delegation
      changed_when: "'CHANGED' in (ad_delegation.output | join(' '))"
```

- [ ] **Step 4: Run `ad.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: svc-idp-ldap TRUSTED_FOR_DELEGATION -> True`, then
`OK: svc-idp-ldap TRUSTED_FOR_DELEGATION=True, no other delegated user accounts`.

- [ ] **Step 5: Flip off, run both, prove the bit is cleared**

```bash
sed -i 's/^    unconstrained_delegation: true$/    unconstrained_delegation: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: svc-idp-ldap TRUSTED_FOR_DELEGATION -> False`, then the
validator passes with `=False`.

- [ ] **Step 6: Commit**

```bash
git add lab.yaml ansible/playbooks/ad.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W5, unconstrained delegation on the bind account

TRUSTED_FOR_DELEGATION is set and cleared with the toggle. The negative
assertion searches user objects only, because every DC carries the bit as
part of being a DC."
```

---

### Task 8: W4 — the logon group can edit the workstation GPO

**Files:**
- Modify: `ansible/playbooks/ad_gpo.yml` (add the task after "Apply the LAPS policy to the LAPS GPO"; add a `pre_tasks` block if Task 1 has not already added one)
- Modify: `ansible/playbooks/ad_validate.yml` (new assertion in the `vms_dc` play)

**Interfaces:**
- Consumes: `lab_ad_weak.weak_gpo_permission` from Task 1; the GPO name
  `gpo_baseline` (`WS - Security Baseline`), already a var in `ad_gpo.yml`, and
  the DN `OU=Workstations,...` which `ad_validate.yml` already builds as part of
  `ad_scope_dns`.

- [ ] **Step 1: Write the failing assertion**

Add to `ansible/playbooks/ad_validate.yml` in the `vms_dc` play:

```yaml
    # --- DELIBERATE WEAKNESS: W4, weak_gpo_permission ---
    #
    # GpoEditDeleteModifySecurity is the permission that matters: it lets the
    # holder change the GPO's settings AND its security filtering, so the
    # holder can grant themselves anything the GPO applies and can edit the
    # ACL that decides who the policy reaches. GpoEdit (settings only, no ACL)
    # would be the softer shape and is deliberately not the one planted.
    #
    # ad_gpo.yml runs before this playbook in deploy.sh, so the GPO exists by
    # the time this asserts. If the order ever changes, this task fails on a
    # missing GPO rather than on a permission — a distinct message.
    - name: Assert the GPO editing permission matches the declared weakness
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module GroupPolicy

          $gpo  = 'WS - Security Baseline'
          $who  = '{{ lab_ad.workstation.logon_group }}'
          $want = [bool]${{ lab_ad_weak.weak_gpo_permission | bool | ternary('true', 'false') }}

          if (-not (Get-GPO -Name $gpo -ErrorAction SilentlyContinue)) {
              throw "the GPO '$gpo' does not exist — run playbooks/ad_gpo.yml first"
          }

          $cur = Get-GPPermission -Name $gpo -TargetName $who -TargetType Group `
                   -ErrorAction SilentlyContinue
          $has = $false
          if ($null -ne $cur -and "$($cur.Permission)" -eq 'GpoEditDeleteModifySecurity') { $has = $true }
          if ($has -ne $want) {
              throw "$who holds $(if ($null -eq $cur) { 'no' } else { $cur.Permission }) on '$gpo'; the declared weakness W4 says $want"
          }

          # Domain Admins must retain the edit permission in both postures —
          # the weakness is the SECOND holder, not a replacement.
          $da = Get-GPPermission -Name $gpo -TargetName 'Domain Admins' -TargetType Group `
                  -ErrorAction SilentlyContinue
          if ($null -eq $da -or "$($da.Permission)" -notmatch 'GpoEdit') {
              throw "'Domain Admins' lost its edit permission on '$gpo' — the baseline was removed"
          }

          "OK: $who on '$gpo' is $has, Domain Admins baseline intact"
      changed_when: false
```

- [ ] **Step 2: Run it with the toggle on and see it fail**

```bash
sed -i 's/^    weak_gpo_permission: false$/    weak_gpo_permission: true/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: FAIL with `GG-Employees holds no on 'WS - Security Baseline'; the
declared weakness W4 says True`.

- [ ] **Step 3: Wire the toggle into `ad_gpo.yml`**

Add to the `vars:` block in `ansible/playbooks/ad_gpo.yml`, after
`gpo_workstations_dn`:

```yaml
    # The AD weakness toggles, from lab.yaml `ad.weaknesses`. See ad.yml for
    # the shared contract; this play owns exactly one of the seven.
    lab_ad_weak: "{{ (lab_ad | default({}, true)).weaknesses | default({}, true) }}"
```

Add as the last task in the play:

```yaml
    # --- DELIBERATE WEAKNESS: W4, weak_gpo_permission ---
    #
    # The all-employees group may edit the workstation security baseline — its
    # settings AND its security filtering. Because the GPO is linked to
    # OU=Workstations and applies to every workstation in the lab, the holder
    # can push an arbitrary startup script or a registry change to every
    # workstation at once, and can edit the ACL that decides who the policy
    # reaches. That is why this is a Domain-Admin-equivalent step rather than a
    # workstation-scoped one.
    #
    # Why a real organization has this: delegated GPO administration is a
    # documented practice, and the delegation is normally made to a role group
    # that has since grown to include everyone.
    #
    # PermissionLevel None removes the trustee's entry outright, which is what
    # makes the flag-off branch an active removal: the entry an earlier run
    # created is gone afterwards. The read guard means the None path is only
    # ever reached when an entry exists, so this cannot create a None entry on
    # a group that never had one.
    #
    # Domain Admins and the default SYSTEM/Enterprise Domain Controllers
    # entries are never touched: Set-GPPermission is called for one target,
    # the logon group, and for nothing else.
    - name: Set the GPO editing permission the logon group holds
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          Import-Module GroupPolicy

          $gpo   = '{{ gpo_baseline }}'
          $who   = '{{ lab_ad.workstation.logon_group }}'
          $want  = [bool]${{ lab_ad_weak.weak_gpo_permission | bool | ternary('true', 'false') }}
          $level = if ($want) { 'GpoEditDeleteModifySecurity' } else { 'None' }

          $cur = Get-GPPermission -Name $gpo -TargetName $who -TargetType Group `
                   -ErrorAction SilentlyContinue
          $has = $false
          if ($null -ne $cur -and "$($cur.Permission)" -eq 'GpoEditDeleteModifySecurity') { $has = $true }
          if ($has -eq $want) { 'OK'; return }

          Set-GPPermission -Name $gpo -TargetName $who -TargetType Group `
              -PermissionLevel $level
          "CHANGED: $who on '$gpo' -> $level"
      register: gpo_perm_weak
      changed_when: "'CHANGED' in (gpo_perm_weak.output | join(' '))"
```

- [ ] **Step 4: Run `ad_gpo.yml` with the toggle on, then the validator**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_gpo.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: GG-Employees on 'WS - Security Baseline' ->
GpoEditDeleteModifySecurity`, then the validator passes with
`OK: GG-Employees on 'WS - Security Baseline' is True, Domain Admins baseline
intact`.

- [ ] **Step 5: Flip off, run both, prove the entry is gone**

```bash
sed -i 's/^    weak_gpo_permission: true$/    weak_gpo_permission: false/' lab.yaml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_gpo.yml
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py ansible/playbooks/ad_validate.yml
```

Expected: `CHANGED: GG-Employees on 'WS - Security Baseline' -> None`, then the
validator passes with `is False`.

- [ ] **Step 6: Commit**

```bash
git add lab.yaml ansible/playbooks/ad_gpo.yml ansible/playbooks/ad_validate.yml
git commit -m "feat(lab): wire W4, weak GPO permission

GG-Employees may edit the workstation baseline GPO, settings and security
filtering, while the toggle is on; PermissionLevel None removes the entry
when it is off. The assertion also proves Domain Admins kept theirs."
```

---

### Task 9: Plant the exfiltration subject

The Chapter 5 impact analysis needs something to steal. APP01's seed documents
are the portal's own data; this task adds the corporate files on the two Windows
hosts that make the last step of the engagement a real one.

**Files:**
- Modify: `lab.yaml` (new top-level `data:` map)
- Create: `ansible/playbooks/lab_data.yml`
- Modify: `deploy.sh` (run `lab_data.yml` after `client.yml`)

**Interfaces:**
- Consumes: the `vms_windows` inventory group, which
  `ansible/inventory/terraform_vms.py` already builds.
- Produces: `\\<host>.clayface\<share>` for every share declared in
  `lab.yaml`, which the Chapter 5 evidence captures from a LAN foothold.

- [ ] **Step 1: Declare the shares in `lab.yaml`**

Add after the `app:` block:

```yaml
# Exfiltration targets for the Chapter 5 impact analysis (roadmap Phase 4).
#
# The engagement's last step is that the attacker takes something, and an
# impact claim with nothing to take is an assertion rather than a result. These
# are fake but realistic business files, planted beyond APP01's own seed.
#
# Placement is a static fact about the lab and lives here; the file CONTENT is
# prose and lives in playbooks/lab_data.yml, which is the same split app01
# uses — paths in lab.yaml, seed material beside the code that plants it.
#
# Shares are reachable from the LAN only. The baseline workstation GPO blocks
# inbound by default except for WinRM from 10.0.0.0/24, and dc01 is not in
# OU=Workstations so the GPO does not apply to it — the DC's own AD DS firewall
# rules permit 445, which is what makes \\dc01.clayface reachable from a
# workstation and a share on a workstation reachable only by someone already
# logged on to it.
data:
  shares:
    dc01:
      - name: Finance
        path: C:\Shares\Finance
        description: Finance department working files
    client01:
      - name: Projects
        path: C:\Shares\Projects
        description: Developer project files
```

- [ ] **Step 2: Write the failing assertion — a playbook that does nothing yet**

Create `ansible/playbooks/lab_data.yml` with only its pre-tasks and the
assertion, no planting tasks. The assertion is what fails first:

```yaml
---
# The lab's planted data: the corporate files the Chapter 5 impact analysis
# exfiltrates. Driven entirely by lab.yaml's `data.shares` map, so adding a
# share is one lab.yaml entry and no playbook edit.
#
# Run after client.yml: the shares live on the joined workstations and on the
# domain controller, and the DC's own share is what the attacker reaches after
# escalation.
#
# What it does (idempotent):
#   1. creates each declared share's directory
#   2. writes the documents for that share, from the content map below
#   3. creates the SMB share with the declared path
#   4. asserts the share exists and the documents are readable
#
# Run:
#   ansible-playbook -i inventory/lab_inventory.py \
#                    -i inventory/terraform_vms.py playbooks/lab_data.yml

- name: Plant the lab's share data
  hosts: vms_windows
  gather_facts: false

  vars:
    lab_data_shares: "{{ (lab_data | default({}, true)).shares | default({}, true) }}"

  pre_tasks:
    - name: Fail if lab.yaml has no usable `data.shares` map
      ansible.builtin.assert:
        that: lab_data_shares | length > 0
        fail_msg: >-
          lab.yaml has no `data.shares` map, so there is nothing to plant.
          The exfiltration subject belongs in lab.yaml: it is a static fact
          about the lab, not a property of this playbook.

    - name: Note hosts with no declared shares
      ansible.builtin.debug:
        msg: >-
          {{ inventory_hostname }} has no entry under lab.yaml `data.shares`,
          so nothing is planted here. That is legitimate (not every VM holds
          data) but it is worth seeing, because a typo'd host key looks
          exactly like this.
      when: lab_data_shares[inventory_hostname] is not defined

    - name: Note shares declared for hosts that are not in the inventory
      ansible.builtin.debug:
        msg: >-
          lab.yaml `data.shares` names {{ item }} but no such host is in the
          inventory. Is it commented out of vm_placements, or did terraform
          apply not run?
      loop: "{{ lab_data_shares.keys() | list }}"
      when: item not in groups['vms_windows'] | default([])

  tasks:
    - name: Assert the declared shares exist with their documents
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          $bad = @()
          {% for host, shares in {} | combine({}) %}{% endfor %}
          foreach ($s in @({% for s in lab_data_shares.get(inventory_hostname, []) %}'{{ s.name }}'{% if not loop.last %}, {% endif %}{% endfor %})) {
              if (-not (Get-SmbShare -Name $s -ErrorAction SilentlyContinue)) {
                  $bad += "share $s is missing"
              }
          }
          if ($bad.Count) { throw ($bad -join '; ') }
          'OK'
      changed_when: false
```

The `{% for host, shares in {} | combine({}) %}{% endfor %}` line above is dead
scaffolding — delete it in the same edit as step 3; it is written here only so
the failing run has a syntactically valid file. (Verify that it parses:
`ansible-playbook --syntax-check`.)

- [ ] **Step 3: Run it and see it fail**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/lab_data.yml
```

Expected: FAIL with `share Finance is missing` on `dc01`.
(If it reports `skipping: no hosts matched`, no Windows VM is in the inventory —
`dc01` must be enabled in `lab.yaml` and terraform applied.)

- [ ] **Step 4: Write the planting tasks**

Replace the placeholder assertion in `ansible/playbooks/lab_data.yml`'s `tasks:`
with the real implementation:

```yaml
  tasks:
    # The content map. Keys are the share names declared in lab.yaml, so a
    # share with no entry here plants an empty directory — which is a
    # legitimate outcome (the directory is the finding) but is reported.
    lab_data_documents:
      Finance:
        - name: 2026-q3-forecast.csv
          content: |
            period,account,amount_eur,confidence
            2026-Q3,Consulting,412500,high
            2026-Q3,Licensing,1180000,high
            2026-Q3,Support,264300,medium
            2026-Q3,Hardware,91800,low
        - name: headcount-plan.txt
          content: |
            Clayface Industries — FY2026 headcount plan (DRAFT, do not circulate)

            Engineering      42 -> 51
            Sales            18 -> 24
            Finance           7 ->  7
            IT                4 ->  6

            Backfill approved for two SRE roles after the March incident.
            Offer band for senior engineers: 78k-92k. Not to be shared with
            candidates before the final round.
        - name: board-minutes-2026-08.txt
          content: |
            Clayface Industries — board minutes, 2026-08-14

            Present: chair, CEO, CFO, two non-executive directors.
            Apologies: CTO.

            1. The Q3 revenue shortfall against the licensing line was
               discussed. Management attributed it to two renewals slipping
               into Q4 and expects no revision to the full-year guidance.
            2. Legal confirmed the supplier dispute is expected to settle
               before the year end. Provision unchanged.
            3. The board approved the acquisition of a small consultancy
               subject to due diligence. Name withheld in this minute.
      Projects:
        - name: portal-roadmap.md
          content: |
            # Portal roadmap

            ## Next
            - Move the customer portal behind the new identity provider
            - Retire the shared service account the batch jobs use
            - Certificate rotation, currently manual

            ## Known issues
            - The reporting query is slow on large accounts
            - Two customers still on the legacy export format
        - name: deployment-runbook.txt
          content: |
            Deployment runbook (internal)

            1. Build the image on the build host
            2. Copy the tarball to the application server
            3. docker compose up --detach
            4. Verify both published ports answer

            Credentials for the database are in the environment file on the
            application server. Ask the platform team before changing them —
            the reporting job uses the same account.

    - name: Create the declared share directories
      ansible.windows.win_file:
        path: "{{ item.path }}"
        state: directory
      loop: "{{ lab_data_shares[inventory_hostname] | default([]) }}"
      loop_control:
        label: "{{ item.path }}"

    - name: Plant this host's documents
      ansible.windows.win_copy:
        dest: "{{ (lab_data_shares[inventory_hostname] | selectattr('name', 'eq', item.0) | first).path }}/{{ item.1.name }}"
        content: "{{ item.1.content }}"
      loop: >-
        {{ lab_data_documents | dict2items
           | selectattr('key', 'in', lab_data_shares[inventory_hostname] | default([]) | map(attribute='name') | list)
           | subelements('value') | list }}
      loop_control:
        label: "{{ item.0.key }}/{{ item.1.name }}"

    - name: Create the SMB shares
      ansible.windows.win_share:
        name: "{{ item.name }}"
        path: "{{ item.path }}"
        description: "{{ item.description }}"
        full: Administrators
        read: "{{ lab_ad_netbios }}\\{{ lab_ad.workstation.logon_group }}"
      loop: "{{ lab_data_shares[inventory_hostname] | default([]) }}"
      loop_control:
        label: "{{ item.name }}"

    - name: Assert the declared shares exist and hold their documents
      ansible.windows.win_powershell:
        script: |
          $ErrorActionPreference = 'Stop'
          $bad = @()
          $dirs = @({% for s in lab_data_shares.get(inventory_hostname, []) %}'{{ s.name }}'{% if not loop.last %}, {% endif %}{% endfor %})
          foreach ($n in $dirs) {
              $s = Get-SmbShare -Name $n -ErrorAction SilentlyContinue
              if (-not $s) { $bad += "share $n is missing"; continue }
              $files = @(Get-ChildItem -LiteralPath $s.Path -File -ErrorAction SilentlyContinue)
              if ($files.Count -eq 0) { $bad += "share $n holds no documents" }
          }
          if ($bad.Count) { throw ($bad -join '; ') }
          "OK: $($dirs.Count) share(s) present and populated on {{ inventory_hostname }}"
      changed_when: false
```

Note that `win_share`'s `read` grant names the logon group: the planted data is
readable by employees by design (it is what makes it a plausible corporate
share), and `full: Administrators` keeps it manageable. This grant is not a
weakness and is not governed by a toggle.

- [ ] **Step 5: Run it and see it pass**

```bash
ansible-playbook -i ansible/inventory/lab_inventory.py \
                 -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/lab_data.yml
```

Expected: `OK: 1 share(s) present and populated on dc01` and the same for
`client01`.

- [ ] **Step 6: Re-run and prove idempotence**

Run the same command again.

Expected: all tasks `ok`, no `changed`.

- [ ] **Step 7: Verify reachability from the LAN**

From a LAN host (the control node, after `hosts.yml` has pointed the resolver at
OPNsense):

```bash
smbclient -L //dc01.clayface -N 2>&1 | head -20
```

Expected: the `Finance` share is listed. Record the output as evidence — the
standing rule is that evidence is captured when it is produced, not afterwards.
From `client01` itself, `Get-ChildItem \\dc01.clayface\Finance` must list the
three planted documents; capture that too, since it is the exfiltration step's
first command.

- [ ] **Step 8: Wire `lab_data.yml` into `deploy.sh`**

After the `client.yml` step (around line 199) and before the `app.yml` step:

```bash
step "Ansible: plant the lab's share data (lab_data.yml)"
ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/lab_inventory.py" \
    -i "$ANSIBLE_DIR/inventory/terraform_vms.py" \
    "$ANSIBLE_DIR/playbooks/lab_data.yml"
```

Match the surrounding style exactly, including how the other steps pass
`--ask-become-pass` or not (this playbook needs no become).

- [ ] **Step 9: Commit**

```bash
git add lab.yaml ansible/playbooks/lab_data.yml deploy.sh
git commit -m "feat(lab): plant the exfiltration subject on dc01 and client01

A Finance share on the domain controller and a Projects share on the
workstation, driven by a data.shares map in lab.yaml with the document
content beside the code that plants it. The Chapter 5 impact analysis
needs a subject, and APP01's own seed is the portal's data rather than
the corporation's."
```

---

### Task 10: The Chapter 4 planted-weakness section

**Files:**
- Create: `docs/planted-weaknesses.md`
- Modify: `CLAUDE.md` (the docs list, so the new document is discoverable)

**Interfaces:**
- Consumes: every toggle name and every assertion message from Tasks 2–8.
- Produces: the section Chapter 4 imports; the per-weakness contracts the
  thesis cites.

- [ ] **Step 1: Write the document**

Create `docs/planted-weaknesses.md`. The table below is the content; the prose
sections expand each row. Follow the five-field shape APP01's toggle table
uses (what it is, why a real organization would have it, which attack it
enables, the correct configuration, the detection signal), with the lab.yaml
toggle and the exercising step named for each.

Header and the full table:

```markdown
# Planted weaknesses — the AD identity layer

The design is `docs/ad-identity-design.md`; section 13 is the authoritative
normal-to-vulnerable table and this document is the Chapter 4 expansion of it.
Each weakness is one entry under `ad.weaknesses` in `lab.yaml`, applied by the
playbook that owns the object, and asserted in both postures by
`ansible/playbooks/ad_validate.yml`. Turning a toggle off is a real
configuration change that restores the baseline — not a skipped task.

| # | Toggle (`lab.yaml`) | Owner | Misconfiguration while `true` | Attack it enables | Detection |
| --- | --- | --- | --- | --- | --- |
| W1 | `kerberoastable_idp_bind` | `ad.yml` | `svc-idp-ldap` holds SPN `HTTP/idp01.clayface`; its password is the shared lab password and never rotates (`MaxPasswordAge 0`) | Request a TGS for the SPN as any authenticated principal and crack it offline | 4769 with RC4 (`0x17`) encryption for a service account |
| W2 | `excessive_local_admin` | `client.yml` | `GG-Employees` is a member of the local `Administrators` group on every workstation | Any employee — or anyone holding an employee credential — is a local administrator; LSASS dump, planted data reachable | 4732/4733 (member added to a security-enabled local group) |
| W3 | `excessive_laps_read` | `ad.yml` | `GG-Employees` holds the LAPS read-password right on `OU=Workstations` | Read the LAPS-managed local administrator password of any workstation | 4662 on the `msLAPS-Password` attribute |
| W4 | `weak_gpo_permission` | `ad_gpo.yml` | `GG-Employees` holds `GpoEditDeleteModifySecurity` on `WS - Security Baseline` | Edit the baseline's settings and security filtering — arbitrary configuration on every workstation at once | 5136 on the GPO's `nTSecurityDescriptor` |
| W5 | `unconstrained_delegation` | `ad.yml` | `TRUSTED_FOR_DELEGATION` set on `svc-idp-ldap` | Capture a forwardable TGT from any host the account authenticates to, then impersonate it forest-wide | 4769 for the account from an unexpected source; 4624 type 3 on the delegation host |
| W6 | `excessive_idp_directory_permissions` | `ad.yml` | `svc-idp-ldap` holds `Replicating Directory Changes` and `Replicating Directory Changes All` on the domain root | DCSync: replicate `krbtgt`'s hash, forge a golden ticket, become Domain Admin | 4662 with the DRSUAPI control-access rights; 4728/4732 if the grant was made by group |
| W7 | `excessive_group_membership` | `ad.yml` | `svc-app-portal` is a member of `GG-Employees`, the group carrying interactive logon on the workstations | The credential recovered from the portal database becomes an interactive logon on `client01` instead of a credential that can only bind LDAP | 4728 (member added to a security-enabled global group) on `GG-Employees` |

## The chain these weaknesses form

The Chapter 5 engagement reaches Domain Admin through one clean path. It is
worth stating which toggles it needs, because the others are not decoration —
they are separate, individually documented findings.

**The escalation (W1 and W6).** The attacker recovers `svc-app-portal`'s
credential from the portal's database, and the DMZ boundary permits that
credential to bind LDAP on `dc01` (tcp 389/636, the documented pivot). What
that credential *cannot* do is reach the LAN: the boundary allows nothing else,
and Kerberos (88) and the RPC traffic DCSync needs do not cross it. So the
attacker needs a foothold inside, and W7 is what provides one: `svc-app-portal`
is a member of `GG-Employees`, which carries `SeInteractiveLogonRight` on
`client01`, so the recovered credential becomes a workstation logon. From there:
W1's SPN makes `svc-idp-ldap` Kerberoastable, the ticket cracks offline, and
W6's replication rights turn that credential into `krbtgt` — Domain Admin.

**The local-privilege branch (W2 and W3).** Local admin on the workstation, and
the LAPS password that gets there, are what make the planted shares on
`client01` reachable and what an LSASS dump needs. They are the §13 first-wave
narrative ("employee credential → local admin → LSASS → Kerberoast the bind
account") and they are exercised in Chapter 5 as a documented alternative
route, not as the escalation.

**W4 and W5 are wired, asserted and offered as adjacent paths.** Neither is
load-bearing for the chain, and this document says so rather than implying
otherwise. W4 is the shortest route to code execution on every workstation at
once; W5 is the classic delegation capture. Both are real findings with real
detection signals, which is why they are planted rather than omitted.

## Why these, and not others

Every row above is a misconfiguration a real organization reaches by doing
something reasonable: a service account added to a group so it could reach a
share; a directory-sync product granted replication rights per its vendor's
install guide; a read right handed to helpdesk on an OU; a delegation bit set
by an installer for a multi-hop scenario; a GPO delegation made to a role group
that has since grown. None of them is exotic, and none requires a flaw in
Active Directory. That is the argument Chapter 4 makes.

## Deliberate deviations

`docs/ad-identity-design.md` section 14 is the register. The three that this
document's weaknesses lean on:

- **Lockout is disabled** (`LockoutThreshold 0`). The built-in Administrator is
  Ansible's transport, and in AD it can be locked out. Password spraying is
  therefore untroubled, which is a deviation this lab declares rather than
  exploits.
- **One shared human password** (`LAB_USER_PASS`), with `svc-app-portal` on its
  own variable (`LAB_SVC_APP_PASS`). W1's cracking step depends on
  `svc-idp-ldap`'s password being the shared one and never rotating.
- **LDAP is plain on 389**, not LDAPS. The DMZ boundary's one allowance is
  DMZ → `dc01` tcp 389/636, so the credential the Chapter 5 chain recovers
  crosses that boundary in the clear when it is used to bind, and the
  firewall's log records it. IDP01, which binds as `svc-idp-ldap` from the
  LAN, does not cross the boundary at all. Documented, not incidental.
```

- [ ] **Step 2: Check the table against the code**

Every toggle name in the table must appear in `lab.yaml`, every owning playbook
must contain the task that applies it, and every detection event ID must be one
`docs/ad-identity-design.md` section 13 names. Run:

```bash
cd /home/kiasoh/crapijat/College/project
for t in kerberoastable_idp_bind excessive_local_admin excessive_laps_read \
         weak_gpo_permission unconstrained_delegation \
         excessive_idp_directory_permissions excessive_group_membership; do
  printf '%s: lab.yaml=%s playbooks=%s\n' "$t" \
    "$(grep -c "$t" lab.yaml)" \
    "$(grep -rc "$t" ansible/playbooks/ | grep -v ':0' | tr '\n' ' ')"
done
```

Expected: every toggle shows a non-zero count in `lab.yaml` and at least one
playbook hit.

- [ ] **Step 3: Add the document to `CLAUDE.md`**

In the "Docs hierarchy" list, add a bullet after the `docs/redclay.yaml` entry:

```markdown
- `docs/planted-weaknesses.md` — the Chapter 4 planted-weakness section: every
  `ad.weaknesses` toggle, its owning playbook, the attack it enables, the
  correct configuration, and the detection signal. Section 13 of the design doc
  is the authoritative table; this is its expansion.
```

- [ ] **Step 4: Commit**

```bash
git add docs/planted-weaknesses.md CLAUDE.md
git commit -m "docs(lab): write the Chapter 4 planted-weakness section

Every AD weakness with its owning playbook, the misconfiguration, the
attack it enables, the correct configuration and the detection event. It
also states which toggles the Chapter 5 chain actually needs, so the
others read as separate findings rather than as an unexplained surplus."
```

---

### Task 11: Prove the whole set in both postures

The per-task runs prove one toggle at a time. This task proves the phase:
all seven on, the validator green, all seven off, the validator green, and the
lab still reachable.

**Files:**
- Modify: `docs/ad-identity-design.md` (section 13 — record the toggles as
  built, with the date)
- Modify: `docs/roadmap.md` (Phase 4 — strike the completed items)

**Interfaces:**
- Consumes: everything above.
- Produces: the evidence Chapter 5's baseline-versus-weakness comparison needs.

- [ ] **Step 1: Turn all seven on**

```bash
cd /home/kiasoh/crapijat/College/project
python3 - <<'PY'
import re
p = 'lab.yaml'
s = open(p).read()
for k in ["kerberoastable_idp_bind", "excessive_local_admin", "excessive_laps_read",
          "weak_gpo_permission", "unconstrained_delegation",
          "excessive_idp_directory_permissions", "excessive_group_membership"]:
    s = re.sub(r'^(    %s): (true|false)$' % k, r'\1: true', s, flags=re.M)
open(p, 'w').write(s)
PY
grep -A9 "^  weaknesses:" lab.yaml | head -12
```

- [ ] **Step 2: Run the full deploy and capture the output**

```bash
./deploy.sh 2>&1 | tee /tmp/phase4-all-on.log
```

Expected: green through `ad_validate.yml`. `./deploy.sh` runs `ad.yml` before
`ad_gpo.yml` before `client.yml`; the ordering matters because W4's GPO must
exist before its assertion and W2's group must exist before the workstation
play asserts it.

- [ ] **Step 3: Turn all seven off and run the full deploy again**

```bash
python3 - <<'PY'
import re
p = 'lab.yaml'
s = open(p).read()
for k in ["kerberoastable_idp_bind", "excessive_local_admin", "excessive_laps_read",
          "weak_gpo_permission", "unconstrained_delegation",
          "excessive_idp_directory_permissions", "excessive_group_membership"]:
    s = re.sub(r'^(    %s): (true|false)$' % k, r'\1: false', s, flags=re.M)
open(p, 'w').write(s)
PY
./deploy.sh 2>&1 | tee /tmp/phase4-all-off.log
```

Expected: green. This is the baseline half of Chapter 5's comparison, and it is
the run that proves the flag-off branches actively remove rather than skip.

- [ ] **Step 4: Prove the lab is still reachable in the hardened posture**

```bash
ansible -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
        vms_windows -m ansible.windows.win_ping
ansible-playbook -i ansible/inventory/lab_inventory.py -i ansible/inventory/terraform_vms.py \
                 ansible/playbooks/ad_validate.yml
```

Expected: both succeed. A weakness that bricks the lab is not a scenario, it is
an outage; the built-in Administrator and WinRM must survive every posture, and
this is the step that proves it.

- [ ] **Step 5: Turn all seven back on and leave them on**

The lab's default posture is vulnerable, matching `app.weaknesses`, which
defaults to `true`.

```bash
python3 - <<'PY'
import re
p = 'lab.yaml'
s = open(p).read()
for k in ["kerberoastable_idp_bind", "excessive_local_admin", "excessive_laps_read",
          "weak_gpo_permission", "unconstrained_delegation",
          "excessive_idp_directory_permissions", "excessive_group_membership"]:
    s = re.sub(r'^(    %s): (true|false)$' % k, r'\1: true', s, flags=re.M)
open(p, 'w').write(s)
PY
./deploy.sh 2>&1 | tee /tmp/phase4-final.log
```

- [ ] **Step 6: Record the result in the design doc**

In `docs/ad-identity-design.md` section 13, replace the paragraph that says the
toggles are all `false` and read by nothing with the observed state: all seven
wired, the owning playbook for each, and the date (2026-09-25). Keep the
normal→vulnerable table itself unchanged — it is the specification.

Also update `lab.yaml`'s comment above the `ad.weaknesses` map, which currently
says "All off, and NOTHING READS THEM YET":

```yaml
  # Deliberate weakness toggles (docs/ad-identity-design.md section 13).
  # Read by ad.yml (W1/W3/W5/W6/W7), client.yml (W2) and ad_gpo.yml (W4);
  # asserted in both postures by ad_validate.yml. On is the lab's default
  # posture, matching `app.weaknesses`.
  weaknesses:
```

- [ ] **Step 7: Strike the completed Phase 4 items in the roadmap**

In `docs/roadmap.md`, annotate the Phase 4 section the way Phase 2's is
annotated: what landed, when, and where the design lives. Leave the
still-outstanding items (nothing in Phase 4 should remain) or add the ones this
plan deliberately did not do — if any — as explicit future work.

- [ ] **Step 8: Commit**

```bash
git add lab.yaml docs/ad-identity-design.md docs/roadmap.md
git commit -m "docs(lab): record Phase 4 as built

All seven AD weakness toggles are wired and asserted in both postures,
and the lab is reachable in either. Section 13's table is unchanged: it
is the specification, and the toggles now implement it."
```

---

## Self-Review

**Spec coverage.** Sections 13's seven weaknesses each own a task: W1 → Task 2,
W7 → Task 3, W6 → Task 4, W2 → Task 5, W3 → Task 6, W5 → Task 7, W4 → Task 8.
Section 4's five mechanisms and the additive-removal rule are enforced by the
flag-off step in every task and stated in Global Constraints. The roadmap's four
Phase 4 bullets map to Task 1–8 (wire the toggles), Tasks 3–4 and the chain
section of Task 10 (one clean path to Domain Admin), Task 9 (plant data to
exfiltrate) and Task 10 (document each planted weakness). Task 11 produces the
baseline-versus-weakness evidence Phase 6 needs.

Two spec requirements are deliberately **not** implemented here and are named in
Task 10 so the omission is visible: W3's second gate (moving the decryption
principal from `GG-IT-Admins` to a broader group) is not used, because opening
one gate is enough and opening both would make the finding indistinguishable
from a broken LAPS deployment; and W5 is not made load-bearing for the chain.

**Placeholder scan.** No `TBD`, `TODO` or "similar to Task N" remains. The one
piece of scaffolding written in Task 9 step 2 is explicitly deleted in step 4
and is present only so the failing run has a parsable file.

**Type consistency.** `lab_ad_weak` (a map) and `lab_ad_weak_keys` (a list) are
defined in Task 1 and read by Tasks 2–8 with the same spellings. The three new
`lab.yaml` values are `ad.idp.spn`, `ad.workstation.weak_logon_member` and
`data.shares`, each read by exactly the playbook and the validator that Task 2,
Task 3 and Task 9 name. The GPO name in Task 8 (`gpo_baseline` in `ad_gpo.yml`,
the literal `WS - Security Baseline` in the validator) is the existing var's
value, not a new one.

**Review Focus.** (1) additive-removal on a second run — steps 6–7 of Tasks 2,
4, 6, 7 and 8, and step 5 of Task 5; (2) `client01` absent — Task 5 step 2;
(3) a typo'd `weak_logon_member` — Task 3 step 7; (4) `LAB_USER_PASS` /
`LAB_SVC_APP_PASS` mismatch — Task 11 step 4 (a hardened-posture run with the
accounts still reachable) plus Task 10's deviations section, which states the
dependency rather than leaving it implied; (5) on → off → on convergence —
Task 5 step 6 and Task 11 steps 1–5.

## Execution Handoff

This plan is eleven tasks, and the tasks are independent of each other's
interfaces — each one adds a self-contained task to an existing playbook plus
its assertion. What a shipped mistake costs is high, though: four of the seven
weaknesses are ACL edits to a live directory, and a wrong `RemoveAccessRule`
call can strip a delegation the lab needs. That argues for the thorough option.

**Recommended: Subagent-driven** — a fresh implementer and a fresh reviewer per
task keeps the ACL work under independent eyes, and the validator output is the
kind of evidence a reviewer can check cheaply.
