# Lesson 7 --- Input Variables

## Learning Objectives

- Declare typed input variables with useful validation.
- Supply values safely and predictably.
- Use variables to describe a VM without duplicating configuration.
- Separate non-secret defaults from environment-specific values.

------------------------------------------------------------------------

## Variables Make Configuration Reusable

Hard-coding a single VM name, memory size, and image works once. Variables let the same module or root configuration describe different roles and environments.

```hcl
variable "vm_name" {
  description = "DNS-safe name for the virtual machine"
  type        = string
}

variable "memory_mib" {
  description = "Guest memory in MiB"
  type        = number
  default     = 2048

  validation {
    condition     = var.memory_mib >= 512 && var.memory_mib % 256 == 0
    error_message = "memory_mib must be at least 512 and divisible by 256."
  }
}

variable "tags" {
  type    = map(string)
  default = {}
}
```

Use `var.vm_name`, `var.memory_mib`, and `var.tags` elsewhere in the configuration.

## Types Matter

Declare a type whenever practical. Common types are `string`, `number`, `bool`, `list(string)`, `set(string)`, `map(string)`, and `object({...})`. Types catch mistakes before the provider receives a request.

For a VM specification, an object keeps related inputs together:

```hcl
variable "vm" {
  type = object({
    name       = string
    vcpu       = number
    memory_mib = number
    role       = string
  })
}
```

## Where Values Come From

Terraform resolves values in roughly this order, with later sources taking precedence:

1. a variable `default`
2. `terraform.tfvars` or `*.auto.tfvars`
3. an explicitly supplied `-var-file=...`
4. `-var 'name=value'`
5. the `TF_VAR_name` environment variable

For a lab, keep non-secret shared defaults in `terraform.tfvars` only if it is safe to commit. Prefer named environment files, such as `lab.tfvars` and `lab.tfvars.example`, and invoke them explicitly:

```bash
terraform plan -var-file=lab.tfvars
```

Do not use command-line `-var` for secrets: command history and process lists may expose them. A `sensitive = true` variable hides values in normal CLI output but does not magically prevent them reaching state or a provider.

## Example: Role-Based Defaults

```hcl
variable "role" {
  type = string
}

locals {
  role_sizes = {
    dns = { vcpu = 1, memory_mib = 1024 }
    web = { vcpu = 2, memory_mib = 2048 }
    ad  = { vcpu = 4, memory_mib = 8192 }
  }
}

# Later: local.role_sizes[var.role].memory_mib
```

Variables describe inputs; locals, covered in Lesson 9, derive repeated internal values from them.

## Common Mistakes

- Using variables for values that never vary inside a module. A literal can be clearer.
- Supplying an undeclared variable and assuming Terraform will use it.
- Storing passwords in committed `terraform.tfvars`.
- Using strings for numbers or booleans simply because CLI input is text.

## Portability Note --- Describe Intent, Not a Hypervisor

Prefer portable inputs such as `vcpu`, `memory_mib`, `network_name`, and `image_id`. A platform-specific module can translate them into provider fields. This makes a later move from libvirt to Proxmox a module implementation change rather than a rewrite of every environment definition.

## Exercises

1. Declare variables for a VM name, vCPU count, memory in MiB, and a management IP address.
2. Add validation so vCPU is at least 1 and memory is at least 512 MiB.
3. Create `lab.tfvars.example` with fake values; keep the real local `lab.tfvars` ignored.
4. Run `terraform console` and evaluate `var.memory_mib * 2`.

## Next Lesson

**Lesson 8 --- Outputs** makes useful information from Terraform available to you and to other automation.
