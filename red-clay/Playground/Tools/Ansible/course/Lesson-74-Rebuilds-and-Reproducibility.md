# Lesson 74 — Destroying and Rebuilding the Lab Reproducibly

## Learning objectives

- Plan a safe destroy/rebuild experiment.
- Distinguish disposable infrastructure from persistent data.
- Use rebuilds to find hidden manual steps.

## Prerequisites

Lessons 56–73 and a lab you are authorized to destroy.

## Concept

Reproducibility is proven by rebuilding, not claimed because files exist. A deliberate destroy/rebuild exposes undocumented steps, stale inventory, missing bootstrap dependencies, and non-idempotent roles. It also exposes which state should persist: backups, certificates, external DNS records, and credentials require planned handling outside a disposable VM lifecycle.

Never run destroy against an unclear or valuable environment. Confirm Terraform workspace/state, exact resources, backup/restore plan, and external dependencies first. A rebuild lab is a controlled experiment, not a recovery plan for irreplaceable data.

## Mental model

Rebuilding is a fire drill for infrastructure knowledge: it proves the written system can replace memory.

## Example

Rebuild evidence checklist:

``` text
- Terraform recreates intended networks and VMs.
- Generated inventory contains current addresses only.
- Readiness succeeds without hand-edited SSH state.
- Base and service roles converge from a clean OS.
- Required services pass client-path verification.
- Exceptions/manual steps are recorded as defects to automate or document.
```

Use a separate lab environment and explicit Terraform workspace. Review a plan before any destroy/apply; state and provider settings determine scope.

## Practical exercise

Do not destroy yet. First create a rebuild plan for one disposable `web01`: backups needed, commands, acceptance checks, expected generated files, and recovery method if access fails. Perform it only when every target is verified disposable.

## Expected result

A clean rebuild either succeeds end-to-end or produces a precise list of manual assumptions to remove.

## Common mistakes

- **Destroying shared networks/resources inadvertently.** Inspect plan and workspace/state.
- **Calling a rebuild successful before client checks.** VM existence is insufficient.
- **Preserving unexplained manual changes.** Convert them into code or documented intentional exceptions.

## Key takeaways

Rebuilds are the strongest test of Terraform/Ansible boundaries and idempotent lab automation.

## Next lesson

Lesson 75 is the final enterprise-scale implementation project.
