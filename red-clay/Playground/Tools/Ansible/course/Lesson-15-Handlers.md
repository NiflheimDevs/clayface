# Lesson 15 --- Handlers

## Learning objectives

- Explain why handlers exist.
- Notify a handler only when a task changes.
- Predict handler timing and deduplication.

## Prerequisites

Lessons 1–14. Understand accurate `changed` reporting and services conceptually.

## Concept

A **handler** is a named task triggered by `notify` from a task that changed. Ansible queues a notified handler and normally runs it once at the end of the play, even if several tasks notify it. This makes configuration deployment safe: a service reload happens only when its configuration changed, and only once.

Handlers are not ordinary task order. Do not rely on an immediate reload after a notifying task unless you later use `meta: flush_handlers` for a justified dependency. An unneeded handler run is usually a sign of inaccurate change detection.

## Mental model

A handler is a maintenance ticket placed only when work altered the system. At the end of the shift, duplicate tickets for the same action become one service reload.

## Example

``` yaml
- name: Write a demonstration service configuration
  ansible.builtin.copy:
    content: "# managed by Ansible\n"
    dest: /tmp/example-service.conf
    mode: "0644"
  notify: Record configuration reload

handlers:
  - name: Record configuration reload
    ansible.builtin.debug:
      msg: A real service would reload now
```

`notify` names the handler. The `handlers:` list is at play level, aligned with `tasks:`. `debug` makes this demonstration harmless; a later service handler will use `ansible.builtin.service` or `ansible.builtin.systemd_service` as appropriate. First run: copy changes and handler runs. Second run: copy is `ok`, so no handler runs.

## Practical exercise

Add the example to a safe playbook and run it twice. Then change the content once and verify the handler runs once. Add a second copy task that notifies the same handler and confirm it still runs only once in that play.

## Expected result

The handler appears only after a relevant change and is deduplicated per host per play.

## Common mistakes

- **Putting `handlers:` inside `tasks:`.** It must be a peer of `tasks` in the play.
- **Notifying after a read-only task.** Handlers should reflect actual configuration change.
- **Restarting rather than reloading without reason.** Reload preserves availability where the service supports it.

## Key takeaways

Handlers connect true configuration changes to a single deferred reaction. They depend on idempotent, accurately reporting tasks.

## Next lesson

Lesson 16 adds tags to run a deliberate subset of a larger playbook.
