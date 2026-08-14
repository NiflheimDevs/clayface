# Lesson 45 — `run_once`

## Learning objectives

- Use `run_once` for a task that must execute once per play.
- Understand the first-host and failure implications.
- Combine it cautiously with delegation.

## Prerequisites

Lesson 44.

## Concept

`run_once: true` tells Ansible to execute a task only for the first available host in the play's current batch. It is useful for a schema initialization, a single API request, or report generation. It is not a distributed lock, leader-election system, or guarantee that a particular hostname performs the task.

When the operation logically belongs on the control node or a specific coordinator, combine `run_once` with `delegate_to: localhost` or an explicit host. Make retry and idempotency behavior clear because a partial run can leave an external action completed even if later hosts fail.

## Mental model

`run_once` is one shared meeting agenda item, assigned to the first attendee who arrives—not a promise about who that attendee will be.

## Example

``` yaml
- name: Generate one deployment report
  ansible.builtin.debug:
    msg: Deployment started for web tier
  run_once: true
  delegate_to: localhost
```

The task appears in a `webservers` play but runs once on the control node. A real report-writing task should use a known path and avoid overwriting meaningful existing data unexpectedly.

## Practical exercise

Add the example to a multi-host play. Change the inventory order and observe why the logical first host is not a suitable place for database initialization. Then keep delegation to localhost and explain the result.

## Expected result

The message appears once per play/batch, rather than once per web host.

## Common mistakes

- **Using it for configuration that every host needs.** Those tasks must run per host.
- **Assuming it survives separate playbook runs.** It is scoped to current execution.
- **Using it without a clear owner for side effects.** Delegate external actions explicitly.

## Key takeaways

`run_once` is for shared, idempotent operations, usually with an explicit execution host.

## Next lesson

Lesson 46 returns to the privilege mechanism used throughout Linux management.
