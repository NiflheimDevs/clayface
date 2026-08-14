# Lesson 2 --- Ansible Architecture

## Learning objectives

- Identify the control node, managed node, inventory, playbook, play, task, and module.
- Explain how Ansible runs a Linux task over SSH.
- State the connection prerequisites for Alpine and Arch targets.

## Prerequisites

Lesson 1. You need basic SSH and IP networking knowledge.

## Concept

The **control node** is the computer on which you run `ansible` and `ansible-playbook`; in this course it is your Arch workstation. A **managed node** is a target machine such as `web01`. An **inventory** maps memorable host names and groups to connection details. A **playbook** contains one or more **plays**; a play selects hosts and gives them ordered **tasks**. A task invokes a **module** to manage one type of resource.

For a normal Linux task, Ansible reads the inventory, opens SSH as the configured remote user, copies a small module program to a temporary directory on the target, executes it with Python, collects JSON-like results, and removes the temporary file. SSH is the transport; it is not the configuration logic itself. This is why the managed VM needs network reachability, an SSH server, an account and credentials, and normally Python 3. Minimal Alpine images sometimes lack Python, which will be handled during bootstrapping.

Ansible is agentless in the important sense that no always-running Ansible daemon has to be installed on every Linux VM. It does not mean “nothing must be configured”: SSH access and privilege permissions are still a security boundary. Windows normally uses WinRM or SSH rather than this Linux SSH/Python path.

Tasks run in playbook order on a host. By default, Ansible works on several hosts concurrently; the exact number is controlled by `forks`. Do not assume all of `webservers` finish task 1 before any host begins task 2 unless a later feature explicitly changes that behavior.

## Mental model

Inventory is the address book. A play is the recipient list plus a checklist. The control node is the administrator who carries out each checklist item over SSH. The module returns whether the item was already true (`ok`) or needed correction (`changed`).

## Example

The final lab inventory will grow toward this shape:

``` ini
[webservers]
web01
web02

[databases]
db01

[dns]
dns01
```

This is **INI inventory syntax**, not a playbook. Brackets create a group; unindented names are hosts. At this point `web01` is only an inventory label. Lesson 4 will connect it to an IP address and SSH user.

When you later run `ansible-playbook -i inventory.ini site.yml`, `-i` chooses that inventory file and `site.yml` is the playbook. A task selected for `webservers` will be sent to `web01` and `web02`, not to `db01`.

## Practical exercise

Draw your initial control path: control-node hostname, management network/subnet, and one target VM (`web01`). Record the SSH user you intend Ansible to use and whether Python 3 is installed on that VM. Then mark which future VMs are Linux and which, if any, will be Windows.

## Expected result

You can explain exactly where Ansible is installed, where it connects, and what must be true before a Linux module can run.

## Common mistakes

- **Installing Ansible on every VM.** Only the control node needs the Ansible package for this model.
- **Confusing an inventory name with DNS.** `web01` is not resolvable unless you provide DNS, `/etc/hosts`, or an address in inventory.
- **Forgetting Python on minimal Linux.** A raw SSH command can work while normal modules fail with an interpreter error.
- **Opening SSH too broadly.** Management connectivity should be restricted to the management network where practical.

## Key takeaways

Ansible executes modules from a control node against inventory-selected managed nodes, normally over SSH. Inventory and reachability come before automation.

## Next lesson

Lesson 3 installs the control-node software and verifies the version you are actually using.
