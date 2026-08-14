# Lesson 5 --- Ad-Hoc Commands

## Learning objectives

- Run a module once from the command line against an inventory pattern.
- Interpret basic success and connection failures.
- Know when to use an ad-hoc command instead of a playbook.

## Prerequisites

Lessons 1–4. Your inventory must contain one reachable Linux VM and correct SSH credentials.

## Concept

An **ad-hoc command** runs one module directly from the command line, without a playbook file. It is useful for discovery, connection checks, and a carefully scoped one-time operation. It is not a replacement for a version-controlled playbook when an action must be repeatable.

The universal first test is `ansible.builtin.ping`. Despite its name, it is not ICMP ping. It establishes the normal Ansible connection and asks the target to run a tiny module. A successful `pong` proves inventory selection, SSH authentication, remote Python, and basic module execution.

## Mental model

An ad-hoc command is a single checked item spoken aloud; a playbook is the reusable, written checklist. Use the former to inspect and validate, the latter to build the lab.

## Example

Run:

``` bash
ansible web01 -i inventory/lab.ini -m ansible.builtin.ping
```

`web01` is the inventory pattern. `-i inventory/lab.ini` selects your project inventory. `-m` selects a module by its fully qualified collection name (**FQCN**); `ansible.builtin` is the collection shipped with Ansible core. Expected success resembles `web01 | SUCCESS => {"ping": "pong"}`.

To collect a safe fact subset from all Linux hosts later, use:

``` bash
ansible linux -i inventory/lab.ini -m ansible.builtin.setup -a 'filter=ansible_distribution*'
```

`linux` selects the parent group. `-a` passes module arguments. Here `filter=...` limits the facts returned, preventing a huge output. `setup` gathers facts; it is an inspection operation and does not normally change the target.

Do not use `-m ansible.builtin.shell -a 'apk add nginx'` as routine configuration. It loses clear intent and usually does not know whether it changed anything. A later playbook will use package modules.

## Practical exercise

Run the ping module against `web01`, then against your `linux` group. If either fails, read the full error and identify whether it is an inventory address, SSH authentication, or missing-Python problem before changing anything. Finally, run the filtered `setup` command and identify the distribution Ansible detected.

## Expected result

Each reachable Linux target returns `pong`. The fact command reports a distribution family/value suitable for later distribution-aware tasks.

## Common mistakes

- **Using system `ping` as the only test.** ICMP reachability does not prove SSH authentication or Ansible module execution.
- **Adding `-k` unnecessarily.** It asks for an SSH password; prefer SSH keys and let the SSH agent supply them.
- **Targeting `all` during experiments.** `all` includes every inventory host, including future Windows or attacker machines.
- **Ignoring a Python error on Alpine.** SSH is working; install Python during the bootstrap phase rather than changing random Ansible settings.

## Key takeaways

Ad-hoc commands are fast, targeted module invocations. `ansible.builtin.ping` validates the real Linux Ansible path, not network ICMP.

## Next lesson

Lesson 6 explains modules, their arguments, and why their state-aware behavior is central to safe automation.
