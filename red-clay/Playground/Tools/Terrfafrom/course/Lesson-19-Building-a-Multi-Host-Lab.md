# Lesson 19 --- Building a Multi-Host Lab

## Learning Objectives

- Design a multi-host Terraform layout with controlled blast radius.
- Connect to remote hypervisors securely.
- Decide when one root module or multiple deployments are appropriate.
- Deploy independent work concurrently without sharing unsafe state.

------------------------------------------------------------------------

## The Goal: One Topology, Several Execution Targets

Your desired lab may place `dns01` on one host, `web01` on another, and shared services elsewhere. Terraform can coordinate independent work, but the architecture must make ownership and failure boundaries clear.

    controller
      ├── shared network / DNS state
      ├── host-a workload state ─► hypervisor A
      └── host-b workload state ─► hypervisor B

This is often safer than one enormous state. A host outage then does not prevent a plan for the other host, and an experiment on host B cannot accidentally target host A's workloads.

## Connectivity and Trust

For remote KVM/libvirt, prove the connection with `virsh` first, then configure the Terraform provider URI. For Proxmox or vSphere, use the platform API with a dedicated least-privilege service identity. In all cases:

- use trusted host certificates or SSH host keys;
- store credentials outside Git;
- restrict the identity to the resources it manages;
- document the controller machine and recovery procedure;
- monitor management network reachability separately from guest networking.

## Two Deployment Patterns

### Explicit aliases in one root module

Useful for a small number of fixed hosts and tightly coupled resources. Configure aliases such as `libvirt.host_a` and `libvirt.host_b`, then pass them into modules. Terraform's graph can create independent VMs on both in parallel.

### One root module/state per host or lifecycle

Useful as the fleet grows or hosts need independent release schedules. Each root module has one target connection and state. A higher-level CI job, Make target, or runbook coordinates the order. Share only stable facts through DNS, IPAM, an inventory service, or carefully controlled outputs.

For most homelabs, begin with the second pattern. It is easier to understand and recover. Move to more central orchestration only after the boundaries are working well.

## Placement Is Data

Keep placement explicit in an inventory-like structure:

```hcl
machines = {
  dns01 = { target = "host-a", vcpu = 1, memory_mib = 1024 }
  web01 = { target = "host-b", vcpu = 2, memory_mib = 2048 }
}
```

Terraform cannot dynamically create provider aliases from `target`. Either route this data into separate roots, or write explicit module calls for the known targets. This limitation is a design cue: provider connections are part of the deployment architecture, not normal data to loop over.

## Network Design Before VM Automation

Decide how guests reach each other across physical hosts. A libvirt NAT network created independently on each host is not automatically one shared L2 segment. You may need a bridge, VLAN, routed network, overlay, or physical network design. The same question exists on Proxmox and vSphere under different terms.

Terraform can declare network objects, but it cannot replace sound routing, switching, DNS, firewall, and IP address design.

## Exercises

1. Draw your current and intended physical hosts, management network, guest networks, and shared services.
2. Choose which state owns shared DNS and which states own per-host VMs.
3. Write the minimum permissions a Terraform identity needs on one target.
4. Explain why two identical NAT networks on separate KVM hosts may not let guests communicate.

## Next Lesson

**Lesson 20 --- Ansible Integration** connects infrastructure provisioning to repeatable guest configuration.
