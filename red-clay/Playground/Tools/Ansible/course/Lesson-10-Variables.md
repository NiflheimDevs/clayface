# Lesson 10 --- Variables

## Learning objectives

- Define variables and reference them with Jinja2 delimiters.
- Put host-appropriate values outside task logic.
- Avoid premature use of high-precedence variable locations.

## Prerequisites

Lessons 1–9. Understand a simple playbook and YAML mappings.

## Concept

A **variable** is a named value supplied to Ansible at runtime. Variables avoid copying a playbook merely because a server name, package list, port, or environment differs. References use `{{ variable_name }}`, a small part of the Jinja2 templating language. At this stage, think of it as “replace this placeholder with a value.”

Variables can come from the playbook, inventory, group variables, host variables, role defaults, command-line extra vars, facts, and more. This flexibility creates **variable precedence**: when names collide, one source wins. We will learn the full precedence model later. Until then, use clear names and put stable environment values in the inventory or a group variable file rather than scattering overrides.

## Mental model

A variable is a labeled slot in a blueprint. The task defines the kind of room; the variable provides the room-specific measurement.

## Example

Use a play-level variable:

``` yaml
---
- name: Create a labeled course marker
  hosts: webservers
  gather_facts: false
  vars:
    lab_name: enterprise-digital-twin
  tasks:
    - name: Write the lab name
      ansible.builtin.copy:
        content: "Lab: {{ lab_name }}\n"
        dest: /tmp/lab-name
        mode: "0644"
```

`vars:` is a YAML mapping belonging to this play. `lab_name` is a string. The quoted `content` allows the `{{ lab_name }}` expression to render before the module writes the file. This value is appropriate here because it is local to this small demonstration; shared environment values later move to `group_vars`.

## Practical exercise

Create the playbook, choose a lab name without spaces, and run it twice. Change only the variable value and predict what the next run will report. Then identify one value that should eventually differ between `webservers` and `databases`.

## Expected result

Changing the variable changes the managed file once. The task remains reusable because the content is not duplicated in its logic.

## Common mistakes

- **Forgetting quotes around strings with `:` or template punctuation.** YAML may parse them unexpectedly.
- **Naming a variable `host` or `user` generically.** Use scoped names such as `web_listen_port` to prevent collisions.
- **Passing every value with `-e`.** Extra variables have very high precedence and make routine configuration hard to audit.

## Key takeaways

Variables separate reusable task logic from environment-specific data. `{{ name }}` renders a variable value.

## Next lesson

Lesson 11 uses facts—information Ansible discovers from each managed node.
