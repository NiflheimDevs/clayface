# Lesson 16 --- Tags

## Learning objectives

- Mark tasks with tags.
- Run or skip tagged task subsets deliberately.
- Avoid using tags as a substitute for correct playbook design.

## Prerequisites

Lessons 1–15.

## Concept

**Tags** are labels on plays, roles, blocks, or tasks. The `--tags` and `--skip-tags` command options filter what Ansible runs. They support targeted work such as applying only SSH hardening after review, or testing only a web role during development.

Tags are execution selection, not access control. A task without a selected tag normally does not run, so a tag-limited run may leave prerequisites unapplied. Full site runs remain the authoritative convergence path.

## Mental model

Tags are index tabs in the same operations manual. They let you open a section quickly; they do not make the rest of the manual irrelevant.

## Example

``` yaml
- name: Create course marker
  ansible.builtin.copy:
    content: "managed\n"
    dest: /tmp/course-marker
  tags:
    - base
    - files
```

`tags` is a YAML list. Run just this class of work:

``` bash
ansible-playbook -i inventory/lab.ini playbooks/base.yml --tags files
```

`--tags files` selects tasks carrying `files`; a task with multiple tags matches any selected tag. Preview known labels using `ansible-playbook ... --list-tags`; it reads playbook structure but does not change hosts.

## Practical exercise

Tag your package loop `packages` and marker task `files`. List tags, run each tag separately, then run the entire playbook. Explain which run proves complete base convergence.

## Expected result

Each filtered run executes only matching work; the unfiltered run executes both. Idempotent tasks stay `ok` when already satisfied.

## Common mistakes

- **Expecting a tag run to include untagged prerequisites.** It will not unless special tags are involved.
- **Using tags to hide destructive work.** Targeting, privilege design, review, and check mode are the safety controls.
- **Creating dozens of inconsistent labels.** Use a small predictable taxonomy such as `base`, `ssh`, `packages`, `web`.

## Key takeaways

Tags select work intentionally, but a complete unfiltered run is what proves the desired whole state.

## Next lesson

Lesson 17 handles expected failures without concealing real ones.
