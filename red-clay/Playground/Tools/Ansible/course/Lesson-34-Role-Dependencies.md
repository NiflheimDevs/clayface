# Lesson 34 — Role Dependencies

## Learning objectives

- Define a role dependency and when to avoid one.
- Declare a dependency in role metadata.
- Prevent hidden ordering assumptions.

## Prerequisites

Lessons 31–33.

## Concept

A role dependency states that another role must run before this role. Dependencies are declared in `meta/main.yml` and are useful when a role cannot function without a small foundational capability. They are not a cure for a role that has grown too broad: explicit composition in a play is often clearer, especially when inventory groups receive different baselines.

Dependencies are static role relationships. They do not automatically solve service readiness, network reachability, or cross-host orchestration; those concerns need explicit tasks and checks.

## Mental model

A dependency is a prerequisite noted in the appliance manual, not an invisible administrator who guarantees every external condition.

## Example

`roles/web_nginx/meta/main.yml`:

``` yaml
dependencies:
  - role: base_linux
```

This YAML list says Ansible should apply `base_linux` before `web_nginx` whenever the latter is used. Only declare it if every possible use of `web_nginx` needs the exact baseline. Alternatively, the site play can make ordering explicit:

``` yaml
roles:
  - base_linux
  - web_nginx
```

## Practical exercise

List the true prerequisites for your future web-server role. Mark each as a role dependency, a play-level composition choice, or an external prerequisite (such as DNS). Do not add metadata until the categories are clear.

## Expected result

You can explain why a dependency exists and avoid duplicate role execution or circular design.

## Common mistakes

- **Creating dependency chains for all lab roles.** It hides the architecture and makes reuse difficult.
- **Expecting dependencies to wait for a remote database.** They only order role application on a host.
- **Duplicating the same base role in many layers.** Review the final role graph.

## Key takeaways

Use dependencies for universal, intrinsic prerequisites; prefer visible play composition for architecture choices.

## Next lesson

Lesson 35 installs and pins external collections.
