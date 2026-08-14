# Lesson 70 — Network and Service Configuration

## Learning objectives

- Separate VM network topology from in-guest service policy.
- Encode service dependencies and firewall flows.
- Verify traffic from the correct network perspective.

## Prerequisites

Lessons 26, 30, and 63–69.

## Concept

Terraform/libvirt owns virtual networks, NIC attachment, and usually initial addressing. Ansible owns in-guest resolver/client settings, service listening addresses, firewall policy, and configuration that consumes topology. Avoid having Ansible overwrite a Terraform/cloud-init-owned interface configuration unless you deliberately move ownership.

For every service, document source, destination, protocol/port, DNS name, authentication, and health check. Test from a source host in the actual VLAN/subnet because the control node may bypass the intended path.

## Mental model

Terraform draws and connects streets; Ansible configures which doors open, which names residents use, and which services answer inside buildings.

## Example

``` yaml
web_nginx_listen_address: 10.10.20.10
web_nginx_listen_port: 443
web_nginx_allowed_sources:
  - 10.10.30.0/24
```

These variables drive a validated server template and firewall implementation. They are not a substitute for creating `10.10.20.0/24` in Terraform. Confirm the listen address exists on the guest before reloading a service; binding to a nonexistent address causes service failure.

## Practical exercise

For web-to-database traffic, make a five-column flow table: source group, destination group, DNS name/address, TCP port, and verification command from the web host. Identify the Terraform-owned and Ansible-owned components of the path.

## Expected result

You can prove intended traffic works and unintended traffic is denied using tests from appropriate VMs.

## Common mistakes

- **Binding a service to a Terraform address that changed.** Generate/consume current topology data.
- **Testing only from the control node.** It may have privileged routing.
- **Configuring DNS after dependent services.** Sequence dependencies explicitly.

## Key takeaways

Network topology and in-guest service policy are complementary layers. Verify flows from actual consumers.

## Next lesson

Lesson 71 prepares Linux systems and lab design for Active Directory integration.
