# Lesson 39 — Ansible Vault

## Learning objectives

- Explain what Vault encrypts and what it does not.
- Create and use an encrypted variable file.
- Keep decryption keys out of Git.

## Prerequisites

Lessons 31–38 and basic Git hygiene.

## Concept

Ansible Vault encrypts files or YAML values at rest in your repository. It is useful for passwords, tokens, private configuration values, and other secrets that must be versioned alongside automation. Vault does not protect a secret after it is decrypted on the control node, displayed in logs, passed to a process, or written insecurely to a target. It is one layer of a secret-management design.

Use a separate vault password source outside the repository, preferably a secure password manager integration or protected local file with restrictive permissions. Team access and key rotation are operational concerns; encryption does not remove them.

## Mental model

Vault locks a document while it is stored. You still must control who has keys, who reads it after opening, and where copies go.

## Example

Create an encrypted file interactively:

``` bash
ansible-vault create group_vars/databases/vault.yml
```

The command opens an editor after requesting a vault password, then encrypts the saved file. Put normal YAML inside, for example `database_backup_password: ...`; do not include plaintext in shell history. Later run a play with `--ask-vault-pass` for a manual course lab, or `--vault-password-file /secure/path` for a protected automation mechanism. The password file itself must not be committed.

## Practical exercise

Create a Vault-encrypted dummy variable with a clearly non-production value. Confirm `git diff` shows ciphertext rather than plaintext. Use `ansible-vault view` to read it, then ensure the value never appears in a debug task.

## Expected result

The file is unreadable without the vault password but is available to authorized Ansible runs.

## Common mistakes

- **Committing a vault password file.** This defeats repository encryption.
- **Assuming Vault encrypts task output.** It does not.
- **Using one shared unmanaged password forever.** Plan access and rotation.

## Key takeaways

Vault encrypts stored secret material, not its entire lifecycle. Protect decryption credentials and output separately.

## Next lesson

Lesson 40 secures the SSH credentials that give Ansible control access.
