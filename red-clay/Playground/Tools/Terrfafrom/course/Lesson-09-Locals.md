# Lesson 9 --- Local Values

## Learning Objectives

- Use locals to give repeated expressions a clear name.
- Distinguish variables, locals, and outputs.
- Build portable naming and tagging conventions.

------------------------------------------------------------------------

## Locals Are Internal Derived Values

Variables are inputs from outside a module. Outputs are values exposed from it. Locals are calculated values used only inside it.

```hcl
variable "environment" { type = string }
variable "service"     { type = string }

locals {
  name_prefix = "${var.environment}-${var.service}"
  common_tags = {
    environment = var.environment
    managed_by  = "terraform"
  }
}
```

Reference them with `local.name_prefix` and `local.common_tags`.

## Why Use Them?

Locals make an expression readable, prevent accidental inconsistencies, and create one place to change a convention. They should clarify intent, not hide a huge amount of logic.

```hcl
locals {
  vm_name = "${local.name_prefix}-01"
}

# Prefer this
name = local.vm_name

# over repeating "${var.environment}-${var.service}-01" everywhere
```

## Portable Metadata

Cloud providers often have native tags; hypervisor providers may use names, descriptions, metadata, folders, or custom attributes instead. Keep your *logical* metadata portable as a local map or object, then adapt it inside a platform module.

```hcl
locals {
  metadata = {
    owner       = "homelab"
    environment = var.environment
    role        = var.service
    managed_by  = "terraform"
  }
}
```

## Common Mistakes

- Using locals to receive environment-specific input; use variables instead.
- Duplicating a local that only wraps a literal once.
- Making names depend on values that change frequently, such as the current date.
- Hiding platform-specific resource decisions in the root configuration instead of a module.

## Exercises

1. Make a local VM name using environment, role, and sequence number.
2. Create a `metadata` local suitable for tagging, naming, or documenting resources on any platform.
3. Refactor one repeated string from an earlier exercise into a local.
4. Explain why a password should not normally be a local.

## Next Lesson

**Lesson 10 --- Expressions and Iteration** turns simple declarations into data-driven infrastructure.
