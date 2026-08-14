# Lesson 73 — Building a Complete Lab Deployment Workflow

## Learning objectives

- Execute a staged full-lab deployment safely.
- Define verification gates and artifacts.
- Roll out roles in dependency-aware order.

## Prerequisites

Lessons 56–72.

## Concept

A complete deployment is a workflow, not a single `site.yml` command. Use explicit gates: validate Terraform, apply infrastructure, produce/inspect inventory, wait for connections, apply baseline, configure foundational services (DNS/identity), configure databases/internal services, configure web/apps, then verify from client networks. Keep logs and result summaries without leaking secrets.

Full automation should remain idempotent. A failed web-role deployment should let you correct and rerun that stage, not require a lab destroy. Staging also limits blast radius while you learn.

## Mental model

Build and commission a small enterprise in dependency order: roads/addresses, access, foundations, shared utilities, applications, then user-facing checks.

## Example

``` text
1. terraform plan/apply
2. Generate inventory; inspect graph and host list
3. Ansible readiness play
4. base_linux on Linux hosts (start with --limit)
5. DNS/domain foundations
6. database and internal-service roles
7. web/application roles
8. Client-path health, firewall, and access verification
```

Each numbered step has an owner, input, expected result, and rerun plan. In CI, gate step 7 on successful step 6; do not simply impose arbitrary sleeps.

## Practical exercise

Turn this sequence into `docs/deployment-runbook.md` for your current one-VM lab. Include exact inventory pattern, check-mode use, post-apply verification, and a rollback/console-access note. Expand it as machines arrive.

## Expected result

You can deploy or troubleshoot the lab through documented checkpoints rather than memory.

## Common mistakes

- **Applying `site.yml` to an unverified generated inventory.** Inspect targeting first.
- **Configuring applications before DNS/DB dependencies.** Failures become misleading.
- **Calling a green Ansible recap end-to-end verification.** Test actual client paths.

## Key takeaways

An enterprise deployment is staged, observable, idempotent, and verified from service consumers.

## Next lesson

Lesson 74 validates reproducibility by destroying and rebuilding intentionally.
