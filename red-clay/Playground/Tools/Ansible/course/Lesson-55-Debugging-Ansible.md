# Lesson 55 — Debugging Ansible

## Learning objectives

- Diagnose failures by layer: selection, transport, privilege, module, service.
- Use safe verbosity and inspection commands.
- Preserve evidence rather than guessing fixes.

## Prerequisites

Lessons 1–54.

## Concept

Debug in order: confirm inventory selection, normal SSH connectivity, remote Python, privilege escalation, module arguments, then the managed service's own logs/status. Changing YAML before knowing which layer failed is slow and can make the problem worse. Ansible's error message, task name, target host, and recap provide initial evidence.

Use `--syntax-check`, `--list-hosts`, `--list-tasks`, `--check`, `-v` through `-vvvv`, `ansible-inventory --graph`, and `ansible-doc`. Higher verbosity can expose credentials, paths, and module arguments; capture and share it carefully.

## Mental model

Troubleshooting is tracing a packet through layers: did you choose the host, reach it, authenticate, elevate, invoke the right module, and leave the service healthy?

## Example

``` bash
ansible-playbook playbooks/base.yml --syntax-check
ansible-playbook playbooks/base.yml --list-hosts
ansible web01 -m ansible.builtin.ping -vv
```

`--syntax-check` parses playbook syntax but does not prove variables, host reachability, or service behavior. `--list-hosts` proves the selected inventory set without connection. `-vv` adds connection/debug detail; increase only as needed and protect the output.

## Practical exercise

Create a controlled failure: use a nonexistent inventory address or a safe nonexistent package on one disposable VM. Follow the layer order, record the evidence, and revert the deliberate mistake. Do not suppress the failure with `ignore_errors`.

## Expected result

You identify the failure layer and its exact cause rather than merely finding a command that appears to work.

## Common mistakes

- **Starting with `-vvvv` for every run.** Noise and secret risk increase.
- **Calling an unreachable host a YAML problem.** Inspect transport first.
- **Testing only the Ansible recap.** Verify service and network behavior after changes.

## Key takeaways

Debug systematically from selection through service verification. Evidence beats trial-and-error.

## Next lesson

Lesson 56 clarifies Terraform and Ansible ownership before connecting the two workflows.
