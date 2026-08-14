# Lesson 33 — Role Variables and Defaults

## Learning objectives

- Distinguish role defaults from role vars.
- Design an override-friendly role interface.
- Explain why variable precedence is a design risk.

## Prerequisites

Lesson 32 and Lesson 10.

## Concept

Role **defaults** (`defaults/main.yml`) have low precedence and are intended as safe, documented values callers may override. Role **vars** (`vars/main.yml`) have much higher precedence and are for internal constants that should rarely be changed. Put a value in defaults unless making it internal is genuinely necessary.

Variable precedence is Ansible's conflict-resolution order. Extra vars are among the highest-precedence sources; role defaults are among the lowest. This power makes an accidental duplicate variable name dangerous. Rather than memorizing every rule now, use namespaced names such as `base_linux_packages`, document overrides, and avoid setting one name in many locations.

## Mental model

Defaults are knobs on the front of an appliance. Role vars are wiring behind the sealed panel. Extra vars are an emergency override switch.

## Example

`roles/base_linux/defaults/main.yml`:

``` yaml
base_linux_packages:
  - curl
  - ca-certificates
```

`roles/base_linux/tasks/main.yml`:

``` yaml
- name: Ensure base packages are installed
  ansible.builtin.package:
    name: "{{ base_linux_packages }}"
    state: present
```

The list default is role-scoped by name. A group can later override it with distribution-appropriate data, but only after you deliberately decide that all group members share the value.

## Practical exercise

Move your package list to `base_linux` defaults. Override it temporarily in the calling play for one run, then remove the override. Explain which value you would put in `vars/main.yml` and why.

## Expected result

The default works unmodified and a deliberate caller override changes only that run's desired package list.

## Common mistakes

- **Using `vars/main.yml` for every setting.** It makes a role unexpectedly hard to customize.
- **Using generic names like `packages`.** Another role can collide with it.
- **Forcing values with `-e` routinely.** High precedence obscures source-of-truth configuration.

## Key takeaways

Defaults define a role's public configuration interface. Precedence is powerful, so design variable ownership rather than relying on overrides.

## Next lesson

Lesson 34 explains explicit dependencies between roles.
