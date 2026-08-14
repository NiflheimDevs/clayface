# Lesson 64 — Configuring Base Linux Machines

## Learning objectives

- Define a minimal secure base role for Alpine and Arch.
- Separate universal controls from service-specific setup.
- Validate the baseline after rebuilding a VM.

## Prerequisites

Lessons 20–38 and 63.

## Concept

`base_linux` should establish only universal host properties: automation account/key, selected baseline packages, time/hostname policy, logging/updates policy, management firewall access, and optionally common monitoring agents. It should not install nginx because database and DNS hosts do not need it. Distribution differences belong in documented variables or specific task files.

Build it incrementally on a disposable VM, use check/diff cautiously, and make a verification checklist part of the role README.

## Mental model

The base role is a trustworthy OS foundation, not a preloaded application image.

## Example

``` yaml
# roles/base_linux/tasks/main.yml
- name: Import package tasks for detected platform
  ansible.builtin.include_tasks: "packages-{{ ansible_facts['distribution'] }}.yml"

- name: Ensure automation account exists
  ansible.builtin.import_tasks: automation-user.yml
```

This illustrates a deliberate split: a known common task file is static, while package implementation can be chosen by discovered distribution. Add an assertion/clear failure for unsupported distributions before relying on the dynamic filename.

## Practical exercise

Write a baseline acceptance checklist: SSH key authentication, correct user privilege, installed tools, firewall management access, and OS-specific service manager. Implement only one additional checklist item at a time.

## Expected result

A fresh supported VM becomes Ansible-ready and meets the same baseline after repeat application.

## Common mistakes

- **Adding a service role's dependencies to base.** It bloats every VM.
- **Forgetting the Alpine/Arch init difference.** Test both before claiming support.
- **Making host-specific policy a default.** Use group/host data.

## Key takeaways

A narrow, validated base role makes every later role safer and easier to reason about.

## Next lesson

Lesson 65 extracts repeated service patterns into reusable server roles.
