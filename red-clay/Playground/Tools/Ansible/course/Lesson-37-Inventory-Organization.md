# Lesson 37 — Inventory Organization

## Learning objectives

- Organize hosts by function and environment.
- Use parent groups without accidental target expansion.
- Inspect final inventory before running changes.

## Prerequisites

Lessons 4 and 31–36.

## Concept

Inventory organization expresses two useful dimensions: function (`webservers`, `databases`, `dns`) and environment/lifecycle (`lab`, `production`, `linux`, `windows`). A host can belong to multiple groups. Parent groups make shared selection easy, but a broad parent can turn a narrow command into a large blast radius.

For the Digital Twin, inventories may later separate environments into directories, while group structure remains consistent. Inventory should identify targets and connection details, not become a dumping ground for application policy.

## Mental model

Groups are overlapping labels, not folders. A server can be both `webservers` and `linux` and `lab`.

## Example

``` ini
[webservers]
web01
web02

[databases]
db01

[linux:children]
webservers
databases

[lab:children]
linux
```

`linux:children` includes group membership, not copies of host definitions. Before applying a broad play, run `ansible-inventory --graph` and `ansible lab --list-hosts`; the latter shows the exact selected hosts without connecting to them.

## Practical exercise

Add only inventory groups corresponding to VMs you intend to build. Create `linux` and `lab` parents, then use `--list-hosts` for `webservers`, `linux`, and `lab`. Explain the blast radius of each.

## Expected result

Each host appears only in appropriate functional and parent groups, and every group expansion is predictable.

## Common mistakes

- **Using groups as VLAN labels and service roles interchangeably.** Name the dimension clearly.
- **Testing a destructive play against `all`.** Prefer a role group plus `--limit` for first application.
- **Duplicating host definitions under parents.** Use `:children`.

## Key takeaways

Inventory groups are targeting policy. Inspect their expansion before changing machines.

## Next lesson

Lesson 38 assigns data to groups and individual hosts.
