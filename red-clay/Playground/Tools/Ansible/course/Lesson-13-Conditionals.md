# Lesson 13 --- Conditionals

## Learning objectives

- Use `when` to run a task only when a condition is true.
- Write conditions without `{{ }}` wrappers.
- Interpret `skipped` outcomes.

## Prerequisites

Lessons 1–12. You should know facts and registered result values.

## Concept

A **conditional** lets Ansible make a task applicable only to hosts meeting a condition. `when` evaluates an expression for each host. Unlike normal template strings, expressions in `when` are already evaluated, so do not wrap variables in `{{ }}`. A false condition gives `skipped`, which is usually a correct, auditable result—not an error.

Conditionals are essential for cross-distribution roles, optional lab services, and actions based on measured state. Keep them understandable. A deeply nested expression is often a signal that values should be organized better in inventory or roles.

## Mental model

`when` is a gate before one checklist item. Every host reaches the gate, but only eligible hosts enter.

## Example

``` yaml
- name: Explain the Arch package manager
  ansible.builtin.debug:
    msg: This is an Arch-family target
  when: ansible_facts['os_family'] == 'Archlinux'
```

The fact-gathering default must be enabled. The string comparison is quoted inside the YAML value. If `web01` is Alpine, the task is skipped; it has not failed and nothing needs fixing. Exact fact values should be verified in your environment with Lesson 11's filtered setup command.

For a registered command result, a useful pattern is:

``` yaml
when: hostname_result.rc == 0
```

This executes only if the earlier command succeeded. It does not repair a failed prerequisite; blocks and error handling will later cover that case.

## Practical exercise

Add one debug task that runs only on your target's reported OS family and one that intentionally uses a different family. Run the playbook and identify the `ok` and `skipped` outcomes. Do not use a conditional merely to hide an actual failure.

## Expected result

One task prints its message and the other is visibly skipped. The recap increments `skipped` but reports no failure.

## Common mistakes

- **Writing `when: "{{ some_var }}"`.** This is unnecessary and can produce warnings or confusing type conversions.
- **Comparing strings to booleans.** Use `when: feature_enabled`, not `when: feature_enabled == 'true'`, when the value is a YAML boolean.
- **Using conditionals to compensate for bad targeting.** Select the correct inventory group first.

## Key takeaways

`when` evaluates a per-host expression. A `skipped` task is intentionally not run; it is not a failed task.

## Next lesson

Lesson 14 repeats a single task over a controlled list with loops.
