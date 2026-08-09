# Lesson 11 --- Functions and the Console

## Learning Objectives

- Use Terraform's built-in functions to transform configuration data.
- Test expressions interactively with `terraform console`.
- Recognize when a function hides an input-design problem.

------------------------------------------------------------------------

## Functions Are Small, Predictable Transformations

Terraform includes functions for strings, collections, numbers, files, encoding, and IP calculations. A few particularly useful ones:

```hcl
lower("WEB01")                         # "web01"
format("%s-%02d", var.role, 1)         # "web-01"
merge(local.common_tags, { role = "dns" })
lookup(var.images, var.os, null)
cidrsubnet("10.10.0.0/16", 8, 12)      # "10.10.12.0/24"
```

Consult the Terraform language documentation for the exact behavior and edge cases of a function before using it in an address plan or a destructive condition.

## The Terraform Console

In a configured working directory, run:

```bash
terraform console
```

Then test expressions without changing infrastructure:

```text
> cidrsubnet("10.10.0.0/16", 8, 12)
"10.10.12.0/24"
> keys({ dns01 = {}, web01 = {} })
[
  "dns01",
  "web01",
]
```

This is an excellent way to verify data transformations before embedding them in a resource.

## File and Template Functions: Treat Inputs as Data

`file()` reads a file; `templatefile()` renders a template with supplied variables. They are useful for cloud-init or configuration payloads, but avoid embedding credentials. A missing file or changed whitespace can alter a plan, so keep templates version-controlled and review their rendered content.

## Avoid Cleverness

Functions should make simple data clearer. Do not turn Terraform into a general-purpose programming language. Complex validation, discovery, or guest configuration generally belongs in a purpose-built tool such as a script, inventory system, Packer, or Ansible.

## Exercises

1. Use the console to calculate three `/24` subnets from a private `/16`.
2. Create a local name with `format` rather than manual string concatenation.
3. Merge base metadata with role-specific metadata.
4. Explain why `templatefile()` content may end up in Terraform state or plans depending on where it is passed.

## Next Lesson

**Lesson 12 --- Dependency Graph** explains how Terraform orders and parallelizes work.
