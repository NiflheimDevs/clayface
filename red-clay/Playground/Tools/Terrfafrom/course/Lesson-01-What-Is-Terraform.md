# Lesson 1 --- What Is Terraform?

## Learning Objectives

By the end of this lesson, you should be able to:

-   Explain what Infrastructure as Code (IaC) is.
-   Describe Terraform's primary purpose.
-   Distinguish declarative from imperative approaches.
-   Understand the concept of **desired state**.

------------------------------------------------------------------------

## What Problem Does Terraform Solve?

Imagine you need to build an enterprise lab containing:

-   2 Ubuntu servers
-   1 Windows Domain Controller
-   1 OPNsense firewall
-   3 virtual networks

You can create everything manually using a GUI.

Now imagine:

-   You accidentally delete a VM.
-   You want to recreate the lab next week.
-   A teammate wants the exact same environment.

Manual configuration quickly becomes slow, error-prone, and difficult to
reproduce.

Terraform solves this by allowing you to **describe infrastructure as
code**.

------------------------------------------------------------------------

## Infrastructure as Code (IaC)

Infrastructure as Code means your infrastructure is described in text
files instead of being created manually.

Example (conceptual):

``` hcl
resource "virtual_machine" "web01" {
  cpu    = 2
  memory = 4096
}
```

The configuration becomes part of your project and can be
version-controlled.

------------------------------------------------------------------------

## Terraform's Real Job

Many beginners think Terraform's job is:

> "Create virtual machines."

That is only one possible outcome.

Terraform's real job is:

> **Reconcile the real infrastructure with the desired infrastructure
> described in code.**

Terraform continually answers:

> "What changes are needed so reality matches the configuration?"

------------------------------------------------------------------------

## Declarative vs Imperative

### Imperative

You specify **how** to perform every step.

Example:

1.  Create network
2.  Create disk
3.  Create VM
4.  Attach network
5.  Boot VM

### Declarative

You specify **what** you want.

Terraform determines the necessary operations.

This is Terraform's model.

------------------------------------------------------------------------

## Desired State

Terraform compares three things:

              Desired State
             (.tf configuration)
                     │
                     ▼
              Terraform Engine
                     │
            Computes Differences
                     │
                     ▼
            Real Infrastructure

If they match:

    No changes.

If they differ:

Terraform plans the required create, update, or destroy actions.

------------------------------------------------------------------------

## Example

Desired configuration:

-   2 Ubuntu VMs
-   1 Firewall

Current infrastructure:

-   1 Ubuntu VM
-   0 Firewalls

Terraform determines that it should:

-   Create 1 Ubuntu VM
-   Create 1 Firewall

Nothing more.

------------------------------------------------------------------------

## Key Terms

  -----------------------------------------------------------------------
  Term                         Meaning
  ---------------------------- ------------------------------------------
  Infrastructure               Servers, networks, storage, firewalls,
                               cloud resources

  IaC                          Infrastructure as Code

  Desired State                The infrastructure described in Terraform
                               files

  Real State                   The infrastructure that currently exists

  Declarative                  Describe *what* you want

  Imperative                   Describe *how* to do it
  -----------------------------------------------------------------------

------------------------------------------------------------------------

## Common Misconceptions

❌ Terraform is a shell script.

No. It computes a plan from the desired state.

❌ Terraform runs commands sequentially.

Not necessarily. It builds a dependency graph and performs independent
operations in parallel when possible.

❌ Terraform only works with cloud providers.

No. It supports many providers, including KVM/libvirt, Docker,
Kubernetes, VMware, and more.

------------------------------------------------------------------------

## Summary

The most important concept in Terraform is:

> **Terraform is a declarative Infrastructure-as-Code tool that
> reconciles the real infrastructure with the desired infrastructure
> described in code.**

Everything else in Terraform builds on this idea.

------------------------------------------------------------------------

## Exercises

1.  In your own words, explain the difference between declarative and
    imperative approaches.
2.  Why is Infrastructure as Code useful for reproducibility?
3.  If your configuration defines three VMs but only two exist, what
    should Terraform do?

------------------------------------------------------------------------

## Next Lesson

**Lesson 2 --- Terraform Architecture**

You'll learn:

-   Providers
-   Resources
-   APIs
-   What happens internally when you run `terraform apply`
