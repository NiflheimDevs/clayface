# Lesson 7 --- Playbooks

## Learning objectives

- Explain the structure and purpose of a playbook.
- Write and run a minimal playbook against a safe group.
- Recognize essential YAML syntax used by Ansible.

## Prerequisites

Lessons 1–6. You need a working `web01` inventory entry.

## Concept

A **playbook** is a YAML file that describes one or more plays. YAML is indentation-sensitive data notation: a colon separates a key from its value, and `-` begins an item in a list. Use spaces, never tabs, and keep related keys aligned. YAML describes data; Ansible gives that data execution meaning.

A play normally contains `name`, `hosts`, optional connection/privilege settings, and a `tasks` list. `hosts` is an inventory pattern, not a DNS hostname. Storing this in Git turns an operation into reviewable, repeatable infrastructure knowledge.

## Mental model

The inventory identifies recipients. A playbook is the reusable operations manual. A play is one chapter addressed to a specific recipient group.

## Example

Create `playbooks/verify.yml`:

``` yaml
---
- name: Verify Linux lab connectivity
  hosts: webservers
  gather_facts: false
  tasks:
    - name: Confirm Ansible can execute a module
      ansible.builtin.ping:
```

`---` is an optional YAML document start marker. The first `-` begins the list of plays. `hosts: webservers` uses the group from Lesson 4. `gather_facts: false` prevents the automatic fact-collection task for this small connectivity test. Under `tasks:`, indentation nests a list of tasks in the play. `ansible.builtin.ping` needs no arguments, so its value is blank.

Run it:

``` bash
ansible-playbook -i inventory/lab.ini playbooks/verify.yml
```

`ansible-playbook` executes a YAML playbook. `-i` selects your non-default inventory. The recap should show `ok=1`, `changed=0`, and `failed=0` for each reachable web server. This check makes no desired configuration change.

## Practical exercise

Create the verification playbook and run it only against your `webservers` group. Add a comment beginning with `#` above `gather_facts` explaining why it is false. Change the play name to describe your actual lab, then run it again.

## Expected result

Ansible displays the play and task names, returns `pong`, and the recap shows no changed tasks. Re-running produces the same result.

## Common mistakes

- **Using tabs.** YAML parsers reject them; configure your editor for spaces.
- **Misaligning `tasks`.** It must be a play key, while each task list item is nested beneath it.
- **Calling the file an inventory.** `-i` takes inventory; the final command argument is the playbook.

## Key takeaways

Playbooks are declarative, version-controlled YAML descriptions. Indentation is structure, and `hosts` limits the intended targets.

## Next lesson

Lesson 8 examines plays and tasks more closely, including their execution order and result reporting.
