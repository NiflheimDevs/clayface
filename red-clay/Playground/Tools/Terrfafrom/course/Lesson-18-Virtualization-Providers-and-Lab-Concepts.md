# Lesson 18 --- Virtualization Providers and Lab Concepts

## Learning Objectives

- Map a generic VM design to common virtualization platforms.
- Understand the KVM/libvirt implementation in useful detail.
- Evaluate a provider before making it the foundation of a lab.
- Identify portability limits before changing platforms.

------------------------------------------------------------------------

## Start With the Common Model

Most virtualization platforms have equivalents for these concepts:

| Lab intent | Generic object | KVM/libvirt example | Other platforms |
| --- | --- | --- | --- |
| VM definition | compute instance | domain | Proxmox VM, vSphere VM, cloud instance |
| Guest disk | block/storage volume | volume in a pool | disk, datastore volume, cloud volume |
| Network segment | virtual network | libvirt network / bridge | Proxmox bridge/VLAN, port group, VPC subnet |
| Base OS | template/image | qcow2 base image | template, ISO, cloud image |
| Guest bootstrap | metadata/user data | cloud-init ISO or supported metadata | cloud-init/config drive/guest customization |

Terraform is portable at the language level, but no provider can erase all platform differences. Treat disk cloning, networking, guest agents, templates, snapshots, and live migration as capabilities to verify, not assumptions.

## KVM/libvirt Detail

With QEMU/KVM, libvirt provides a stable management API above QEMU. Virt-manager is also a libvirt client, so Terraform and virt-manager can see the same host—provided they use the same libvirt connection scope.

    Terraform provider ─┐
    virt-manager       ├─► libvirt daemon ─► QEMU/KVM domains
    virsh              ─┘

Common libvirt objects include storage pools, volumes, networks, domains, and cloud-init media. `qemu:///system` is typical for system-wide host management; a remote `qemu+ssh` URI can reach another host after you establish SSH and libvirt authorization correctly.

The exact resource fields and provider behavior are version-dependent. Pin the provider, use its current documentation, and test an actual plan against a disposable VM. Do not assume a snippet written for one release or fork works unchanged in another.

## Evaluating Any Provider

Before committing a real lab to a provider, answer:

1. Is it actively maintained and compatible with your Terraform version?
2. Does it support the VM, disk, network, clone, and cloud-init features you require?
3. Does it document imports, updates, replacements, and known limitations?
4. Can it connect with least-privilege credentials and secure authentication?
5. Can you test its create, change, destroy, and import lifecycle safely?
6. What happens if the provider cannot report a guest IP or waits for a guest agent?

This is particularly important for Proxmox: multiple community providers and forks may exist, with different resource schemas and maturity. Select based on current documentation and testing, rather than a provider name copied from an old tutorial.

## A Portable Environment Contract

Define an environment in terms of intent:

```hcl
machines = {
  dns01 = { vcpu = 1, memory_mib = 1024, image = "ubuntu-24.04", network = "lab" }
  web01 = { vcpu = 2, memory_mib = 2048, image = "ubuntu-24.04", network = "lab" }
}
```

Then let a chosen module translate the map to a platform. A migration is not a switch of one resource type: it normally creates new VM identities, moves data, updates DNS, and retires old VMs. But this contract prevents your desired topology from being trapped in vendor-specific syntax.

## Exercises

1. Inventory your current virt-manager lab as compute, storage, network, image, and bootstrap objects.
2. Test `virsh -c qemu:///system list --all` and compare it with what virt-manager displays.
3. Choose a possible Proxmox provider and evaluate it using the six questions above—without applying anything yet.
4. Write a provider-neutral machine map for three roles in your lab.

## Next Lesson

**Lesson 19 --- Building a Multi-Host Lab** designs deployment across several machines safely.
