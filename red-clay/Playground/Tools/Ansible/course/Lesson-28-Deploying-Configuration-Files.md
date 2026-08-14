# Lesson 28 — Deploying Configuration Files

## Learning objectives

- Deploy an owned static configuration file safely.
- Use `validate` before replacing syntax-sensitive files.
- Notify a service handler only after accepted content changes.

## Prerequisites

Lessons 15 and 23–25.

## Concept

Configuration files are code for services. Ansible should own the source in Git, deploy exact content, set secure metadata, validate syntax where the service supports it, and reload only when content changes. `copy` is correct when every target gets identical content; `template` is introduced next for values that differ per host.

`validate` runs a command against a temporary file before the final replacement. It is not run through a shell, so pipes and redirection do not work. `%s` is substituted with the temporary candidate file path.

## Mental model

Deploying config is publishing a signed draft: check it before replacing the live edition, then tell the service about the new edition once.

## Example

``` yaml
- name: Deploy SSH daemon policy
  ansible.builtin.copy:
    src: files/sshd_config
    dest: /etc/ssh/sshd_config
    owner: root
    group: root
    mode: "0600"
    validate: /usr/sbin/sshd -t -f %s
  notify: Reload SSH daemon
```

The validation path may differ on Alpine and Arch; discover it on the actual target. Validation protects syntax, not policy correctness or safe login behavior. The handler must use the target's actual service manager and should be tested with a console fallback available.

## Practical exercise

Deploy a noncritical static config file first. If its service has a documented syntax-check command, add `validate`; otherwise explain why no validation is possible. Use a debug handler until you are ready to control the real service.

## Expected result

Changed content is installed and triggers one notification. Invalid candidate content fails before it replaces the live file.

## Common mistakes

- **Assuming `validate` invokes a shell.** It does not.
- **Putting mutable per-host addresses into a copied static file.** Use a template next lesson.
- **Reloading without syntax validation.** A bad config can take down management access or a service.

## Key takeaways

Configuration deployment needs source control, correct metadata, validation where possible, and change-triggered reloads.

## Next lesson

Lesson 29 introduces Jinja2 templates for carefully variable configuration.
