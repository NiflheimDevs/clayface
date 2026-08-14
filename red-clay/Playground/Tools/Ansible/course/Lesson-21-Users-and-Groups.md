# Lesson 21 — Users and Groups

## Learning objectives

- Create a Linux user and supplementary group declaratively.
- Separate human, service, and automation accounts.
- Preserve least privilege in the lab.

## Prerequisites

Lesson 20 and remote privilege escalation.

## Concept

Accounts are security principals, not merely convenient names. Your Ansible SSH account should be a dedicated, auditable automation user with only the privilege it requires. Human administrators, application services, and automation should normally have separate identities. The `ansible.builtin.user` and `ansible.builtin.group` modules inspect and modify account databases idempotently.

Password hashes, not plaintext passwords, belong in user management. SSH keys are preferable for your administration account; secret handling comes later. On Alpine and Arch, sudo may be absent on minimal images, and the privileged escalation mechanism may instead be `doas` or a bootstrapped root connection.

## Mental model

Users are named keys; groups are permissions shared by several keys. Do not give every key the master-key role.

## Example

``` yaml
- name: Ensure the automation group exists
  ansible.builtin.group:
    name: automation
    state: present

- name: Ensure the Ansible account exists
  ansible.builtin.user:
    name: ansible
    group: automation
    groups: wheel
    append: true
    shell: /bin/sh
    state: present
```

`group` is the primary group. `groups` is a supplementary-group list; `append: true` prevents Ansible from removing other existing supplementary groups. `wheel` is common but not universal; its sudo policy must be separately configured. Alpine's default shell is often BusyBox `ash`; Arch normally has `/bin/sh` linked to Bash.

## Practical exercise

On a disposable VM, declare an `ansible` account and a dedicated primary group. Before applying, inspect how your distribution grants administrative access. Add only the exact group or sudo/doas policy you can justify.

## Expected result

The account exists with the intended group membership, and a second run makes no changes.

## Common mistakes

- **Replacing supplementary groups accidentally.** Omit `append: true` only when you intentionally manage the complete group list.
- **Using plaintext `password:`.** Use a hash protected by Vault if passwords are necessary.
- **Granting passwordless root to every account.** Scope automation privilege to its needs.

## Key takeaways

Account configuration is security configuration. Declare identities and privileges deliberately, not with opaque shell commands.

## Next lesson

Lesson 22 puts the automation account and SSH transport on a secure foundation.
