# Lesson 20 --- Ansible Integration

## Learning Objectives

- Divide responsibilities between Terraform and Ansible.
- Pass reliable inventory information between the tools.
- Build an idempotent infrastructure-to-configuration workflow.
- Avoid secret and timing pitfalls.

------------------------------------------------------------------------

## Complementary Tools

Terraform manages the infrastructure boundary: virtual networks, instances, disks, load balancers, and provider-level metadata. Ansible manages the operating system and applications inside reachable guests: packages, users, files, services, and configuration.

    Terraform apply
       │ creates VM and exposes stable facts
       ▼
    inventory / DNS / IPAM
       │ provides reachable hostnames or addresses
       ▼
    Ansible playbook
       │ configures guest role
       ▼
    validated service

This boundary remains useful regardless of whether the VMs are on libvirt, Proxmox, vSphere, or a cloud.

## Choose a Reliable Inventory Source

Good options include:

- DNS records created as part of the infrastructure workflow;
- an IPAM or inventory system;
- a generated Ansible inventory from Terraform outputs;
- an Ansible dynamic inventory plugin for the target platform, if it is maintained and accurate.

Avoid assuming Terraform can always discover DHCP addresses. An address assigned through predictable addressing, DNS, guest-agent reporting, or IPAM is more dependable than scraping an incidental provider field.

## A Simple Handoff

Terraform can output a structured map:

```hcl
output "ansible_hosts" {
  value = {
    web = { hostname = "web01.lab.example", groups = ["web"] }
    dns = { hostname = "dns01.lab.example", groups = ["dns"] }
  }
}
```

An inventory-generation step can consume `terraform output -json`. Keep that bridge small, version-controlled, and free of credentials. Another option is to let both tools use DNS as their interface, which is often simpler.

## Readiness Is Not Creation

A successful Terraform apply means the provider completed its requested operations. It does not guarantee that SSH is accepting connections, cloud-init has finished, or your service is healthy. Let the configuration stage explicitly wait/retry in an idempotent, bounded way, then validate the service with a health check.

## Secrets

Use SSH keys, Ansible Vault, or a secret manager appropriate to your environment. Do not export secrets through Terraform outputs solely to make Ansible convenient. Prefer each tool retrieving its own authorized secrets.

## Exercises

1. Assign Terraform and Ansible ownership for a DNS server, from VM disk through BIND configuration.
2. Design a host naming/DNS convention Ansible can use across hypervisors.
3. Produce a fake `ansible_hosts` output map and sketch the inventory it would generate.
4. Identify a guest readiness check and an application health check; explain why they differ.

## Next Lesson

**Lesson 21 --- Capstone Project** combines the course into a portable, multi-host lab design.
