# Infrastructure

The physical + IaC layer — the hypervisor substrate and the toolchain that
provisions everything drawn in [[Overview]]. This is the IaC side objective of
the project.

## Physical + overlay

VMs are distributed across two hypervisor hosts and joined into a single
Layer 2 network by a VXLAN overlay, so placement is a scheduling decision
rather than a network one: any VM can sit on any host and still share the lab
LAN.

```mermaid
flowchart TB

    subgraph HA["host_a · 192.168.1.161 · Arch / KVM"]
        OPN["opnsense01<br/><i>edge — DHCP / DNS / NAT / VPN</i>"]
        DC01["dc01<br/><i>Windows Server — AD</i>"]
    end

    subgraph HB["host_b · 192.168.1.134 · Arch / KVM"]
        CLIENT01["client01<br/><i>Windows 11</i>"]
        APP01["app01<br/><i>Ubuntu — Docker host</i>"]
    end

    HA <-->|"VXLAN 100 over wlan0<br/>bridge vm-br0"| HB

    class OPN,DC01,CLIENT01,APP01 built
    classDef built fill:#dff5e1,stroke:#2f9e44,color:#111827
```

> [!note] Edge placement
> The gateway VM is pinned to whichever host `edge.host` names in `lab.yaml`
> — exactly one host holds it. Everything else is free to move between hosts
> by changing one line in the placement map.

## IaC toolchain

How a change flows from a single edit to running infrastructure. `lab.yaml` is
the single source of truth for static facts.

```mermaid
flowchart LR

    LAB["lab.yaml<br/><i>hosts · network · VM placements</i>"]
    PKR["Packer<br/><i>builds base images</i>"]
    TF["Terraform<br/><i>libvirt provider</i>"]
    ANS["Ansible<br/><i>host + VM config</i>"]
    VMS["Running VMs"]

    LAB -->|"yamldecode (locals.tf)"| TF
    LAB -->|"dynamic inventory"| ANS
    PKR -.->|"base images"| TF
    TF  -->|"vms output → inventory"| ANS
    TF  -->|"defines"| VMS
    ANS -->|"configures / promotes / joins"| VMS

    class LAB,TF,ANS,VMS built
    class PKR planned

    classDef built fill:#dff5e1,stroke:#2f9e44,color:#111827
    classDef planned fill:#f1f3f5,stroke:#adb5bd,color:#495057,stroke-dasharray:5 5
```

- **Terraform** reads `lab.yaml`, builds VMs via the `dmacvicar/libvirt`
  provider, and emits a `vms` output that becomes Ansible's inventory.
- **Ansible** configures the hypervisors (bridge + VXLAN), starts VMs, promotes
  `dc01` to the forest, and joins `client01`.
- **Packer** will build the Windows and Linux base images that Terraform clones
  as backing files. Images are hand-built today (`base-image/README.md`).
- `deploy.sh` wraps the whole flow (terraform apply → ansible playbooks).
