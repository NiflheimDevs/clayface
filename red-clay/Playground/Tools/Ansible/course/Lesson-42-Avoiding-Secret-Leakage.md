# Lesson 42 — Avoiding Secret Leakage

## Learning objectives

- Identify common Ansible secret-leak paths.
- Use `no_log` narrowly.
- Review logs, diffs, and artifacts as sensitive data.

## Prerequisites

Lessons 39–41.

## Concept

Secrets can leak through `debug`, failed task output, registered results, command arguments, `--diff`, CI logs, fact caches, rendered files, and Git history. Vault only protects source files before decryption. Design tasks so sensitive values do not need to be displayed at all.

`no_log: true` suppresses task arguments and results in output. Apply it to the specific task that handles a secret, not broadly to whole plays: broad suppression removes evidence needed for troubleshooting. `no_log` cannot erase secrets a command itself writes to external logs or files.

## Mental model

Treat automation output like a shared incident room: anyone who can see it may see what a task reveals.

## Example

``` yaml
- name: Create database application credential
  community.postgresql.postgresql_user:
    name: appsvc
    password: "{{ vault_appsvc_database_password }}"
    state: present
  no_log: true
```

The collection/module must be installed and configured for the database context; this snippet illustrates output protection. `no_log: true` is at task level. Do not add a subsequent debug task that prints the registered result, because the secret should never be registered or displayed unnecessarily.

## Practical exercise

Search your repository for `password`, `token`, `secret`, and `debug`. Review each match for whether it could expose a value. Run a dummy secret-handling task with `no_log: true` and observe the redacted output; do not use a real password for the experiment.

## Expected result

Sensitive task details are redacted while unrelated tasks remain diagnosable. No plaintext secret is added to Git, terminal history, or a diff.

## Common mistakes

- **Using `--diff` on secret templates.** It can display the exact secret.
- **Adding `no_log` after a failure leaked data.** Rotation may be required; output cannot be unshared.
- **Suppressing all logs.** It makes diagnosis and audit impractical.

## Key takeaways

Secret safety includes source, execution, output, and destination. Use narrow redaction and never rely on Vault alone.

## Next lesson

Lesson 43 consolidates the security practices into a baseline review.
