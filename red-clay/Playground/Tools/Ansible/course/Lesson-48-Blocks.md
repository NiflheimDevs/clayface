# Lesson 48 — Blocks

## Learning objectives

- Group related tasks with `block`.
- Use `rescue` and `always` for controlled recovery/reporting.
- Avoid pretending rollback is automatic.

## Prerequisites

Lessons 17 and 46–47.

## Concept

A **block** groups tasks so they can share directives such as `when`, `become`, and tags. `rescue` runs when a task in its block fails; `always` runs whether the block succeeds or fails. This resembles exception handling, but it does not magically restore pre-change state. Real rollback must be explicitly designed, safe, and tested.

Blocks are useful for a risky configuration update that needs validation and an explicit fallback action. Avoid rescuing every failure; a rescue should leave the machine in a known condition or report clearly that human intervention is required.

## Mental model

A block is a controlled work zone: perform planned work, take a defined emergency action if it fails, and always complete the safety log.

## Example

``` yaml
- name: Validate a candidate configuration
  block:
    - name: Deploy candidate
      ansible.builtin.copy:
        src: files/app.conf
        dest: /tmp/app.conf
    - name: Validate candidate
      ansible.builtin.command: /usr/local/bin/app --check-config /tmp/app.conf
      changed_when: false
  rescue:
    - name: Report validation failure
      ansible.builtin.debug:
        msg: Candidate was not accepted
  always:
    - name: Record validation completion
      ansible.builtin.debug:
        msg: Validation attempt finished
```

The command's exact validator is fictional; use only a documented service validator. `rescue` is not entered for every unreachable-host condition, so external connectivity still needs separate treatment.

## Practical exercise

Build a block around a harmless validation command that you can deliberately make fail. Ensure the rescue does not hide the reason in production; capture useful safe status without exposing secrets.

## Expected result

Success follows the block; an intentional validation failure triggers rescue and always. You can explain final file state.

## Common mistakes

- **Calling rescue a rollback without restoring anything.** It is only an alternate task sequence.
- **Using `ignore_errors` instead of a planned rescue.** It loses context and safety.
- **Putting unrelated tasks in one block.** Group one recoverable operation.

## Key takeaways

Blocks express grouped policy and controlled failure paths. Recovery must be explicitly designed.

## Next lesson

Lesson 49 contrasts task-file inclusion and import behavior.
