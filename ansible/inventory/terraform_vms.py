#!/usr/bin/env python3
"""Dynamic Ansible inventory fed by Terraform outputs.

Reads the `vms` output from the terraform/ directory and builds inventory
groups so playbooks can target the VMs created by Terraform:

    terraform_vms : every VM
    vms_gateway   : edge VMs, role "gateway" (started first, they provide DHCP/DNS)
    vms_linux     : every other VM

Each VM host exposes:
    libvirt_hypervisor : hypervisor alias (matches a host in lab_inventory.py)
    ansible_host       : <name>.<domain> (resolved via OPNsense DNS)

The VM placement itself comes from lab.yaml (via terraform locals); this
script only reads back what terraform derived from it.

The DNS domain is NOT read from terraform/lab.yaml — it is configured in
the OPNsense image, so it is a mirrored constant here. If you change it,
change the OPNsense DNS config too.

Usage together with the lab.yaml host inventory:

    ansible-playbook \
        -i ansible/inventory/lab_inventory.py \
        -i ansible/inventory/terraform_vms.py \
        ansible/playbooks/start_vms.yml

Environment overrides:
    LAB_DOMAIN  DNS domain of the lab (default: clayface)
"""

import json
import os
import subprocess
import sys

LAB_DOMAIN = os.environ.get("LAB_DOMAIN", "clayface")

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
TERRAFORM_DIR = os.path.join(BASE_DIR, "..", "..", "terraform")


def terraform_output(name):
    """Return a terraform output value, or None on any failure."""
    cmd = ["terraform", "-chdir=" + TERRAFORM_DIR, "output", "-json", name]
    try:
        result = subprocess.run(
            cmd, capture_output=True, text=True, check=True)
        data = json.loads(result.stdout)
        # `terraform output -json <name>` returns the value directly; the
        # {"value": ..., "type": ...} wrapper only appears when dumping
        # all outputs. Handle both shapes.
        if isinstance(data, dict) and "value" in data and "type" in data:
            return data.get("value")
        return data
    except FileNotFoundError:
        print("WARNING: terraform binary not found", file=sys.stderr)
    except subprocess.CalledProcessError as exc:
        msg = exc.stderr.strip() if exc.stderr else exc
        print(f"WARNING: terraform output failed: {msg}", file=sys.stderr)
    except json.JSONDecodeError as exc:
        print(f"WARNING: could not parse terraform output: {exc}", file=sys.stderr)
    return None


def read_terraform_vms():
    """Return the terraform `vms` output: {name: {hypervisor, role}}."""
    vms = terraform_output("vms")
    return vms or {}


def build_inventory():
    vms = read_terraform_vms()

    inventory = {
        "_meta": {"hostvars": {}},
        "all": {"children": ["terraform_vms"]},
        "terraform_vms": {"children": ["vms_gateway", "vms_linux"], "hosts": []},
        "vms_gateway": {"hosts": []},
        "vms_linux": {"hosts": []},
    }

    for name, attrs in sorted(vms.items()):
        hypervisor = (attrs or {}).get("hypervisor", "")
        role = (attrs or {}).get("role", "linux")
        if not hypervisor:
            print(f"WARNING: VM '{name}' has no hypervisor attribute, skipping",
                  file=sys.stderr)
            continue

        inventory["terraform_vms"]["hosts"].append(name)
        group = "vms_gateway" if role == "gateway" else "vms_linux"
        inventory[group]["hosts"].append(name)

        inventory["_meta"]["hostvars"][name] = {
            "libvirt_hypervisor": hypervisor,
            "ansible_host": f"{name}.{LAB_DOMAIN}",
        }

    return inventory


def main():
    inventory = build_inventory()

    if len(sys.argv) > 1 and sys.argv[1] == "--host":
        host = sys.argv[2] if len(sys.argv) > 2 else ""
        print(json.dumps(inventory["_meta"]["hostvars"].get(host, {})))
        return 0

    print(json.dumps(inventory))
    return 0


if __name__ == "__main__":
    sys.exit(main())
