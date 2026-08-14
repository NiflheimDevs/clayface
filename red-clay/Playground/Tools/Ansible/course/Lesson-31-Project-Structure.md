# Lesson 31 — Project Structure

## Learning objectives

- Organize inventory, playbooks, roles, variables, and collections by responsibility.
- Explain why structure supports safe reuse.
- Create an initial repository layout.

## Prerequisites

Lessons 1–30 and Git basics.

## Concept

Structure is an organizational abstraction, not a required collection of folders. As automation grows, separation makes ownership visible: inventory describes targets, playbooks orchestrate intent, roles implement reusable capabilities, and group/host variables carry environment data. Avoid one giant playbook because it couples unrelated services and makes small changes difficult to review.

## Mental model

The repository is an operations library: a catalog (inventory), books for deployments (playbooks), reusable chapters (roles), and environment data cards (variables).

## Example

``` text
ansible/
├── ansible.cfg
├── inventory/
│   └── lab.ini
├── group_vars/
├── host_vars/
├── playbooks/
│   ├── site.yml
│   └── base.yml
├── roles/
└── collections/requirements.yml
```

`site.yml` will orchestrate the complete lab, while a focused playbook permits narrow work. `group_vars` and `host_vars` are Ansible-recognized locations when adjacent to inventory/playbooks according to the chosen layout. `collections/requirements.yml` pins external functionality.

## Practical exercise

Move your course files into this layout using Git-aware moves. Keep the existing inventory working and revise relative paths only after checking `ansible-inventory --graph`.

## Expected result

Commands run from the repository root discover project configuration and inventory. A new reader can locate each responsibility without reading every file.

## Common mistakes

- **Creating folders merely because a tutorial did.** Each must have a clear ownership purpose.
- **Putting secrets beside playbooks unencrypted.** Reserve protected variable files for Vault later.
- **Mixing Terraform state with Ansible roles.** Keep tool boundaries clear even if both share a parent repository.

## Key takeaways

Good structure makes intent, reuse, and review easier; it does not replace clear automation.

## Next lesson

Lesson 32 introduces roles as reusable units of configuration behavior.
