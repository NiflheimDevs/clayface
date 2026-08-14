# Lesson 46 — `become`

## Learning objectives

- Explain remote privilege escalation in Ansible.
- Apply `become` at the narrowest suitable scope.
- Distinguish login user from effective task user.

## Prerequisites

Lessons 20–25 and SSH access with an authorized escalation method.

## Concept

Ansible first connects as `ansible_user` (or a configured SSH user). `become: true` asks it to execute a task or play as another user, normally root, through an escalation method such as sudo or doas. It does not make SSH login itself root and does not bypass the target's policy.

Place `become: true` at play level when nearly every task needs it, or task/block level when only a small section does. The least scope that is clear is usually best. `become_user` defaults to root but can select a service account for specific tasks.

## Mental model

SSH gets Ansible through the building entrance; `become` is the controlled key checkout for a room requiring higher access.

## Example

``` yaml
- name: Create root-owned configuration directory
  ansible.builtin.file:
    path: /etc/digital-twin
    state: directory
    mode: "0755"
  become: true
```

This task-level setting avoids elevating unrelated inspection tasks. If sudo needs a password, run with `--ask-become-pass` for an interactive lab test; do not place the password in inventory. Noninteractive automation should use a deliberately designed privilege policy.

## Practical exercise

Run one read-only task without become and one directory task with become. Verify the effective owner of the created directory. Inspect your target's `/etc/sudoers` or doas configuration before broadening access.

## Expected result

Only the privileged task runs as root; SSH identity remains the automation account.

## Common mistakes

- **Setting `remote_user: root` just to manage files.** It loses account separation.
- **Using become globally without review.** Every module then has more power than it may need.
- **Confusing `--ask-pass` with `--ask-become-pass`.** One is SSH login, the other escalation.

## Key takeaways

`become` is controlled remote escalation, separate from SSH authentication. Scope it intentionally.

## Next lesson

Lesson 47 designs escalation policy for the lab.
