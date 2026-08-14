# Lesson 65 — Creating Reusable Server Roles

## Learning objectives

- Define a service-role contract.
- Structure package, config, service, and verification tasks.
- Support configuration through variables rather than forks.

## Prerequisites

Lessons 32–36 and 64.

## Concept

Most server roles follow a dependable sequence: validate inputs; install packages; create service account/directories; render configuration; validate it; notify service reload; verify health. This pattern is reusable, while package names and configuration syntax differ. Give each role defaults for safe common behavior and named variables for topology-specific values.

## Mental model

A service role is a small product: inputs, installation, configuration, controlled activation, and acceptance test.

## Example

``` text
roles/web_nginx/
├── defaults/main.yml
├── handlers/main.yml
├── tasks/main.yml
├── templates/site.conf.j2
└── README.md
```

Typical `tasks/main.yml` order is assert → package → directory → template (`validate` if supported) → service. The template notifies a reload handler; the service task ensures enabled/running independently.

## Practical exercise

Create a blank role skeleton for either `web_nginx` or `database_postgresql`. Write its inputs, supported platforms, opened ports, stateful-data risks, and verification command before filling tasks.

## Expected result

The role's purpose and safety boundaries are clear before it changes any host.

## Common mistakes

- **Combining database and web setup in one role.** Different lifecycle and data risks deserve separation.
- **No validation/verification task.** Deployment success is not service health.
- **Forcing every variation via opaque extra vars.** Document normal defaults and group vars.

## Key takeaways

Reusable server roles standardize the service lifecycle while exposing a small documented interface.

## Next lesson

Lesson 66 applies that pattern to the web tier.
