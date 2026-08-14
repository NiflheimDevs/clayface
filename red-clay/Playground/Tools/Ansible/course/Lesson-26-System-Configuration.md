# Lesson 26 — System Configuration

## Learning objectives

- Manage a small OS setting safely and persistently.
- Distinguish runtime state from boot-time configuration.
- Validate platform-specific configuration paths.

## Prerequisites

Lessons 20–25.

## Concept

System configuration includes kernel parameters, timezone, hostname, resolver settings, limits, and logging. These settings are often both runtime state and persistent configuration. Do not assume one module controls every operating system: Alpine, Arch, Debian, and RHEL systems use different files and init integrations.

Prefer a dedicated module such as `ansible.builtin.hostname` or `community.general.timezone` when it correctly models the setting. For a configuration file, manage a narrowly scoped drop-in you own rather than editing a vendor file blindly. Validate a setting on a disposable machine before spreading it.

## Mental model

System configuration is the building code beneath applications. Runtime inspection tells you what exists now; persistence determines what survives reboot.

## Example

``` yaml
- name: Set a stable inventory-aligned hostname
  ansible.builtin.hostname:
    name: "{{ inventory_hostname }}"
```

The hostname module changes the system identity, not the inventory label. The use of `inventory_hostname` is appropriate only if that naming convention is intentional. DNS, `/etc/hosts`, certificates, and applications may need separate changes; a hostname task does not automatically reconcile them.

## Practical exercise

Choose one harmless, reversible system setting for your disposable VM, such as hostname or timezone. Identify both the runtime verification command and the persistent configuration mechanism on its OS before writing a task.

## Expected result

The setting remains correct after a reboot test when that is relevant, and a second run is `ok`.

## Common mistakes

- **Changing a hostname without considering DNS.** Name resolution is a separate network system.
- **Editing package-managed files.** Use a documented drop-in or template you own.
- **Applying kernel settings copied from another distribution.** Verify semantics and persistence path.

## Key takeaways

Manage system settings with platform-aware intent and test persistence, not merely current output.

## Next lesson

Lesson 27 handles environment variables without confusing shell setup with service configuration.
