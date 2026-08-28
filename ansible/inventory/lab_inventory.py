#!/usr/bin/env python3
"""Dynamic Ansible inventory built from lab.yaml (single source of truth).

lab.yaml at the repo root holds every static lab fact: hypervisor
addresses/users, the single edge host, shared network settings, VM
placement. This script turns it into inventory groups:

    hypervisors : every lab host (target of playbooks/hosts.yml)
    edge        : the single host named by the top-level `edge.host` key in
                  lab.yaml (target of playbooks/edge.yml)

VM-level inventory (VM names, placement, gateway/linux role) is derived
state owned by Terraform and served by inventory/terraform_vms.py instead.

Run alone:

    ansible-inventory -i inventory/lab_inventory.py --list

or combined with the terraform inventory, as deploy.sh does:

    ansible-playbook \
        -i inventory/lab_inventory.py \
        -i inventory/terraform_vms.py \
        playbooks/start_vms.yml
"""

import json
import os

import yaml

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
LAB_YAML = os.path.join(BASE_DIR, "..", "..", "lab.yaml")

# Network settings shared by every host (top-level `network` in lab.yaml).
NETWORK_VARS = (
    "bridge_name",
    "wan_name",
    "vxlan_name",
    "vxlan_id",
    "vxlan_port",
    "vxlan_parent_interface",
)


def load_lab():
    with open(LAB_YAML, "r", encoding="utf-8") as fh:
        return yaml.safe_load(fh)


def build_inventory(lab):
    hosts = lab.get("hosts", {})
    network = lab.get("network", {})

    hypervisor_hosts = {}

    for name, attrs in sorted(hosts.items()):
        peers = [h["address"] for h in hosts.values()
                 if h is not attrs]

        host_vars = {
            "ansible_host": attrs["address"],
            "ansible_user": attrs["user"],
            # Single VXLAN peer: fine while the lab has exactly two hosts.
            # More hosts need a mesh (multiple FDB entries) - see the
            # "more than two hosts" note in red-clay/Report/Journal.md.
            "vxlan_remote_ip": peers[0] if len(peers) == 1 else "",
            "vxlan_peer_ips": peers,
        }
        host_vars.update({k: network[k] for k in NETWORK_VARS if k in network})

        hypervisor_hosts[name] = host_vars

    edge = lab.get("edge") or {}
    if not isinstance(edge, dict) or not edge.get("host"):
        raise ValueError(
            "lab.yaml must set `edge.host` to a hypervisor name from `hosts`")
    edge_host_name = edge["host"]
    try:
        edge_hosts = {edge_host_name: dict(hypervisor_hosts[edge_host_name])}
    except KeyError:
        raise ValueError(
            f"lab.yaml `edge.host` ({edge_host_name!r}) does not match any "
            "host in the `hosts` map"
        ) from None

    return {
        "_meta": {"hostvars": {**hypervisor_hosts, **edge_hosts}},
        "all": {"children": ["hypervisors", "edge"]},
        "hypervisors": {"hosts": list(hypervisor_hosts)},
        "edge": {"hosts": list(edge_hosts)},
    }


def main():
    inventory = build_inventory(load_lab())

    if len(sys.argv) > 1 and sys.argv[1] == "--host":
        host = sys.argv[2] if len(sys.argv) > 2 else ""
        print(json.dumps(inventory["_meta"]["hostvars"].get(host, {})))
        return 0

    print(json.dumps(inventory))
    return 0


if __name__ == "__main__":
    import sys
    sys.exit(main())
