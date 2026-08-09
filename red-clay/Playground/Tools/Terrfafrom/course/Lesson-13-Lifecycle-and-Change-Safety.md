# Lesson 13 --- Lifecycle and Change Safety

## Learning Objectives

- Read Terraform's proposed change actions carefully.
- Use lifecycle rules only for a documented reason.
- Protect critical lab infrastructure from accidental replacement or deletion.
- Separate disposable workloads from persistent data.

------------------------------------------------------------------------

## Every Plan Is a Change Review

Terraform plans may create (`+`), update in place (`~`), delete (`-`), or replace (`-/+`) resources. A plan is the safety boundary between editing code and changing infrastructure. Always review the resource address, action, and values before applying.

For a VM, distinguish its compute definition from its data. Replacing a VM may be acceptable if its data lives on protected storage; replacing a disk may be catastrophic. Provider schemas decide which arguments force replacement, so verify them with an actual plan.

## Lifecycle Meta-Arguments

Lifecycle controls Terraform's behavior for a resource:

```hcl
resource "example_vm" "important" {
  name = "directory-01"

  lifecycle {
    prevent_destroy = true
  }
}
```

Useful rules include:

| Rule | Use |
| --- | --- |
| `prevent_destroy` | A guardrail for critical objects; remove it intentionally when retirement is approved. |
| `create_before_destroy` | Reduces downtime where two objects can coexist temporarily. Requires enough quota, IPs, and names. |
| `ignore_changes` | Temporarily ignores a specific externally managed field. Use sparingly. |
| `replace_triggered_by` | Makes replacement explicit when another object changes. |

## `ignore_changes` Is Not a Drift Strategy

Ignoring changes can be appropriate when another controller owns a field, such as an automatically assigned value. It is dangerous when used to silence unwanted plans: Terraform will no longer enforce that part of the desired state. Record the external owner and a review date whenever you use it.

## Practical Lab Boundaries

- Put disposable training VMs in a separate state from shared DNS, routers, or storage.
- Keep persistent data volumes separate from replaceable VM definitions where the platform supports it.
- Require a saved plan or a second review before shared-infrastructure applies.
- Test provider upgrades and replacement behavior in a noncritical lab first.

These practices work whether the implementation is libvirt, Proxmox, VMware, or a cloud provider.

## Exercises

1. Read a plan containing `-/+` and identify the attribute forcing replacement.
2. Add `prevent_destroy` to a throwaway resource, confirm the planned destroy fails, then remove the guard intentionally.
3. Describe when `create_before_destroy` could fail for a VM with a fixed name or IP.
4. Classify your lab's objects as disposable, recoverable, or critical.

## Next Lesson

**Lesson 14 --- Modules** packages reusable infrastructure patterns without sacrificing clarity.
