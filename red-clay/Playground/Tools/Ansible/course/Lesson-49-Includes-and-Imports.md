# Lesson 49 — Includes vs Imports

## Learning objectives

- Distinguish static imports from dynamic includes.
- Choose `import_tasks` or `include_tasks` deliberately.
- Avoid premature conditional file splitting.

## Prerequisites

Lessons 32 and 48.

## Concept

`ansible.builtin.import_tasks` is **static**: Ansible parses the referenced tasks when it parses the playbook. `ansible.builtin.include_tasks` is **dynamic**: it decides at execution time to include tasks, often based on variables or facts. Static content is easier for listing tasks/tags and has predictable structure; dynamic content handles a genuinely runtime-selected file.

Roles have related import/include mechanisms. Do not split a three-task role merely to use a feature; files should reflect a coherent responsibility.

## Mental model

An import is a chapter bound into the manual before printing. An include is a sealed appendix chosen while the operator is already following the manual.

## Example

``` yaml
- name: Import common package tasks
  ansible.builtin.import_tasks: packages.yml

- name: Include distribution-specific tasks
  ansible.builtin.include_tasks: "{{ ansible_facts['os_family'] }}.yml"
```

The first file is known at parse time. The second depends on gathered facts, so the corresponding filenames must exist and be tightly controlled. Do not derive an include filename from untrusted external input.

## Practical exercise

Split `base_linux` packages into a static task file. Then sketch—not necessarily implement—a dynamic selection for Alpine and Arch package names. Explain why an undefined or unsupported OS family should fail clearly.

## Expected result

You can predict whether task listing sees a file before execution and why dynamic selection needs guardrails.

## Common mistakes

- **Using dynamic includes for every file.** It makes review and tag behavior harder to reason about.
- **Expecting static imports to defer conditions.** Parse-time structure differs from execution-time conditions.
- **Using facts before they are gathered.** Dynamic OS selection needs facts.

## Key takeaways

Imports are static and predictable; includes are dynamic and runtime-selected. Use the simplest model matching the need.

## Next lesson

Lesson 50 makes the static/dynamic distinction concrete in execution behavior.
