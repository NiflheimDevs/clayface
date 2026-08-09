# Lesson 6 --- Terraform State

## Learning Objectives

- Explain why Terraform needs state.
- Inspect state safely and recognize sensitive information risk.
- Understand refresh, drift, import, and state locking.
- Choose a safe state layout for a personal multi-host lab.

------------------------------------------------------------------------

## State Is Terraform's Inventory and Memory

Terraform state maps configuration addresses to real infrastructure identities and their last known attributes.

    libvirt_domain.web01  ---> domain UUID on kvm-01
    libvirt_volume.web01  ---> volume path / ID on kvm-01

Without state, Terraform cannot reliably know whether `web01` already exists, which disk belongs to it, or how to delete it later. The default local state file is `terraform.tfstate`.

State is not your desired configuration, and it is not a backup of every VM disk. It is Terraform's operational record.

## Inspect, Do Not Edit

Useful read-only commands are:

```bash
terraform state list
terraform state show libvirt_volume.practice_disk
terraform show
```

Do not hand-edit `terraform.tfstate`. If you must make a state-only change, use `terraform state mv`, `terraform state rm`, or `terraform import`, after taking a copy and understanding the result. State commands can sever Terraform's connection to real objects without deleting them.

## Refresh and Drift

Drift means reality changed outside Terraform: for example, an administrator alters a VM's vCPU count in virt-manager.

At planning time Terraform normally asks providers to refresh managed objects, then compares current reality, recorded state, and configuration. The plan may propose restoring the declared setting, accepting an update, or replacing an object depending on the provider schema.

Use `terraform plan -refresh-only` to inspect and record discovered changes without applying configuration changes. Decide deliberately whether to:

1. change the Terraform code to represent the intended new reality, or
2. apply the code to restore the declared design.

## Sensitive Data

State can contain passwords, private addresses, cloud-init content, connection details, and provider-returned secrets even when an output is marked `sensitive`. Treat state as confidential.

- Never commit `terraform.tfstate` or its backups to Git.
- Limit read/write access to the state location.
- Avoid putting secrets in user-data or resource arguments where possible.
- Use a secret manager or a protected deployment mechanism as your lab matures.

Add this to `.gitignore` in Terraform projects:

```gitignore
.terraform/
*.tfstate
*.tfstate.*
crash.log
*.tfvars
!example.tfvars
```

Keep `.terraform.lock.hcl` committed; it is not state and is useful for reproducibility.

## State Layout for Your Lab

Avoid one enormous state for every experiment and every host. Split by lifecycle and blast radius, for example:

    terraform/
    ├── shared-network/      # networks, base images, shared storage
    ├── kvm-01-workloads/    # VMs owned by host 1
    └── kvm-02-workloads/    # VMs owned by host 2

Separate state means a mistake in a disposable VM exercise cannot plan changes to your shared network. It also permits independent work. When multiple people or automation runs share a state, use a remote backend with locking; that is covered in Lesson 15.

## Import Is Adoption, Not Discovery

`terraform import` associates an existing real object with a Terraform resource address. It does not generate a complete configuration for you. First write the intended resource block, then import, inspect the plan, and adjust code until it is stable.

Before importing a VM that virt-manager currently manages, make a backup and understand how the provider maps its disks, network interfaces, and XML. Import is best treated as a migration project, not a first exercise.

## Portability Note --- State Is Provider-Neutral

State works the same for libvirt, Proxmox, VMware, DNS, and cloud resources. The stored IDs differ, but the operational rules do not: secure it, lock it when shared, split it by lifecycle, and use imports deliberately.

## Exercises

1. Run `terraform state list` after creating a practice resource. What address is shown?
2. Change a harmless setting outside Terraform, run `terraform plan -refresh-only`, and explain the observed difference.
3. Create a `.gitignore` using the pattern above. Which Terraform file should still be committed?
4. Sketch a state split for shared networks, KVM host 1, and KVM host 2.

## Next Lesson

**Lesson 7 --- Input Variables** lets one configuration describe many consistent machines without copying files.
