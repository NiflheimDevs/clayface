# Lesson 56 — Terraform vs Ansible

## Learning objectives

- State the core responsibility of Terraform and Ansible.
- Identify overlaps and choose one owner.
- Explain why both tools are declarative but not interchangeable.

## Prerequisites

Terraform course knowledge and Lessons 1–55.

## Concept

Terraform reconciles infrastructure resources through providers and maintains state to plan lifecycle changes. Ansible converges reachable systems through modules and usually does not maintain a Terraform-like global state file. Both describe desired results, but their APIs, lifecycle models, and best use differ.

Terraform is normally authoritative for libvirt networks, VM definitions, disks, NICs, and boot-time cloud-init. Ansible is normally authoritative for OS packages, users, files, services, application configuration, and firewall policy inside those VMs. Choose one owner for any overlap.

## Mental model

Terraform assembles connected computers; Ansible configures the computers after they can be managed.

## Example

| Requirement | Owner | Why |
| --- | --- | --- |
| Create `web01` VM and its vNIC | Terraform | libvirt provider resource lifecycle |
| Install nginx | Ansible | OS package desired state |
| Add initial SSH key through cloud-init | Terraform/bootstrap | makes Ansible reachable |
| Nginx virtual-host configuration | Ansible | host/service configuration |

The exact boundary can vary, but never have both tools continually write `/etc/nginx/nginx.conf` or both declare competing network addresses.

## Practical exercise

Review one existing Terraform VM resource. Create a two-column ownership table for every setting it currently applies. Mark any provisioner or cloud-init command that should eventually become Ansible work after reachability.

## Expected result

You can explain a clear handoff point: Terraform completes a reachable VM, then Ansible configures it.

## Common mistakes

- **Using Terraform provisioners for full configuration management.** They are harder to converge, rerun, and test.
- **Expecting Ansible to recreate destroyed VM infrastructure.** It needs a reachable target.
- **Letting tools overwrite one another.** Establish an ownership boundary.

## Key takeaways

Terraform and Ansible complement each other: infrastructure lifecycle first, system configuration next.

## Next lesson

Lesson 57 turns that ownership rule into a staged operating model.
