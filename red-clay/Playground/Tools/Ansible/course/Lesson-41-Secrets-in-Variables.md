# Lesson 41 — Secrets in Variables

## Learning objectives

- Classify a variable as secret or non-secret.
- Keep secret values separate from ordinary configuration.
- Pass a secret to the smallest necessary scope.

## Prerequisites

Lessons 10, 38–40.

## Concept

A variable is not inherently safe because it has a secret-looking name. Passwords, API tokens, private keys, connection strings, and some internal topology details need different handling from ports or package names. Keep encrypted values in Vault-backed files and reference them from ordinary configuration by clear names. Avoid copying secrets into several templates, command arguments, and debug messages.

The best secret is one that is not needed. Prefer certificate/key-based authentication, short-lived credentials, scoped service accounts, and generated per-environment values. When a secret must reach a service, write it only to the protected destination with least-privilege ownership and mode.

## Mental model

A secret variable is a sealed component with a labeled connector. Connect it only to the exact service that needs it; do not scatter copies through the project.

## Example

`group_vars/databases/main.yml` may contain a non-secret reference name:

``` yaml
database_backup_user: backup
```

The encrypted `group_vars/databases/vault.yml` contains:

``` yaml
vault_database_backup_password: replace-with-a-real-secret
```

A protected template can refer to `vault_database_backup_password`. Prefixing with `vault_` is a convention that communicates sensitivity; it does not encrypt anything by itself. Keep the secret source file encrypted and restrict the rendered target file to the service account.

## Practical exercise

Classify five values in your planned lab as public configuration, operationally sensitive, or secret. Create an encrypted dummy variable file for one secret class and design the target file owner/mode it would require—do not print it.

## Expected result

Secret sources are distinct from ordinary variables and every intended consumer has a clear, minimal path to its value.

## Common mistakes

- **Encrypting every variable.** This makes normal review needlessly opaque without improving actual secret handling.
- **Putting a secret in a process command line.** Process listings and logs may expose it.
- **Using secret names as security.** Naming is documentation, not protection.

## Key takeaways

Separate secret data, encrypt it at rest, and minimize the scope and copies of decrypted values.

## Next lesson

Lesson 42 prevents accidental secret exposure during normal automation work.
