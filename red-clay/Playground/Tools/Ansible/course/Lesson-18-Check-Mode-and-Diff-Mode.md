# Lesson 18 --- Check Mode and Diff Mode

## Learning objectives

- Run a playbook in check mode.
- Use diff mode safely for text-file changes.
- State why a dry run is evidence, not a guarantee.

## Prerequisites

Lessons 1–17. Have an idempotent file playbook.

## Concept

**Check mode** asks supporting modules to predict changes without applying them. Run it with `--check`. **Diff mode** asks supporting modules to show before/after text differences, using `--diff`. Together they are a review tool before making real changes.

Support is module-specific. A command or shell task cannot generally predict its outcome. Check mode may also be unable to discover a resource that would be created earlier in the same run. Never treat a green dry run as proof that production will succeed; use it to find likely changes, then apply to a narrow safe target first.

## Mental model

Check mode is a construction estimate based on current plans. It is valuable before work starts, but it is not a completed building inspection.

## Example

``` bash
ansible-playbook -i inventory/lab.ini playbooks/idempotency.yml --check --diff
```

`--check` prevents supported tasks from modifying managed nodes. `--diff` displays textual changes for modules such as `copy` and `template`. Do not pass `--diff` around secrets: changed file content may be displayed in terminal history, CI logs, or an artifact. Use `--limit web01` during a first narrow review to target one inventory host.

## Practical exercise

Change your marker content, run the file playbook with `--check --diff`, and predict whether the remote file changes. Verify it does not. Then perform a normal run and re-run check mode to confirm no predicted difference remains.

## Expected result

The first check predicts a change and may show the text diff; the remote file stays unchanged until the normal run. The final check reports no change.

## Common mistakes

- **Assuming all tasks honor check mode.** Read module documentation and test the exact playbook.
- **Leaking passwords with `--diff`.** Treat output as sensitive.
- **Skipping normal validation after apply.** A syntax-valid config may still fail its service reload.

## Key takeaways

Check mode estimates supported changes; diff mode makes file review visible. Both reduce risk but cannot replace staged validation.

## Next lesson

Lesson 19 establishes project-level defaults with `ansible.cfg`.
