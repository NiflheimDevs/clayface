# Lesson 44 — Delegation

## Learning objectives

- Run one task on a host other than the current play host.
- Explain `delegate_to` and variable context.
- Avoid accidental centralization of per-host work.

## Prerequisites

Lessons 1–43.

## Concept

`delegate_to` changes where a task executes. The task still iterates over the hosts selected by the play, but its module runs on the delegated host. Common uses are adding a newly configured web server to a load balancer, asking a DNS server to create a record, or creating a local report on the control node with `delegate_to: localhost`.

The current inventory host remains the logical subject, so variables normally refer to it. This distinction prevents a common mistake: delegation does not automatically mean “use the delegate's facts.”

## Mental model

Each web server fills out its own request form, but sends it to the load balancer's office for processing.

## Example

``` yaml
- name: Register each web server in a local report
  ansible.builtin.debug:
    msg: "Would register {{ inventory_hostname }}"
  delegate_to: localhost
```

The play's selected `webservers` cause one delegated task per web host. `localhost` is the control node. `debug` is harmless here; a real role could call a load-balancer API with an appropriate module.

## Practical exercise

Run the debug example against `webservers` and verify that the task says it executed on localhost while its message identifies each web host. Explain why delegation is appropriate for a DNS update but not for installing nginx.

## Expected result

One delegated result appears for every selected web host; the remote web hosts are still the data source.

## Common mistakes

- **Expecting delegation to select hosts.** `hosts:` controls selection; delegation controls execution location.
- **Using delegated facts accidentally.** Be explicit with `hostvars` when cross-host data is truly required.
- **Delegating privileged work to localhost.** The control node's privileges are also a security boundary.

## Key takeaways

Delegation separates the logical host from the execution host, enabling controlled cross-system orchestration.

## Next lesson

Lesson 45 limits an otherwise per-host task to one execution.
