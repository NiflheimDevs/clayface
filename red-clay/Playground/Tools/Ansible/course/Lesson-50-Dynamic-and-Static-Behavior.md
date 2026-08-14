# Lesson 50 — Dynamic vs Static Behavior

## Learning objectives

- Predict parse-time versus run-time evaluation.
- Apply tags and conditions with the right inclusion mechanism.
- Diagnose “task not listed” versus “task skipped.”

## Prerequisites

Lesson 49.

## Concept

Static Ansible content is expanded during parse time, before managed-host execution. Dynamic content is resolved during execution for each host. This affects `--list-tasks`, tag inheritance, condition placement, and whether a task is visible before variables/facts exist. Neither is universally superior; predictability is usually preferable until runtime variation is necessary.

## Mental model

Static behavior plans the route before departure. Dynamic behavior chooses a turn after seeing the road condition on each host.

## Example

``` yaml
- name: Import firewall tasks for all Linux hosts
  ansible.builtin.import_tasks: firewall.yml
  tags: firewall

- name: Include an OS implementation
  ansible.builtin.include_tasks: "firewall-{{ ansible_facts['os_family'] }}.yml"
  when: ansible_facts['os_family'] in ['Alpine', 'Archlinux']
```

The import's tags apply to the statically imported tasks. The dynamic include itself is conditioned per host; files under it are not known until that point. Verify actual fact values—Alpine/Arch family strings differ from friendly distribution names.

## Practical exercise

Use `--list-tasks` on a play with one import and one include. Run it against one host and compare “not listed” with “skipped.” Write which behavior you want for your base role and why.

## Expected result

You can explain where the decision is made and why an included file may not appear in static listing.

## Common mistakes

- **Calling skipped tasks missing.** They existed and were evaluated.
- **Using dynamic filenames without an allowlist.** This can create broken or unsafe behavior.
- **Expecting identical tag inheritance.** Test the mechanism and Ansible version.

## Key takeaways

Static structure improves visibility; dynamic structure handles per-host runtime choice. Diagnose based on when selection occurs.

## Next lesson

Lesson 51 expands Jinja2 only after the data model is clear.
