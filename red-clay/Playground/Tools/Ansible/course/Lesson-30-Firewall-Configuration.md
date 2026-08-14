# Lesson 30 — Firewall Configuration

## Learning objectives

- Design firewall rules from required traffic, not convenience.
- Identify Linux firewall implementation differences.
- Apply management-safe changes in narrow stages.

## Prerequisites

Lessons 20–29 and console access to any firewall test target.

## Concept

Firewall policy is networking and security configuration. Before automation, identify each service's listening address, protocol, port, source networks, and required return traffic. A rule such as “allow SSH” is incomplete without deciding from where. Test first on one host while preserving console or out-of-band access.

Ansible is general; firewall tooling is OS-specific. Arch typically uses nftables directly or via a chosen frontend. Alpine commonly uses nftables or iptables tooling. Debian/Ubuntu may use UFW; RHEL systems may use firewalld. Select one implementation per platform role and manage it with the matching module/collection, rather than mixing commands.

## Mental model

A firewall is a door policy: who may arrive, at which door, for which purpose. Ansible enforces the documented policy; it does not invent it.

## Example

Your web-server policy table should precede YAML:

| Flow | Source | Destination | Protocol/port | Reason |
| --- | --- | --- | --- | --- |
| Management SSH | management subnet | web01 | TCP/22 | Ansible administration |
| Web | client subnet | web01 | TCP/80,443 | application access |

For an nftables-based role, deploy a reviewed complete ruleset template, validate it with `nft -c -f %s` where supported, and apply it through the appropriate service. Do not paste a generic rule file: interface names, address families, and management subnets are lab-specific.

## Practical exercise

Write a traffic table for `web01` with only SSH and one future web service. Mark source subnet and direction. Do not apply firewall changes until you can explain how you will recover if SSH is denied.

## Expected result

You have a reviewable network policy that can become a template in a later role. It permits only explicitly justified flows.

## Common mistakes

- **Allowing SSH from `0.0.0.0/0` by habit.** Scope it to management where the lab design permits.
- **Mixing UFW, firewalld, iptables, and nftables.** Multiple owners create confusing final state.
- **Applying a full firewall before testing management access.** This can isolate every managed node.

## Key takeaways

Firewall automation starts with a traffic model. Distribution tooling differs, but least-privilege policy and staged validation do not.

## Next lesson

Lesson 31 reorganizes these growing playbooks into a maintainable repository.
