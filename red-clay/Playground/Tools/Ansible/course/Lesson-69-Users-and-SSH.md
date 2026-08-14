# Lesson 69 — Configuring Users and SSH

## Learning objectives

- Apply a consistent enterprise account/SSH policy.
- Distinguish automation, human-admin, and service identities.
- Stage SSH hardening without lockout.

## Prerequisites

Lessons 21–22 and 39–43.

## Concept

At lab scale, account policy must remain explicit: which human administrators exist, how access is granted/revoked, which automation key is accepted, whether root login/password authentication are allowed, and how emergency console access works. Roles can implement a common policy while group variables select appropriate exceptions. Keep attacker VMs distinct: they may require intentionally different accounts and should never inherit broad production-like administration keys by accident.

## Mental model

Identity configuration is an access map for the entire enterprise, not a convenience task for SSH.

## Example

``` yaml
linux_admin_users:
  - name: labadmin
    groups: [wheel]
    ssh_keys:
      - "ssh-ed25519 AAAA... labadmin"
```

This is data, not a full task implementation. A role can loop over validated dictionaries to create accounts and keys. The public-key text is abbreviated; never invent/fabricate keys. For real policy, separate individual access data from shared role code and use reviewable changes.

## Practical exercise

Write an access matrix for `web01`, `db01`, `kali01`, and the control node: allowed identities, authentication method, elevation, and emergency access. Implement only the automation account on a disposable target and test a new session before hardening.

## Expected result

You can revoke or add an identity through reviewed data, while SSH remains reachable through a tested path.

## Common mistakes

- **One shared account for every person.** Accountability and revocation are poor.
- **Pushing personal keys to attacker hosts automatically.** Scope groups carefully.
- **Hardening SSH before validating the new account.** Preserve console recovery.

## Key takeaways

Enterprise SSH automation is identity and privilege policy, applied in a lockout-safe order.

## Next lesson

Lesson 70 turns topology requirements into network and service configuration.
