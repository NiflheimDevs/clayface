# Domain Users


We have four Identity planes.
1. AD
2. IDP
3. APP
4. DB
Why?
AD for centralized authentication of users in our domain. duh
IDP for applications to use/query centralized authentication of users in our domain. duh
APP because a non domain user want access to app (after all, this is an service exposed to outside world)
DB. it doesn't fall into any of the categories so it has it's own plane

This is the perfect world im chasing. But im failing because of the deadline.
The deadline is pushing me to remove IDP from the whole lab.
me removing that, will break the APP to AD relation i need for a good scenario and satisfy my thesis objective. So i just placed a AD user in database reachable by APP. that way, the link still exists. a bit doggy i know but that's the best i can do for now


Now let's move on to actual Domain users, policies and OUs.

## OUs

```
DC=clayface,DC=local
├── OU=Clayface
│   ├── OU=Users
│   ├── OU=Groups
│   ├── OU=ServiceAccounts
│   └── OU=Workstations
└── OU=Domain Controllers   built-in, untouched
```

- **why an `OU=Clayface` anchor** — it is the only link point for a lab-level GPO
  that is not `Default Domain Policy`, it keeps lab objects out of the built-in
  containers, and it avoids a real ambiguity: an `OU=Users` directly under the
  root sits confusingly beside the un-deletable `CN=Users`.
- **`OU=Users` is flat on purpose.** Five accounts do not need a department tree
  and no baseline policy targets a subset. It gains sub-OUs the day a
  per-department GPO has to target one — not before.
- **`Groups` and `ServiceAccounts` are separate** because a group object and a
  service identity have different lifecycles and different blast radii. Keeping
  the bind account out of `OU=Users` means a future human-targeted GPO, or a bulk
  password reset, cannot sweep it up by accident. It is also the natural scope
  boundary for the IDP delegation, which is granted per-OU.
- **no `OU=Servers`** — zero members today. It gets created when a member server
  exists, not in anticipation of one.
- **no GPO on `Users`, `Groups` or `Clayface`** — fine-grained password policy is
  a PSO, not a GPO, and nothing else targets those scopes. They are containers,
  not policy boundaries.

### where the workstation lives

`client01$` is created in `CN=Computers` by the join and moved into
`OU=Workstations` immediately after, in `client.yml`. Not cosmetic: the GPOs
linked to `OU=Workstations` are what deliver the workstation baseline, and a
machine in `CN=Computers` receives none of them.

## users

Five named accounts plus the built-in `Administrator`.

| account          | who                                           | OU              | groups                          | power                                                                 |
| ---------------- | --------------------------------------------- | --------------- | ------------------------------- | --------------------------------------------------------------------- |
| `ron.weas`       | unprivileged employee — the baseline identity | Users           | `GG-Employees`                  | none                                                                  |
| `bob.sing`       | employee who is also a developer              | Users           | `GG-Employees`, `GG-Developers` | none in AD                                                            |
| `abed.nad`       | IT admin / helpdesk                           | Users           | `GG-Employees`, `GG-IT-Admins`  | local admin + LAPS read on CLIENT01                                   |
| `adm-hermione`   | tier-0 admin (`tier0: true`)                  | Users           | `Domain Admins`                 | domain-wide                                                           |
| `svc-idp-ldap`   | IDP01 directory bind account                  | ServiceAccounts | none                            | read-only on two OUs                                                  |
| `svc-app-portal` | APP01's service identity                      | ServiceAccounts | none                            | none in AD — its credential is planted in the portal and its database |
| `Administrator`  | built-in: Ansible transport + break-glass     | CN=Users        | `Domain Admins`                 | domain-wide                                                           |

Why each, and what was rejected:

- **`ron.weas`** — every negative test needs an identity that must *fail*.
  Without an unprivileged account you cannot show the privilege boundary holds.
- **`bob.sing` is not justified by a job title.** He is the second role that makes
  the IDP group-to-role mapping *testable*: with one role you can demonstrate an
  allow but never a denial. `GG-Developers` grants nothing inside AD.
- **`abed.nad` and `adm-hermione` are two different people**, not two accounts for
  one. `abed.nad` is daily-use helpdesk; `adm-hermione` is tier-0 and must never
  touch a workstation. A tier-0 account is only interesting *as a boundary* if a
  lower tier exists to compare it against.
- note the asymmetry: `adm-hermione` is in `Domain Admins` but deliberately **not**
  in `GG-IT-Admins`. Putting her there would quietly make every tier-0 account a
  local admin on every workstation.

## account policy

| setting                       | value            | why                                       |
| ----------------------------- | ---------------- | ----------------------------------------- |
| `MinPasswordLength`           | 12               | realistic corporate floor                 |
| `ComplexityEnabled`           | true             |                                           |
| `PasswordHistoryCount`        | 24               | no cycling back to a known password       |
| `MinPasswordAge`              | 1 day            | makes the history count enforceable       |
| `MaxPasswordAge`              | 0 (never)        | NIST 800-63B: no forced rotation          |
| `ReversibleEncryptionEnabled` | false            | a true value stores passwords recoverably |
| `LockoutThreshold`            | **0 (disabled)** | see below                                 |

**lockout is off** . for personal reasons :). it is temp but still.

## groups

| group           | purpose                                                               | members                      | grants                                                                    |
| --------------- | --------------------------------------------------------------------- | ---------------------------- | ------------------------------------------------------------------------- |
| `GG-Employees`  | every human employee — the "may use a corporate workstation" boundary | ron.weas, bob.sing, abed.nad | `Allow log on locally` on CLIENT01; IDP maps it to `app_user`             |
| `GG-Developers` | the developer role. **grants nothing in AD**                          | bob.sing                     | nothing in AD; IDP maps it to `app_developer`                             |
| `GG-IT-Admins`  | workstation administration                                            | abed.nad                     | local `Administrators` on CLIENT01, LAPS read + decryption principal, RDP |

## GPOs

| GPO                      | linked at         | purpose                                                                  |
| ------------------------ | ----------------- | ------------------------------------------------------------------------ |
| `WS - Security Baseline` | `OU=Workstations` | firewall, Defender, OS hardening — the boundary an attacker has to cross |
| `WS - LAPS`              | `OU=Workstations` | Windows LAPS policy: where to back up, password shape, which account     |
| `Default Domain Policy`  | domain root       | account policy above                                                     |

### WS - Security Baseline

Firewall, via `Set-NetFirewallProfile -PolicyStore`: all three profiles enabled,
inbound default **block**, outbound allow, allowed and blocked connections logged.
Plus one explicit inbound rule — TCP 5985 (WinRM) from `10.0.0.0/24`. Ansible
reaches CLIENT01 over WinRM, and depending on the *locally* created WinRM rules
would make Ansible's access a property of the image rather than of the policy.

Two traps that break the lab if hit: do **not** set `AllowLocalFirewallRules` to
false (the base image enables WinRM via `winrm quickconfig`, which makes *local*
rules — disabling merging removes Ansible's access to the workstation), and do
**not** ship inbound default-block without the 5985 rule. Same outcome, harder to
diagnose.

Hardening, via `Set-GPRegistryValue`:

| setting                                             | effect                                                          |
| --------------------------------------------------- | --------------------------------------------------------------- |
| `LmCompatibilityLevel = 5`                          | refuse LM and NTLMv1, send NTLMv2 only                          |
| `NoLMHash = 1`                                      | stop storing the LM hash                                        |
| `RestrictAnonymousSAM = 1`                          | hides local account enumeration                                 |
| SMB signing required (server and client)            | blocks NTLM **relay** to SMB                                    |
| `NoDriveTypeAutoRun = 255`                          | kills autorun as an initial-access vector                       |
| `InactivityTimeoutSecs = 900`                       | workstation locks after 15 minutes                              |
| UAC `ConsentPromptBehaviorAdmin = 2`, `...User = 0` | elevation re-authenticates; standard users are refused outright |
| Defender real-time, cloud and PUA protection        | Defender stays live                                             |

The signing rows deserve a precise note: signing blocks relay **to SMB**, it does
**not** block lateral movement with a stolen credential. `docs/redclay.yaml` lists
`future_SMB_NTLM_RDP_WinRM_lateral_movement` as a target path, and that path
survives this baseline intact — what signing removes is the relay variant.

**Deliberately omitted: Defender ASR**, notably *Block credential stealing from
lsass.exe*. Enabling it would break the primary attack path (local admin → LSASS
dump), so a real corporate baseline would have it and this one does not. That is
**omitted hardening — a gap, not a planted weakness**: nothing was changed to
create it, and turning it on is a hardening task rather than a scenario toggle.
The distinction matters for the writeup; conflating the two would overstate how
intentional this lab is.

### WS - LAPS

| setting                                        | value                                              |
| ---------------------------------------------- | -------------------------------------------------- |
| `BackupDirectory`                              | 2 — back up to AD only                             |
| `PasswordLength` / `PasswordComplexity`        | 20 / 4 (upper, lower, digits, special)             |
| `PasswordAgeDays`                              | 30                                                 |
| `ADPasswordEncryptionEnabled` / `...Principal` | 1 / `CLAYFACE\GG-IT-Admins`                        |
| `AutomaticAccountManagement*`                  | on, new custom account, named `lapsadmin`, enabled |

So each workstation gets a unique, 30-day-rotated local admin password instead of
a shared one. `AutomaticAccountManagement*` needs Windows 11 24H2 or later and
CLIENT01 runs 26H1 — which is why no manual local-account step appears anywhere in
this design: LAPS creates the managed account itself. (Fallback if it misbehaves
on the built image: `win_user` plus `AdministratorAccountName` — specifying a name
does *not* create the account.)

### why no GptTmpl.inf and no LGPO.exe

`microsoft.ad.gpo` manages **links only** — it does not create GPOs, by its own
documentation. So creation is `New-GPO`, and content has three possible homes:
`registry.pol` (Administrative Templates, Defender, LAPS, firewall),
`GptTmpl.inf` (security policy: user rights, security options), and GPP XML.

Registry and firewall content have first-party cmdlets (`Set-GPRegistryValue`, the
NetSecurity cmdlets with `-PolicyStore`), and those cmdlets **bump the GPO version
number correctly**, so clients notice the change. `GptTmpl.inf` has no such
cmdlet: writing the file into SYSVOL by hand does not bump the version, so clients
may not reapply, and poking the GPO's `versionNumber` attribute is a hack. LGPO.exe
from the Microsoft Security Compliance Toolkit handles it properly, at the cost of
an external binary the lab would have to fetch.

The only `GptTmpl.inf` content this design needs is user rights and local group
membership — both per-machine *local security* — so they are applied on the host
by Ansible instead. No `GptTmpl.inf`, no LGPO.exe, no version hack. The trade,
recorded honestly: those settings are then host configuration, not Group Policy.
They do not inherit, they do not reassert themselves every 90 minutes, and a
rebuilt workstation needs `client.yml` re-run (which `deploy.sh` does anyway). The
trigger to move them into a real GPO is a second workstation or a member server.

## local security on client01

Applied by `client.yml` on the host, after the join and the move.

- **local `Administrators`** gets `CLAYFACE\GG-IT-Admins` via
  `win_group_membership`. **Additive, deliberately.** The `GptTmpl.inf`
  alternative — Restricted Groups — *replaces* the membership on every refresh,
  which would strip the built-in local `Administrator` (breaking the Ansible
  transport) and the account LAPS creates (breaking LAPS). Both failures are
  silent until the next refresh.
- **user rights** go through `secedit /export` → edit `[Privilege Rights]` →
  `secedit /configure /areas USER_RIGHTS`. That `/areas` flag is load-bearing: it
  restricts the template to the one section, so the template's group-membership
  section cannot replace local group membership.

| right                                                              | assigned to                                               |
| ------------------------------------------------------------------ | --------------------------------------------------------- |
| `SeInteractiveLogonRight`                                          | `GG-Employees`, `BUILTIN\Administrators`, `Domain Admins` |
| `SeDenyInteractiveLogonRight`, `SeDenyRemoteInteractiveLogonRight` | `svc-idp-ldap` and every `tier0: true` account            |
| `SeRemoteInteractiveLogonRight`                                    | `GG-IT-Admins`                                            |
| `SeDenyNetworkLogonRight`                                          | `svc-idp-ldap`                                            |

## why the structure matters for attacks

The point of all of the above is that the relationships are traversable. Each edge
maps to a hop in the chain:

| edge                                                             | hop it enables                                                                                                                                       |
| ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GG-Employees` → logon right on CLIENT01                         | **initial access** — a stolen employee credential becomes a real foothold, not just a valid password                                                 |
| CLIENT01 → AD                                                    | **enumeration** — users, groups, SPNs and ACLs from any foothold; the input to a BloodHound graph                                                    |
| `GG-IT-Admins` → local `Administrators`                          | **credential discovery** — the hinge: local admin is the precondition for an LSASS dump; without it the foothold stays unprivileged                  |
| LSASS on CLIENT01                                                | **credential discovery** — domain credentials of everyone who logged on, including admins                                                            |
| LAPS → AD                                                        | **privilege escalation** — a readable LAPS password is a local admin credential, retrieved from the directory rather than cracked                    |
| `svc-idp-ldap` → LDAP read                                       | **lateral movement into the identity layer** — a Windows foothold becomes directory-wide read, and the IDP's data                                    |
| `svc-idp-ldap` ⇄ IDP01 config                                    | **IDP misuse** — the bind credential lives in IDP01's config, so the edge is worth drawing both ways                                                 |
| IDP01 → OIDC → APP01                                             | **application access** without touching the application's own authentication                                                                         |
| AD groups → IDP roles                                            | **authorization bypass** — the one edge where an AD change has consequences outside AD                                                               |
| `adm-hermione` in `Domain Admins` + the deny-logon rights on her | **the escalation target, and the boundary that makes it interesting** — without the deny, tier-0 on a workstation is normal behaviour, not an attack |
| `svc-app-portal` → APP01 config and database                     | **the Linux-to-AD pivot** — the portal's own weaknesses expose a credential for a real directory account                                             |
| APP01 → PostgreSQL                                               | **data access**, deliberately outside AD — a separate identity plane, which is why that boundary is worth preserving                                 |

The shape: an employee credential gives a foothold, local admin on the workstation
converts it into domain credentials, one of those reaches into the identity layer,
and everything else follows. The AD design makes each hop possible; it does not
itself perform any of them.

## baseline, weakness, and the third thing

- **baseline** — what this design builds by default. Reasonably secure, realistic.
- **planted weakness** — a deliberate, documented degradation of the baseline,
  behind a toggle. **None exist yet** (see below).
- **omitted hardening** — a control a real baseline would have, left out because it
  would break a scenario (Defender ASR). A *gap*, not a weakness: nothing was
  changed to create it, and enabling it is a hardening task rather than a scenario.

Keeping these apart matters for the writeup — a planted weakness is a design
decision with an attacker-facing consequence, omitted hardening is scope.

### the weakness wave (designed, not built)

Each is a var-gated overlay on the baseline, not a second code path, which is what
makes "enable it later with Ansible" cheap. The mechanism: `microsoft.ad.user` and
`.computer` already expose `spn`, `delegates` (RBCD), `trusted_for_delegation` and
`password_never_expires` as first-class options, so a weakness is extra keys on
the same declarative object, gated by the `ad.weaknesses` map in `lab.yaml` (all
`false`, and nothing reads them yet).

| id  | category                             | vulnerable configuration                                                                                                         |
| --- | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| W1  | poorly protected service credentials | SPN + human-chosen password + `password_never_expires` on `svc-idp-ldap` → Kerberoast, crack, recover the bind credential        |
| W2  | excessive local admin                | add `GG-Employees` to local `Administrators` on CLIENT01 → any employee credential is local admin, and an LSASS dump yields more |
| W3  | excessive LAPS read                  | widen either gate — read to `GG-Employees`, or decryption principal to `Domain Users`                                            |
| W4  | weak GPO permissions                 | grant a lower group edit rights on `WS - Security Baseline` → they can edit policy for every workstation                         |
| W5  | weak delegation                      | `trusted_for_delegation` on the bind account, or RBCD on CLIENT01 → coerce, capture the forwarded TGT, DCSync                    |
| W6  | excessive IDP directory permissions  | add it to `Domain Admins`, or grant *Replicating Directory Changes All* → DCSync every hash in the domain                        |
| W7  | excessive group membership           | add `Domain Users` or a stale account to `GG-Employees` → every account gains the logon right                                    |

## deviations from a secure baseline

Recorded so they are choices rather than accidents. Not weakness candidates —
baseline compromises the lab accepts.

| deviation                                                    | why                                                                        | upgrade path                                                        |
| ------------------------------------------------------------ | -------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| domain account lockout disabled                              | a locked-out `Administrator` stalls the Ansible transport                  | PSO for tier-0, then raise the threshold                            |
| one shared password for all humans                           | makes credential-reuse scenarios reproducible                              | per-user values in `lab.yaml`, one at a time                        |
| built-in `Administrator` keeps `Admin@123` and never expires | it *is* the Ansible transport; LAPS cannot manage it without breaking that | move Ansible to a delegated account, then let LAPS own the built-in |
| IDP01 binds over plain LDAP:389                              | LDAPS needs a cert, which needs AD CS, which is out of core scope          | deploy AD CS, issue the cert, move to 636                           |
| DSRM password equals the domain admin password               | inherited from `dc.yml`, which already flags it                            | set a distinct DSRM password at promotion                           |
| Defender ASR LSASS rule omitted                              | enabling it breaks the W2 attack path                                      | enable once the scenarios are documented as detections              |


## open questions

1. **Vendor LGPO.exe?** Not needed by this design (see the GPO section), but it
   becomes necessary the moment user rights have to scale past one workstation.
