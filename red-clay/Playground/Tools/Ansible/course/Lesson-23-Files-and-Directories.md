# Lesson 23 — Files and Directories

## Learning objectives

- Create directories and files with state-aware modules.
- Choose `file`, `copy`, or `template` based on content ownership.
- Avoid destructive replacement of unknown data.

## Prerequisites

Lessons 20–22.

## Concept

`ansible.builtin.file` manages paths and metadata: it can ensure a directory exists, a symlink points somewhere, or an unwanted path is absent. It does not create arbitrary file content with `state: touch` in a content-aware way. Use `copy` for static content stored in the repository and `template` when content must be rendered from variables (Lesson 29).

Idempotency includes preserving surrounding state: `state: directory` creates missing parents only where appropriate, while `state: absent` deletes the exact target. Never use broad recursive deletion in automation without a tightly controlled target and explicit authority.

## Mental model

The file module builds the shelves; copy puts a fixed document on one shelf; template prints a customized document from a form.

## Example

``` yaml
- name: Ensure application configuration directory exists
  ansible.builtin.file:
    path: /etc/digital-twin
    state: directory
    owner: root
    group: root
    mode: "0755"

- name: Install static ownership notice
  ansible.builtin.copy:
    src: files/ownership-notice.txt
    dest: /etc/digital-twin/ownership-notice.txt
    owner: root
    group: root
    mode: "0644"
```

`path` is the remote target. `src` is relative to the playbook directory for a simple playbook; roles later give it a role-relative meaning. `copy` creates/replaces content only if its checksum differs.

## Practical exercise

Create `/etc/digital-twin` on a disposable target and deploy a static text file you wrote. Check it after the run with SSH and run the play twice.

## Expected result

The directory and file exist with the requested attributes. Only the first relevant run reports changes.

## Common mistakes

- **Using `touch` to manage content.** It updates timestamps and defeats a clean convergence signal.
- **Assuming `src` is remote.** `copy.src` is normally on the control node.
- **Making application files world-writable.** Choose modes based on the consuming service's needs.

## Key takeaways

Use the module matching the resource: path metadata, static content, or rendered content. Manage exact targets conservatively.

## Next lesson

Lesson 24 examines ownership and permissions as security controls.
