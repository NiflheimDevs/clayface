# Lesson 32 — Roles

## Learning objectives

- Explain roles as reusable configuration units.
- Create a minimal `base_linux` role.
- Apply a role from a play.

## Prerequisites

Lesson 31 and prior task/module knowledge.

## Concept

A **role** packages related tasks, defaults, handlers, templates, files, and metadata under one capability. It is not simply a directory tree: its purpose is to make a repeatable promise such as “this host receives a secure Linux baseline” or “this host is an nginx web server.” A role should have a focused interface and avoid assuming every host is identical.

## Mental model

A role is a reusable appliance specification. A play decides which machines receive it; role variables customize the installation.

## Example

``` text
roles/base_linux/
├── defaults/main.yml
├── handlers/main.yml
├── tasks/main.yml
├── templates/
└── vars/main.yml
```

`roles/base_linux/tasks/main.yml` begins with a task list—no play header:

``` yaml
- name: Ensure base diagnostic packages are installed
  ansible.builtin.package:
    name: "{{ base_packages }}"
    state: present
```

Apply it from `playbooks/base.yml`:

``` yaml
- name: Configure Linux baseline
  hosts: linux
  become: true
  roles:
    - base_linux
```

`roles:` is a list at play level. Ansible finds `roles/base_linux` relative to the configured roles path; task, handler, template, and default locations gain predictable role-relative lookup behavior.

## Practical exercise

Create `base_linux` and move only the package task into it. Define no service logic yet. Apply it from a dedicated base play and prove idempotency before moving more tasks.

## Expected result

The role executes its task on the `linux` group. The playbook stays a readable statement of which capabilities each group receives.

## Common mistakes

- **Putting a `hosts:` header in a role task file.** Roles contain task lists; plays choose hosts.
- **Making one “common” role configure every application.** Keep roles coherent and composable.
- **Hiding mandatory behavior in undocumented defaults.** Define the role interface next lesson.

## Key takeaways

Roles organize reusable behavior; plays compose them for selected inventory groups.

## Next lesson

Lesson 33 designs role defaults and variables without losing control of precedence.
