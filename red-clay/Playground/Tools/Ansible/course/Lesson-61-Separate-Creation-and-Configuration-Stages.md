# Lesson 61 — Handling VM Creation and Configuration as Separate Stages

## Learning objectives

- Design a repeatable staged deployment workflow.
- Identify failure ownership by stage.
- Rerun configuration without recreating infrastructure.

## Prerequisites

Lessons 56–60.

## Concept

Keep infrastructure and configuration commands independently executable. This makes troubleshooting clear: a libvirt/provider failure belongs to Terraform; a guest boot/SSH failure belongs to bootstrap/readiness; a package/service failure belongs to Ansible. CI or a Makefile can orchestrate the stages, but should not conceal their logs or state boundaries.

Destroy/rebuild needs the inverse awareness: Terraform destroys infra, while Ansible's role state disappears with the VM. Persistent external resources—DNS, backups, certificates—need explicit lifecycle ownership so rebuilds do not leak or collide.

## Mental model

The pipeline has gates: construct, connect, configure, verify. Each gate can be inspected and rerun without pretending the others completed.

## Example

``` text
terraform init && terraform apply
generate inventory from terraform output
ansible-playbook playbooks/readiness.yml
ansible-playbook playbooks/site.yml
run service and network verification
```

The sequence communicates boundaries. In an automated runner, preserve nonzero exit codes and stop at a failed gate; do not continue to configuration after failed inventory generation.

## Practical exercise

Write your own deployment runbook with the five stages above. For each, specify input, successful evidence, and safe rerun command. Test only the readiness-to-base portion on one VM.

## Expected result

You can resume after an Ansible failure without unnecessarily destroying/recreating a VM.

## Common mistakes

- **Putting all steps in one opaque script.** Diagnosis becomes hard.
- **Ignoring nonzero status in a wrapper.** It creates a false successful deployment.
- **Destroying Terraform to fix an nginx config.** Rerun the owning Ansible role instead.

## Key takeaways

Stage separation gives repeatability and clear failure ownership across Terraform and Ansible.

## Next lesson

Lesson 62 designs the combined repository deliberately.
