# Lesson 3 --- Installing Ansible

## Learning objectives

- Install `ansible-core` on an Arch Linux control node.
- Verify the installed version and configuration location.
- Understand the difference between `ansible-core` and optional community collections.

## Prerequisites

Lessons 1–2. Use an Arch Linux machine as the control node and have sudo access.

## Concept

`ansible-core` supplies the execution engine, commands, and built-in modules used throughout this course. It is the appropriate base package for an automation repository. A **collection** is a versioned package of modules, plugins, roles, and documentation; vendor-specific functionality is often installed later as a collection rather than bundled into the core.

Use the operating system package manager for the initial control-node installation. Do not install Ansible on managed nodes merely because they are managed. Python is needed on most Linux targets, but Ansible itself lives on the control node.

## Mental model

Install the workshop on your desk, not a workshop inside every house you maintain. Collections are specialist toolkits you add only when your project needs them.

## Example

On Arch Linux, run:

``` bash
sudo pacman -Syu ansible-core
ansible --version
```

`sudo` runs the package installation as root. `pacman -Syu` synchronizes package databases, upgrades installed packages, and installs the named package; this is normal Arch maintenance, so review its proposed upgrade before confirming. `ansible-core` supplies the CLI.

`ansible --version` does not contact any server. It shows the core version, Python version, collection paths, and the active configuration file. Keep this output when diagnosing inconsistent behavior between machines.

On Debian/Ubuntu the package is commonly installed with `sudo apt update && sudo apt install ansible-core`; RHEL-derived systems commonly use `sudo dnf install ansible-core`. These commands are for the control node only. Package versions can differ by distribution repository, so use `ansible --version` instead of assuming a tutorial's version.

## Practical exercise

Install `ansible-core` on your Arch control node. Run `ansible --version`, then record the core version, Python version, and `config file` value in your course notes. Do not create a configuration file yet.

## Expected result

The command exits successfully and reports an Ansible core version. If no project configuration exists, the configuration-file line may say `None`; that is normal at this stage.

## Common mistakes

- **Installing package `ansible` when you only need core.** It can be a broader convenience bundle; this course starts deliberately with `ansible-core`.
- **Running `pip install` into the system Python.** That can conflict with pacman-managed packages. Use a virtual environment only when you later need isolated versions.
- **Treating a warning as an unreachable-host error.** `ansible --version` is local and cannot test SSH.

## Key takeaways

`ansible-core` belongs on the control node. Confirm the real version and configuration path before troubleshooting anything else.

## Next lesson

Lesson 4 creates the inventory that tells Ansible where your lab machines are.
