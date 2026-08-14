# Lesson 60 — Managing Terraform-Created VMs

## Learning objectives

- Sequence apply, inventory generation, readiness, and configuration.
- Use `wait_for_connection` for an Ansible-ready VM.
- Limit the first configuration rollout safely.

## Prerequisites

Lessons 56–59 and a bootstrapped VM.

## Concept

A successful Terraform apply means the provider accepted the VM resources; it does not guarantee guest boot, DHCP, SSH, cloud-init completion, package database availability, or Ansible Python. Begin configuration with a targeted readiness play, then apply the base role and server role. Retain Terraform output and Ansible recap as deployment evidence.

## Mental model

Provisioning says the machine was delivered. `wait_for_connection` confirms the operator can enter and start setup.

## Example

``` yaml
---
- name: Wait for new Linux VMs
  hosts: linux
  gather_facts: false
  tasks:
    - name: Wait until Ansible can connect
      ansible.builtin.wait_for_connection:
        timeout: 300
        connect_timeout: 10
```

`timeout` is the total wait in seconds. `connect_timeout` bounds one connection attempt. This module tests Ansible's connection mechanism, not merely TCP port availability. Follow it with ping/fact gathering and the base role. Adjust timeouts to realistic image boot behavior rather than arbitrary sleep commands.

## Practical exercise

Run the readiness play against a single new or rebooted `web01`. Then apply `base_linux` with `--limit web01`. Verify SSH, package presence, and service state independently.

## Expected result

The play waits through boot delay, then configuration begins only after the connection becomes usable.

## Common mistakes

- **Using fixed `sleep 60`.** It is slow when unnecessary and insufficient when boot takes longer.
- **Treating port 22 open as a ready guest.** Authentication/Python may still be unavailable.
- **Configuring all hosts first time without a limit.** Stage one representative host.

## Key takeaways

VM creation and configuration are separate observable stages. Wait for the actual Ansible transport, then roll out deliberately.

## Next lesson

Lesson 61 formalizes the multi-stage deployment workflow.
