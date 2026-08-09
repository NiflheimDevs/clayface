# Lesson 10 --- Expressions and Iteration

## Learning Objectives

- Reference values and combine strings, collections, and conditions.
- Use `for_each` for stable repeated resources.
- Know when `count` is appropriate and why it can be fragile.
- Model a small fleet as data rather than copied blocks.

------------------------------------------------------------------------

## Expressions Connect Configuration

An expression computes a value:

```hcl
name   = "${var.environment}-${var.role}"
memory = var.memory_mib * 1024
```

Terraform also supports conditionals:

```hcl
enable_agent = var.environment == "production" ? true : false
```

Keep expressions understandable. If an expression needs a paragraph to explain, give it a local name or simplify the input data.

## `for_each`: Repeated Objects With Stable Identities

Use a map when each machine has a meaningful stable key.

```hcl
variable "machines" {
  type = map(object({
    vcpu       = number
    memory_mib = number
    role       = string
  }))
}

# This is provider-neutral pseudocode: substitute your platform resource type.
resource "example_vm" "machine" {
  for_each = var.machines

  name       = each.key
  vcpu       = each.value.vcpu
  memory_mib = each.value.memory_mib
}
```

Its addresses are `example_vm.machine["dns01"]` and `example_vm.machine["web01"]`. Removing `dns01` affects only that instance. A real platform module can accept the same machine map and create libvirt domains, Proxmox VMs, or cloud instances.

## `count`: Good for Identical Numbered Things

```hcl
resource "example_vm" "worker" {
  count = var.worker_count
  name  = "worker-${count.index + 1}"
}
```

`count` uses numeric addresses. Removing an item from the middle of a list can shift indexes and cause unwanted replacement. Prefer `for_each` for named servers or records that may be added and removed independently.

## Dynamic Blocks

A `dynamic` block can generate repeated nested configuration. Use it only when the provider requires nested blocks; it cannot generate arbitrary resource blocks. Often a small module or normal expression is easier to read.

## Lab Mapping

Your portable `machines` map might describe `dns01`, `monitor01`, and `web01`. A libvirt implementation translates it into domains, volumes, and network interfaces. A Proxmox implementation translates it into Proxmox VM resources. Keep VM intent in the root layer and provider details in the implementation layer.

## Exercises

1. Write a `machines` map for three lab roles with different sizes.
2. Use `terraform console` to evaluate `[for name, vm in var.machines : name]`.
3. Explain why `for_each` is safer than `count` for a named fleet.
4. Add one machine, run a plan, then remove a different one. Which addresses should change?

## Next Lesson

**Lesson 11 --- Functions and the Console** helps you transform data safely.
