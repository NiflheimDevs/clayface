# Lesson 6 --- Modules

## Learning objectives

- Define an Ansible module and its desired-state arguments.
- Choose a module instead of `command` or `shell` where possible.
- Read module documentation locally.

## Prerequisites

Lessons 1–5. You should have successfully run `ansible.builtin.ping`.

## Concept

A **module** is a focused unit of Ansible code that manages or queries a resource: a package, user, file, service, firewall rule, or API object. Its arguments state the desired result. Modules return structured results, including `changed`, `failed`, and useful details.

`ansible.builtin.command` executes a program without a shell; it does not interpret pipes, redirects, or variables. `ansible.builtin.shell` executes through a shell and supports those features, but adds quoting, injection, portability, and idempotency risks. Use them only when no purpose-built module can express the operation, then explicitly define safe change/failure conditions later in the course.

Read the documentation before using a module. The module's `state` values, platform support, check-mode behavior, and return values determine whether it models your intent correctly.

## Mental model

A module is a specialist who understands one kind of object. A shell command is asking a general assistant to follow a fragile sentence. Prefer the specialist when one exists.

## Example

Inspect the package module with:

``` bash
ansible-doc ansible.builtin.package
```

`ansible-doc` displays installed documentation; it needs no network or target host. The FQCN avoids ambiguity with a collection module of a similar name.

This future task expresses a package state:

``` yaml
- name: Ensure curl is installed
  ansible.builtin.package:
    name: curl
    state: present
```

`present` means installed. The module determines whether the package already exists. The first run may be `changed`; the second should be `ok`. It is more declarative than `ansible.builtin.shell: pacman -S --noconfirm curl`, which always attempts an installation and is Arch-specific.

Use the distribution-specific `ansible.builtin.pacman`, `ansible.builtin.apk`, `ansible.builtin.apt`, or `ansible.builtin.dnf` module when you need package-manager-specific behavior such as cache updates. `package` is useful only for the portable common subset.

## Practical exercise

Run `ansible-doc` for `ansible.builtin.package`, `ansible.builtin.user`, and `ansible.builtin.command`. For each, find its required arguments and whether it reports change prediction in check mode. Write one desired-state sentence that would use each of the first two modules.

## Expected result

You can explain why a user module can safely manage an existing account while a bare `useradd` command cannot automatically make the same guarantee.

## Common mistakes

- **Assuming every module works on every OS.** Read platform notes, especially for services and firewalls.
- **Using `latest` casually.** It may change a machine on every package update; pin or choose `present` when reproducibility matters.
- **Writing shell pipelines where a module exists.** This obscures intent and makes reporting less trustworthy.

## Key takeaways

Modules are Ansible's state-aware interface to systems. Prefer FQCNs and documentation-driven arguments over guessed shell commands.

## Next lesson

Lesson 7 stores several tasks in a repeatable playbook.
