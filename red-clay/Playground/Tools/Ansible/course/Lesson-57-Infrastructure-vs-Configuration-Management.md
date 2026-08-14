# Lesson 57 — Infrastructure vs Configuration Management

## Learning objectives

- Separate infrastructure, bootstrap, and configuration stages.
- Identify the minimum bootstrap contract.
- Design safe rerun boundaries.

## Prerequisites

Lesson 56.

## Concept

The pipeline has three stages: infrastructure provisioning creates the VM/network; bootstrap makes it reachable (IP, SSH, user/key, Python where needed); configuration converges its role. Keeping bootstrap minimal reduces duplicated configuration and makes failures easier to locate. Terraform cloud-init is a practical bootstrap mechanism; Ansible then owns ongoing state.

Each stage should be independently rerunnable. Terraform apply should not need to configure nginx; Ansible should not need to guess whether a VM disk exists. A readiness check bridges the stages because a VM definition existing does not mean SSH is ready.

## Mental model

Build the house, install a locked front door and address, then send the interior crew. Do not ask the interior crew to build the foundation.

## Example

Bootstrap contract for a Linux VM:

``` text
- Management IP/DNS is reachable from the control node.
- SSH daemon runs and accepts the automation public key.
- Automation user can elevate as designed.
- Python 3 exists for normal Ansible modules.
```

This contract is testable with `ssh` and `ansible.builtin.ping`. Keep it stable across Alpine and Arch images even if cloud-init implementation differs.

## Practical exercise

Write your current image bootstrap contract. For each point, state whether Terraform/cloud-init, image build, or Ansible owns it. Test it manually on one newly created VM.

## Expected result

You have a precise definition of “ready for Ansible,” not merely “VM is running.”

## Common mistakes

- **Making cloud-init configure every application.** It creates a second configuration system.
- **Starting Ansible immediately after Terraform without readiness.** SSH and package locks may not be ready.
- **Using image customization and Ansible for the same baseline.** Decide source of truth.

## Key takeaways

Minimal bootstrap is the contract between provisioning and configuration. It should make Ansible possible, not replace it.

## Next lesson

Lesson 58 consumes Terraform outputs as structured Ansible inputs.
