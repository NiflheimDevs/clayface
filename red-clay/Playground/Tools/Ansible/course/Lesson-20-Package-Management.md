# Lesson 20 — Package Management

## Learning objectives

- Install packages declaratively on Linux hosts.
- Choose portable or distribution-specific package modules.
- Avoid uncontrolled package upgrades.

## Prerequisites

Lessons 1–19; a disposable Linux VM with privilege escalation available.

## Concept

Package management is operating-system specific: Alpine uses `apk`, Arch uses `pacman`, Debian/Ubuntu uses `apt`, and RHEL-family systems use `dnf`. Ansible's `ansible.builtin.package` provides a common `name`/`state` interface, but only for operations all package managers share. Use the specific module when cache refreshes, repositories, versions, or package manager behavior matters.

`state: present` means installed, without demanding the newest version. `state: latest` permits updates and should be an intentional maintenance policy, not a default in a reproducible lab. Package installation requires root privileges, introduced here with `become: true` and explained deeply in Lessons 46–47.

## Mental model

Say “this software must be installed,” not “replay this package-manager command.” The module checks the installed database before changing it.

## Example

``` yaml
---
- name: Install base packages
  hosts: linux
  become: true
  tasks:
    - name: Ensure diagnostic tools are installed
      ansible.builtin.package:
        name:
          - curl
          - ca-certificates
        state: present
```

`become: true` asks Ansible to elevate the remote account for the play. `name` accepts a YAML list, avoiding a shell loop. Verify package names on every distribution you target; portable module syntax does not guarantee portable package names.

## Practical exercise

Add a `base`-tagged package task for two tools available on your target OS. Run it with `--check`, then normally, then once more. Record which package manager Ansible used from verbose output only if you need to diagnose it.

## Expected result

The initial run installs missing packages and reports `changed`; the repeated run reports `ok`.

## Common mistakes

- **Using `latest` to mean installed.** It can silently change your lab months later.
- **Forgetting `become`.** The failure is a permissions issue, not a package-module issue.
- **Assuming `package` can add repositories.** Use the platform-specific module or explicit repository configuration.

## Key takeaways

Manage packages as desired state and be explicit about distribution differences and update policy.

## Next lesson

Lesson 21 creates the administration accounts that run Ansible safely.
