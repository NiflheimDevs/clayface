# Lesson 58 — Terraform Outputs as Ansible Inputs

## Learning objectives

- Design Terraform outputs that are useful but non-secret.
- Convert VM identifiers/IPs into Ansible connection data.
- Avoid coupling Ansible to Terraform state internals.

## Prerequisites

Terraform outputs knowledge and Lesson 57.

## Concept

Terraform outputs are an intentional interface. Output stable VM names, management addresses, role/group data, and perhaps SSH user—not state-file internals or secrets. Ansible needs inventory-shaped data, not a direct dependency on every resource attribute. JSON output (`terraform output -json`) is suitable for a controlled generator or inventory plugin.

Sensitive Terraform outputs are still visible to state and authorized operators; marking an output sensitive only redacts normal CLI display. Do not casually feed secrets into generated inventories.

## Mental model

Terraform outputs are the handoff manifest between the construction crew and operations, not a dump of every construction record.

## Example

Conceptual Terraform output:

``` hcl
output "ansible_hosts" {
  value = {
    web01 = { address = "10.10.10.11", groups = ["webservers", "linux"] }
  }
}
```

An inventory generator can turn `address` into `ansible_host` and `groups` into inventory membership. Use real calculated addresses in Terraform; the literal here explains the shape. The stable key `web01` becomes the Ansible inventory identity.

## Practical exercise

Add or sketch one non-sensitive `ansible_hosts` Terraform output for your existing VM. Decide whether hostname, address, and group are Terraform-owned facts or Ansible desired configuration. Do not expose private keys or passwords.

## Expected result

The output gives Ansible exactly enough connection/topology information to target the VM.

## Common mistakes

- **Parsing human-oriented `terraform output` text.** Use JSON or a documented generated file format.
- **Using ephemeral resource IDs as inventory names.** Prefer stable logical names.
- **Publishing sensitive outputs.** Redaction is not a complete secret boundary.

## Key takeaways

Design a small, stable output interface that maps infrastructure facts into inventory data.

## Next lesson

Lesson 59 generates or discovers inventories from that interface.
