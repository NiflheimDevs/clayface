# Lesson 71 — Preparing Machines for Active Directory

## Learning objectives

- Identify prerequisites for AD-dependent clients/services.
- Explain why DNS and time are critical to AD.
- Prepare Linux machines without pretending domain join is universal.

## Prerequisites

Lessons 63–70 and basic Active Directory concepts.

## Concept

Active Directory relies heavily on correct DNS, time synchronization (Kerberos), reachable domain controllers, service ports, and controlled credentials. Prepare clients by configuring trusted DNS servers, hostname policy, time synchronization, and firewall flows first. Linux integration may use realmd/SSSD/Kerberos/Samba depending on distribution and lab goal; Windows domain join uses Windows-native modules later.

Do not embed a domain-join password in a normal variable or join every host blindly. Domain joins alter identity and trust state and must be idempotently detected, scoped, and protected with Vault/no_log.

## Mental model

AD is an identity ecosystem. DNS tells clients where its authorities are, time proves tickets are current, and network policy permits the conversation.

## Example

Preparation checklist:

``` text
- Client resolves AD DNS zones using intended internal DNS.
- Client and domain controller time are synchronized.
- Required Kerberos/LDAP/SMB/RPC flows are allowed according to your design.
- Computer name is stable and unique.
- Join credential is vaulted and minimally privileged.
```

This is intentionally not a one-line domain-join task: commands, packages, and policy differ substantially between Alpine, Arch, Debian, RHEL, and Windows. Alpine may be unsuitable for some full AD client stacks without extra work; choose a supported client image for realistic integration.

## Practical exercise

Create an AD readiness checklist for one future `client01`. Test DNS resolution and time from a Linux VM once a DC/DNS service exists. Do not attempt a join until those checks pass.

## Expected result

You understand why AD failures often originate in DNS/time/network rather than the join command itself.

## Common mistakes

- **Using public DNS on a domain member.** It cannot locate AD service records.
- **Ignoring clock skew.** Kerberos authentication can fail.
- **Testing a Linux join on an unsupported/minimal image first.** Use a deliberate platform choice.

## Key takeaways

Prepare DNS, time, names, network flows, and credential handling before any domain-join automation.

## Next lesson

Lesson 72 introduces the different transport and module model for Windows.
