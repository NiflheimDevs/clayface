# Lesson 17 --- Project Design and Best Practices

## Learning Objectives

- Structure a Terraform repository for safe change.
- Apply a repeatable plan, review, and apply workflow.
- Handle versions, formatting, secrets, and documentation.
- Design for a future platform migration.

------------------------------------------------------------------------

## A Practical Repository Shape

```text
terraform-lab/
├── modules/
│   ├── virtual-machine-libvirt/
│   └── network-libvirt/
├── environments/
│   ├── shared/
│   ├── kvm-01/
│   └── kvm-02/
├── docs/
└── README.md
```

Each environment is a root module with a focused state and a clear owner. Modules contain repeatable implementation patterns. Keep module inputs and outputs documented; they are the stable boundary that makes a future Proxmox implementation feasible.

## The Normal Change Loop

```bash
terraform fmt -recursive
terraform init
terraform validate
terraform plan -out=tfplan
terraform show tfplan
terraform apply tfplan
```

A saved plan is useful when the exact reviewed change must be applied soon afterward in an unchanged environment. Do not keep plans indefinitely: credentials, state, and real infrastructure can change.

## Non-Negotiable Habits

- Pin Terraform and provider versions; commit `.terraform.lock.hcl`.
- Keep state and secret variable files out of Git.
- Use a formatter and validation in every change.
- Review replacement and destroy actions line by line.
- Use least-privilege provider identities.
- Back up state and document recovery access.
- Prefer modules for proven repetition, not premature abstraction.
- Test a provider upgrade or a new platform module in a disposable environment.

## Migration-Friendly Decisions

Separate *intent* from *implementation*. Environment inputs can describe VMs, networks, roles, and desired capacity. Platform modules translate those descriptions to libvirt, Proxmox, or another provider. Do not hide capabilities that matter: document differences, migrate in stages, and expect resource identities to change across platforms.

## Exercises

1. Create a README for a small environment covering target platform, state location, prerequisites, and safe apply command.
2. Add the course's `.gitignore` rules to a Terraform practice project.
3. Review a plan and write a one-sentence reason for every replacement or destroy.
4. List the module contracts you would retain during a KVM-to-Proxmox move.

## Next Lesson

**Lesson 18 --- Virtualization Providers and Lab Concepts** applies the portable model to hypervisors without locking you into one.
