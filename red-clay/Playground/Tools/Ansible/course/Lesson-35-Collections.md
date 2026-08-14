# Lesson 35 — Collections

## Learning objectives

- Define an Ansible collection.
- Install collection dependencies from a requirements file.
- Use FQCNs to make module origins clear.

## Prerequisites

Lessons 1–34 and `ansible-galaxy` from `ansible-core`.

## Concept

Collections package modules, plugins, roles, and documentation under a namespace such as `community.general` or `ansible.posix`. They provide functionality that is not part of `ansible-core`. Pinning a collection version records a dependency contract much like provider-version constraints in Terraform.

Collection content can be powerful and runs in your automation context. Review publisher, documentation, version, license, and update changes. Prefer fully qualified collection names, for example `ansible.posix.authorized_key`, so a reader knows where behavior comes from.

## Mental model

Collections are versioned toolboxes. A requirements file is the inventory of approved toolboxes for this repository.

## Example

Create `collections/requirements.yml`:

``` yaml
collections:
  - name: ansible.posix
    version: ">=1.5.0,<2.0.0"
```

Install to a project-local path:

``` bash
ansible-galaxy collection install -r collections/requirements.yml -p collections
```

`-r` reads the requirements file. `-p collections` chooses a local installation path; configure `collections_paths` in `ansible.cfg` if needed so normal runs find it. A bounded version range permits compatible updates; an exact version offers maximum repeatability.

## Practical exercise

Add `ansible.posix` requirement, install it, and use `ansible-doc ansible.posix.authorized_key` to confirm discovery. Record the installed version. Do not install random collections solely to make a warning disappear.

## Expected result

Documentation resolves for the FQCN and the repository declares why the dependency exists.

## Common mistakes

- **Using a short module name after adding a collection.** Origins become ambiguous.
- **Leaving collection versions unconstrained in a lab meant to rebuild.** Future downloads may behave differently.
- **Committing a private key with collection files.** Dependencies and credentials are separate concerns.

## Key takeaways

Collections extend Ansible and should be explicit, reviewed, and version-controlled like other automation dependencies.

## Next lesson

Lesson 36 turns a role into a reusable, documented interface.
