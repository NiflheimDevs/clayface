# Lesson 38 — Group Variables and Host Variables

## Learning objectives

- Put shared values in `group_vars` and exceptions in `host_vars`.
- Understand the relationship to inventory grouping.
- Avoid using host vars as a default dumping ground.

## Prerequisites

Lessons 10, 33, and 37.

## Concept

`group_vars/<group>.yml` supplies variables to every host in a group. `host_vars/<hostname>.yml` supplies values to one inventory host. These files separate data from task logic and keep the inventory itself readable. They participate in variable precedence; specific data can override broad data, so duplicate names need deliberate ownership.

Group variables are appropriate for a web port shared by web servers; host variables are appropriate for `web01`'s unique IP-bound virtual host name. If every host requires a distinct exception, reconsider whether the grouping or role interface is wrong.

## Mental model

Group vars are a team policy; host vars are an approved individual exception.

## Example

`group_vars/webservers.yml`:

``` yaml
web_nginx_listen_port: 80
```

`host_vars/web01.yml`:

``` yaml
web_nginx_server_name: web01.lab.example
```

The file names must match the inventory group and host names. A role should consume `web_nginx_*` values, not know where they came from. Use `ansible-inventory --host web01` to inspect resolved inventory data, while remembering it may show only inventory-derived values rather than every play/role source.

## Practical exercise

Create one group variable for the Linux baseline and one host variable for `web01`. Render both into your harmless template and verify only `web01` receives the host-specific setting.

## Expected result

Shared values are centralized and exceptional values remain visibly attached to the host that needs them.

## Common mistakes

- **Writing group vars for a nonexistent group.** File naming must match inventory.
- **Duplicating the same value at every precedence level.** This makes the winning source unclear.
- **Storing secrets unencrypted in `host_vars`.** Vault applies equally to all variable locations.

## Key takeaways

Place data at the broadest scope that is true, then use host vars only for real host-specific differences.

## Next lesson

Lesson 39 protects sensitive variable files with Ansible Vault.
