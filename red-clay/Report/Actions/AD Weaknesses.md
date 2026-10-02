# AD Weaknesses

Seven weaknesses are now planted in the AD layer. Each one is a toggle under `ad.weaknesses` in `lab.yaml`, and all seven default to `true`.

`ad_validate.yml` asserts all seven, in both postures. The detection column below is the event Windows would log, not a built detection. There is no SIEM (security information and event management) in the lab yet, so nothing sees any of this.

| #   | toggle                                | owner        | what it plants                              |
| --- | ------------------------------------- | ------------ | ------------------------------------------- |
| W1  | `kerberoastable_idp_bind`             | `ad.yml`     | SPN on the IDP bind account                 |
| W2  | `excessive_local_admin`               | `client.yml` | employees are local admins                  |
| W3  | `excessive_laps_read`                 | `ad.yml`     | employees can read LAPS passwords           |
| W4  | `weak_gpo_permission`                 | `ad_gpo.yml` | employees can edit the workstation baseline |
| W5  | `unconstrained_delegation`            | `ad.yml`     | delegation bit on the bind account          |
| W6  | `excessive_idp_directory_permissions` | `ad.yml`     | DCSync rights for the bind account          |
| W7  | `excessive_group_membership`          | `ad.yml`     | the portal's service account is an employee |

## W1 — Kerberoastable bind account

`svc-idp-ldap` gets an SPN (service principal name): `HTTP/idp01.clayface`. It is the bind account for the identity provider (IDP), which is where the name comes from.

`HTTP` in that SPN is the Kerberos service class. It is the class name for the Hypertext Transfer Protocol, and it does not mean the service actually speaks HTTP.

Any authenticated user can ask the DC (domain controller) for a service ticket for that SPN, and the ticket is encrypted with the account's password hash. Take the ticket away and crack it offline.

It works because the account's password is the shared lab password and it never rotates (`MaxPasswordAge 0`). The weak password is not planted here, it comes from the identity design. This one just makes it reachable.

Detection: 4769 with RC4 (`0x17`, Rivest Cipher 4, an old stream cipher) for a service account.

Worth saying: the SPN names IDP01, which does not exist yet. It is planted for a service that is still only a plan.

## W2 — Every employee is a local admin

`GG-Employees` (the `GG-` prefix marks a global group in this lab) goes into the local `Administrators` group on every workstation. So any employee credential is a local administrator, and so is any credential an attacker steals from an employee.

Local admin is the hinge for the rest. It is what an LSASS dump (Local Security Authority Subsystem Service, the process holding credentials) needs, and it is what makes the planted share on `client01` readable.

Detection: 4732/4733, member added to a security-enabled local group. The removal event matters too, because that is what the hardened posture comparison observes.

## W3 — Employees can read LAPS

`GG-Employees` holds the LAPS (Local Administrator Password Solution) read-password right on `OU=Workstations`. That means the local admin password is not something to crack, it is something to read out of the directory.

The LAPS module has no revoke cmdlet, so removing this one goes through the ACL (access control list) directly. The assertion also proves the `GG-IT-Admins` baseline ACE (access control entry) survived, otherwise I would not know if I broke the baseline while removing the weakness.

Detection: 4662 on the `msLAPS-Password` attribute.

## W4 — Editable workstation baseline

`GG-Employees` gets `GpoEditDeleteModifySecurity` on the "WS - Security Baseline" GPO (Group Policy Object; the `WS` in the name is short for Workstation). Settings and security filtering both.

This is the shortest route to code execution on every workstation at once. One policy edit and the baseline an attacker was supposed to fight through is gone. It also lets them change who the policy applies to.

Detection: 5136 on the GPO's `nTSecurityDescriptor`.

## W5 — Unconstrained delegation

`TRUSTED_FOR_DELEGATION` is set on `svc-idp-ldap`. Whatever host that account authenticates to, an attacker who owns that host can capture the forwarded TGT (ticket-granting ticket). Then they impersonate the account anywhere in the forest.

The negative assertion only searches user objects, because every DC carries this bit just by being a DC.

Detection: 4769 for the account from an unexpected source, 4624 type 3 on the delegation host.

## W6 — DCSync rights

`svc-idp-ldap` holds `Replicating Directory Changes` and `Replicating Directory Changes All` on the domain root. Those are the rights a directory sync product asks for in its install guide.

DCSync is a replication request. It speaks the Directory Replication Service Remote Protocol (DRSUAPI), which is the same path a real DC uses to sync, so the account reads the directory without touching the DC's disk.

With it, the account can replicate `krbtgt`'s hash (the Kerberos ticket-granting-ticket account) and forge a golden ticket. That is Domain Admin. Both explicit ACEs are removed when the toggle is off.

Detection: 4662 with the DRSUAPI control-access rights.

## W7 — The portal account is an employee

`svc-app-portal` joins `GG-Employees`. On its own this is not a way in. It is an authorization.

`GG-Employees` carries the console logon right on `client01`, and with W2 it also carries local admin. So the credential the portal database hands over stops being a credential with no rights and becomes a workstation administrator.

The console right is `SeInteractiveLogonRight`, not remote. The remote path is W2 over WinRM (Windows Remote Management), which the workstation baseline allows from `10.0.0.0/24`. Without W2, W7 grants a logon right on a host you cannot reach.

Detection: 4728, member added to a security-enabled global group.

## The chain

W1 and W6 are the escalation. The bind account is Kerberoastable, the ticket cracks offline, and the replication rights turn that credential into `krbtgt`.

That escalation runs from a foothold inside the LAN (local area network). The DMZ (demilitarized zone) gives the attacker the same password, from the `api_keys` table, but the firewall lets it do exactly one thing: bind to LDAP (Lightweight Directory Access Protocol) on `dc01`. Kerberos and the RPC (remote procedure call) traffic that DCSync needs do not cross that boundary. The lab does not model how the LAN foothold is taken, and I am not going to write as if it does.

W2 and W3 are the local branch. Employee credential, local admin, LSASS, then Kerberoast the bind account. That is the documented alternative route, not the escalation.

W4 and W5 are wired, asserted and offered as adjacent paths. Neither is load-bearing. Both are real findings with real detection events, so they stayed in.

The design is in [[Domain Users]], the source of truth is `docs/planted-weaknesses.md`, and the VM (virtual machine) side is in [[AD]].
