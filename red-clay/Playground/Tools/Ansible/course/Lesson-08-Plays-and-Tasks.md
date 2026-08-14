# Lesson 8 --- Plays and Tasks

## Learning objectives

- Distinguish a play from a task.
- Predict task ordering and per-host outcomes.
- Use descriptive task names to make runs auditable.

## Prerequisites

Lessons 1–7. Keep the verification playbook available.

## Concept

A **play** maps a set of hosts to a sequence of work. A **task** is one invocation of a module with arguments. Plays run in file order; tasks run in listed order for each selected host. Ansible commonly processes multiple hosts in parallel, but on any one host it preserves task order. Later lessons cover strategies and serial execution when ordering between hosts matters.

Task names are not decoration. They are the primary readable trace of what a playbook was trying to do when a large lab run fails. Name the intended result, not an implementation accident: “Ensure nginx is installed” is better than “Run package thing.”

Ansible's final **recap** counts outcomes per host. `ok` includes successful checks and tasks that found compliance. `changed` means the module claims it altered state. `failed` stops that host's normal execution unless you intentionally handle the error. `skipped` occurs when a conditional excludes the task.

## Mental model

A play is a route assigned to selected machines. Tasks are its ordered stops. Each host keeps its own completion record, so one failed VM does not erase the evidence from the others.

## Example

Extend `playbooks/verify.yml`:

``` yaml
---
- name: Verify Linux lab connectivity
  hosts: webservers
  gather_facts: false
  tasks:
    - name: Confirm Ansible can execute a module
      ansible.builtin.ping:

    - name: Show the remote hostname
      ansible.builtin.command: hostname
      changed_when: false
```

The blank line is for readability, not semantics. `command: hostname` runs the `hostname` executable without a shell. Query commands may not modify the machine, but generic command modules cannot always know that; `changed_when: false` tells Ansible this task is a read-only check. We will explain this expression more fully in Lessons 12–13.

Run the same `ansible-playbook -i inventory/lab.ini playbooks/verify.yml` command. The task's standard output is available with more verbosity (`-v`) if needed; avoid adding verbosity by habit because it can expose sensitive connection details later.

## Practical exercise

Add a second read-only command task that reports the current user (`id -un` is suitable). Ensure it cannot report `changed`. Run the playbook twice and compare both recaps.

## Expected result

Both tasks report `ok` on every reachable web host. The second run has the same recap, demonstrating that inspection tasks should not manufacture changes.

## Common mistakes

- **Thinking a task runs once total.** It runs once per selected host unless a later `run_once` setting changes that.
- **Using vague names.** They become expensive to debug across dozens of hosts.
- **Ignoring an unexpected `changed` from a command.** Decide whether the command mutates state and set `changed_when` only when the evidence supports it.

## Key takeaways

Plays target hosts; tasks are ordered module operations. The recap is a per-host audit trail, not decorative output.

## Next lesson

Lesson 9 makes idempotency concrete and shows why configuration should converge instead of merely execute.
