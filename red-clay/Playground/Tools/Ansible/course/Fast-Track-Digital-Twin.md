# Ansible Fast Track --- Digital Twin

This is the recommended route for getting productive with Ansible in your Terraform/libvirt Digital Twin without first studying every advanced feature.

The full 75-lesson course remains a reference. Follow the linked lessons below in order, perform their exercises on one disposable Linux VM, then begin building your actual lab. You do not need every fictional host (`web02`, `db01`, `dc01`, and so on) before starting.

## Outcome

After this track, you should be able to take a requirement such as:

> Every web server needs nginx, a dedicated user, a rendered configuration file, an enabled service, and narrowly permitted network access.

…and naturally design it as:

``` text
inventory → play → variables → modules → template → handler → role
```

You will also understand the handoff: Terraform creates reachable VMs; Ansible configures what runs inside them.

## Phase 1 --- Fundamentals

1. [What Is Ansible?](Lesson-01-What-Is-Ansible.md) — configuration management and the Terraform boundary.
2. [Ansible Architecture](Lesson-02-Ansible-Architecture.md) — control node, managed nodes, SSH, and remote Python.
3. [Installing Ansible](Lesson-03-Installing-Ansible.md) — set up the Arch Linux control node.
4. [Inventory](Lesson-04-Inventory.md) — define the VMs in your lab.
5. [Ad-Hoc Commands](Lesson-05-Ad-Hoc-Commands.md) — test reachability and diagnose the SSH path.
6. [Modules](Lesson-06-Modules.md) — use state-aware modules instead of blind shell commands.
7. [Playbooks](Lesson-07-Playbooks.md) — write repeatable YAML automation.
8. [Plays and Tasks](Lesson-08-Plays-and-Tasks.md) — understand execution order and the recap.
9. [Idempotency](Lesson-09-Idempotency.md) — make re-runs safe and predictable.

## Phase 2 --- Configure Linux Machines

10. [Variables](Lesson-10-Variables.md) — keep configuration reusable.
11. [Facts](Lesson-11-Facts.md) — adapt carefully to Alpine, Arch, and other targets.
12. [Conditionals](Lesson-13-Conditionals.md) — run platform- or role-specific work only where appropriate.
13. [Loops](Lesson-14-Loops.md) — manage repeated packages, users, or files clearly.
14. [Handlers](Lesson-15-Handlers.md) — reload a service only after its configuration changes.
15. [Privilege Escalation](Lesson-46-Become.md) — use `become` safely for root-owned system state.
16. [Package Management](Lesson-20-Package-Management.md) — install software declaratively across Linux distributions.
17. [Users and Groups](Lesson-21-Users-and-Groups.md) — establish dedicated administration and service identities.
18. [Files and Directories](Lesson-23-Files-and-Directories.md) and [Services](Lesson-25-Services.md) — manage configuration locations and active daemons.
19. [Templates and Jinja2](Lesson-29-Templates-and-Jinja2.md) — render per-host service configuration.
20. [Roles](Lesson-32-Roles.md) — organize the growing lab into reusable capabilities.

## Phase 3 --- Your Infrastructure

These are the final essential integrations. Read them immediately after Lesson 20 above; they are grouped here because they turn the core mechanics into your Digital Twin workflow.

- [Inventory Organization](Lesson-37-Inventory-Organization.md) and [Group Variables and Host Variables](Lesson-38-Group-Variables-and-Host-Variables.md) — separate web, database, DNS, clients, and other lab functions.
- [Ansible Vault](Lesson-39-Ansible-Vault.md) — protect credentials before your roles begin using them.
- [Terraform vs Ansible](Lesson-56-Terraform-vs-Ansible.md), [Terraform Outputs as Ansible Inputs](Lesson-58-Terraform-Outputs-as-Ansible-Inputs.md), [Generating Inventories from Terraform](Lesson-59-Generating-Inventories-from-Terraform.md), and [Managing Terraform-Created VMs](Lesson-60-Managing-Terraform-Created-VMs.md) — build the Terraform-to-Ansible handoff.

At this point, stop studying generic features and build one vertical slice:

``` text
Terraform creates web01
        ↓
Generated inventory supplies its management address
        ↓
Ansible base_linux role creates administration baseline
        ↓
web_nginx role installs nginx, renders config, enables service
        ↓
Health check proves the web service is reachable
```

Then add `db01`, DNS, clients, Windows/AD, and attacker simulation one verified slice at a time.

## Defer Until Needed

Use the full course as a just-in-time reference for these topics:

- [Tags](Lesson-16-Tags.md), [Check Mode and Diff Mode](Lesson-18-Check-Mode-and-Diff-Mode.md), [Registering Results](Lesson-12-Registering-Command-Results.md), [Error Handling](Lesson-17-Error-Handling.md), and [Ansible Configuration](Lesson-19-Ansible-Configuration.md)
- [Advanced Jinja2](Lesson-51-Advanced-Jinja2.md), [Delegation](Lesson-44-Delegation.md), [`run_once`](Lesson-45-Run-Once.md), [Blocks](Lesson-48-Blocks.md), and [Includes vs Imports](Lesson-49-Includes-and-Imports.md)
- [Dynamic Inventories](Lesson-53-Dynamic-Inventories.md), [Collections](Lesson-35-Collections.md), [Custom Facts](Lesson-52-Custom-Facts.md), and [Windows Automation](Lesson-72-Windows-Automation.md)

Read one when your current task creates the need. For example: learn delegation when a web-server play genuinely must update a DNS or load-balancer host; do not memorize it in advance.
