# Lesson 59 — Generating Inventories from Terraform

## Learning objectives

- Generate a deterministic static inventory from Terraform outputs.
- Validate generated inventory before Ansible runs.
- Understand generated-file ownership.

## Prerequisites

Lessons 4, 53, and 58.

## Concept

For a small libvirt lab, generating an inventory file after `terraform apply` is often clearer than a live dynamic plugin. Terraform owns the generated file; humans should not hand-edit it because the next generation overwrites manual edits. Keep human-owned group variables and Ansible configuration separate.

Use a script/template in the Terraform project or CI pipeline to render valid INI/YAML inventory from structured outputs. Validate it with `ansible-inventory --graph` before calling a playbook. If the generator is custom code, test it with sample JSON, including empty groups and changed addresses.

## Mental model

Terraform writes the address book from its build ledger; Ansible reads it. Notes about how to configure people in the address book live elsewhere.

## Example

Generated `inventory/generated.ini` can look like:

``` ini
[webservers]
web01 ansible_host=10.10.10.11 ansible_user=ansible

[linux:children]
webservers
```

Run:

``` bash
ansible-inventory -i inventory/generated.ini --graph
ansible linux -i inventory/generated.ini --list-hosts
```

The first validates group shape; the second validates a target pattern without SSH. Add the generated file to `.gitignore` if it represents ephemeral environment state, while committing its generation template/code.

## Practical exercise

Create a manually generated sample inventory matching one Terraform output. Mark it generated in a comment, inspect it with both commands, and decide the exact command/stage that will replace it after Terraform apply.

## Expected result

Ansible sees Terraform-created host names, addresses, and groups without manual copying.

## Common mistakes

- **Editing generated inventory to add configuration variables.** Store those in `group_vars`/`host_vars` instead.
- **Generating malformed YAML/INI with unescaped values.** Validate every output.
- **Running Ansible before regeneration after a rebuild.** It may target stale IP addresses.

## Key takeaways

Generated inventory is a controlled Terraform-to-Ansible handoff and should be validated and treated as generated state.

## Next lesson

Lesson 60 applies a role to Terraform-created VMs with readiness checks.
