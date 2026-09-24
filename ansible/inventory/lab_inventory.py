#!/usr/bin/env python3
"""Dynamic Ansible inventory built from lab.yaml (single source of truth).

lab.yaml at the repo root holds every static lab fact: hypervisor
addresses/users, the single edge host, the network segments, VM placement.
This script turns it into inventory groups:

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


def load_lab():
    with open(LAB_YAML, "r", encoding="utf-8") as fh:
        return yaml.safe_load(fh)


def build_inventory(lab):
    hosts = lab.get("hosts", {})
    networks = lab.get("networks", {})

    # Overlay settings shared by every segment, declared once at the top
    # level. `vxlan_port` is the standard; `vxlan_parent_interface` is a
    # property of the host NIC. Both ride on the hypervisor host vars below.
    vxlan_port = lab.get("vxlan_port")
    vxlan_parent_interface = lab.get("vxlan_parent_interface")

    hypervisor_hosts = {}

    for name, attrs in sorted(hosts.items()):
        peers = [h["address"] for h in hosts.values()
                 if h is not attrs]

        host_vars = {
            "ansible_host": attrs["address"],
            "ansible_user": attrs["user"],
            # The whole segment map, so a playbook can loop over it rather
            # than over a set of flattening keys that had to be kept in sync
            # with lab.yaml by hand.
            #
            # Each entry is {bridge, subnet, gateway, dns, monitor,
            # vxlan_id}; only `bridge` is required, because the WAN leg has
            # none of the others. A segment WITHOUT a `vxlan_id` is the
            # statement "this leg has no overlay" - that is what makes `wan`
            # self-describing and what keeps an attacker VM on the WAN pinned
            # to the host that holds OPNsense.
            "lab_networks": networks,
            "vxlan_port": vxlan_port,
            "vxlan_parent_interface": vxlan_parent_interface,
            # Single VXLAN peer: fine while the lab has exactly two hosts.
            # More hosts need a mesh (multiple FDB entries) - see the
            # "more than two hosts" note in red-clay/Report/Journal.md.
            # `vxlan_peers` carries the whole list so the eventual mesh work
            # has it; `vxlan_remote_ip` stays the "can a two-host overlay be
            # built, and to whom" answer that hosts.yml guards on, and is ""
            # whenever the peer count is not exactly one.
            "vxlan_remote_ip": peers[0] if len(peers) == 1 else "",
            "vxlan_peers": peers,
        }

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

    # The AD identity facts are consumed by playbooks that target the VMs
    # (dc01, client01), not the hypervisors - so they go on `all`, which
    # every host inherits. See CLAUDE.md's data-flow rule: users and group
    # membership are static lab facts and belong in lab.yaml.
    ad = lab.get("ad") or {}

    # Same reasoning for the application server: playbooks/app.yml renders the
    # weakness toggles into the container environment from this map, and
    # playbooks/app_validate.yml asserts against them. Both target the VM, so
    # the map rides on `all` rather than on the hypervisor that hosts it.
    app = lab.get("app") or {}

    # The remote-access VPN pool. Published so the fact has a home in the
    # machine-read data, exactly as the target topology describes it.
    #
    # NOTHING READ IT YET, and nothing answers on it: there is no VPN server
    # in this lab, no WireGuard config, and no OPNsense VPN instance. It is
    # declared so that no document can quietly imply the lab supports VPN
    # initial access when it does not.
    vpn = lab.get("vpn") or {}

    return {
        "_meta": {"hostvars": {**hypervisor_hosts, **edge_hosts}},
        "all": {
            "children": ["hypervisors", "edge"],
            # `lab_networks` rides on `all` as well as on each hypervisor,
            # because it is a lab-wide fact and two of its consumers are not
            # hypervisors: start_vms.yml waits for the LAN segment's DNS
            # address, and that play targets VMs, which come from the
            # terraform inventory and would not otherwise see the map.
            #
            # The overlay transport facts stay on the hypervisors: a VXLAN
            # parent interface is a property of a host NIC and means nothing
            # to a guest.
            "vars": {
                "lab_ad": ad,
                "lab_app": app,
                "lab_vpn": vpn,
                "lab_networks": networks,
            },
        },
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
