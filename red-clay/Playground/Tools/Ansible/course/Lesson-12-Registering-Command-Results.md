# Lesson 12 --- Registering Command Results

## Learning objectives

- Save a task result with `register`.
- Inspect result fields safely with `debug`.
- Distinguish `stdout`, `rc`, and changed status.

## Prerequisites

Lessons 1–11. You should understand read-only command tasks.

## Concept

`register` stores the complete result of one task in a variable for that host. It is not only for commands: modules return structured data too. Common result fields include `stdout` (text output), `stderr`, `rc` (exit code), `changed`, and `failed`. Their availability depends on the module.

Registered values are temporary execution data for the current play run; they are not automatically durable host configuration. Use them to make a measured decision, not to turn every task into shell parsing.

## Mental model

Registering is taking a measurement and labeling the notebook page so a later task can consult it.

## Example

``` yaml
- name: Read the current hostname
  ansible.builtin.command: hostname
  register: hostname_result
  changed_when: false

- name: Show the measured hostname
  ansible.builtin.debug:
    msg: "Remote hostname is {{ hostname_result.stdout }}"
```

`register: hostname_result` assigns the result object, separately for every target host. The task is read-only, so `changed_when: false` corrects the generic command module's conservative change status. `hostname_result.stdout` accesses its standard-output field. Do not call `stdout` a normal variable string until the task has actually run.

## Practical exercise

Add these tasks to your verification playbook. Add a debug task that shows `hostname_result.rc`, then run it. Identify which output is Ansible's task recap and which is the remote command's standard output.

## Expected result

The hostname command and both debug tasks are `ok`; `rc` is `0` when the command succeeds.

## Common mistakes

- **Using a registered variable before its task.** Ansible executes top-to-bottom.
- **Assuming a command exit code defines changed state.** Successful query commands can still be marked changed unless you state otherwise.
- **Printing a whole registered result in normal production runs.** It can be large and may expose sensitive output.

## Key takeaways

`register` preserves structured task evidence for a later task on the same host. Use `changed_when` carefully for read-only commands.

## Next lesson

Lesson 13 uses registered values and facts to conditionally include or skip work.
