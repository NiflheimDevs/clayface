# Lesson 5 --- Resources

## Learning Objectives

- Define a Terraform resource and resource address.
- Read a resource schema and distinguish arguments from attributes.
- Create a safe first resource in an isolated practice directory.
- Understand create, update, replace, and destroy actions.

------------------------------------------------------------------------

## A Resource Is One Managed Object

A resource block declares an object Terraform should manage. The general form is:

```hcl
resource "TYPE" "LOCAL_NAME" {
  argument = value
}
```

For example, this asks the libvirt provider for a storage volume:

```hcl
resource "libvirt_volume" "practice_disk" {
  name   = "tf-practice-disk.qcow2"
  pool   = "default"
  format = "qcow2"
}
```

Its Terraform address is `libvirt_volume.practice_disk`. The first label is provider-defined; the second is your local, stable name. Do not change the local name casually: Terraform will otherwise interpret it as removing one address and adding another.

## Arguments and Attributes

Arguments are values you set. Attributes are values the provider reports after creation. Documentation for each resource tells you which is which.

```hcl
resource "libvirt_volume" "practice_disk" {
  name = "tf-practice-disk.qcow2" # argument
}

output "volume_id" {
  value = libvirt_volume.practice_disk.id # computed attribute
}
```

Reference attributes rather than copying values. A later lesson explains how this also creates dependencies.

## The CRUD Lifecycle

Provider resource implementations generally support four operations:

| Action | Meaning |
| --- | --- |
| Create | Make an object that is declared but absent. |
| Read | Query the real object to refresh Terraform's view. |
| Update | Change an object in place when the provider supports it. |
| Delete | Remove an object Terraform manages when it is no longer declared. |

Some changes cannot be updated safely. Terraform will show `-/+` (replace): it destroys the old object and creates a new one. Treat a replacement of a VM disk or domain as a consequential change and inspect the plan closely.

## A Safe First Apply

Make a dedicated directory, not your existing virt-manager production-like lab. Add the volume resource, then run:

```bash
terraform init
terraform fmt
terraform validate
terraform plan
terraform apply
terraform state list
terraform destroy
```

Read the plan before each apply or destroy. `destroy` affects only resources recorded in the current state, but that can still be real infrastructure.

## Data Sources Are Read-Only Lookups

Resources manage objects; `data` blocks look up existing objects without claiming ownership. Use data sources for a pre-existing image, network, or DNS zone when supported by the provider.

```hcl
data "libvirt_network" "existing_lab" {
  name = "lab-net"
}
```

Do not declare an existing manually created object as a new resource just to reference it. Terraform may try to create a duplicate. Importing an existing object into management is a separate, intentional operation.

## Portability Note --- Resource Names Change, Model Stays

`libvirt_volume` and `libvirt_domain` are KVM-specific types. Proxmox, vSphere, and cloud providers use different resource names and fields, but the resource model stays the same: read the schema, use stable Terraform addresses, reference attributes instead of copying IDs, and inspect replacement actions.

## Common Mistakes

- Naming a resource after an IP address or date that will change.
- Assuming every change is in-place.
- Mixing resources Terraform owns with manually managed copies of the same object.
- Applying a plan without reading destructive or replacement actions.

## Exercises

1. Create and destroy the practice volume above. Confirm it appears and disappears in your libvirt storage pool.
2. What is the address of a resource declared as `resource "libvirt_domain" "web01"`?
3. In a plan, what does `-/+` mean? Why is it important for VM disks?
4. Find the libvirt provider documentation for `libvirt_volume` and list two configurable arguments and two computed attributes.

## Next Lesson

**Lesson 6 --- State** explains how Terraform remembers which real objects belong to those resource addresses.
