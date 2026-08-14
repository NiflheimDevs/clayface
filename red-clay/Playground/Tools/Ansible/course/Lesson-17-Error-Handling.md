# Lesson 17 --- Error Handling

## Learning objectives

- Read a failure before suppressing it.
- Use `failed_when` and `changed_when` for a known command contract.
- Explain `ignore_errors` risks.

## Prerequisites

Lessons 1–16, especially registered results and conditionals.

## Concept

By default, a task failure stops later normal tasks for that host. This protects the machine from running dependent work against an unknown state. Good error handling models an expected exceptional outcome precisely; it does not turn red output green indiscriminately.

`failed_when` defines which result values count as failure. `changed_when` defines which results count as change. `ignore_errors: true` continues after a failure but records it; use it rarely, only when later tasks remain genuinely safe and you will inspect the result. `any_errors_fatal` and `max_fail_percentage` later control behavior across groups.

## Mental model

Failure handling is a written exception policy. “Ignore everything” is not a policy; it is discarding the instrument panel.

## Example

``` yaml
- name: Test whether an optional marker exists
  ansible.builtin.command: test -f /tmp/optional-marker
  register: marker_test
  changed_when: false
  failed_when: marker_test.rc not in [0, 1]
```

`test -f` exits `0` when the file exists and `1` when it does not. For this task both are expected observations, not Ansible failure. Other return codes remain failures. The list `[0, 1]` is YAML's inline-list notation. This is safer than `ignore_errors` because it documents exactly the accepted contract.

## Practical exercise

Run the example once with the file absent and once after creating it with Ansible. Add a debug message conditioned on `marker_test.rc == 1`; do not create the marker with a shell command.

## Expected result

Both expected states complete successfully with no false `changed`; an unexpected command error still fails.

## Common mistakes

- **Using `ignore_errors` to continue through SSH or privilege failures.** Later configuration cannot be trusted.
- **Accepting every nonzero return code.** It hides syntax errors, missing commands, and permission problems.
- **Forgetting `changed_when: false` for a test command.** Tests observe; they do not change state.

## Key takeaways

Preserve failures by default. When an exception is expected, define its exact evidence with `failed_when`.

## Next lesson

Lesson 18 previews playbook behavior with check and diff modes.
