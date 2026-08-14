# Lesson 29 — Templates and Jinja2

## Learning objectives

- Explain why `template` differs from `copy`.
- Render simple variables safely in a `.j2` file.
- Keep templates readable and avoid logic overload.

## Prerequisites

Lessons 10, 23, and 28.

## Concept

The `ansible.builtin.template` module renders a Jinja2 template on the control node and copies its output to a target. A template uses `{{ variable }}` for expressions and `{% ... %}` for control statements. Start with simple substitution; Jinja2 is powerful, but an unreadable template can hide business logic and undermine review.

Template input is version-controlled source; rendered output is the host-specific desired file. The module compares rendered content, so changing a variable can correctly yield `changed` even when the template file itself is unchanged.

## Mental model

A template is a printed form with blank fields. Variables fill each host's fields before the final document is delivered.

## Example

Create `templates/lab.conf.j2`:

``` text
# Managed by Ansible
server_name={{ inventory_hostname }}
listen_port={{ web_listen_port }}
```

Then render it:

``` yaml
- name: Render web service configuration
  ansible.builtin.template:
    src: templates/lab.conf.j2
    dest: /etc/digital-twin/web.conf
    owner: root
    group: root
    mode: "0644"
```

`src` is local to the control node. The template expression has no extra quotes inside the template because it is plain text. Define `web_listen_port` in an appropriate variable location. A missing variable is a useful failure: do not replace it with a silent empty value.

## Practical exercise

Render a harmless `/tmp` configuration containing `inventory_hostname` and one play variable. Change the variable, use `--check --diff`, then apply and inspect the remote result. Avoid templates for a file that is actually identical everywhere.

## Expected result

Each host receives text with its own inventory name and declared value. A second unchanged run is `ok`.

## Common mistakes

- **Putting `{{ }}` around a whole YAML value when passing a non-string.** Understand whether you need a rendered string or a YAML type.
- **Adding complex loops and decisions before data design.** Keep policy in variables and tasks where reviewers can see it.
- **Using `copy` with unresolved template syntax.** `copy` does not render Jinja2.

## Key takeaways

Use templates for intentional per-host rendering. Jinja2 starts as clear placeholder substitution, not magic.

## Next lesson

Lesson 30 applies the same desired-state discipline to firewalls and management-network safety.
