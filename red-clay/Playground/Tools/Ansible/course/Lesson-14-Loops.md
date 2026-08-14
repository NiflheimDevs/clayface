# Lesson 14 --- Loops

## Learning objectives

- Use `loop` to repeat a task over a YAML list.
- Reference the current item and label output clearly.
- Recognize when a loop hides unrelated responsibilities.

## Prerequisites

Lessons 1–13. Understand variables and task results.

## Concept

`loop` repeats one task for every item in a list. It replaces older `with_items` syntax in new playbooks. During each iteration, the special variable `item` contains the current list value. A loop reduces duplication when every item has the same desired state, such as a collection of packages or directories.

Each loop iteration has its own outcome. Use `loop_control.label` when items are complex dictionaries so task output identifies the meaningful value rather than printing an entire data structure.

## Mental model

One task is a rubber stamp; the loop is the stack of cards beneath it. It applies the same rule to each card, not a different workflow disguised as a list.

## Example

``` yaml
- name: Ensure base diagnostic packages are installed
  ansible.builtin.package:
    name: "{{ item }}"
    state: present
  loop:
    - curl
    - ca-certificates
```

The `loop` value is a YAML list. Each pass renders `{{ item }}` as one package name. Whether both names exist depends on the distribution: validate package names on Alpine and Arch before applying a portable-looking list. The module reports change for only the missing packages.

## Practical exercise

On a disposable VM, make a two-item list of packages available for its actual distribution. Add the task to a new `playbooks/base.yml`, run it twice, and observe the per-item output. Do not add packages you do not understand to a security-sensitive machine.

## Expected result

Missing packages become installed once; the second run is `ok` for each item. The recap remains readable as one task with item results.

## Common mistakes

- **Using a comma-separated string instead of a YAML list.** `loop` expects a list; write each `-` item separately.
- **Assuming portable package names.** `python3`, `openssh`, and service packages differ by distribution.
- **Looping unrelated actions.** Separate package installation from user creation even if both list two names.

## Key takeaways

Use `loop` for repeated application of one resource model. Each item remains independently idempotent.

## Next lesson

Lesson 15 introduces handlers, which react once when a configuration genuinely changes.
