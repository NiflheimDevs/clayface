
# The AD Attack Techniques

### Kerberoasting

Any domain user can ask the KDC for a TGS ticket for any service. The KDC will give it to you — it doesn't check if you're authorized to *use* it, only that you're a valid domain user.

The TGS is encrypted with the service account's password hash. You take it offline and brute-force crack it. If the service account has a weak password, you crack it and get the plaintext password of a potentially privileged account.

**Why it works:** the KDC handing out TGS tickets to anyone is by design. The assumption was you can't crack the encryption. With modern GPUs and weak passwords, that assumption fails.

### AS-REP Roasting

Normally in Kerberos, when you request a TGT, you first must prove you know your password (called "pre-authentication"). But some accounts have this disabled.

For accounts without pre-auth, the KDC will happily return an AS-REP response encrypted with the user's password hash — without you proving who you are first. You take that encrypted blob offline and crack it.

**Why it works:** same as Kerberoasting, but targets users instead of services. The account flag `DONT_REQUIRE_PREAUTH` is a misconfiguration.

### Pass-the-Hash (PtH)

NTLM authentication never requires the plaintext password — it only needs the NT hash. If you compromise a machine and dump its memory (where Windows caches hashes), you can take that hash and authenticate as that user to other services without ever knowing the actual password.

**Why it works:** it's a design property of NTLM. The hash *is* the credential.

### DCSync

Domain Controllers replicate their AD database with each other (for redundancy). The replication protocol is called **DRSUAPI**. A DC can tell another DC "give me all password hashes for these accounts."

If an attacker gets an account with the right AD permissions (`DS-Replication-Get-Changes`, `DS-Replication-Get-Changes-All`), they can pretend to be a DC and issue those replication requests. The real DC will comply and send back all the password hashes — including the `KRBTGT` account hash, which is the master key of Kerberos.

**Why it works:** it's not exploiting a bug, it's abusing legitimate replication permissions being assigned to the wrong account.

### Golden Ticket

The KDC signs all TGTs using the `KRBTGT` account's password hash. If you get that hash (via DCSync), you can **forge a TGT yourself**, for any user, with any privileges, valid for any duration. No need to authenticate. The DC will accept it because the signature is cryptographically valid.

A forged TGT is called a Golden Ticket. It essentially gives you permanent, god-mode access to the domain even if all passwords are changed — because the ticket is self-signed and the DC has no way to distinguish it from a legitimate one unless they rotate the KRBTGT hash twice.

**Why it works:** the DC trusts the math, not the origin.

### NTLM Relay

Rather than cracking a hash, you intercept an NTLM authentication in progress and **relay it to another service in real time**. The victim authenticates to you (via a spoofed service), and you simultaneously use their authentication to log into a real target.

You never see the password or even the hash — you're just forwarding the handshake. If the victim has admin rights on the target, you get admin access.

**Why it works:** NTLM doesn't bind the authentication session to the specific connection, so it can be relayed.

### Delegation (Unconstrained and RBCD)

Kerberos **delegation** is a feature that lets a service authenticate to another service *on behalf of* a user. Example: a web server that needs to query a database using the user's identity.

**Unconstrained delegation** means the service can impersonate the user to *any* service. When a user connects to a server with unconstrained delegation enabled, their full TGT is sent to that server and stored in memory. If you compromise that server, you steal all those TGTs and can impersonate those users anywhere.

**RBCD (Resource-Based Constrained Delegation)** is a newer, more targeted version. But it has its own abuse path: if you can write the `msDS-AllowedToActOnBehalfOfOtherIdentity` attribute on a computer object, you can configure delegation yourself and gain the ability to impersonate any user to that machine.

**Why it matters:** delegation was designed for convenience. The security assumptions it makes break in attacker-controlled scenarios.

### ACL Abuses (WriteDACL, GenericAll)

Every object in AD (users, groups, computers, OUs) has an **ACL (Access Control List)** that defines who can do what to it. These permissions include things like:
- `GenericAll` — full control over the object
- `WriteDACL` — can modify the object's ACL
- `GenericWrite` — can write to certain attributes
- `ForceChangePassword` — can reset the password without knowing the current one

These are supposed to be tightly controlled. But in large organizations that have grown over years, these permissions accumulate messily. If your compromised low-privilege account has `GenericAll` over a Domain Admin account (because someone misconfigured it years ago), you can reset that admin's password or add yourself to their group.

**Why it works:** pure misconfiguration. The permissions are doing exactly what they're set to do — they're just set wrong.

### GPO Misconfigurations

Group Policy Objects (GPOs) are policies pushed from the DC to machines in the domain — controlling things like password policy, software installs, startup scripts, user rights. They're powerful and applied automatically to all machines in their scope.

If a low-privilege user has write access to a GPO (another ACL misconfiguration), they can modify it to run arbitrary code on every machine that GPO applies to — including DCs.
