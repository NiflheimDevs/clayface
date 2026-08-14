# Lesson 27 — Environment Variables

## Learning objectives

- Identify the correct scope for an environment variable.
- Manage service environments separately from login shells.
- Avoid placing secrets in broadly readable environment files.

## Prerequisites

Lesson 26 and basic process-environment knowledge.

## Concept

An environment variable belongs to a process and its children; it is not automatically a global operating-system setting. Decide whether a value is needed by an interactive user, a service, a single Ansible task, or every process. A shell profile affects interactive shells; it usually does not affect systemd services. Systemd units commonly use an `EnvironmentFile=` drop-in, while OpenRC has its own service configuration conventions.

Use Ansible's `environment:` keyword for process variables needed while Ansible runs a task. Use an owned configuration file/template for persistent service values. Do not use global profiles for application secrets.

## Mental model

Environment variables are notes handed to a process when it starts, not labels painted on the whole machine.

## Example

``` yaml
- name: Query an internal service through its proxy
  ansible.builtin.command: curl -fsS https://status.example.invalid
  environment:
    https_proxy: http://proxy01.lab:3128
  changed_when: false
```

`environment` applies only to this module invocation. `curl -f` fails for HTTP error responses, `-sS` suppresses progress while preserving errors, and `-S` is meaningful with `-s`. This is an observation command, so `changed_when: false` is appropriate.

## Practical exercise

Choose a non-secret variable required by one test command. Use task-level `environment`, run it, and verify the variable is not permanently added to a login profile. Then identify how your selected init system would supply a persistent variable to a service.

## Expected result

The command sees the variable; a new interactive shell does not gain it merely from this task.

## Common mistakes

- **Using `/etc/profile` for daemon settings.** Services are often not launched by login shells.
- **Putting passwords in a world-readable environment file.** Environment values are often exposed through process or service inspection.
- **Assuming task environment persists to later tasks.** It is explicitly scoped.

## Key takeaways

Match environment-variable scope to the consumer. Persistent service configuration and temporary task environment are different tools.

## Next lesson

Lesson 28 deploys static configuration files with validation and handlers.
