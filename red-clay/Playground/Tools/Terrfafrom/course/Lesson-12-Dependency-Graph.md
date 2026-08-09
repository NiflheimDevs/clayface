# Lesson 12 --- Dependency Graph

## Learning Objectives

- Explain implicit and explicit dependencies.
- Predict why Terraform can run independent work in parallel.
- Avoid unnecessary `depends_on`.
- Apply dependency reasoning to networks, disks, VMs, and post-provisioning.

------------------------------------------------------------------------

## Terraform Builds a Graph, Not a Script

Terraform derives an execution graph from references between objects. If a VM uses a volume ID, Terraform knows the volume must exist first.

```hcl
resource "example_disk" "web" {
  name = "web01-disk"
}

resource "example_vm" "web" {
  name    = "web01"
  disk_id = example_disk.web.id
}
```

    disk ───► VM ───► output

This reference is an **implicit dependency**. Terraform can create an unrelated `dns` VM at the same time, subject to provider and API limits.

## Explicit Dependencies Are Rare but Valid

Use `depends_on` when a real dependency exists but no value reference can express it:

```hcl
resource "example_vm" "web" {
  name       = "web01"
  depends_on = [example_network.lab]
}
```

Do not use it as a ritual. Broad explicit dependencies reduce parallelism and can cause Terraform to treat more values as unknown during planning. Prefer referencing the specific ID or attribute the resource actually needs.

## Parallelism Across Hosts

Terraform's graph allows independent resources on KVM host 1 and host 2 to proceed concurrently. Provider aliases direct calls to separate targets; the graph provides the order. Terraform's default parallelism is limited, and you can tune it with `-parallelism=N` after observing API capacity and storage/network bottlenecks. Higher is not automatically better.

This pattern is platform-neutral: it applies equally to Proxmox nodes, vSphere clusters, or cloud regions.

## Debugging Graph Problems

When a plan reports a dependency cycle, look for resources that reference each other directly or indirectly. Break the cycle by moving a shared value earlier in the design, using a data source, or splitting the lifecycle into separate layers. Never add `depends_on` to solve a cycle; it only adds edges.

## Exercises

1. Draw dependencies for a network, base image, disk clone, VM, and DNS record.
2. Identify which two objects can be created in parallel.
3. Explain why referencing a disk ID is preferable to a blanket `depends_on` on every disk.
4. Run `terraform graph` in a practice configuration and inspect the generated graph text.

## Next Lesson

**Lesson 13 --- Lifecycle and Change Safety** teaches how Terraform handles consequential changes.
