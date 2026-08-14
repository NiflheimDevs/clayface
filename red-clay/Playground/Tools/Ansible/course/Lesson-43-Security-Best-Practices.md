# Lesson 43 — Security Best Practices

## Learning objectives

- Review an Ansible project for basic security properties.
- Apply least privilege, controlled targeting, and dependency hygiene.
- Identify security boundaries Terraform and Ansible share.

## Prerequisites

Lessons 1–42.

## Concept

Secure automation protects both the control node and every managed node. Key practices are least-privilege remote accounts, dedicated SSH credentials, host-key verification, encrypted secret storage, minimized log exposure, pinned dependencies, reviewed changes, narrow inventory targeting, and staged rollout. Treat playbooks as privileged code: a compromised collection, inventory entry, or CI runner can control servers.

Terraform and Ansible meet at sensitive boundaries such as cloud-init keys, management IPs, and generated inventory. Ensure Terraform outputs do not print secrets and that Ansible does not overwrite infrastructure-owned network configuration without an explicit design.

## Mental model

Your Ansible control node is an administration workstation at scale. Its project files, keys, dependencies, and targeting rules are all part of the attack surface.

## Example

Pre-flight checklist before an impactful play:

``` text
1. Confirm the Git diff and collection versions.
2. Confirm `ansible-inventory --graph` and `--list-hosts` target only intended hosts.
3. Use `--check --diff` only when output cannot expose secrets.
4. Start with `--limit web01` and retain console access.
5. Verify service health and firewall reachability after apply.
```

These are operational controls, not Ansible syntax. They prevent a valid playbook from being deployed to the wrong targets or with unsafe credentials.

## Practical exercise

Perform this checklist on your current base playbook. Write down one control you already have and one missing control to implement before your lab grows beyond two VMs.

## Expected result

You can describe who can run automation, what it can reach, where secrets reside, and how changes are reviewed and verified.

## Common mistakes

- **Assuming an isolated lab needs no security.** It is training ground for habits and may still bridge to your home network.
- **Trusting unpinned third-party content.** Automation dependencies run with your privileges.
- **Using a broad group because it is convenient.** Target selection is a security decision.

## Key takeaways

Secure Ansible combines least privilege, secret lifecycle control, dependency review, verified identity, and disciplined rollout.

## Next lesson

Lesson 44 begins advanced execution patterns with delegation.
