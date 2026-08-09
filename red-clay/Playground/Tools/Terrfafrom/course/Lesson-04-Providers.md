# Lesson 4 --- Providers

## Learning Objectives

By the end of this lesson, you should be able to:

- Explain why Terraform needs providers.
- Declare, install, and pin a provider.
- Configure the libvirt provider without storing secrets in code.
- Distinguish a provider requirement from a provider configuration.

------------------------------------------------------------------------

## Providers Are Terraform's Drivers

Terraform Core knows how to read HCL, calculate a plan, and record state. It does **not** know how to create a QEMU virtual machine. A provider is the plugin that translates Terraform's requests into calls for a specific platform.

For your lab, that platform is libvirt, the API used by QEMU/KVM and tools such as virt-manager.

    Terraform configuration
              |
              v
    libvirt provider plugin
              |
              v
    libvirtd / virtqemud on a KVM host
              |
              v
    networks, volumes, and domains

## Provider Requirement vs Configuration

The `required_providers` block says *which plugin and versions the project needs*. The `provider` block says *how this project connects to that platform*.

```hcl
terraform {
  required_version = ">= 1.6, < 2.0"

  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "~> 0.8"
    }
  }
}

provider "libvirt" {
  # For a local KVM host, this is normally enough.
  uri = "qemu:///system"
}
```

`terraform init` reads the requirement, downloads the provider, and records the selected exact version in `.terraform.lock.hcl`. Commit that lock file: it makes a teammate or a second controller use the same provider build.

## Connecting to a Remote KVM Host

Libvirt supports URIs over SSH. A typical remote URI is:

```hcl
provider "libvirt" {
  uri = "qemu+ssh://terraform@kvm-01.example.internal/system"
}
```

Before Terraform is involved, verify that your normal SSH account and libvirt permissions work:

```bash
virsh -c qemu+ssh://terraform@kvm-01.example.internal/system list --all
```

Use SSH keys and host-key verification. Do not put passwords, private keys, or tokens in `.tf` files. Where a provider has a sensitive setting, pass it through an environment variable or an ignored `.tfvars` file instead.

## Provider Aliases

An alias creates a second named connection to the same provider. This is the foundation for a small, fixed set of KVM hosts.

```hcl
provider "libvirt" {
  alias = "kvm01"
  uri   = "qemu+ssh://terraform@kvm-01.example.internal/system"
}

provider "libvirt" {
  alias = "kvm02"
  uri   = "qemu+ssh://terraform@kvm-02.example.internal/system"
}

resource "libvirt_network" "lab_on_kvm01" {
  provider = libvirt.kvm01
  name     = "lab-net"
  mode     = "nat"
}
```

Provider aliases are declared statically; Terraform cannot create provider configurations dynamically with `for_each`. For a few known hypervisors, explicit aliases are clear and reliable. For a larger fleet, use an orchestration design that gives each host or tenant its own Terraform working directory and state.

## Portability Note --- Changing Platforms

The Terraform concepts here do not change with the platform. A libvirt provider talks to libvirt; a Proxmox provider talks to the Proxmox API; vSphere and cloud providers use their own control planes. In every case, pin versions, protect credentials, test connectivity, and use aliases for distinct targets.

## Useful Commands

```bash
terraform init
terraform providers
terraform validate
terraform version
```

Run `terraform init -upgrade` deliberately when you intend to consider newer allowed provider versions. Review its lock-file change and test the plan before applying it.

## Common Mistakes

- **Leaving provider versions unconstrained:** a later provider release can change behavior underneath a stable configuration.
- **Treating `qemu:///session` and `qemu:///system` as interchangeable:** they use different libvirt scopes, permissions, and network capabilities. For a managed lab, `qemu:///system` is usually appropriate.
- **Assuming a successful SSH login proves libvirt access:** test with `virsh -c ...` too.
- **Putting a remote URI into every resource:** configure it once in the provider.

## Exercises

1. Create a new empty directory and write a `versions.tf` with the Terraform and libvirt requirements above.
2. Run `terraform init`, then locate `.terraform.lock.hcl`. What exact version was selected?
3. Use `virsh` to prove you can list domains on your intended local or remote KVM host.
4. Explain why an alias is needed when one configuration manages two hosts.

## Next Lesson

**Lesson 5 --- Resources** introduces the objects Terraform can create, read, update, and destroy.
