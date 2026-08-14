# Lesson 36 — Reusable Roles

## Learning objectives

- Define a role interface, assumptions, and outputs.
- Make roles portable across appropriate Linux hosts.
- Test a role with a small, disposable inventory.

## Prerequisites

Lessons 31–35.

## Concept

A reusable role has a narrow responsibility, documented defaults, explicit platform support, predictable handlers, and no hidden dependence on a particular lab hostname. Reuse does not mean universal abstraction: a good `web_nginx` role may support Alpine and Arch but intentionally refuse Windows or a different web stack.

Document required variables, optional variables and defaults, created users/files/ports, dependencies, tested distributions, and verification steps in `roles/<name>/README.md`. A tiny test inventory prevents your only test from being a complete enterprise deployment.

## Mental model

A reusable role is an appliance with an installation guide, supported voltage range, controls, and acceptance test—not a pile of copied tasks.

## Example

Role interface excerpt:

``` markdown
## Inputs

- `web_nginx_listen_port` (default: `80`): TCP port for the virtual host.
- `web_nginx_server_name` (required): DNS name written to the configuration.

## Supported platforms

Arch Linux and Alpine Linux; service and package names are selected by platform variables.
```

Namespacing protects the role from collision. A required setting should fail clearly with `ansible.builtin.assert` rather than producing an incomplete template later.

## Practical exercise

Write a README for `base_linux`: purpose, default packages, supported distributions, privileges required, and a manual verification command. Ask whether every statement is true on a clean VM.

## Expected result

A future you can apply the role without reading all implementation tasks, and knows what must be supplied.

## Common mistakes

- **Hard-coding `web01` in a role.** Put topology data in inventory variables.
- **Claiming portability without testing package/service names.** State tested scope honestly.
- **Including secret values in examples.** Use placeholders and Vault references.

## Key takeaways

Reuse comes from a clear contract, not merely a standard directory structure.

## Next lesson

Lesson 37 grows the inventory into environment and functional layers.
