# Lesson 75 — Capstone: Enterprise-Scale Digital Twin

## Learning objectives

- Design and implement a complete Digital Twin automation repository.
- Demonstrate safe Terraform-to-Ansible handoff.
- Verify idempotency, security controls, and rebuild reproducibility.

## Prerequisites

Lessons 1–74.

## Concept

The capstone combines every course concept into one coherent lab. Start small and expand only after each layer works. The goal is not maximum VM count; it is a documented, reproducible enterprise-shaped environment where Terraform owns infrastructure and Ansible owns machine/service configuration.

## Mental model

This is your systems-integration exercise: topology, identities, services, policy, automation, and verification must agree.

## Example

Target fictional inventory:

``` ini
[webservers]
web01
web02
[databases]
db01
[dns]
dns01
[domain_controllers]
dc01
[clients]
client01
client02
[attackers]
kali01
```

Suggested composition: Terraform provisions segregated libvirt networks and VM identities; a generated inventory hands off current management addresses; `base_linux` configures Linux; service roles configure DNS, database, internal service, and web tiers; Windows roles manage DC/client prerequisites; firewall roles implement the documented traffic matrix. `kali01` is intentionally isolated and receives only roles you explicitly authorize.

## Practical exercise

Build the capstone in milestones:

1. One Terraform-created `web01` plus generated inventory and Ansible baseline.
2. Add `db01`, a database role, least-privilege firewall flow, and client-path health check.
3. Add `dns01` and replace hard-coded service addresses with DNS names.
4. Add Windows/DC/client stages only after DNS/time/WinRM design is ready.
5. Add segmented attacker simulation and document boundaries/authorization.
6. Perform a controlled rebuild of the disposable environment.

For every milestone, commit code, run syntax/inventory checks, apply to a limit first, verify services, rerun for idempotency, and update documentation. Do not immediately use solutions from a template; use your role contracts and evidence.

## Expected result

Your repository can provision and configure the lab from documented source, with a clear Terraform/Ansible boundary, protected secrets, scoped access, testable network flows, and repeatable results.

## Common mistakes

- **Adding every enterprise component before one end-to-end slice works.** Grow by verified milestones.
- **Treating an attack lab as permission to weaken your home network.** Keep segmentation and authorization explicit.
- **Declaring completion without a rebuild/idempotency test.** Reproducibility needs evidence.

## Key takeaways

You can now use Ansible as a deliberate configuration-management system for an Enterprise Digital Twin: state-aware, secure, structured, and integrated cleanly with Terraform.

## Next lesson

The course is complete. Revisit any lesson as you implement each capstone milestone, then evolve roles through small reviewed changes and repeatable rebuild tests.
