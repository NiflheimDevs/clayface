# Lesson 68 — Configuring Internal Services

## Learning objectives

- Model internal service dependencies and configuration inputs.
- Deploy a service using health checks and internal-only firewall policy.
- Separate service discovery from hard-coded addresses.

## Prerequisites

Lessons 63–67.

## Concept

Internal services—DNS, APIs, monitoring, queues, artifact repositories—need explicit consumers, ports, identities, and dependency order. Put stable service names in DNS where possible rather than embedding every VM IP in templates. Ansible can configure each component, but it is not a universal service-discovery system; DNS/registry design remains part of the infrastructure architecture.

Health checks should test the actual service behavior, not merely process existence. Use an appropriate URI, socket, command, or module and mark read-only checks `changed_when: false`.

## Mental model

An internal service is a contract: providers announce a reachable named endpoint; consumers are configured to use it; firewall policy permits exactly that contract.

## Example

``` yaml
- name: Verify internal HTTP health endpoint
  ansible.builtin.uri:
    url: "http://{{ internal_service_name }}:{{ internal_service_port }}/health"
    status_code: 200
    return_content: false
```

`uri` makes an HTTP request from the managed host executing the task. `status_code: 200` declares accepted response. `return_content: false` avoids collecting unnecessary body output. DNS must already resolve `internal_service_name` from that host.

## Practical exercise

Choose one internal service for the lab and write its provider, consumers, DNS name, port, protocol, health check, and firewall flows. Implement only a harmless health check if the service already exists.

## Expected result

The service contract is documented and testable from the actual consumer network.

## Common mistakes

- **Testing from the control node only.** A consumer VM may have different DNS/firewall reachability.
- **Hard-coding an address that Terraform will change.** Use DNS or generated variables deliberately.
- **Calling a TCP open port healthy.** Test a meaningful protocol response.

## Key takeaways

Internal automation succeeds when service names, dependencies, health checks, and allowed flows are explicit.

## Next lesson

Lesson 69 consolidates users and SSH policy across the enterprise lab.
