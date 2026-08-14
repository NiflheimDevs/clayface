# Lesson 4 --- Inventory

## Learning objectives

- Create a static inventory for one or more lab VMs.
- Use host variables and groups to describe connections.
- Select targets with basic inventory patterns.

## Prerequisites

Lessons 1–3. Have one Linux VM reachable over SSH.

## Concept

An **inventory** is Ansible's source of managed-node identities, groups, and connection variables. It can be an INI file, YAML file, directory, or dynamic source. Start with a readable static INI inventory; it makes the relationship between the lab topology and targets visible.

A host's inventory name is an Ansible identifier. `ansible_host` is the address actually used for SSH. Keeping them separate lets playbooks say `web01` even if Terraform later assigns a different IP. `ansible_user` is the remote login account. Group names such as `webservers` select a role-like class of hosts, not necessarily a network segment.

An **inventory pattern** is a selector in a command or play. `webservers` selects the group, `web01` selects a host, and `webservers:databases` selects their union. `webservers:!web02` selects web servers except `web02`. Start with simple group names; patterns become more important as the lab grows.

## Mental model

An inventory is an address book plus labels. A group is a label applied to hosts, and a pattern is a search over those labels.

## Example

Create `inventory/lab.ini` in your future project directory:

``` ini
[webservers]
web01 ansible_host=192.168.122.101 ansible_user=ansible

[databases]
db01 ansible_host=192.168.122.102 ansible_user=ansible

[linux:children]
webservers
databases
```

`[webservers]` starts a group. The first token, `web01`, is the stable inventory name. The key/value pairs are host variables. `[linux:children]` creates a parent group whose members are groups, not individual hosts; a play aimed at `linux` reaches both groups. Substitute your real management IP and login user. Do not invent a `db01` entry until you have that VM; one host is enough.

Inspect the parsed result with:

``` bash
ansible-inventory -i inventory/lab.ini --graph
```

`ansible-inventory` reads inventory without making SSH connections. `-i` supplies a non-default inventory path. `--graph` renders group membership, making accidental nesting easy to spot.

## Practical exercise

Create an inventory containing only your reachable `web01`. Add its actual management address and a non-root SSH account. Run the graph command, then add a parent `linux` group even if it has only `webservers` as a child.

## Expected result

The graph lists `@linux`, `@webservers`, and `web01`; it does not need to resolve the hostname through DNS because `ansible_host` supplies the address.

## Common mistakes

- **Indenting host lines.** INI inventory host entries should begin at the left margin.
- **Using the VM's NAT address from the wrong network namespace.** Test the exact address with normal SSH from the control node first.
- **Putting credentials into inventory.** Use SSH keys; Vault will cover protected secrets later.
- **Making every host part of every group.** Groups should express meaningful shared configuration.

## Key takeaways

Inventory separates stable host names and roles from changing connection addresses. It is the safety boundary for which hosts a play may affect.

## Next lesson

Lesson 5 uses an ad-hoc command to test the inventory and SSH connection before writing playbooks.
