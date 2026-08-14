# Lesson 63 — Designing the Ansible Architecture for the Lab

## Learning objectives

- Map lab machine roles to Ansible roles and inventory groups.
- Define management and service dependency boundaries.
- Design a safe site-play composition.

## Prerequisites

Lessons 31–62.

## Concept

Your Digital Twin should represent enterprise functions without turning each VM into a bespoke snowflake. Start with inventory groups—`webservers`, `databases`, `dns`, `domain_controllers`, `clients`, `attackers`—then compose focused roles: `base_linux`, `web_nginx`, `database_postgresql`, `dns_server`, and later Windows roles. Base roles establish common controls; service roles manage only their service.

Network placement is Terraform-owned topology, but Ansible roles must know service dependencies: DNS must resolve before domain join, databases before applications, and management access must remain available during firewall changes.

## Mental model

Inventory describes who each machine is in the enterprise; roles describe the repeatable capabilities it receives; site plays assemble them.

## Example

``` yaml
- name: Configure web tier
  hosts: webservers
  become: true
  roles:
    - base_linux
    - web_nginx

- name: Configure database tier
  hosts: databases
  become: true
  roles:
    - base_linux
    - database_postgresql
```

The role order is explicit and reviews the intended composition. Do not put attacker tooling in `base_linux`; attacker VMs have a distinct purpose and risk profile.

## Practical exercise

Draw your desired inventory groups and assign each one zero or more roles. Mark service dependencies and management access paths. Start with only `web01` if that is all you have.

## Expected result

Your architecture can grow from one VM to the full fictional inventory without rewriting base assumptions.

## Common mistakes

- **One role per hostname.** Reuse disappears.
- **One giant `site.yml` full of implementation tasks.** Roles become impossible to test independently.
- **Treating attackers as normal managed servers.** Keep their credentials and allowed automation purpose constrained.

## Key takeaways

Architecture starts with functional groups and focused roles, then composes them into explicit deployment plays.

## Next lesson

Lesson 64 implements the common Linux baseline as a production-like role.
