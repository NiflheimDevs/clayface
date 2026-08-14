# Lesson 1 --- What Is Ansible?

## Learning objectives

By the end of this lesson, you should be able to:

- Explain the problem Ansible solves in a virtual enterprise lab.
- Distinguish configuration management from infrastructure provisioning.
- Describe Ansible as an agentless automation tool.
- Explain desired state and idempotency at a high level.
- Identify work that belongs in Terraform versus Ansible.

## Prerequisites

No previous Ansible knowledge is required. You should be comfortable with Linux command lines, SSH, and virtual machines. Terraform is useful context but not required.

## Concept

### The problem: repeatable machine configuration

Terraform can create a VM, attach its disks and virtual NICs, and connect it to libvirt networks. A freshly created VM is still only a computer with an operating system. It may need users, SSH hardening, packages, services, firewall rules, configuration files, and application settings.

Doing that manually works for one short-lived VM. It fails as a lab practice because the steps become undocumented, vary between machines, and are hard to repeat after a rebuild. In an Enterprise Digital Twin, that is especially harmful: the same lab topology should produce the same usable environment every time.

**Ansible** is a configuration-management and automation tool. You describe the intended configuration in text files, called **playbooks**, and Ansible connects to target machines to make their configuration match it.

For example, a base-server configuration can state that every Linux server should have:

- an `ansible` administration user;
- SSH configured to disallow direct root login;
- a firewall enabled;
- packages required by the server's role.

This is not merely a list of commands to run. The important goal is the final condition of each machine.

### Configuration management versus provisioning

Terraform and Ansible often appear in the same repository, but they work at different layers.

| Concern | Terraform is usually responsible for | Ansible is usually responsible for |
| --- | --- | --- |
| Virtual infrastructure | VM, disk, virtual NIC, libvirt network, IP assignment integration | — |
| Operating system configuration | Initial cloud-init data only when needed to make the machine reachable | users, packages, files, services, firewall rules |
| Application configuration | — | web server, database, DNS service, application settings |
| Lifecycle | create, modify, destroy infrastructure resources | converge a reachable system to its intended configuration |

There is overlap at the boundary: Terraform may use cloud-init to place an SSH public key on a new VM, because Ansible needs a way to connect. Once SSH works, Ansible should own the ongoing OS and service configuration.

As a useful rule: **Terraform builds the house and connects its utilities; Ansible furnishes and operates the rooms.** This is a guideline, not a law. The boundary should be clear enough that the two tools do not fight over the same setting.

### Desired state and convergence

In Ansible, a task expresses an intended result, such as “the `nginx` package is installed” or “this file has these contents.” On each run, Ansible checks the target machine and takes action only if it needs to bring reality closer to that result. This process is called **convergence**.

``` text
Playbook (intended configuration)
              |
              v
       Ansible control node
              |
          SSH / WinRM
              |
              v
Managed nodes (actual configuration)
```

Ansible reports the outcome of each task for each host:

- `ok`: the host was already in the requested state.
- `changed`: Ansible changed something to reach that state.
- `failed`: the task could not achieve the requested result.
- `skipped`: Ansible deliberately did not run the task, commonly because a condition was false.

You will use these words throughout the course. Seeing `changed` on the first run and `ok` on the second is often evidence that a task is correctly idempotent.

### Idempotency

An operation is **idempotent** when running it repeatedly produces the same final result as running it once.

Suppose you declare that an SSH configuration file must have particular content. The first Ansible run may replace it and report `changed`. The second run should see the requested content is already present and report `ok`; it should not rewrite the file or restart services without need.

This matters because you want to safely re-run automation after Terraform rebuilds a VM, an interrupted deployment, or a change to only one role. Good automation is repeatable, predictable, and makes the smallest necessary change.

Ansible cannot automatically make every action idempotent. A task that blindly runs a shell command like `useradd alice` may fail on a second run because Alice already exists. Later you will use the purpose-built `ansible.builtin.user` module, which can determine whether the user exists and report `ok` when no change is needed.

### Agentless automation

For ordinary Linux targets, Ansible normally uses SSH. You install Ansible on one machine, the **control node**, and it uses normal SSH access to manage the target machines. The targets generally do not need a permanently running Ansible agent.

That design fits your lab well: your Arch Linux workstation can be the control node and Alpine/Arch VMs can be managed nodes. Windows is different: Ansible commonly connects using WinRM or SSH and has extra setup requirements; the general playbook model remains the same.

This lesson does not require you to install anything yet. The next lessons explain how the connection and execution model works before we use it.

## Mental model

Treat Ansible as a careful remote systems administrator with a written checklist.

For each machine selected by the checklist, it asks: “Is this desired condition already true?” If yes, it records `ok`. If not, it makes only the required correction and records `changed`. If it cannot verify or apply the condition, it records `failed`.

Terraform has a related declarative idea, but it reconciles infrastructure resources using provider APIs and state. Ansible usually reconciles inside an already-running operating system over a management connection.

## Example

Imagine Terraform has created `web01` with an Alpine Linux disk and attached it to the `server` libvirt network. Your intended configuration is:

``` text
web01
  - reachable over SSH
  - nginx installed
  - nginx service enabled and running
  - company web configuration present
```

Terraform should normally create the VM, virtual disk, NIC, and network attachment. Ansible should install and configure nginx after the VM is reachable.

A future Ansible task will look roughly like this:

``` yaml
- name: Ensure nginx is installed
  ansible.builtin.package:
    name: nginx
    state: present
```

Do not copy this into a file yet; we will explain YAML, playbooks, tasks, and modules before using it.

For now, read it as a sentence:

- `- name:` is a human-readable label for one task. The leading `-` begins an item in a YAML list.
- `ansible.builtin.package` is an Ansible **module**: code designed to manage packages rather than merely execute a command.
- `name: nginx` identifies the package.
- `state: present` expresses the desired result: nginx must be installed. It does not prescribe a command such as `apk add nginx` or `pacman -S nginx`.

The `package` module selects the appropriate package manager where possible. Alpine uses `apk`; Arch uses `pacman`; Debian/Ubuntu normally use `apt`; RHEL-derived systems use `dnf` (or sometimes `yum` on older systems). In a real project, package names and service names can still differ by distribution, so portable intent does not eliminate platform-specific knowledge.

On a first run against a clean `web01`, Ansible would likely report `changed`. On the next run, it should report `ok`. That is the value of a module expressing desired state instead of a blind shell command.

## Practical exercise

Design the first two machines of your Digital Twin on paper or in a Markdown note. You do not need to create them yet.

1. Choose one Linux VM to become `web01` and one optional second VM to become `db01`. If you only have one VM, use only `web01`.
2. For each machine, write two lists:
   - infrastructure properties Terraform should own (for example: VM CPU, RAM, disk, NIC, network);
   - operating-system or service properties Ansible should own (for example: package, user, SSH, firewall, service).
3. Pick three desired configuration conditions for `web01`. Phrase each as a stable result, not a command. For example, “the Nginx service is enabled and running,” rather than “run `rc-service nginx start`.”
4. For one condition, explain what should happen on the first and second Ansible run using the words `changed` and `ok`.

Do not automate these conditions yet. The next lessons provide the connection and inventory foundations needed to do so safely.

## Expected result

You should have a small, explicit boundary between Terraform and Ansible. Your `web01` list should contain desired end states that remain meaningful even if you rebuild the VM. You should be able to explain why the second run should normally make no change.

## Common mistakes

- **Treating Ansible as a remote shell-command launcher.** It can run commands, but commands often lack reliable change detection and repeatability. Prefer a module when one manages the desired resource.
- **Giving Terraform and Ansible ownership of the same configuration.** For example, if both continuously write the same application configuration file, later runs can undo each other. Choose one owner.
- **Assuming “agentless” means no preparation.** Linux targets still need an SSH service, network reachability, credentials, and a usable Python interpreter for most modules. We will cover these requirements soon.
- **Expecting all distributions to behave identically.** The Ansible concept can be portable while package names, init systems, filesystem paths, and firewall tools differ. Alpine commonly uses OpenRC; Arch commonly uses systemd.
- **Equating `changed` with success in every case.** `ok` is also a successful result; it often proves the machine was already configured correctly. `failed` is the result that requires investigation.

## Key takeaways

- Ansible configures reachable machines; Terraform usually provisions the infrastructure around them.
- A playbook describes desired configuration and Ansible converges managed nodes toward it.
- `ok`, `changed`, `failed`, and `skipped` are task outcomes you will read on every run.
- Idempotent automation can be safely run repeatedly and should make no unnecessary changes.
- Modules express system intent more reliably than blindly executing shell commands.

## Next lesson

**Lesson 2 --- Ansible Architecture** explains the control node, managed nodes, SSH transport, Python requirement, and what actually happens when Ansible runs a task on one or many machines.
