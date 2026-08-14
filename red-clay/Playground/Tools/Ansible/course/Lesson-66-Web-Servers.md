# Lesson 66 — Configuring Web Servers

## Learning objectives

- Configure an nginx-style web tier declaratively.
- Template virtual-host data safely.
- Verify service and network reachability.

## Prerequisites

Lessons 25, 29, 30, and 65.

## Concept

A web role installs the server package, deploys a validated configuration, ensures the service is enabled/running, and exposes only required ports to intended networks. Nginx package/configuration paths vary across distributions. Keep application upstreams, DNS names, TLS settings, and listening ports as variables, not hard-coded lab hostnames.

TLS private keys and certificates require explicit secret/file permissions; do not show them in diff. For an internal red-team lab, decide whether TLS is realistic enough to model and document its certificate authority/lifecycle.

## Mental model

The web role turns a general Linux host into a controlled network entry point: content/configuration in, limited HTTP(S) exposure out.

## Example

``` yaml
- name: Deploy web virtual host
  ansible.builtin.template:
    src: site.conf.j2
    dest: /etc/nginx/conf.d/digital-twin.conf
    mode: "0644"
    validate: nginx -t -c %s
  notify: Reload nginx
```

Verify the exact `nginx -t` behavior/path on the installed distribution; some configurations need a full main config context. The template is changed only when rendered output differs, so the handler reloads only then.

## Practical exercise

Deploy a minimal non-sensitive HTTP page to `web01`. Define the listener port and server name through group/host variables. Test locally on the VM, then from a permitted client network, and only then add the matching firewall policy.

## Expected result

Nginx remains enabled/running, configuration validates, and HTTP is reachable only from intended networks.

## Common mistakes

- **Assuming config path is the same on Alpine and Arch.** Consult package layout.
- **Opening a firewall before confirming listener behavior.** Verify both host and network layers.
- **Reloading invalid config.** Use validation and preserve access.

## Key takeaways

Web automation combines package, template, service handler, and least-privilege network policy.

## Next lesson

Lesson 67 handles databases with extra care for persistent data and credentials.
