# Lesson 19 --- Ansible Configuration (`ansible.cfg`)

## Learning objectives

- Locate the active Ansible configuration file.
- Create a minimal project-local `ansible.cfg`.
- Explain why configuration defaults affect safety and reproducibility.

## Prerequisites

Lessons 1–18. You should have a project directory with `inventory/` and `playbooks/`.

## Concept

`ansible.cfg` controls client behavior such as the default inventory, remote temporary directory, forks, and privilege defaults. Ansible searches several locations; a project-local configuration is generally preferable because Git records the team's intentional settings. Use `ansible --version` to see which file is active.

Configuration is not a substitute for explicit playbook safety. For example, setting a default inventory is convenient, but you must still use narrow `hosts` patterns and inspect `--limit` values. Avoid disabling security checks merely to quiet warnings; understand the warning first.

## Mental model

`ansible.cfg` is the project-wide operating policy. A playbook is the work order. Keep policy visible, small, and predictable.

## Example

Create a project-root `ansible.cfg`:

``` ini
[defaults]
inventory = inventory/lab.ini
interpreter_python = auto_silent
retry_files_enabled = False
```

`[defaults]` is an INI section. `inventory` lets commands omit `-i` when run from this project. `interpreter_python = auto_silent` asks Ansible to discover a usable Python interpreter while suppressing the normal discovery warning; it does not install Python. `retry_files_enabled = False` avoids old-style retry-file clutter; diagnose failures from the real output instead.

Verify resolution:

``` bash
ansible --version
ansible-inventory --graph
```

The first should name your project configuration. The second can now omit `-i`; do not omit it in scripts until the project's execution directory is controlled.

## Practical exercise

Create this minimal configuration, verify it is active, and run your verification playbook without `-i`. Then run the same command from another directory and observe why current working directory matters. Do not add `host_key_checking = False`; instead learn the SSH host-key workflow in a later lesson.

## Expected result

Within the project directory, Ansible finds the inventory automatically. Outside it, behavior depends on configuration discovery, demonstrating why automation commands should have a defined working directory.

## Common mistakes

- **Using global configuration for project behavior.** It makes results depend on an individual workstation.
- **Disabling host-key checking in a lab by default.** It hides possible man-in-the-middle and rebuilt-host identity problems.
- **Adding every possible setting.** Minimal explicit policy is easier to review.

## Key takeaways

Project-local `ansible.cfg` makes defaults reproducible. Verify which configuration Ansible actually loaded.

## Next lesson

Lesson 20 begins Linux management with distribution-aware package installation.
