# Lesson 16 --- Provisioners and Bootstrapping

## Learning Objectives

- Explain why provisioners are a last resort.
- Use image building, cloud-init, and configuration management in the right roles.
- Bootstrap a VM without embedding fragile SSH scripts in Terraform.

------------------------------------------------------------------------

## Terraform Creates Infrastructure; It Is Not Your Main Config Manager

Terraform is strongest at declaring infrastructure objects and their relationships. Provisioners run commands or copy files during resource creation or destruction. They can be useful, but Terraform cannot reliably track the changes a script makes inside a guest.

Avoid designing a deployment around long `remote-exec` or `local-exec` scripts. Network timing, SSH availability, retries, partial failure, and idempotency turn a simple apply into a fragile orchestration system.

## Prefer This Layering

    Image builder (optional)  -> reusable OS image
    Terraform                -> VM, network, disk, metadata
    cloud-init / first boot  -> minimal bootstrap
    Ansible or similar       -> repeatable guest configuration

This works on KVM/libvirt, Proxmox, vSphere, and many clouds because the boundaries are conceptual, not provider-specific.

## Cloud-Init as Bootstrap

Where supported by your image and platform, cloud-init can set an initial user, SSH key, hostname, and small first-boot setup. Keep it minimal and avoid long-lived passwords. Terraform can pass user-data, metadata, or a rendered template through the provider-specific mechanism.

Cloud-init availability and syntax differ across platforms. Verify that the guest image includes it and that your hypervisor exposes the correct metadata channel before relying on it.

## If You Must Use a Provisioner

Use it for a narrowly scoped action that cannot be modeled otherwise, document why, and make the action safe to retry. Do not use a provisioner to manage a guest's whole lifecycle. Never pass secrets on the command line or leave them in scripts that enter state.

## Exercises

1. Draw the division of responsibility between Terraform, cloud-init, and Ansible for one web VM.
2. Create a minimal cloud-init file that only creates an SSH user and sets a hostname.
3. List three reasons a `remote-exec` script can fail even after a VM is created.
4. Identify one task in your lab that belongs in Terraform and one that belongs in configuration management.

## Next Lesson

**Lesson 17 --- Project Design and Best Practices** brings the language lessons into a maintainable workflow.
