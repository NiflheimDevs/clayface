# Lesson 52 — Custom Facts

## Learning objectives

- Define a custom fact and its appropriate use.
- Store a simple static fact on Linux.
- Avoid using facts as a hidden configuration database.

## Prerequisites

Lesson 11 and file management knowledge.

## Concept

Custom facts let a managed node report local, organization-defined information through the fact system. On Linux, static facts can be INI or JSON files in `/etc/ansible/facts.d/*.fact`; executable fact scripts can calculate values but add maintenance and trust risk. They are useful for information owned by the node, such as an image build identifier or installed agent capability.

Inventory/group variables remain better for desired topology. Do not use custom facts to hide the desired configuration of the lab inside each VM; that makes rebuilds and review harder.

## Mental model

Custom facts are a machine's self-description badge, not a remote-control instruction.

## Example

``` ini
[digital_twin]
image_family=lab-linux
```

Deploy this as `/etc/ansible/facts.d/digital_twin.fact`, then run `ansible.builtin.setup` (or a later play with fact gathering) to expose it under `ansible_local`. The exact nested key should be inspected with `debug` rather than guessed.

## Practical exercise

Create a static fact identifying a disposable VM's base image family. Gather facts again and print only that fact. Explain why a web server's desired listen port should remain a group variable instead.

## Expected result

The fact is visible after gathering and survives normal Ansible runs/reboots as a file you manage.

## Common mistakes

- **Expecting a newly written fact to appear without gathering it.** Refresh facts after deployment.
- **Making executable facts writable by untrusted users.** They execute during fact gathering.
- **Storing secrets in custom facts.** Facts may be logged/cached.

## Key takeaways

Custom facts describe local observed identity or capability. Keep desired state in the repository.

## Next lesson

Lesson 53 replaces static host lists with controlled dynamic inventory sources.
