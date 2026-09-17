#!/usr/bin/env python3
"""Dynamic Ansible inventory fed by Terraform outputs.

Reads the `vms` output from the terraform/ directory and builds inventory
groups so playbooks can target the VMs created by Terraform:

    terraform_vms : every VM
    vms_gateway   : edge VMs, role "gateway" (started first, they provide DHCP/DNS)
    vms_windows   : Windows guests — os_family "windows" (WinRM connection vars)
    vms_linux     : Linux guests — os_family "linux" (SSH)
    vms_dc        : Windows domain controllers, role "dc"
    vms_client    : Windows workstations, role "client"
    vms_pinned    : VMs with both a pinned MAC and a static IP, i.e. the ones
                    playbooks/opnsense.yml manages as DHCP reservations

Every VM is in exactly one of vms_gateway / vms_windows / vms_linux, then in
zero or more role groups. "Which OS is it" and "which playbook owns it" are
separate questions: dc.yml must never see a workstation, and a future member
server would be a third role that is still os_family "windows".

Each VM host exposes:
    libvirt_hypervisor : hypervisor alias (matches a host in lab_inventory.py)
    ansible_host       : <name>.<domain> (resolved via OPNsense DNS)
    libvirt_mac        : pinned NIC MAC (when the module derives one)
    libvirt_ip         : static address from lab.yaml `ip:` (when set)

Windows hosts additionally expose WinRM connection variables:
    ansible_connection/port/user/password

The connection is chosen by os_family, which terraform derives from the VM's
module (see terraform/locals.tf `module_os`).

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
    LAB_DOMAIN           DNS domain of the lab (default: clayface)
    LAB_WIN_ADMIN_PASS   Windows Administrator password for os_family
                         "windows" VMs (default: Admin@123 — lab-only,
                         deliberately weak)
"""

import json
import os
import subprocess
import sys

LAB_DOMAIN = os.environ.get("LAB_DOMAIN", "clayface")
LAB_WIN_ADMIN_PASS = os.environ.get("LAB_WIN_ADMIN_PASS", "Admin@123")

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
        print(f"WARNING: could not parse terraform output: {exc}",
              file=sys.stderr)
    return None


def read_terraform_vms():
    """Return the terraform `vms` output: {name: attrs}."""
    vms = terraform_output("vms")
    return vms or {}


def build_inventory():
    vms = read_terraform_vms()

    inventory = {
        "_meta": {"hostvars": {}},
        "all": {"children": ["terraform_vms"]},
        "terraform_vms": {
            "children": ["vms_gateway", "vms_linux", "vms_windows"],
            "hosts": [],
        },
        "vms_gateway": {"hosts": []},
        "vms_linux": {"hosts": []},
        "vms_windows": {"hosts": []},
        "vms_dc": {"hosts": []},
        "vms_client": {"hosts": []},
        "vms_pinned": {"hosts": []},
    }

    for name, attrs in sorted(vms.items()):
        attrs = attrs or {}
        hypervisor = attrs.get("hypervisor", "")
        role = attrs.get("role", "linux")
        os_family = attrs.get("os_family", "")

        if not hypervisor:
            print(f"WARNING: VM '{name}' has no hypervisor attribute, skipping",
                  file=sys.stderr)
            continue
        if not os_family:
            print(f"WARNING: VM '{name}' has no os_family attribute - stale "
                  "terraform state, run `terraform apply`; skipping",
                  file=sys.stderr)
            continue

        inventory["terraform_vms"]["hosts"].append(name)

        if role == "gateway":
            inventory["vms_gateway"]["hosts"].append(name)
        elif os_family == "windows":
            inventory["vms_windows"]["hosts"].append(name)
        else:
            inventory["vms_linux"]["hosts"].append(name)

        if role == "dc":
            inventory["vms_dc"]["hosts"].append(name)
        elif role == "client":
            inventory["vms_client"]["hosts"].append(name)

        hostvars = {
            "libvirt_hypervisor": hypervisor,
            "ansible_host": f"{name}.{LAB_DOMAIN}",
        }
        # Pinned MAC — lets playbooks/opnsense.yml tie name -> MAC -> DHCP
        # lease without hardcoding IPs.
        if attrs.get("mac"):
            hostvars["libvirt_mac"] = attrs["mac"]
        # Static address — the `ip:` key in lab.yaml, passed through terraform.
        if attrs.get("ip"):
            hostvars["libvirt_ip"] = attrs["ip"]
        # Only a VM that has both can be turned into a DHCP reservation.
        if hostvars.get("libvirt_mac") and hostvars.get("libvirt_ip"):
            inventory["vms_pinned"]["hosts"].append(name)

        if os_family == "windows":
            hostvars.update({
                "ansible_connection": "winrm",
                "ansible_port": 5985,
                "ansible_user": "Administrator",
                "ansible_password": LAB_WIN_ADMIN_PASS,
                # NTLM, not the default plaintext: a stock guest enables only
                # Negotiate, and NTLM also works before the domain exists,
                # which Kerberos does not.
                "ansible_winrm_transport": "ntlm",
                "ansible_winrm_scheme": "http",
            })

        inventory["_meta"]["hostvars"][name] = hostvars

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
