# Lesson 8 --- Outputs

## Learning Objectives

- Publish useful values after an apply.
- Mark confidential values as sensitive.
- Use outputs as a contract between Terraform and people or automation.
- Avoid using outputs as a substitute for state ownership.

------------------------------------------------------------------------

## Outputs Answer: “What Did We Build?”

Outputs expose selected values from a Terraform root module after an apply.

```hcl
output "web_server_name" {
  description = "Name assigned to the web server"
  value       = var.vm_name
}

output "web_server_ip" {
  description = "Guest address, when the chosen provider can report it"
  value       = libvirt_domain.web01.network_interface[0].addresses
}
```

Use `terraform output` to view values and `terraform output -json` for another tool to consume them. An output does not create anything; it makes an existing value intentionally visible.

## Outputs Form an Interface

A well-designed environment publishes stable, useful facts: IP addresses, FQDNs, VM IDs, connection endpoints, or inventory data. Downstream tools can consume those facts without needing to know every internal resource name.

    Terraform provisions VM
              |
              v
       output: address / ID
              |
              v
      Ansible, monitoring, or an operator

Do not expose every provider attribute. Outputs are an interface, so keep them small and meaningful.

## Sensitive Is Helpful, Not Encryption

```hcl
output "bootstrap_token" {
  value     = var.bootstrap_token
  sensitive = true
}
```

Terraform redacts a sensitive value in routine CLI output, but it may still exist in state. A sensitive output is not a secret-management system. Prefer to output a secret reference, endpoint, or identifier when consumers can retrieve the secret themselves.

## Platform Note

The exact attributes available vary. A cloud VM may report a public IP; a Proxmox or libvirt provider may not reliably know a DHCP address until the guest agent, leases, or a separate IPAM system provides it. Design automation around a dependable source of truth rather than assuming every provider can discover every guest address.

## Exercises

1. Add outputs for a VM name and the ID of a resource you manage.
2. Run `terraform output` and `terraform output -json` after an apply.
3. Decide which two facts an Ansible inventory generator would need from your lab.
4. Mark a test output sensitive and observe the CLI behavior. Then explain why state still needs protection.

## Next Lesson

**Lesson 9 --- Locals** shows how to name and reuse derived values inside a configuration.
