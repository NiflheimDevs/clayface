# Lesson 62 — Designing a Clean Terraform + Ansible Repository

## Learning objectives

- Organize infrastructure and configuration code with a clear interface.
- Keep generated state and secrets out of source control.
- Define a readable deployment entry point.

## Prerequisites

Lessons 31 and 56–61.

## Concept

A clean repository makes tool ownership apparent. Terraform directories hold provider, VM, network, module, and output definitions. Ansible directories hold inventory sources/generated outputs, group data, playbooks, roles, and collection requirements. Shared documentation defines the bootstrap contract and deployment stages. Generated inventories and Terraform state should normally not be edited by hand.

Monorepo versus separate repositories is a workflow choice. A single Digital Twin repository often helps students keep topology and configuration changes reviewed together; separate repos can fit independently released infrastructure. In either case, the interface must be versioned and documented.

## Mental model

Two workshops share a signed handoff shelf. Each owns its tools; the handoff manifest is explicit.

## Example

``` text
digital-twin/
├── terraform/
│   ├── modules/
│   ├── environments/lab/
│   └── outputs.tf
├── ansible/
│   ├── inventory/generated/
│   ├── group_vars/
│   ├── playbooks/
│   └── roles/
├── docs/bootstrap-contract.md
└── Makefile
```

The `Makefile` may offer explicit targets such as `infra-apply`, `inventory`, `configure`, and `verify`, but each should expose its real command and preserve failure status. Do not make it a hiding place for credentials.

## Practical exercise

Sketch your final repository tree and label the owner of every top-level directory. Add `.gitignore` entries for Terraform state, generated inventory if appropriate, vault-password files, and private keys; do not ignore the templates and requirements needed to rebuild them.

## Expected result

A reader can locate infrastructure, automation, generated state, protected secrets, and the documented stage handoff.

## Common mistakes

- **Committing `.tfstate` or private keys.** Treat both as sensitive state.
- **Hiding output contract in a script.** Document fields and lifecycle.
- **Mixing Terraform resource files into Ansible role folders.** It obscures ownership.

## Key takeaways

Repository design preserves the Terraform/Ansible boundary and makes full lab rebuilds explainable.

## Next lesson

Lesson 63 applies this design to the Enterprise Digital Twin architecture.
