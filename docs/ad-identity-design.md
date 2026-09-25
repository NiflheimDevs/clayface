# Clayface — AD Identity Design

Design for the Active Directory identity layer of the Clayface lab: the OUs,
users, groups, memberships, policies, and the contract between AD and IDP01.

Status: **the playbook is written; the layer has not been run.** Everything in
this document that is implemented lives in `ansible/playbooks/ad.yml`, which
`deploy.sh` now runs after `dc.yml` and `client.yml`. It has never executed
against a live forest — the lab was down when it was written — so treat this as
a specification with an implementation attached, not as a description of a
working domain. `dc01` is still, as far as anyone has observed, a bare forest:
promoted by `ansible/playbooks/dc.yml`, with `client01` joined by `client.yml`
and sitting in the default `CN=Computers` container. There are no custom OUs,
users, groups, or GPOs on disk. This document specifies them.

Sections 1, 1.1, 4, 4.1, 4.2 and 10 were corrected when APP01 was built — see
the note in section 1. The IDP01 half remains a contract for a later cycle.

The forest is **`clayface.local`**, NetBIOS **`CLAYFACE`**. It is derived, not
declared: `dc.yml` builds it from `LAB_DOMAIN` (default `clayface`). Any other
domain name appearing in earlier drafts of this project is obsolete.

---

## 1. Scope and non-goals

**In scope.** The AD identity layer on `dc01`: OU structure, user accounts,
security groups, group nesting, domain account policy, Group Policy Objects,
the CLIENT01 workstation policy set, Windows LAPS, and the AD-side half of the
IDP01 integration.

**Out of scope.** IDP01 is designed here but not built. No VM, no terraform
module, no playbook. The IDP01 section is a contract for a later cycle.

> **Update — APP01 is built, and it changed one thing in this document.**
> `app01` now exists (`terraform/modules/app`, `playbooks/app.yml`, see
> `docs/app01-design.md`), and it *does* consume AD: the portal carries a
> directory credential for the service account `svc-app-portal`, and `ad.yml`
> creates that account. Sections 1.1, 4 and 10 were written when it did not,
> and have been corrected in place. Everything else here is unchanged — in
> particular, **APP01 is still not domain-joined**, which is the rejection that
> mattered.

**Explicitly rejected.** Domain-joining any Linux component. Making DB01 depend
on AD. Creating AD service accounts for components that do not consume AD. Any
permission whose only justification is "to make AD matter".

### 1.1 The identity planes

Four separate identity systems, deliberately not merged:

```
AD    corporate identity        clayface.local        humans on Windows
IDP   application identity      IDP01 · OIDC/SAML     apps trust this
APP   application authorization APP01 roles           its own store
DB    database authorization    PostgreSQL roles      its own store
```

Only two components consume AD as an *authentication* dependency:

- **CLIENT01** — employees authenticate directly against AD (Kerberos/NTLM
  domain logon). This is the traditional enterprise identity plane.
- **IDP01** — uses AD as an upstream *user store* and *credential validator*.
  It does not federate with AD in the protocol sense; see section 11.

APP01 is a third consumer, of a different kind: it is **not** domain-joined and
nothing on it authenticates *to* AD, but the portal holds a directory
credential for `svc-app-portal` and uses it as a service identity. That account
is a real directory object, created by `ad.yml`. The design deliberately makes
that credential reachable through the application's own weaknesses — a database
dump yields it — which is what turns a Linux foothold into a domain credential.
`docs/app01-design.md` is the authority on the application side; section 10
below lists the edge in the attack graph.

DB01 does not consume AD. Its "service account" is a PostgreSQL role — a local
artefact, not a directory object. OPNsense does not authenticate against AD in
this design. Kali and any attacker system are outside the identity plane
entirely.

The consequence that matters: **there are two AD service accounts in this
design** — `svc-idp-ldap` for IDP01, and `svc-app-portal` for APP01 — because
there are two non-Windows consumers with a genuine reason to hold a directory
identity. Neither has an SPN, group membership, or any right beyond what its
own function needs; see sections 4 and 10.

---

## 2. Baseline facts this design respects

| Fact | Value | Source of truth |
|---|---|---|
| AD forest | `clayface.local` | derived in `dc.yml` from `LAB_DOMAIN` |
| NetBIOS name | `CLAYFACE` | derived in `dc.yml` |
| DC | `dc01` — `10.0.0.10` | `lab.yaml` |
| Workstation | `client01` — `10.0.0.20` | `lab.yaml` |
| DC OS | Windows Server 2025 (build 26100) | `base-image/README.md` |
| Workstation OS | Windows 11 26H1 | `base-image/README.md` |
| Domain admin password | `Admin@123` (or `LAB_WIN_ADMIN_PASS`) | env + unattend |
| DNS | `clayface.local` authoritative on `dc01`, forwarder to OPNsense | `dc.yml` |

Server 2025 and Windows 11 26H1 both ship the Windows LAPS client, and a
Server 2025 forest has at least one 2025 DC, which is the prerequisite for the
`msLAPS-CurrentPasswordVersion` attribute. One caveat, verified against
Microsoft's LAPS schema reference: the remaining `msLAPS-*` attributes are
**not** present in a new forest automatically and require running
`Update-LapsADSchema`. That is a task in the plan, not an assumption.

---

## 3. OU architecture

```
DC=clayface,DC=local
├── OU=Clayface
│   ├── OU=Users
│   ├── OU=Groups
│   ├── OU=ServiceAccounts
│   └── OU=Workstations
└── OU=Domain Controllers          built-in, created by promotion, untouched
```

Four OUs, each with a named consumer. The built-in
`OU=Domain Controllers` already exists with `Default Domain Controllers Policy`
linked; this design adds nothing to it.

**Why an `OU=Clayface` anchor.** It is the only link point for a lab-level GPO
that is not `Default Domain Policy`, and it keeps lab objects out of the
domain's built-in containers. It also avoids a genuine ambiguity: a
`OU=Users,DC=clayface,DC=local` would sit confusingly beside the un-deletable
built-in `CN=Users,DC=clayface,DC=local`.

**Why `OU=Users` is flat.** Six accounts do not need a department tree, and no
baseline policy targets a subset of users. Splitting it now would be hierarchy
for its own sake. It gains sub-OUs the day a per-department GPO has to target a
subset — not before.

**Why `OU=Groups` and `OU=ServiceAccounts` are separate.** A group object and a
service identity have different lifecycles and different blast radii. Keeping
the bind account out of `OU=Users` means a future human-targeted GPO linked
there, or a bulk password reset, cannot sweep it up by accident. It is also the
natural scope boundary for the IDP delegation, which is granted per-OU.

**No `OU=Servers`.** Zero members today. It gets created when a member server
exists, not in anticipation of one.

**No GPO on `OU=Users`, `OU=Groups`, or `OU=Clayface`.** Fine-grained password
policy is a PSO, not a GPO, and nothing else in the baseline targets those
scopes. Each of these OUs is a container, not a policy boundary.

### 3.1 Computer placement

CLIENT01's machine account is created in `CN=Computers` by the join and is
moved into `OU=Workstations` immediately afterwards, in `client.yml`. This is
not cosmetic: GPOs linked to `OU=Workstations` are what deliver the workstation
baseline, and a machine sitting in `CN=Computers` receives none of them.
`client.yml` already reports `OrgUnit` in its verification step, so the move is
also self-checking.

`microsoft.ad.computer` supports this: when it finds an existing object by
`identity`, the `path` option **moves** it. No `Move-ADObject` needed.

One constraint decides where that task runs. Every `microsoft.ad` object module
documents that it "must be run on a Windows target host with the
`ActiveDirectory` module installed" — and a stock Windows 11 client has no
RSAT. So the move cannot execute on CLIENT01 even though the play is about
CLIENT01. It runs as a **second play inside `client.yml`, targeting `vms_dc`**,
which keeps the ordering guarantee (the machine account exists only after the
join) without needing RSAT on the workstation. That constraint applies to every
directory-object task in this design: they all run on `dc01`.

---

## 4. Users

Five named accounts plus the built-in `Administrator`. Six objects total.

| Username | Purpose | OU | Groups | Privilege level | Used by |
|---|---|---|---|---|---|
| `ron.weas` | Unprivileged employee. The baseline identity | `OU=Users` | `GG-Employees` | none | CLIENT01 logon; negative tests |
| `bob.sing` | Employee who is also a developer | `OU=Users` | `GG-Employees`, `GG-Developers` | none in AD | CLIENT01 logon; IDP role mapping |
| `abed.nad` | IT admin / helpdesk | `OU=Users` | `GG-Employees`, `GG-IT-Admins` | local admin on CLIENT01; LAPS read | CLIENT01 logon; workstation support |
| `adm-hermione` | Tier-0 administrator | `OU=Users` | `Domain Admins` | domain-wide | DC administration |
| `svc-idp-ldap` | Directory bind account for IDP01 | `OU=ServiceAccounts` | none | read-only, scoped to two OUs | IDP01 |
| `svc-app-portal` | Service identity for APP01's portal | `OU=ServiceAccounts` | none | none in AD; its credential is planted in the portal and its database | APP01 |
| `Administrator` | Built-in. Ansible transport + break-glass | `CN=Users` (built-in) | `Domain Admins` | domain-wide | Ansible over WinRM |

### 4.1 Why these, and what was rejected

**`ron.weas`** exists because every negative test in section 16 needs an
identity that should *fail*. Without an unprivileged account you cannot
demonstrate that the privilege boundary holds.

**`bob.sing` is not justified by a job title.** He is the second role needed to
make the IDP group-to-role mapping *testable*: with one role you cannot
demonstrate a denial, only an allow. `GG-Developers` grants nothing inside AD
itself — see section 6.

**`abed.nad` and `adm-hermione` are two different people, not two accounts for
one person.** `abed.nad` is a daily-use helpdesk account with local admin on
the workstation and LAPS read. `adm-hermione` is a tier-0 account that must
never touch a workstation. The separation is the point: a tier-0 account is
only interesting as a *boundary* if a lower tier exists to compare it against.

Note the asymmetry: `adm-hermione` is in `Domain Admins` but deliberately **not**
in `GG-IT-Admins`. Tier-0 does not belong in the group that grants workstation
local admin — that membership would quietly make every tier-0 account a local
admin everywhere.

**`svc-idp-ldap` and `svc-app-portal` are the two AD service accounts.** Each
exists because a non-Windows component genuinely holds a directory identity.
`svc-idp-ldap` reads the directory; `svc-app-portal` does not read it at all —
its account exists so that the portal's *planted* credential is a real one, and
so that stealing it is worth something. Both are plain user objects rather than
gMSAs, for two reasons: a gMSA's 240-character machine-managed password cannot
be transported to a Linux host as a bind credential without a keytab, and a
gMSA cannot be Kerberoasted, which would remove a scenario rather than add one.
See weakness W1.

**Rejected accounts:**

- AD accounts for DB01 — it does not consume AD. Its service identity is a
  PostgreSQL role.
- A distinct application account per portal role — APP01's roles are its own
  store, and multiplying directory objects for them would invent an AD
  dependency the application does not have.
- A dedicated security/admin account distinct from `adm-hermione` — a third
  privileged identity with no consumer.
- Departmental users (HR, finance, …) — no OU, no resource, and no GPO targets
  them.
- A `krbtgt`-adjacent or backup account — `krbtgt` exists; nothing else is
  needed.

### 4.2 Passwords

Account passwords come from `LAB_USER_PASS` (new env var, read by
`ad.yml`, default in `deploy.env.example`), consistent with how
`LAB_WIN_ADMIN_PASS` is already handled — **never committed to `lab.yaml`**.
`lab.yaml` holds usernames, OUs, groups, and memberships (facts); the secrets
come from the environment.

**One account is the exception.** `svc-app-portal` names its own password
variable in `lab.yaml` (`password_env: LAB_SVC_APP_PASS`) so that it rotates
separately from the human accounts. The reason is on the application side:
APP01 plants that same string in its configuration and in its database, so the
planted copy and the real account must be the same string, and rotating it must
not mean rotating `ron.weas`. When the variable is unset the account falls back
to the shared lab password like everyone else — which works, but only as long
as it is set *before* `ad.yml` first creates the account.

The baseline gives every human the same lab password. That is a deliberate
deviation and it is recorded in section 14, not an oversight: it is what makes
a credential-reuse scenario reproducible, and `credential_reuse` is already a
named conceptual attack in `docs/redclay.yaml`.

---

## 5. Domain account policy

Set on `Default Domain Policy` via `Set-ADDefaultDomainPasswordPolicy`. This
cmdlet writes the domain NC directly, which is what the Default Domain Policy
GPO's `[System Access]` section would set anyway, and it is declarative and
idempotent with no GPO-content problem.

| Setting | Value | Rationale |
|---|---|---|
| `MinPasswordLength` | 12 | realistic corporate floor |
| `ComplexityEnabled` | true | matches the image's expectations |
| `PasswordHistoryCount` | 24 | prevents cycling back to a known password |
| `MinPasswordAge` | 1 day | makes the history count actually enforceable |
| `MaxPasswordAge` | 0 (never) | NIST 800-63B: no forced rotation |
| `ReversibleEncryptionEnabled` | false | a true value stores passwords recoverably |
| `LockoutThreshold` | **0 (disabled)** | deliberate — see below |

### 5.1 Lockout is disabled, and how to turn it on

**Why disabled.** In AD, unlike a local SAM, the built-in `Administrator` is
*not* exempt from account lockout. `Ansible` authenticates as
`CLAYFACE\Administrator` over NTLM for every playbook run. Five failed attempts
from any source locks the transport out of the domain for the duration of the
lockout window, which turns a mistake into a stalled demo. Lockout is therefore
off in the baseline.

**How to enable it.** Two steps, in this order:

1. Create a PSO that exempts the privileged accounts. `microsoft.ad.pso`
   carries its own `subjects` list, so no separate delegation step is needed:

   ```yaml
   - name: PSO exempting tier-0 from lockout
     microsoft.ad.pso:
       name: PSO-Tier0-NoLockout
       precedence: 10
       lockout_threshold: 0        # 0 = never lock out
       subjects:
         - Domain Admins
       state: present
   ```

   `subjects` has replace semantics, so the list is the whole truth: any
   subject removed from it stops being exempt. A PSO takes precedence over
   domain policy, which keeps the Ansible transport and break-glass usable
   while everyone else is subject to lockout.

2. Raise the domain threshold:

   ```powershell
   Set-ADDefaultDomainPasswordPolicy -Identity clayface.local `
     -LockoutThreshold 5 -LockoutDuration 00:15:00 `
     -LockoutObservationWindow 00:15:00
   ```

Verify with `Get-ADDefaultDomainPasswordPolicy` and
`Get-ADUserResultantPasswordPolicy -Identity adm-hermione`.

Enabling lockout also makes password spraying against `GG-Employees` a
*detectable* event, which is why it is worth doing before the SIEM phase rather
than after.

---

## 6. Groups

Three groups. Each is used for authorization; the two role groups are also what
IDP01 maps to application roles.

| Group | Scope | Purpose | Members | Permissions |
|---|---|---|---|---|
| `GG-Employees` | Global, Security | Every human employee. The "may use a corporate workstation" boundary | `ron.weas`, `bob.sing`, `abed.nad` | `Allow log on locally` on CLIENT01; mapped by IDP01 to `app_user` |
| `GG-Developers` | Global, Security | The developer role. **Grants nothing in AD** — exists for the IDP role mapping | `bob.sing` | none in AD; mapped by IDP01 to `app_developer` |
| `GG-IT-Admins` | Global, Security | Workstation administration | `abed.nad` | member of local `Administrators` on CLIENT01; LAPS read + LAPS decryption principal on `OU=Workstations`; `Allow log on through Remote Desktop` on CLIENT01 |

### 6.1 Nesting, and where it is not used

Two nest edges, both load-bearing:

- `GG-Developers` ⊂ `GG-Employees`
- `GG-IT-Admins` ⊂ `GG-Employees`

They exist so that `Allow log on locally` is granted **once**, to one group,
instead of being repeated per role. Adding a fourth role then means adding a
group and one nest edge, not editing a user-rights list. Both are also the
edges a weakness widens — see W7.

**No nesting beyond this.** There is no resource-side layer of `DL-*` groups.
The AGDLP pattern pays off when many resources each need to grant to many
roles; at one workstation it is pure ceremony. The upgrade trigger is a second
resource: when a file server or a second workstation needs the same grant, the
grant moves to a `DL-*` group and the role groups nest into it — the role
groups do not change.

**No group for the service account.** `svc-idp-ldap` is in no group at all. Its
only rights are an ACE on two OUs, granted directly to the account. One
consumer, one grant: a group would add a layer with nothing in it.

**No `GG-Deny-*` group.** Deny-logon rights are assigned to the two principals
that need them, directly.

### 6.2 Authentication versus authorization

Worth being precise, because the two are easy to conflate:

- **`GG-Employees`** is used for **both**. It is an *authorization* input (the
  interactive logon right) and, via IDP01, an *authentication-adjacent* input —
  IDP01 will only attempt to validate a password for a user it has synced from
  the scoped OUs.
- **`GG-Developers`** and **`GG-IT-Admins`** are **authorization only**. Neither
  is consulted when a user proves an identity; they are only read after the fact
  to decide what that identity may do.

No group in this design is used for authentication in the sense of "proving who
you are". Proving identity is the user's password or Kerberos ticket. A group
never authenticates.

---

## 7. Group Policy Objects

Two GPOs, plus `Default Domain Policy` for account policy (section 5) and the
built-in `Default Domain Controllers Policy`, which is untouched.

Both new GPOs contain **registry-based policy only**. That is a deliberate
constraint, and section 7.3 explains why.

| GPO | Linked at | Purpose | Security impact |
|---|---|---|---|
| `WS - Security Baseline` | `OU=Workstations` | Firewall profiles, Windows Defender, and OS hardening for workstations | Establishes the boundary an attacker must cross: no unsigned SMB, no LM hash, no autorun, Defender live |
| `WS - LAPS` | `OU=Workstations` | Windows LAPS policy: where to back up, password shape, and which account is managed | Gives each workstation a unique, rotated local admin password instead of a shared one |
| `Default Domain Policy` | domain root | Domain account policy (section 5) | Password floor, history, no reversible encryption |

### 7.1 `WS - Security Baseline` settings

**Windows Defender Firewall** — delivered with `Set-NetFirewallProfile
-PolicyStore "clayface.local\WS - Security Baseline"`:

| Setting | Value |
|---|---|
| Domain / Private / Public profile state | Enabled |
| Default inbound action | Block |
| Default outbound action | Allow |
| Log allowed connections | enabled |
| Log blocked connections | enabled |

Plus one explicit inbound rule: TCP 5985 (WinRM) permitted from `10.0.0.0/24`.
Ansible reaches CLIENT01 over WinRM, and relying on the locally-created WinRM
firewall rules is fragile — an explicit management-plane rule makes Ansible's
access a property of the policy rather than of the image.

**Two firewall traps, both of which break the lab if hit:**

- Do **not** set `AllowLocalFirewallRules` to false. The base image enables
  WinRM via `winrm quickconfig`, which creates *local* rules; disabling local
  rule merging removes Ansible's access to the workstation.
- Do **not** leave inbound default-block without the 5985 rule above. Same
  outcome, harder to diagnose.

**Hardening (registry, via `Set-GPRegistryValue`):**

| Setting | Key | Value | Effect |
|---|---|---|---|
| LM compatibility level | `HKLM\SYSTEM\CurrentControlSet\Control\Lsa` | `LmCompatibilityLevel = 5` | refuse LM and NTLMv1, send NTLMv2 only |
| No LM hash | `HKLM\SYSTEM\CurrentControlSet\Control\Lsa` | `NoLMHash = 1` | stop storing the LM hash |
| Anonymous SAM enumeration | `HKLM\SYSTEM\CurrentControlSet\Control\Lsa` | `RestrictAnonymousSAM = 1` | hides local account enumeration |
| SMB server signing | `...\Services\LanmanServer\Parameters` | `RequireSecuritySignature = 1` | blocks NTLM relay *to SMB* |
| SMB client signing | `...\Services\LanmanWorkstation\Parameters` | `RequireSecuritySignature = 1` | same, client side |
| Autorun | `...\Policies\Explorer` | `NoDriveTypeAutoRun = 255` | kills autorun as an initial-access vector |
| Machine inactivity limit | `...\Policies\System` | `InactivityTimeoutSecs = 900` | workstation locks after 15 minutes |
| UAC admin prompt | `...\Policies\System` | `ConsentPromptBehaviorAdmin = 2` | admin elevation re-authenticates |
| UAC standard user | `...\Policies\System` | `ConsentPromptBehaviorUser = 0` | standard users are denied elevation outright |
| Defender real-time protection | `HKLM\SOFTWARE\Policies\Microsoft\Windows Defender\Real-Time Protection` | `DisableRealtimeMonitoring = 0` | Defender stays on |
| Defender cloud protection | `...\Windows Defender\SpyNet` | `SpynetReporting = 2`, `SubmitSamplesConsent = 2` | cloud-delivered protection |
| Defender PUA protection | `...\Windows Defender` | `PUAProtection = 1` | blocks potentially unwanted apps |

A precise note on the signing rows, because the distinction matters for the
attack graph: requiring SMB signing **blocks relay to SMB** but does **not**
block lateral movement using a stolen credential. `docs/redclay.yaml` lists
`future_SMB_NTLM_RDP_WinRM_lateral_movement` as a target path, and that path
survives this baseline intact. What signing removes is the *relay* variant.

**Deliberately omitted hardening — Defender ASR.** The Windows 11 security
baseline includes Attack Surface Reduction rules, notably *Block credential
stealing from lsass.exe*. It is **not** in this baseline, because enabling it
would break the primary attack path at weakness W2 (local admin → LSASS dump).
This is a documented gap, not a planted weakness — they are different things,
and section 13 keeps them separate. A real corporate baseline would have it on.

Likewise `FilterAdministratorToken` is left at its default. Setting it to 1
interacts with the Ansible transport and with the LAPS-managed account, and
built-in `Administrator` is already a documented deviation.

### 7.2 `WS - LAPS` settings

Delivered with `Set-GPRegistryValue` under
`HKLM\SOFTWARE\Policies\Microsoft\Windows\LAPS`.

| Setting | Value | Meaning |
|---|---|---|
| `BackupDirectory` | `2` | back the password up to Active Directory only |
| `PasswordLength` | `20` | well past the crackable range |
| `PasswordComplexity` | `4` | upper + lower + digits + special |
| `PasswordAgeDays` | `30` | rotation interval |
| `ADPasswordEncryptionEnabled` | `1` | store encrypted, not clear-text |
| `ADPasswordEncryptionPrincipal` | `CLAYFACE\GG-IT-Admins` | who may decrypt |
| `AutomaticAccountManagementEnabled` | `1` | LAPS creates and manages its own account |
| `AutomaticAccountManagementTarget` | `1` | manage a new custom account, not built-in Administrator |
| `AutomaticAccountManagementNameOrPrefix` | `lapsadmin` | the managed account's name |
| `AutomaticAccountManagementEnableAccount` | `1` | LAPS enables the account it creates |

`AutomaticAccountManagement*` requires Windows 11 24H2 or later; CLIENT01 runs
26H1, so this is available. It is the reason no manual local account creation
step appears anywhere in this design: LAPS creates the managed account itself.
If it misbehaves on the built image, the fallback is `ansible.windows.win_user`
to create the account plus `AdministratorAccountName = lapsadmin` — the CSP
documentation is explicit that specifying a name does **not** create the
account.

Values verified against the LAPS CSP reference, which is the same setting set
the GPO exposes.

### 7.3 Why no `GptTmpl.inf`, and no LGPO.exe

The options were weighed and the constraint is deliberate.

`microsoft.ad.gpo` manages **links only** — its own documentation states it
does not create or delete GPOs. So GPO creation is `New-GPO`, and GPO *content*
has three possible homes: `registry.pol` (Administrative Templates, Defender,
LAPS, firewall), `GptTmpl.inf` (security policy: user rights, security options),
and GPP XML (Preferences).

Registry and firewall content have clean first-party cmdlets —
`Set-GPRegistryValue` and the NetSecurity cmdlets with `-PolicyStore` — and
those cmdlets **bump the GPO version number correctly**, so clients notice the
change. `GptTmpl.inf` has no such cmdlet. Writing that file into SYSVOL by hand
is possible but does not bump the version, so clients may not reapply; the
workaround is poking the GPO's `versionNumber` attribute, which is a hack.
LGPO.exe from the Microsoft Security Compliance Toolkit does handle it
properly, at the cost of an external binary the lab would have to fetch.

Since the only `GptTmpl.inf` content this design needs is user rights and local
group membership — both of which are *per-machine local security* settings —
they are applied on the host by Ansible instead (section 8). No `GptTmpl.inf`,
no LGPO.exe, no version hack.

The trade is recorded honestly: these settings are then host configuration, not
Group Policy. They do not inherit, they do not reassert themselves every 90
minutes, and a rebuilt workstation needs `client.yml` re-run (which `deploy.sh`
does). The upgrade trigger is a second workstation or a member server: at that
point the user-rights half moves to a `GptTmpl.inf` GPO and LGPO.exe is
vendored alongside the ISOs in the gitignored `vm/images/`, which is already
how this repo handles large external artefacts.

---

## 8. Local security settings on CLIENT01

Applied by `client.yml` on the host, after the domain join and the machine
account move.

### 8.1 Local Administrators membership

`ansible.windows.win_group_membership` adds `CLAYFACE\GG-IT-Admins` to the
local `Administrators` group.

**Additive, deliberately.** The `GptTmpl.inf` alternative — Restricted Groups
under `[Group Membership]` — *replaces* the group's membership on every policy
refresh. That would strip the built-in local `Administrator`, breaking the
Ansible transport, and it would also strip the account LAPS creates and
enables, breaking LAPS. Both failures are silent until the next refresh.
`win_group_membership` only adds, so neither can happen.

### 8.2 User rights

Applied with `secedit`, scoped to user rights only:

```
secedit /export /cfg <tmp>.inf
# edit [Privilege Rights]
secedit /configure /db secedit.sdb /cfg <tmp>.inf /areas USER_RIGHTS
```

The `/areas USER_RIGHTS` flag is load-bearing: it restricts the template to the
`[Privilege Rights]` section, so the group-membership section of the template
cannot replace local group membership. Without it this would have the same
failure mode described above.

| Right | Assigned to | Effect |
|---|---|---|
| `SeInteractiveLogonRight` | `CLAYFACE\GG-Employees`, `BUILTIN\Administrators`, `CLAYFACE\Domain Admins` | replaces the default `Users` grant, so only known staff may log on at this machine |
| `SeDenyInteractiveLogonRight` | `CLAYFACE\svc-idp-ldap`, `CLAYFACE\adm-hermione` | console logon refused outright |
| `SeDenyRemoteInteractiveLogonRight` | `CLAYFACE\svc-idp-ldap`, `CLAYFACE\adm-hermione` | RDP refused outright |
| `SeRemoteInteractiveLogonRight` | `CLAYFACE\GG-IT-Admins` | RDP permitted for workstation support |
| `SeDenyNetworkLogonRight` | `CLAYFACE\svc-idp-ldap` | the bind account has no reason to touch a workstation over SMB or WinRM |

**`Allow log on locally` replaces the list; it does not extend it.** Setting it
to `GG-Employees` alone would lock `Domain Admins` and `Administrators` off
CLIENT01 entirely. All three principals must be listed. That is also why
`Domain Admins` appears in the allow list *and* `adm-hermione` appears in the
deny list: **deny always beats allow**, so the deny is the control that holds
even if someone later widens the allow list. The two controls do different jobs
and neither is redundant.

Two rights are deliberately **left at their defaults**:

- `SeNetworkLogonRight` — restricting it would cut Ansible's WinRM access to
  the workstation and brick the lab.
- `SeRemoteInteractiveLogonRight` is narrowed, not removed, for the same class
  of reason.

`SeInteractiveLogonRight` does not affect network logons at all, which is why
narrowing it is safe for Ansible even though it restricts who can use the
keyboard.

---

## 9. Directory permissions

### 9.1 The IDP delegation

`svc-idp-ldap` is granted a read-only ACE on exactly two containers:

- `OU=Users,OU=Clayface,DC=clayface,DC=local`
- `OU=Groups,OU=Clayface,DC=clayface,DC=local`

Access: `Read`, `List contents`, `Read all properties` — applied to **this
object and all descendants**, which is what lets it enumerate users and read
their group memberships.

Scoped to those two OUs means it cannot read `OU=Workstations` (where the LAPS
attributes live), `OU=ServiceAccounts` (its own siblings), or the domain root.
The LAPS attributes are additionally confidential (`SearchFlags 904`) and only
reachable through an extended right it does not hold, so the delegation is
defence in depth rather than the only barrier. Both facts are worth stating,
because the *baseline* is the thing W6 weakens.

### 9.2 LAPS directory permissions

Three operations, all on the DC, all idempotent:

| Operation | Target | Purpose |
|---|---|---|
| `Update-LapsADSchema` | forest | add the `msLAPS-*` attributes |
| `Set-LapsADComputerSelfPermission` | `OU=Workstations` | let computers write their own LAPS password |
| `Set-LapsADReadPasswordPermission` | `OU=Workstations`, principal `GG-IT-Admins` | let IT admins read it |

`ADPasswordEncryptionEnabled` is on and `ADPasswordEncryptionPrincipal` is
`GG-IT-Admins`, so reading a password requires **both** the read permission and
membership of the decryption principal. Two independent gates, which is why W3
is a two-line weakness rather than one.

`Update-LapsADSchema` requires Schema Admins and Enterprise Admins. The domain
`Administrator` — which is what Ansible uses — holds both in a fresh forest.

---

## 10. IDP01 integration — the AD contract

Designed here; built in a later cycle. This section is the specification for
that cycle, and it is the part of the design most likely to be got wrong,
because four different things are routinely called "integrating with AD".

### 10.1 Four distinct concepts

| Concept | What it means | Used here? |
|---|---|---|
| **LDAP directory lookup** | IDP01 binds as a service account and *reads* users and groups. No user password is involved | **Yes** |
| **LDAP authentication (bind-as-user)** | IDP01 receives the user's password at its own login form and performs an LDAP simple bind *as that user* to validate it | **Yes** |
| **Protocol federation** | AD acts as a SAML or OIDC identity provider and issues tokens IDP01 trusts | **No — impossible with AD DS alone** |
| **Domain membership** | IDP01 is joined to `clayface.local` as a machine | **No — and it would buy nothing** |

**On federation.** AD DS is not a SAML or OIDC provider. It has no such
endpoint. That role belongs to AD FS or Entra ID, neither of which is in this
lab. So there is no protocol federation between AD and IDP01, and the design
should not pretend otherwise. What is commonly *called* "federating with AD" is
concept 2: IDP01 validates credentials against AD and then issues **its own**
OIDC/SAML tokens. The federation is at the application layer; AD's role is
credential validation and attribute source.

**On domain membership.** IDP01 is Linux. Joining it via `realmd`/SSSD would
give the *host* Kerberos and let `sssd` resolve AD users for SSH logon. It would
do nothing whatsoever for OIDC or SAML — a web application cannot ride `sssd`.
Joining it costs a machine account, a keytab, and a new failure mode, and buys
exactly zero progress on the stated goal. So IDP01 is **not** domain-joined.

### 10.2 The model

```
                ┌──────────────────────────────┐
                │  clayface.local (AD DS)      │
                │  dc01 · 10.0.0.10            │
                └──────────────────────────────┘
                   ▲                        ▲
    bind as svc-idp-ldap (read)   bind as <user> (validate password)
                   │                        │
                ┌──┴────────────────────────┴──┐
                │  IDP01                       │
                │  syncs users + groups        │
                │  maps AD groups → app roles  │
                │  issues its OWN OIDC tokens  │
                └──────────────────────────────┘
                            │ OIDC
                            ▼
                ┌──────────────────────────────┐
                │  APP01                       │
                │  trusts IDP01 as issuer      │
                │  authorizes on app roles     │
                └──────────────────────────────┘
                            │ its own credentials
                            ▼
                ┌──────────────────────────────┐
                │  PostgreSQL                  │
                │  its own roles               │
                └──────────────────────────────┘
```

Three flows, all using the one bind account:

1. **Directory sync (read).** IDP01 binds as `svc-idp-ldap` and reads users and
   groups from the two delegated OUs, on a schedule. No user password involved.
2. **Authentication (bind-as-user).** A user submits their domain credentials
   to IDP01's login form. IDP01 performs an LDAP simple bind as that user
   against `dc01`. AD validates. IDP01 never stores the password; it discards it
   after the bind and issues its own session.
3. **Attribute and role mapping.** IDP01 reads the user's group memberships
   from the synced data and maps AD groups to application roles.

### 10.3 What AD provides, precisely

| AD provides | Mechanism | Consumed by |
|---|---|---|
| User lookup | LDAP search in `OU=Users` | sync flow |
| Group lookup | LDAP search in `OU=Groups` | sync flow |
| Group membership | `memberOf` / `member` | role mapping |
| Authentication | LDAP **simple bind as the user** | login flow |
| Account state | `userAccountControl` (disabled, locked) | login flow — a disabled account must not authenticate |
| Password policy outcome | the bind succeeding or failing | login flow |

AD does **not** provide: tokens, sessions, MFA, application roles, or
authorization decisions. Those are IDP01's.

### 10.4 The group-to-role mapping — the explicit contract

This is the only place an AD group becomes an application role, and it is
explicit and one-directional:

| AD group | IDP01 role | Meaning in APP01 |
|---|---|---|
| `GG-Employees` | `app_user` | may use the application |
| `GG-Developers` | `app_developer` | may use developer functions |
| *(anyone else)* | none | no application access |

`GG-IT-Admins` is deliberately absent from this table. Administering the
workstation is not an application role, and mapping it to one would be exactly
the conflation this design exists to avoid.

`bob.sing` is in both mapped groups, which is what makes the mapping
demonstrable: he receives two roles, `ron.weas` receives one, and a user in
neither group receives none.

### 10.5 Transport

For the baseline, IDP01 binds over **LDAP on port 389**, unencrypted, and this
is a **documented deviation** (section 14). LDAPS on 636 requires a certificate
for `dc01.clayface.local` with that name in the SAN, which means either AD CS —
which `docs/redclay.yaml` decision `AD_CS_is_optional` places outside the core —
or an externally issued certificate. Neither is in scope here.

The consequences are stated rather than hidden: a simple bind over 389 exposes
the bind credential to anything on the wire path, and that is a real finding.
It also makes "sniff the LDAP bind" a legitimate scenario, and "deploy AD CS
and move to LDAPS" a legitimate hardening task with a measurable before and
after. Both are better outcomes than silently pretending the transport is fine.

Regardless of transport, **the credential's protection is IDP01's
responsibility, not AD's.** AD stores a password hash; how IDP01 stores the
bind credential in its own configuration or database is an APP01/IDP01 concern
and the boundary is worth keeping explicit.

---

## 11. Attack-path relevance

The point of the structure above is that the relationships are traversable.
This maps each one to the path it enables, using the chain from the project
brief:

```
Kali → initial access → credential discovery → AD enumeration
     → privilege escalation → IDP misuse → application access → data access
```

| Relationship | Attack path it enables |
|---|---|
| `GG-Employees` → `Allow log on locally` on CLIENT01 | **Initial access.** A stolen employee credential becomes a real foothold: interactive logon on a domain-joined workstation, not just a valid password |
| CLIENT01 → AD (Kerberos/NTLM) | **AD enumeration.** From any foothold, unauthenticated and then authenticated enumeration of users, groups, SPNs and ACLs — the input to a BloodHound-style graph |
| `GG-IT-Admins` → local `Administrators` on CLIENT01 | **Credential discovery.** Local admin is the precondition for an LSASS dump. This is the hinge of the whole chain: without it the foothold stays unprivileged |
| LSASS on CLIENT01 | **Credential discovery.** Domain credentials of every user who has logged on, including admins who used the workstation |
| LAPS → AD password storage | **Privilege escalation.** A readable LAPS password is a local admin credential, retrieved from the directory rather than cracked |
| `svc-idp-ldap` → LDAP read on `OU=Users`/`OU=Groups` | **Lateral movement into the identity layer.** Compromising this account converts a Windows foothold into directory-wide read access, and into the IDP's own data |
| `svc-idp-ldap` → IDP01 configuration | **IDP misuse.** The bind credential lives in IDP01's config. Compromising IDP01 yields the credential; compromising the credential yields directory read. The edge is bidirectional and worth drawing both ways |
| IDP01 → OIDC → APP01 | **Application access.** A forged or mis-issued token is application access without touching the application's own authentication |
| AD groups → IDP01 roles | **Authorization bypass.** If the mapping is widened, group membership becomes application privilege. This is the one edge where an AD change has consequences outside AD |
| `adm-hermione` in `Domain Admins` | **Privilege escalation target.** The tier-0 account is what the chain is ultimately climbing toward |
| Deny-logon rights on `adm-hermione` | **The boundary that makes the previous row interesting.** Without it, tier-0 on a workstation is normal behaviour, not an attack |
| `svc-app-portal` → APP01 configuration and database | **The Linux-to-AD pivot.** The portal's own weaknesses expose a credential for a real directory account, so compromising APP01 yields an AD identity. This is the edge that makes APP01 matter to *this* document rather than only to the application design |
| APP01 → PostgreSQL | **Data access.** Deliberately outside AD — it is a separate identity plane, which is precisely why the boundary is worth preserving. It is also where `svc-app-portal`'s credential is planted, which is what makes the row above reachable |

The chain has a clean shape: an employee credential gives a foothold, local
admin on the workstation converts that into domain credentials, and one of
those credentials reaches into the identity layer, from which everything else
follows. The AD design is what makes each hop possible; it does not itself
perform any of them.

---

## 12. Baseline versus weakness

Two different things, kept separate on purpose:

- **Baseline** — what this design builds by default. Reasonably secure, and
  realistic. Sections 3 through 10.
- **Planted weakness** — a deliberate, documented degradation of the baseline,
  enabled by a configuration toggle. Section 13. **All seven exist as of
  2026-09-25**, each applied by the playbook that owns the object it changes.

A third category also exists and is worth distinguishing from both:

- **Omitted hardening** — a control a real corporate baseline would have, that
  this lab leaves out because it would break a scenario. The Defender ASR
  LSASS rule is the example. Omitted hardening is a *gap*, not a planted
  weakness; nothing was changed to create it, and enabling it is a hardening
  task rather than a scenario toggle.

The distinction matters for the writeup: a planted weakness is a design
decision with an attacker-facing consequence, whereas omitted hardening is
scope. Conflating them would overstate the lab's intentionality.

---

## 13. Deliberate weakness candidates

**Built, 2026-09-25.** All seven are wired, each by the playbook that owns the
object it changes — `ad.yml` (W1, W3, W5, W6, W7), `client.yml` (W2) and
`ad_gpo.yml` (W4) — and each is asserted in both postures by
`ansible/playbooks/ad_validate.yml`. On is the lab's default posture, matching
`app.weaknesses`. Each maps to one of the seven categories in the brief, and
each is expressible as a var-gated overlay on the baseline rather than a second
code path — which is what made "enable it later with Ansible" cheap. The
Chapter 4 expansion is `docs/planted-weaknesses.md`. **The table below is
unchanged: it is the specification these toggles implement.**

The enabling mechanism: `microsoft.ad.user` and `microsoft.ad.computer` already
expose `spn`, `delegates` (the `msDS-AllowedToActOnBehalfOfOtherIdentity`
attribute, i.e. RBCD), `trusted_for_delegation`, and `password_never_expires`
as first-class options. So a weakness is additional keys on the same declarative
object, gated by a toggle in `lab.yaml`'s `ad.weaknesses` map.

| ID | Category | Normal configuration | Vulnerable configuration | Attacker gain | How it is observed and proved |
|---|---|---|---|---|---|
| **W1** | Poorly protected service credentials | `svc-idp-ldap` has no SPN; a long random password | Give it an SPN, a human-chosen password, and `password_never_expires` | Request a service ticket for its SPN, crack it offline, recover the directory bind credential | **Observe:** Kerberos TGS request (4769) with RC4 encryption type, from a host that is not the service. **Prove:** crack the hash offline, then bind to LDAP with the recovered password and enumerate `OU=Users` |
| **W2** | Excessive local administrator privileges | Only `GG-IT-Admins` is in local `Administrators` on CLIENT01 | Add `GG-Employees` (or `Domain Users`) to the same local group | Any employee credential becomes local admin; LSASS dump yields further credentials | **Observe:** 4732 — a member added to a security-enabled local group — naming the group; plus the GPO/host change itself. **Prove:** log on as `ron.weas`, `whoami /groups` shows `Administrators`, then perform a privileged operation |
| **W3** | Excessive LAPS read permission | `GG-IT-Admins` holds both the read permission and is the decryption principal, scoped to `OU=Workstations` | Widen either gate: grant read to `GG-Employees`, or set the decryption principal to `Domain Users` | Any low-privileged user reads the workstation's local admin password in clear text from AD | **Observe:** 4662 on the `msLAPS-Password` attribute naming the *reading* account — a precise, low-noise signal. **Prove:** `Get-LapsADPassword -Identity CLIENT01 -AsPlainText` as `ron.weas`, then log on locally with it |
| **W4** | Weak GPO permissions | Only `Domain Admins` may edit GPOs | Grant a lower group *Edit settings, delete, modify security* on `WS - Security Baseline` | That group can edit policy for every workstation: add itself to local Administrators, or plant a computer startup script running as SYSTEM | **Observe:** 5136 — a directory service object modified — on the GPO object with a changed `nTSecurityDescriptor`, and the GPO version number incrementing in SYSVOL. **Prove:** `Get-GPPermission` shows the ACE; then edit and observe the setting land on CLIENT01 |
| **W5** | Weak delegation | No account is trusted for delegation | Set `trusted_for_delegation` on `svc-idp-ldap`, or `delegates` (RBCD) on the CLIENT01 computer object | Coerce the DC into authenticating to the delegating principal, capture the forwarded TGT, then DCSync | **Observe:** 4624 type 3 followed by 4769 for the coerced SPN, plus a 5136 on the account's `userAccountControl`. **Prove:** `Get-ADObject -Filter 'userAccountControl -band 524288'` finds the flag; then perform the coercion and capture |
| **W6** | Excessive IDP directory permissions | `svc-idp-ldap` has read-only on two OUs, nothing else | Add it to `Domain Admins`, or grant it *Replicating Directory Changes* / *Replicating Directory Changes All* on the domain root | Compromise IDP01, recover the bind credential, and DCSync every password hash in the domain | **Observe:** 4662 carrying the DRSUAPI control access right — the canonical DCSync detection — or a 4728/4732 group-membership event naming the service account. **Prove:** `lsadump::dcsync` or `Get-ADReplAccount` against `krbtgt` |
| **W7** | Excessive group membership | `GG-Employees` contains three humans | Add `Domain Users`, or add a stale or contractor account that should have been removed | Every domain account gains the interactive logon right and anything else `GG-Employees` is granted | **Observe:** 4728/4732 on `GG-Employees`, and the members list diverging from `lab.yaml`. **Prove:** log on at CLIENT01 with an account that should have been refused |

**Recommended first wave: W2, W3, W1 — in that order.** This orders the
presentation, not the build: all seven are built and switched on, as this
section's opening records. They chain into a single coherent story (employee
credential → local admin on the workstation → LSASS → Kerberoast the bind
account → directory read), each is a small configuration change, and each has a
crisp detection. W4, W5, W6 and W7 are independent branches rather than steps in
that spine, which is why they come after it; W5 in particular is the most
involved to demonstrate and is best presented last, once the SIEM phase can show
the coercion.

---

## 14. Documented deviations from a secure baseline

Recorded explicitly so they are choices rather than accidents. These are not
weakness candidates — they are baseline compromises the lab accepts.

| Deviation | Why | Upgrade path |
|---|---|---|
| Domain account lockout disabled | A locked-out `Administrator` stalls the Ansible transport for the lockout window | Section 5.1 — PSO for tier-0, then raise the domain threshold |
| One shared password for all human accounts | Makes credential-reuse scenarios reproducible | Per-user values in `lab.yaml`, one at a time |
| Built-in `Administrator` keeps `Admin@123` and `password_never_expires` | It is the Ansible transport. LAPS cannot manage it without breaking that | Move Ansible to a delegated domain account, then let LAPS own the built-in account |
| IDP01 binds over plain LDAP:389 | LDAPS needs a certificate, which needs AD CS, which is out of core scope | Deploy AD CS, issue a cert for `dc01.clayface.local`, move to 636 |
| DSRM password equals the domain admin password | Inherited from `dc.yml`, which already flags it as a shortcut to revisit | Set a distinct DSRM password at promotion |
| Defender ASR LSASS rule omitted | Enabling it breaks the W2 attack path | Enable it once the scenarios are documented as detections rather than attacks |

---

## 15. Automation plan

### 15.1 Where the data lives

CLAUDE.md's data-flow rule names **users** explicitly as content that belongs
in `lab.yaml` and nowhere else. So identity *facts* go there; GPO *settings*
stay in the playbooks, as configuration rather than lab fact.

`lab.yaml` gains an `ad:` section: the OU names, the user list with their OUs
and group memberships, the group definitions with their members, the IDP bind
account name, the LAPS managed-account name, and the `weaknesses:` toggle map.
No passwords — those come from the environment.

One change is needed for this to reach a playbook. `lab_inventory.py` currently
reads only `hosts`, `edge`, and `network`, and no other inventory script reads
`lab.yaml` at all. The sanctioned path is therefore to extend `lab_inventory.py`
to publish the `ad` section into the `all` group's vars, so every host in the
merged inventory can see `lab_ad`. The alternative — having the playbook do
`lookup('file', ...) | from_yaml` on `lab.yaml` directly — is a smaller diff but
routes around the convention that Ansible reads lab.yaml through the inventory
script, and would leave the fact path inconsistent with how every other lab
fact is distributed.

`lab.yaml` is read by `terraform/locals.tf` via `yamldecode()`, which ignores
keys it does not index, so adding `ad:` does not affect `terraform plan`.

### 15.2 Files

| File | Change |
|---|---|
| `lab.yaml` | add `ad:` — OUs, users, groups, memberships, LAPS account name, weakness toggles |
| `ansible/inventory/lab_inventory.py` | publish `lab_ad` into the `all` group vars |
| `ansible/playbooks/ad.yml` | **new** — the directory: OUs, groups, users, nesting, account policy, OU delegation, LAPS schema and permissions |
| `ansible/playbooks/ad_gpo.yml` | **new** — GPOs: create, link, and write all content |
| `ansible/playbooks/ad_validate.yml` | **new** — read-only assertions; the executable form of section 16 |
| `ansible/playbooks/client.yml` | a second play targeting `vms_dc` that moves the machine account into `OU=Workstations` after the join; plus local Administrators membership, user rights and `gpupdate` on the client |
| `deploy.sh` | two new steps between `dc.yml` and `client.yml` |
| `deploy.env.example` | document `LAB_USER_PASS` |

The split between `ad.yml` and `ad_gpo.yml` follows the repo's existing
one-playbook-per-concern style (`hosts`, `edge`, `start_vms`, `opnsense`, `dc`,
`client`) and the natural seam: `ad.yml` changes the *directory*, `ad_gpo.yml`
changes *policy*.

### 15.3 Order

The order is not arbitrary — each step's precondition is the one before it.

1. **`dc.yml`** — the forest must exist. *Already in `deploy.sh`.*
2. **`ad.yml`, in this task order:**
   1. OUs — everything else is placed into them
   2. Groups — before users, because `microsoft.ad.user` can add a user to
      groups as it creates them
   3. Users
   4. Nesting — the two subset edges, after both endpoints exist
   5. Domain account policy — `Set-ADDefaultDomainPasswordPolicy`
   6. `Update-LapsADSchema` — before any LAPS permission is granted
   7. LAPS OU permissions — `Set-LapsADComputerSelfPermission`,
      `Set-LapsADReadPasswordPermission`
   8. IDP delegation — the read-only ACE on `OU=Users` and `OU=Groups`
3. **`ad_gpo.yml`:**
   1. `New-GPO` for both GPOs
   2. Link both to `OU=Workstations` with `microsoft.ad.gpo`
   3. `WS - Security Baseline` content — firewall profiles and rule, then the
      registry hardening settings
   4. `WS - LAPS` content — the LAPS registry policy
4. **`client.yml`** — unchanged up to the join, then:
   1. *(play 2, on `vms_dc`)* move the machine account into `OU=Workstations`
   2. *(play 1, on the client)* add `GG-IT-Admins` to local `Administrators`
   3. *(play 1)* apply the user rights with `secedit /areas USER_RIGHTS`
   4. *(play 1)* `gpupdate /force` and wait for the policy to land

   The split is forced by the RSAT constraint in section 3.1: the directory-side
   task needs the `ActiveDirectory` module, which the workstation does not have.
   It stays in this playbook rather than moving to `ad.yml` because `ad.yml` runs
   *before* the join, when `client01$` does not exist yet.
5. **`ad_validate.yml`** — the assertions in section 16. Read-only, safe to
   re-run at any time.

`ad_gpo.yml` must precede `client.yml` so the GPOs are linked to
`OU=Workstations` *before* CLIENT01 is moved into it — otherwise the first
policy refresh finds an empty OU and the workstation comes up unmanaged until
the next cycle.

LAPS passwords will not appear in AD until CLIENT01 has the LAPS policy applied
and has run a policy refresh, which happens after `client.yml`. Validation of
LAPS therefore asserts *policy delivery* and the *OU permissions*, not the
presence of a password — a password assertion belongs to a later run.

### 15.4 Idempotency

Every task must be safe to re-run, which is the property `dc.yml` and
`client.yml` already have. The modules handle it: `microsoft.ad.ou`,
`.group`, `.user`, `.computer`, and `.gpo` compare current state before acting.

The settings that need explicit read-first logic, because the underlying cmdlet
does not:

- `Set-GPRegistryValue` — read with `Get-GPRegistryValue`, write only on a
  difference. This is exactly the idiom `dc.yml` already uses for the DNS
  forwarder.
- `Set-NetFirewallProfile` and `New-NetFirewallRule` — read the current profile
  and rule first.
- `Set-ADDefaultDomainPasswordPolicy` — read with
  `Get-ADDefaultDomainPasswordPolicy` first.
- `secedit` — export, compare, configure only on a difference.
- `Update-LapsADSchema` — re-running it is harmless, but it should be gated on
  a schema check rather than run blind on every deploy.

### 15.5 The weakness toggles

Each weakness in section 13 is a key under `ad.weaknesses` in `lab.yaml`. All
seven are built and ship `true` — on is the lab's default posture, matching
`app.weaknesses` — and `ansible/playbooks/ad_validate.yml` asserts each of them
in both postures. The playbooks read them and conditionally add or remove the
offending option on the object the baseline already declares:

```yaml
# The shape this design described, not a verbatim copy of the shipped task.
# The shipped tasks that need an active removal are
# `ansible.windows.win_powershell` scripts, because the flag-off branch must
# actively REMOVE the configuration rather than merely not add it; the shipped
# W1 task reads the toggle as `lab_ad_weak.<key>`.
- name: Give the IDP bind account an SPN (weakness W1)
  microsoft.ad.user:
    identity: "{{ lab_ad.idp_bind_account }}"
    spn:
      - "HTTP/idp01.{{ lab_ad.dns_domain }}"
  when: lab_ad.weaknesses.kerberoastable_idp_bind | bool
```

This is the reason the candidates were chosen: none of them needs a new code
path. The baseline object and the weakened object are the same object with
different keys. The shipped W1 task (`ad.yml`, "Set the SPN that makes the IDP
bind account Kerberoastable") is the worked example: `setspn -S` and
`setspn -D` are the two branches of that one toggle.

---

## 16. Validation

`ad_validate.yml` runs these as assertions so a regression fails the deploy
rather than being noticed later. Each is also runnable by hand.

### 16.1 Users can authenticate

Prove a real authentication event against AD, not a successful bind of a
service account. From CLIENT01:

```powershell
net use \\dc01.clayface.local\SYSVOL /user:CLAYFACE\ron.weas "<password>"
```

Exit code 0 means AD accepted the credential. A non-zero exit with *The user
name or password is incorrect* means it did not — and the same command with a
deliberately wrong password proves the test can fail.

Check the account's own state:

```powershell
Get-ADUser ron.weas -Properties Enabled, PasswordLastSet, LastLogonDate |
  Select-Object Name, Enabled, PasswordLastSet, LastLogonDate
```

### 16.2 Group memberships are correct

```powershell
Get-ADGroupMember GG-Employees -Recursive | Select-Object SamAccountName
Get-ADGroupMember GG-Developers -Recursive | Select-Object SamAccountName
Get-ADGroupMember GG-IT-Admins -Recursive | Select-Object SamAccountName
Get-ADPrincipalGroupMembership ron.weas | Select-Object Name
```

The nested cases are the ones worth asserting explicitly, because they are what
the design relies on:

```powershell
# bob.sing must resolve to enterprise access THROUGH the nest, not directly
(Get-ADPrincipalGroupMembership bob.sing).Name -contains 'GG-Employees'
# adm-hermione must NOT be a member of GG-IT-Admins
-not (Get-ADPrincipalGroupMembership adm-hermione).Name -contains 'GG-IT-Admins'
```

Compare the live membership against `lab.yaml` as the authority — a divergence
is exactly what W7 looks like.

### 16.3 CLIENT01 receives the expected policies

```powershell
gpresult /scope computer /v
# The applied list must name both lab GPOs
Get-GPO -All | Where-Object DisplayName -like 'WS - *' |
  Select-Object DisplayName, Id, GpoStatus
```

Prove the settings actually *landed*, rather than trusting that the GPO is
linked:

```powershell
Get-NetFirewallProfile -PolicyStore ActiveStore |
  Select-Object Name, Enabled, DefaultInboundAction
Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Control\Lsa' -Name LmCompatibilityLevel
Get-ItemProperty 'HKLM:\SOFTWARE\Policies\Microsoft\Windows\LAPS'
whoami /groups        # GG-Employees present, Administrators absent for ron.weas
```

Every one of these must come from the resolved policy store
(`ActiveStore`, the registry, `whoami`), never from the GPO's own definition —
a GPO that is defined but not applying is the failure this is designed to
catch.

The machine account placement is asserted too, since every workstation GPO
depends on it:

```powershell
(Get-CimInstance Win32_ComputerSystem).OrganizationalUnit
# expect: OU=Workstations,OU=Clayface,DC=clayface,DC=local
```

### 16.4 Unauthorized users cannot perform privileged operations

The negative half, which is what actually proves the boundary holds.

```powershell
# ron.weas on CLIENT01 — must NOT be a local admin
whoami /groups | Select-String 'S-1-5-32-544'      # expect: no match
# and a privileged operation must fail
net localgroup Administrators ron.weas /add        # expect: Access is denied

# The interactive logon rights must be the designed set, and adm-hermione
# must appear in the deny — deny beats allow regardless of the allow list.
# Logon rights are not privileges, so whoami /priv does not show them;
# the resultant set has to be read from local policy.
secedit /export /cfg C:\Windows\Temp\secpol-init.inf /areas USER_RIGHTS
Select-String 'SeInteractiveLogonRight|SeDenyInteractiveLogonRight' `
  C:\Windows\Temp\secpol-init.inf
# expect: SeInteractiveLogonRight lists GG-Employees, Administrators,
#         Domain Admins — and SeDenyInteractiveLogonRight lists adm-hermione

# the bind account must not authenticate interactively anywhere
secedit /export /cfg C:\Windows\Temp\secpol.inf /areas USER_RIGHTS
Select-String 'SeDenyInteractiveLogonRight|SeDenyNetworkLogonRight' `
  C:\Windows\Temp\secpol.inf
# expect: svc-idp-ldap in both

# ron.weas must NOT be able to read the LAPS password — the W3 boundary
Get-LapsADPassword -Identity CLIENT01 -AsPlainText
# expect: Access is denied
```

The LAPS check is the sharpest of these: reading the password as `abed.nad`
must succeed and as `ron.weas` must fail. Both directions, or the test proves
nothing about the permission scope.

### 16.5 IDP can consume the intended AD identity information

This splits into a positive case and a negative one, and they need different
instruments. The positive case has to bind *as the bind account* rather than
merely inspect its ACEs, because an ACE that looks right can still fail to
grant what it appears to. The negative case cannot use a bind at all — see
below. For the positive case, two tools were considered and one was rejected:

- **`microsoft.ad.debug_ldap_client` is the wrong tool.** It reports the
  *Ansible host's* LDAP client capabilities — dnspython, Kerberos configuration,
  discovered SRV records. It has no options at all and its own documentation
  says the return values are not a contract. It proves nothing about a bind.
- **`Get-ADUser` as the bind account is also awkward**: the account is denied
  interactive and network logon everywhere (section 8.2), so it cannot be used
  to open a session, and `runas /netonly` cannot be driven cleanly from
  Ansible.

The tool that fits is a raw LDAP bind with explicit credentials, which needs no
session and no profile:

```powershell
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.DirectoryServices.Protocols

# Three arguments, not two. The two-argument form stores the whole
# 'CLAYFACE\svc-idp-ldap' string in UserName, which Negotiate cannot parse,
# and every bind then fails with "The supplied credential is invalid."
$conn = New-Object System.DirectoryServices.Protocols.LdapConnection('clayface.local:389')
$conn.Credential = New-Object System.Net.NetworkCredential(
  'svc-idp-ldap', '<password>', 'CLAYFACE')
# Negotiate, not Basic. A Basic bind is refused outright by the DC with
# "Strong authentication is required for this operation", and it would put
# the password on the wire in the clear over 389.
$conn.AuthType = [System.DirectoryServices.Protocols.AuthType]::Negotiate
$conn.Bind()

function Test-Search([string]$Base, [string]$Filter) {
    $req = New-Object System.DirectoryServices.Protocols.SearchRequest(
        $Base, $Filter, [System.DirectoryServices.Protocols.SearchScope]::Subtree)
    try   { $conn.SendRequest($req) | Out-Null; 'ALLOWED' }
    catch { "DENIED: $($_.Exception.Message)" }
}

Test-Search 'OU=Users,OU=Clayface,DC=clayface,DC=local'  '(objectClass=user)'
Test-Search 'OU=Groups,OU=Clayface,DC=clayface,DC=local' '(objectClass=group)'
```

Expected: `ALLOWED` for both. That is the whole of what a bind can prove here.

**The negative case cannot be a search, and an earlier draft of this section
was wrong about it.** That draft added a third probe against the domain root
and expected `DENIED`. Measured against the live lab, it returns `ALLOWED` —
and would for *any* account, because a default AD grants Authenticated Users
read across the naming context, and nothing in `ad.yml` restricts that; the
delegation only *adds* an ACE. So the probe passes whether the delegation is
scoped or wide open. It proves nothing, and worse, it reads as if it does.

What the lab actually controls is which objects the bind account holds an
**explicit** ACE on. That is the property to assert, and unlike the search it
can fail:

- each OU in `idp.read_ous`: exactly one explicit `Allow` ACE carrying
  `GenericRead`, and
- every other object in scope — the remaining child OUs, the anchor OU, and
  the domain root — zero explicit ACEs for that account.

Inherited ACEs are skipped, since every OU inherits from the domain root by
construction. An explicit ACE appearing on the domain root is the DCSync-shaped
W6 this test exists to catch, and granting one is how the assertion was
confirmed to fail rather than merely pass.

`ad_validate.yml` splits this into two tasks. The ACL half stays visible in the
play output; the bind half is `no_log: true`, because its rendered script
carries the password.

The LAPS attribute is deliberately not part of this test: it is confidential
(`SearchFlags 904`), so reading it requires a control access right the bind
account does not hold, and a plain search would confirm the OU scope rather
than the attribute's protection. It is asserted separately, from `dc01`.

---

## 17. Open questions

1. **Whether to vendor LGPO.exe.** Not needed by this design (section 7.3), but
   it becomes necessary the moment user rights have to scale past one
   workstation. Deciding now would avoid reworking the playbooks later.
2. **Whether AD CS comes in scope.** It gates LDAPS for IDP01 and the
   certificate-based attacks already listed as future paths
   (`future_AD_CS_certificate_authentication`). `docs/redclay.yaml` currently
   marks it optional.
3. **Which weakness wave to build first.** Section 13 recommends W2, W3, W1.
   That is a recommendation, not a decision.
