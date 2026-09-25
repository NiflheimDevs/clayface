# Planted weaknesses — the AD identity layer

The design is `docs/ad-identity-design.md`; section 13 is the authoritative
normal-to-vulnerable table and this document is the Chapter 4 expansion of it.
Each weakness is one entry under `ad.weaknesses` in `lab.yaml`, applied by the
playbook that owns the object, and asserted in both postures by
`ansible/playbooks/ad_validate.yml`. Turning a toggle off is a real
configuration change that restores the baseline — not a skipped task. The
correct configuration each entry departs from is section 13's **Normal
configuration** column, one row per weakness; this document delegates that half
to it rather than duplicating the table.

| # | Toggle (`lab.yaml`) | Owner | Misconfiguration while `true` | Attack it enables | Detection |
| --- | --- | --- | --- | --- | --- |
| W1 | `kerberoastable_idp_bind` | `ad.yml` | `svc-idp-ldap` holds SPN `HTTP/idp01.clayface`; its password is the shared lab password and never rotates (`MaxPasswordAge 0`) | Request a TGS for the SPN as any authenticated principal and crack it offline | 4769 with RC4 (`0x17`) encryption for a service account |
| W2 | `excessive_local_admin` | `client.yml` | `GG-Employees` is a member of the local `Administrators` group on every workstation | Any employee — or anyone holding an employee credential — is a local administrator; LSASS dump, planted data reachable | 4732/4733 (member added to a security-enabled local group) |
| W3 | `excessive_laps_read` | `ad.yml` | `GG-Employees` holds the LAPS read-password right on `OU=Workstations` | Read the LAPS-managed local administrator password of any workstation | 4662 on the `msLAPS-Password` attribute |
| W4 | `weak_gpo_permission` | `ad_gpo.yml` | `GG-Employees` holds `GpoEditDeleteModifySecurity` on `WS - Security Baseline` | Edit the baseline's settings and security filtering — arbitrary configuration on every workstation at once | 5136 on the GPO's `nTSecurityDescriptor` |
| W5 | `unconstrained_delegation` | `ad.yml` | `TRUSTED_FOR_DELEGATION` set on `svc-idp-ldap` | Capture a forwardable TGT from any host the account authenticates to, then impersonate it forest-wide | 4769 for the account from an unexpected source; 4624 type 3 on the delegation host |
| W6 | `excessive_idp_directory_permissions` | `ad.yml` | `svc-idp-ldap` holds `Replicating Directory Changes` and `Replicating Directory Changes All` on the domain root | DCSync: replicate `krbtgt`'s hash, forge a golden ticket, become Domain Admin | 4662 with the DRSUAPI control-access rights; 4728/4732 if the grant was made by group |
| W7 | `excessive_group_membership` | `ad.yml` | `svc-app-portal` is a member of `GG-Employees`, the group carrying interactive logon on the workstations | The credential recovered from the portal database becomes an interactive logon on `client01` instead of a credential that can only bind LDAP | 4728 (member added to a security-enabled global group) on `GG-Employees` |

One note on the row for W2. Section 13 names 4732 — a member was added to a
security-enabled local group — and not 4733, a member was removed from one.
Both are kept in the detection column, because 4733 is the removal complement
of the event section 13 names: Chapter 5 compares the hardened posture against
the vulnerable one, and it is the removal that comparison observes.

## The chain these weaknesses form

The Chapter 5 engagement reaches Domain Admin through one clean path. It is
worth stating which toggles it needs, because the others are not decoration —
they are separate, individually documented findings.

**The escalation (W7, W1 and W6).** The attacker recovers `svc-app-portal`'s
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
