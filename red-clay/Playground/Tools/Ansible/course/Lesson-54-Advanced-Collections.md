# Lesson 54 — Advanced Collections

## Learning objectives

- Pin, document, and update collections safely.
- Identify collection-provided modules and plugins.
- Avoid collection namespace confusion.

## Prerequisites

Lesson 35.

## Concept

As the Digital Twin adds Windows, databases, virtualization, and vendor integrations, collections become first-class dependencies. Pin exact versions when reproducibility is most important; use controlled ranges only when you test updates. Record why each collection exists and remove unused dependencies. Collection plugins can affect inventory, connections, filters, and execution—not only modules.

## Mental model

Collections are application dependencies for your automation code. They need version review, not one-time installation.

## Example

``` yaml
collections:
  - name: ansible.windows
    version: "==2.5.0"
  - name: community.windows
    version: "==2.3.0"
```

Exact version constraints make a rebuild resolve the same release at the time of writing. Verify current compatibility with your installed `ansible-core`; collection documentation specifies supported versions. Windows collections do not make a Linux SSH transport manage Windows automatically—the connection setup remains separate.

## Practical exercise

List every collection in your requirements file with one sentence: capability, role/playbook consumer, and tested core version. Create a deliberate update process: install into a test environment, run syntax/check tests, then update the pin.

## Expected result

Your repository's external behavior is reviewable and reproducible enough for a lab rebuild.

## Common mistakes

- **Installing from Galaxy ad hoc on one workstation.** Another control node will not reproduce it.
- **Assuming collection version equals core compatibility.** Check both.
- **Using unqualified short module names.** FQCNs make dependency ownership visible.

## Key takeaways

Treat collections like versioned code dependencies, including compatibility and update testing.

## Next lesson

Lesson 55 provides a disciplined debugging workflow for Ansible failures.
