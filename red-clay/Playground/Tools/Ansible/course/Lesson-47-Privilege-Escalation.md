# Lesson 47 — Privilege Escalation

## Learning objectives

- Choose an escalation method compatible with target Linux systems.
- Design a least-privilege automation policy.
- Diagnose common become failures.

## Prerequisites

Lesson 46 and basic sudo/doas administration knowledge.

## Concept

`become_method` selects an escalation mechanism. `sudo` is common on Arch, Debian/Ubuntu, and RHEL systems; Alpine often uses `doas` on minimal installations, though sudo can be installed. The connection account needs policy permission on the target, and noninteractive execution must not unexpectedly require a TTY or password.

For early bootstrap, broad privilege can be acceptable on disposable VMs. Mature roles should minimize privileges where practical and document their requirement. A restrictive command-only sudo policy can be difficult because Ansible modules execute temporary programs; test it thoroughly rather than guessing which command must be permitted.

## Mental model

Privilege policy is a contract between the target OS and automation. Ansible asks; the OS decides according to configured rules.

## Example

``` yaml
- name: Configure Alpine base host
  hosts: alpine
  become: true
  become_method: doas
  roles:
    - base_linux
```

`become_method: doas` is play-level because all role tasks need it in this example. The target's `/etc/doas.conf` must permit the SSH user; Ansible cannot escalate merely because YAML requests it. Use a separate group such as `alpine` only when inventory facts/organization genuinely warrant it.

## Practical exercise

For one target, document SSH user, escalation method, target user, and whether a password/TTY is required. Intentionally run `ansible ... -b` once without permission only if it is safe, then interpret the exact failure instead of disabling security controls.

## Expected result

You can prove the automation account escalates only through declared OS policy.

## Common mistakes

- **Editing sudoers without `visudo` validation.** A syntax error can remove administrative access.
- **Assuming sudo exists on Alpine.** Verify package and policy.
- **Solving failures with passwordless root everywhere.** Start with targeted policy and document the tradeoff.

## Key takeaways

Privilege escalation is target-side policy plus Ansible settings. Platform and security design both matter.

## Next lesson

Lesson 48 groups related work with blocks, rescue, and always sections.
