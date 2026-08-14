# Lesson 40 — Managing SSH Credentials

## Learning objectives

- Use dedicated SSH keys for Ansible access.
- Protect private keys and verify host identity.
- Separate credentials from inventory data.

## Prerequisites

Lessons 22 and 39.

## Concept

Ansible normally delegates authentication to OpenSSH. A dedicated key pair identifies the control workflow and allows revocation without breaking a human administrator's key. The private key stays on the control node or secure credential system; the public key is installed on managed accounts. Inventory can select a key path with `ansible_ssh_private_key_file`, but key material never belongs in it.

Host key checking verifies that the server you reached is the one you expected. Rebuilt VMs legitimately change host keys, but blindly disabling checking makes a management network vulnerable to impersonation. Maintain `known_hosts` deliberately or use a controlled bootstrap process.

## Mental model

The private key is your signing stamp, the public key is the allowed-stamp list on a server, and `known_hosts` is your verified-address ledger.

## Example

``` bash
ssh-keygen -t ed25519 -f ~/.ssh/ansible_lab -C 'ansible lab automation'
```

`-t ed25519` selects a modern key type. `-f` chooses the private-key file path and creates a `.pub` public counterpart. `-C` adds an identifying comment. Use a passphrase unless a secure noninteractive credential mechanism protects the key. This shell command runs only on the control node; install the public half through the managed `authorized_key` task.

## Practical exercise

Create or identify one dedicated lab automation key. Confirm private-key permissions are restrictive, add its public key to the `ansible` account, and test a fresh connection using `ssh -i` before making it the Ansible default.

## Expected result

Ansible can authenticate noninteractively through normal SSH-agent/key configuration, while the private key never enters Git or a VM.

## Common mistakes

- **Reusing a personal key everywhere.** Revocation and attribution become difficult.
- **Disabling host key checking permanently.** Resolve intended rebuild key changes explicitly.
- **Placing private key text in a Vault file by default.** Prefer a dedicated credential system; if unavoidable, control file permissions and exposure rigorously.

## Key takeaways

SSH credentials are a high-value control path. Use separate keys, secure private material, and verify server identity.

## Next lesson

Lesson 41 maps secret values to safe variable design.
