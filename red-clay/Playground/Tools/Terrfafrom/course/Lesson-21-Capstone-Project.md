# Lesson 21 --- Capstone: A Portable Multi-Host Lab

## Goal

Design and implement a repeatable lab deployment that can create a small service topology across two or more virtualization targets, then configure the guests with Ansible. Your first implementation may use libvirt/KVM; the design must make the platform-specific parts replaceable.

------------------------------------------------------------------------

## Suggested Topology

Build a minimal, useful environment:

| Role | Example responsibility | Initial size |
| --- | --- | --- |
| `dns01` | Internal DNS or inventory endpoint | 1 vCPU, 1 GiB |
| `web01` | Simple application/service | 2 vCPU, 2 GiB |
| `monitor01` | Monitoring or logging practice | 2 vCPU, 2 GiB |

Place at least two guests on different targets. Choose a network design that genuinely permits the required communication; do not assume same-named local virtual networks span physical hosts.

## Required Deliverables

1. A repository README with prerequisites, topology, provider choice, state location, safe commands, and recovery notes.
2. A root/module layout that separates portable machine intent from provider-specific implementation.
3. Typed variables, validation, locals, and outputs that provide a clear interface.
4. Separate state boundaries for shared infrastructure and per-target workloads, or a written justification for another layout.
5. Version constraints and a committed dependency lock file.
6. A secret-safe authentication method and `.gitignore`.
7. An Ansible inventory handoff through DNS, IPAM, a structured output, or a documented dynamic inventory.
8. A tested create, no-change reapply, intentional update, and destroy of **disposable** resources.

## Recommended Build Sequence

1. Draw networking, host placement, and ownership boundaries.
2. Create one disposable VM on one target manually with Terraform.
3. Extract the repeatable VM pattern into a provider-specific module.
4. Define platform-neutral VM intent in a root environment.
5. Add a second target and deploy an independent VM there.
6. Establish a dependable addressing/inventory method.
7. Use Ansible to configure each role.
8. Add state protection, documentation, and a recovery rehearsal.
9. Only then add more roles, storage, firewall policy, or automated pipeline execution.

## Acceptance Tests

Run these tests and record the result in your README or lab notes:

```bash
terraform fmt -check -recursive
terraform validate
terraform plan
terraform apply
terraform plan             # expected: no changes
terraform destroy          # disposable environment only
```

Also test that Ansible can reach the declared hosts and converge twice without changing a healthy system. Verify one planned change, such as a tag/metadata or harmless VM setting, and understand whether your provider updates or replaces it.

## Migration Exercise: KVM to Proxmox

Do not attempt a blind state migration. Instead:

1. preserve the platform-neutral environment input contract;
2. implement and test a Proxmox-specific module in a disposable environment;
3. create new workloads from known images/templates;
4. migrate application data using backups or replication, not Terraform state;
5. update DNS/load-balancing/inventory deliberately;
6. retire old resources only after validation and backup.

Terraform state records provider-specific identities, so rebuilding in the new platform is usually safer than trying to make a libvirt domain become a Proxmox VM in place.

## Final Reflection

You are ready to extend the lab when you can answer these questions for every resource: Who owns it? Which state records it? How is it reached? Where is its data protected? What does a replacement do? How would you rebuild it on a different platform?

If those answers are clear, Terraform is giving you the durable automation foundation you wanted—not just a faster way to click “Create VM.”
