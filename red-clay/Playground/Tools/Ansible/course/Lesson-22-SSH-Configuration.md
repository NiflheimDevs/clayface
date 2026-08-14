# Lesson 22 — SSH Configuration

## Learning objectives

- Install an SSH public key without overwriting authorized access.
- Understand SSH hardening order and lockout risk.
- Validate configuration before reloading SSH.

## Prerequisites

Lesson 21. Keep a console or a second SSH session available during testing.

## Concept

SSH is Ansible's normal Linux control channel, so hardening it can lock out the automation that applies the hardening. Establish the automation user and its public key first, verify a fresh SSH login, then change daemon policy in small reviewed stages. SSH configuration paths and service names vary: Alpine's OpenRC service is commonly `sshd`; Arch uses `sshd.service` through systemd.

`ansible.posix.authorized_key` manages individual public-key lines. It is supplied by the `ansible.posix` collection, which you will install before using it. File changes should notify a syntax validation/reload handler only after a known-good configuration is in place.

## Mental model

SSH hardening is changing the bridge you are standing on. Build and test an alternate path before closing an old one.

## Example

After installing `ansible.posix`, an intended key task is:

``` yaml
- name: Authorize the control-node key for Ansible
  ansible.posix.authorized_key:
    user: ansible
    key: "{{ lookup('ansible.builtin.file', 'keys/ansible.pub') }}"
    state: present
```

`user` owns the target account. `lookup('ansible.builtin.file', ...)` reads the public key on the control node, not the managed host. `state: present` adds it if absent. Public keys are not secrets, but their corresponding private keys are.

## Practical exercise

Create a non-production key pair if you do not have an appropriate dedicated automation key. Install only its public half for the new user, then prove a new SSH login works before designing any root-login or password-authentication changes.

## Expected result

The public key is present exactly once and new SSH authentication succeeds for the automation account.

## Common mistakes

- **Using `exclusive: true` without understanding it.** It can remove other authorized keys.
- **Disabling password authentication before key testing.** This causes avoidable lockouts.
- **Copying an unvalidated `sshd_config` wholesale.** Use `sshd -t` before reload later.

## Key takeaways

Treat SSH changes as access-critical. Establish, test, and preserve a safe management path first.

## Next lesson

Lesson 23 manages directories and files safely.
