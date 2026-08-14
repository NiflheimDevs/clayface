# Lesson 25 — Services

## Learning objectives

- Enable and start a service declaratively.
- Handle systemd and OpenRC differences.
- Connect configuration changes to a reload handler.

## Prerequisites

Lessons 15 and 20–24. Have an installed test service such as nginx.

## Concept

A service has at least two independent desired states: whether it starts at boot and whether it is running now. `ansible.builtin.service` is a portable abstraction for `state: started`, `stopped`, `restarted`, or `reloaded` and `enabled: true/false`. Use `ansible.builtin.systemd_service` when systemd-specific features are required. Alpine commonly uses OpenRC, so systemd-only tasks must be conditional.

Service management is most reliable after package, configuration directory, config file, and permissions are correct. A static `state: restarted` task creates disruption every run; handlers should reload or restart only after configuration changes.

## Mental model

Installation provides the program; enabling schedules it at boot; starting makes it live now. Reloading is a reaction to altered configuration.

## Example

``` yaml
- name: Ensure nginx is enabled and running
  ansible.builtin.service:
    name: nginx
    enabled: true
    state: started
```

`enabled: true` asks the init system to persist startup behavior. `state: started` starts it if not already active, rather than restarting a healthy service. Service names vary by package and OS; verify them with the target's service manager.

## Practical exercise

Install and manage one safe service on a disposable VM. Stop it manually once, then run the playbook and observe Ansible restore only the declared running state. Add a handler design note explaining whether the service supports reload.

## Expected result

The service is enabled and active after the first run, and an unchanged healthy service becomes `ok` on the second.

## Common mistakes

- **Confusing `enabled` and `started`.** One is boot policy, the other current runtime state.
- **Using `systemd_service` on Alpine.** Select an init-appropriate module or conditional behavior.
- **Restarting after every play.** Notify a handler from the actual configuration task.

## Key takeaways

Declare boot and runtime state separately. Keep service restarts tied to genuine configuration changes.

## Next lesson

Lesson 26 manages operating-system configuration while preserving distribution boundaries.
