# Lesson 9 --- Idempotency

## Learning objectives

- Define idempotency precisely in Ansible terms.
- Verify an idempotent task by running it twice.
- Explain why `changed` accuracy matters to services and safe rebuilds.

## Prerequisites

Lessons 1–8. Have one disposable Linux VM; this lesson changes a file in `/tmp` only.

## Concept

An idempotent task reaches a requested final state and then makes no further change when the same inputs are applied again. This does not mean it never changes anything. It means the first corrective run may be `changed`, while subsequent equivalent runs are `ok`.

Idempotency is a safety property. It lets you rerun a complete base configuration after VM replacement, recover from a partial run, and apply a web role to ten hosts without unnecessarily restarting every service. Accurate `changed` reporting later controls handlers, so a falsely changed task can create real disruption.

## Mental model

An idempotent task is a thermostat: it acts only while the measured state differs from the set point. A blind command is a switch flipped every time whether the room is already correct or not.

## Example

Create `playbooks/idempotency.yml`:

``` yaml
---
- name: Demonstrate safe desired state
  hosts: webservers
  gather_facts: false
  tasks:
    - name: Create a lab marker with exact content
      ansible.builtin.copy:
        content: "managed by Ansible\n"
        dest: /tmp/ansible-course-marker
        mode: "0644"
```

`ansible.builtin.copy` manages a file's content and metadata. `content` is the literal source text; `\n` ensures a trailing newline. `dest` is the remote path. Quote `"0644"` so YAML does not interpret it as an unrelated numeric representation. On a clean VM, the file is created and reports `changed`; without a change to content or mode, run two reports `ok`.

Run it twice:

``` bash
ansible-playbook -i inventory/lab.ini playbooks/idempotency.yml
```

The command is identical on purpose. Change the `content` once, run again, and predict the result before executing it.

## Practical exercise

Run the example twice. Verify the marker using normal SSH (`cat /tmp/ansible-course-marker`) rather than assuming output proves file content. Then alter exactly one desired property—content or mode—and run it once more. Record the expected and actual recap each time.

## Expected result

Run one is `changed`; run two is `ok`; changing a declared property produces `changed` once, followed by `ok` on the next unchanged run.

## Common mistakes

- **Calling a task idempotent because it exits zero.** A command may succeed while rewriting files or restarting a service every run.
- **Leaving test artifacts in an important path.** This lesson deliberately uses `/tmp`; later roles should use explicit application paths and ownership.
- **Suppressing `changed` to make output look clean.** Fix the task's resource model instead of hiding genuine changes.

## Key takeaways

Idempotency is repeatable convergence plus accurate change reporting. It is what makes full-lab reapplication safe.

## Next lesson

Lesson 10 introduces variables so one playbook can describe several hosts without copying hard-coded values.
