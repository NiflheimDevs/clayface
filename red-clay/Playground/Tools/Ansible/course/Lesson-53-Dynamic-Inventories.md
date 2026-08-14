# Lesson 53 — Dynamic Inventories

## Learning objectives

- Explain why dynamic inventory exists.
- Inspect a dynamic source before running a play.
- Compare Terraform-generated files with inventory plugins.

## Prerequisites

Lessons 4 and 37; Terraform familiarity is useful.

## Concept

Dynamic inventory obtains hosts from an external system at run time, commonly a cloud provider, virtualization API, CMDB, or custom script/plugin. It is valuable when VM addresses and membership change frequently. Its output must still be treated as targeting policy: inspect it, authenticate it securely, and understand its grouping rules.

A Terraform-generated static inventory is often simpler for a small libvirt lab because Terraform is already authoritative for VM addresses. A dynamic plugin is appropriate when the source API is the authoritative discovery system and its collection/plugin is supported and pinned.

## Mental model

Static inventory is a checked-in address book. Dynamic inventory is a live directory lookup. Both can be wrong; live is not automatically safer.

## Example

Inventory sources can be passed as directories:

``` bash
ansible-inventory -i inventory/lab/ --graph
```

`-i` may identify a file or directory. Before executing, use `--graph` and `ansible all --list-hosts` to confirm current discovery. Never point production-like automation at an uninspected dynamic source merely because an API query succeeded.

## Practical exercise

Write down whether your current Terraform setup has stable VM names and IP outputs. Decide whether a generated static inventory or a libvirt/dynamic plugin is simpler today, and state the data owner and refresh moment for your choice.

## Expected result

You can justify an inventory source based on authoritative data and lab size, not fashion.

## Common mistakes

- **Treating dynamic inventory as an authorization system.** It discovers hosts; your play targeting and credentials still control impact.
- **Generating inventory with secrets embedded.** Use secure connection mechanisms.
- **Failing to inspect changed membership.** A new VM may suddenly match a broad group.

## Key takeaways

Dynamic inventory is live discovery, while Terraform-generated inventory is a useful controlled handoff. Both require inspection.

## Next lesson

Lesson 54 manages collection dependencies at a more production-like level.
