# Lesson 15 --- Remote State and Backends

## Learning Objectives

- Explain what a backend does.
- Choose between local and remote state for a homelab stage.
- Understand locking, access control, and remote-state sharing.
- Migrate state carefully.

------------------------------------------------------------------------

## A Backend Stores State

The default backend writes state to the local working directory. A remote backend stores it in a shared service or storage system. The backend is configured in the root module:

```hcl
terraform {
  backend "s3" {
    bucket = "example-terraform-state"
    key    = "lab/kvm-01/terraform.tfstate"
    region = "example-region"
  }
}
```

The exact supported arguments depend on the backend. Do not copy this example blindly: choose a backend that provides appropriate durability, authentication, encryption, and locking for your environment.

## Why Remote State?

Remote state helps when more than one machine, person, or CI job can apply the same configuration. Locking prevents two applies from writing conflicting state at once. Central access control and backups reduce the risk of losing the controller laptop.

For a one-person learning exercise, local state can be reasonable. The moment an automated runner or second operator can apply, design remote state before concurrent use begins.

## State Sharing Is an API

The `terraform_remote_state` data source can read outputs from another state. Use it sparingly: it couples two deployments and grants the reader access to the complete source state, not only the selected output. For broadly shared data such as DNS, image IDs, or network prefixes, a dedicated inventory/IPAM/DNS source may be a safer long-term interface.

## Safe Migration

1. Make sure no one else is applying.
2. Back up the existing state securely.
3. Add the backend configuration.
4. Run `terraform init` and follow the migration prompt.
5. Run `terraform plan` and expect no infrastructure changes.
6. Verify state access and locking from the intended controller.

Never put backend credentials directly in version-controlled configuration. Use the backend's supported environment, credential, or identity mechanism.

## Lab Topology Guidance

Use different state keys or workspaces only where they represent genuinely separate lifecycles. A simple split can be shared network, per-hypervisor workloads, and disposable exercises. Backends are platform-neutral: your KVM-to-Proxmox migration changes providers and modules, not the need for guarded state.

## Exercises

1. Write down who and which machines are allowed to apply to each current state.
2. Choose a remote backend you can operate safely, and identify how it locks state.
3. Explain why reading one remote output can still expose more state than expected.
4. Rehearse a migration in a disposable test directory first.

## Next Lesson

**Lesson 16 --- Provisioners and Bootstrapping** separates infrastructure creation from guest configuration.
