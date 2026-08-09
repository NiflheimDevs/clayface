# Lesson 14 --- Modules

## Learning Objectives

- Explain root modules and child modules.
- Design a small module with clear inputs and outputs.
- Keep platform-specific details behind a reusable interface.
- Pass provider aliases into modules for multi-target deployments.

------------------------------------------------------------------------

## A Module Is a Reusable Terraform Package

Every Terraform directory is a module. The directory you run Terraform from is the **root module**. A child module is called from another module:

```hcl
module "web01" {
  source = "../../modules/virtual-machine"

  name       = "web01"
  vcpu       = 2
  memory_mib = 2048
  image_id   = var.ubuntu_image_id
}
```

A module should have a small, documented contract:

    inputs  ──► module implementation ──► outputs
    name                               vm_id
    vcpu                               address
    network                            metadata

## A Portable Module Design

Keep the root environment intent-oriented. For example, its machine map can specify name, vCPU, memory, image, and network. A `virtual-machine-libvirt` module may turn those into a volume and domain; a future `virtual-machine-proxmox` module may turn them into a Proxmox VM. They can expose similar outputs without pretending that platforms have identical capabilities.

Do not make an abstraction so generic that it hides important differences such as disk buses, cloud-init support, or live migration. A good module standardizes the 80% you repeat and exposes deliberate escape hatches for the rest.

## Module Layout

```text
modules/
└── virtual-machine-libvirt/
    ├── main.tf       # resources
    ├── variables.tf  # public input contract
    ├── outputs.tf    # public output contract
    └── README.md     # assumptions and examples
```

Run `terraform validate` and a representative plan from a real calling root module. A module has no provider configuration by default; the caller normally supplies it.

## Passing an Alias to a Module

```hcl
module "web_on_kvm02" {
  source = "../../modules/virtual-machine-libvirt"

  providers = {
    libvirt = libvirt.kvm02
  }

  name = "web02"
}
```

This lets one module implementation be used against multiple explicitly configured targets. The equivalent approach works with other providers that support aliases.

## Common Mistakes

- Creating a module before a pattern has actually repeated.
- Letting callers reach into a module's internal resource addresses.
- Adding provider configurations inside reusable modules, which makes multi-target use harder.
- Changing module input names or output meanings without treating it as an interface change.

## Exercises

1. Extract your practice volume or VM into a module with `name` as an input and `id` as an output.
2. Write a README that states the module's platform assumptions.
3. List which inputs would remain portable if moving from libvirt to Proxmox.
4. Call the module twice with different names and inspect the resource addresses in the plan.

## Next Lesson

**Lesson 15 --- Remote State and Backends** makes state safer for shared and automated work.
