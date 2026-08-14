# Lesson 67 — Configuring Database Servers

## Learning objectives

- Design a database role that respects persistent data.
- Manage service configuration and credentials safely.
- Limit database network exposure.

## Prerequisites

Lessons 39–43, 65, and Linux service management.

## Concept

Databases are stateful systems. Package/service configuration is idempotent, but schema migration, initial data, password rotation, and deletion need explicit lifecycle design. Never treat a production-like database data directory as disposable merely because a VM role runs successfully. In the lab, declare whether rebuilds recreate data or restore it from a controlled snapshot/backup.

Database ports should generally accept only application subnets/hosts and management paths, never every lab network. Store service credentials in Vault and use `no_log` for tasks that handle them.

## Mental model

The database role builds and operates the vault; application data and schema are separate valuables with their own lifecycle.

## Example

``` yaml
- name: Ensure PostgreSQL is enabled and running
  ansible.builtin.service:
    name: postgresql
    enabled: true
    state: started
```

The service name and initialization procedure vary by OS/package. Some PostgreSQL packages require an initialized cluster before start; implement that with an idempotent module/guard after consulting current platform documentation, not a blind `initdb` every run.

## Practical exercise

Write a database role design document: package, service name, data directory, initialization detection, backup target, permitted client network, and secret sources. Do not automate destructive initialization until you can test its second-run behavior.

## Expected result

You can state what data survives, who can connect, and how the role avoids overwriting an initialized database.

## Common mistakes

- **Running initialization commands unconditionally.** They may fail or destroy state.
- **Exposing port 5432/3306 broadly.** Limit by source and role.
- **Debugging credentials.** Use Vault/no_log and rotate leaked values.

## Key takeaways

Database automation needs lifecycle and data-safety design beyond package/service idempotency.

## Next lesson

Lesson 68 configures internal services and their dependencies.
