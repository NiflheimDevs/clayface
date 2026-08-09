# Lesson 2 --- Terraform Architecture

## Learning Objectives

-   Understand Terraform's major components.
-   Explain the role of providers.
-   Understand resources and APIs.
-   Follow the lifecycle of `terraform apply`.

------------------------------------------------------------------------

## High-Level Architecture

    +----------------------+
    |  Terraform Config    |  (.tf files)
    +----------+-----------+
               |
               v
    +----------------------+
    | Terraform Core       |
    | - Parses HCL         |
    | - Builds graph       |
    | - Creates plan       |
    +----------+-----------+
               |
               v
    +----------------------+
    | Provider Plugin      |
    | (AWS, libvirt, etc.) |
    +----------+-----------+
               |
               v
    +----------------------+
    | Target API           |
    | Hypervisor / Cloud   |
    +----------------------+

Terraform Core never talks directly to AWS, KVM, Docker, or Kubernetes.
It delegates that work to **providers**.

------------------------------------------------------------------------

## Terraform Core

Terraform Core is responsible for:

-   Reading `.tf` files
-   Validating syntax
-   Loading state
-   Building a dependency graph
-   Comparing desired vs current state
-   Creating an execution plan
-   Coordinating providers

Terraform Core is infrastructure-agnostic.

------------------------------------------------------------------------

## Providers

A provider is a plugin that knows how to communicate with a platform.

Examples:

-   AWS
-   Azure
-   Google Cloud
-   Docker
-   Kubernetes
-   libvirt (KVM/QEMU)
-   VMware

Providers translate generic Terraform operations into API calls.

For example:

    resource "libvirt_domain" "web"

becomes a sequence of libvirt API calls to create a virtual machine.

------------------------------------------------------------------------

## Resources

A **resource** is a single managed object.

Examples:

-   Virtual machine
-   Network
-   Disk
-   Firewall rule
-   DNS record

General syntax:

``` hcl
resource "<TYPE>" "<NAME>" {
  # arguments
}
```

The **type** comes from the provider. The **name** is only meaningful
within your Terraform configuration.

------------------------------------------------------------------------

## What Happens During `terraform apply`?

1.  Read all `.tf` files.
2.  Validate the configuration.
3.  Download required providers (if needed).
4.  Load the state file.
5.  Query providers for current infrastructure.
6.  Compare desired and current state.
7.  Build a dependency graph.
8.  Produce an execution plan.
9.  Ask for confirmation (unless auto-approved).
10. Call provider APIs.
11. Update the state file.

This process is why Terraform is idempotent: running it again without
changes should produce no work.

------------------------------------------------------------------------

## Dependency Graph

Terraform automatically determines resource order.

    Virtual Network
          |
          v
    Virtual Machine
          |
          v
    DNS Record

If two resources are independent, Terraform can create them in parallel.

------------------------------------------------------------------------

## Common Misconceptions

-   **Terraform is not the provider.** Providers perform
    platform-specific work.
-   **Resources are not VMs only.** Anything manageable through a
    provider can be a resource.
-   **`terraform apply` is not a script runner.** It computes changes
    before acting.

------------------------------------------------------------------------

## Summary

Remember these responsibilities:

  Component        Responsibility
  ---------------- ----------------------------------------------
  Terraform Core   Planning and orchestration
  Provider         Platform-specific implementation
  Resource         Individual managed object
  API              Interface to the target platform
  State            Terraform's record of managed infrastructure

------------------------------------------------------------------------

## Exercises

1.  Why does Terraform need providers?
2.  Which component builds the dependency graph?
3.  Which component actually communicates with libvirt or AWS?
4.  Explain the difference between a provider and a resource.

------------------------------------------------------------------------

## Next Lesson

**Lesson 3 --- HCL Basics**

You'll learn the syntax of Terraform's configuration language and begin
writing your first real configurations.
