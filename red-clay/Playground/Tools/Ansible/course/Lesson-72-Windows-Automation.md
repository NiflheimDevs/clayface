# Lesson 72 — Windows Automation with Ansible

## Learning objectives

- Explain how Windows management differs from Linux Ansible management.
- Configure a Windows inventory group conceptually.
- Use Windows collection modules rather than Linux modules/commands.

## Prerequisites

Lessons 2, 35, 54, and 71; basic Windows/AD knowledge.

## Concept

Windows targets commonly use WinRM (HTTP/HTTPS) or SSH, and do not follow the Linux remote-Python module model. Install and pin `ansible.windows` and `community.windows` collections. Bootstrap WinRM securely—prefer HTTPS, certificates, limited firewall exposure, and an appropriate administrator/service account. Windows has its own package, service, feature, registry, and domain modules.

Do not apply Linux `package`, `service`, paths, or sudo concepts to Windows. The Ansible play/task/variable/handler concepts remain general; the connection and module implementation changes.

## Mental model

Ansible is still the control plane, but Windows is a different managed operating environment with a different remote-management door and toolkit.

## Example

``` ini
[windows]
client01 ansible_host=10.10.40.11

[windows:vars]
ansible_connection=winrm
ansible_winrm_transport=ntlm
ansible_winrm_server_cert_validation=validate
```

This inventory declares the connection type. Authentication, HTTPS listener configuration, port exposure, and certificate trust must be built and tested before a production-like use. `validate` is safer than ignoring server certificates; use the actual supported transport suited to your lab identity design.

## Practical exercise

Add a `windows` group without targeting it from Linux plays. Research and document your intended secure WinRM bootstrap method, required ports, certificate strategy, and collection versions. Only then create a disposable Windows target.

## Expected result

Your inventory and architecture keep Windows separate from Linux assumptions, while enabling future shared site orchestration.

## Common mistakes

- **Trying `ansible.builtin.ping` as the Windows proof.** Use `ansible.windows.win_ping` after Windows connection setup.
- **Disabling certificate validation to make WinRM work.** Fix trust/bootstrap design.
- **Using local Administrator credentials everywhere.** Use scoped accounts and protect them.

## Key takeaways

Ansible concepts transfer to Windows, but connection security and modules are Windows-specific.

## Next lesson

Lesson 73 composes every stage into one auditable lab deployment workflow.
