# Lesson 11 --- Facts

## Learning objectives

- Explain what facts are and when Ansible gathers them.
- Inspect facts without overwhelming output.
- Use a fact for simple distribution-aware logic.

## Prerequisites

Lessons 1–10. Your Linux target needs Python for normal fact gathering.

## Concept

**Facts** are variables discovered from a managed node, such as its OS family, hostname, interfaces, addresses, memory, and Python details. By default, each play gathers facts with `ansible.builtin.setup` before the first task. This is convenient but costs time and requires a working remote Python interpreter.

Facts describe observed state, while normal variables describe desired choices. Do not confuse them: `ansible_facts['os_family']` tells you what the host reports; `web_listen_port` tells Ansible what you want to configure.

## Mental model

Facts are the intake form filled out by each machine before work begins. Variables are instructions you bring with you.

## Example

``` yaml
---
- name: Inspect lab hosts
  hosts: linux
  gather_facts: true
  tasks:
    - name: Show operating system family
      ansible.builtin.debug:
        msg: "{{ inventory_hostname }} is in the {{ ansible_facts['os_family'] }} family"
```

`gather_facts: true` is the default but is explicit for teaching. `ansible.builtin.debug` prints a value and changes nothing. `inventory_hostname` is the inventory label, whereas facts come from the remote machine. Square-bracket fact access is robust when a key includes punctuation or when you want an explicit lookup.

Inspect selected facts ad-hoc:

``` bash
ansible linux -i inventory/lab.ini -m ansible.builtin.setup -a 'filter=ansible_*distribution*'
```

The quoted module argument remains one shell argument. The filter limits output to distribution-related fact keys.

## Practical exercise

Run the filtered setup command against your Linux group. Create the debug play and compare `inventory_hostname` with the host's reported hostname. If you only have one host, intentionally make these names different in inventory and explain why that can be useful.

## Expected result

You see a stable inventory identity and a discovered OS family. The debug task reports `ok`, never `changed`.

## Common mistakes

- **Using facts with `gather_facts: false`.** They may be undefined unless another task explicitly gathered them.
- **Branching on exact distribution names needlessly.** Prefer `os_family` where that correctly captures the required behavior.
- **Putting secrets in debug output.** Debug is useful, but all output can end up in logs.

## Key takeaways

Facts are discovered variables, usually gathered at play start. They enable informed portability but do not replace deliberate inventory design.

## Next lesson

Lesson 12 saves a task's structured result for later decisions.
