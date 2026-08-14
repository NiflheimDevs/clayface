# Lesson 24 — File Permissions and Ownership

## Learning objectives

- Apply owner, group, and mode intentionally.
- Explain numeric modes and quoting in YAML.
- Recognize when recursive permission changes are unsafe.

## Prerequisites

Lesson 23 and basic Unix permission knowledge.

## Concept

Unix access is determined by owner, group, and mode bits, subject to ACLs, MAC systems, and service-specific behavior where applicable. Ansible can enforce these through `file`, `copy`, and `template`. Describe permissions in the task that owns the file so a rebuild restores the security property with its content.

Quote numeric modes—`"0640"`—because YAML numeric parsing can otherwise create surprising values. Do not use `recurse: true` casually: it changes every descendant and may break executables, key files, or service-owned state.

## Mental model

Ownership says who is responsible; mode says who may read, write, or execute. A configuration is incomplete if its security envelope is accidental.

## Example

``` yaml
- name: Deploy private service configuration
  ansible.builtin.copy:
    src: files/service.conf
    dest: /etc/digital-twin/service.conf
    owner: root
    group: appsvc
    mode: "0640"
```

Mode `0640` gives the owner read/write, the group read, and others no access. The group must exist before this task. A service that runs as `appsvc` can read but not modify the configuration.

## Practical exercise

Choose a non-secret static configuration file. Declare a least-privilege owner/group/mode, apply it, and use `stat` over SSH to verify actual metadata. Explain why `0644` or `0600` would be more or less appropriate.

## Expected result

The file matches the declared ownership and mode, with no repeated changes.

## Common mistakes

- **Writing unquoted `0640`.** YAML interpretation can make the result unclear.
- **Using `0777` to cure a service failure.** Diagnose the service user and required access instead.
- **Recursively chowning application data.** Databases and package-managed files can be damaged.

## Key takeaways

Permissions are declared infrastructure security. Quote modes and limit changes to the resource you own.

## Next lesson

Lesson 25 makes services enabled, running, and safely restarted only when needed.
