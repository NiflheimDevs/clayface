# Clayface — Overview

Clayface is an IaC-provisioned virtual corporate network with intentional,
documented misconfigurations, built for practicing the ethical-hacking
lifecycle (initial access, privilege escalation, lateral movement, data
exposure).

This page is the **network / zone map** only — the one-glance picture of what
sits where. It deliberately shows *no* attack arrows and *no* service
internals, because those are different graphs and mixing them is what makes a
diagram unreadable. See instead:

- [[Attack Paths]] — how an attacker moves through the lab.
- [[Identity]] — AD / IDP authentication and trust.
- [[Infrastructure]] — the physical hosts and the IaC toolchain that builds all this.

## Target network (zones)

Solid boxes are **built today**; dashed boxes are **planned** (this diagram is
the target vision — we are mid-development).

```mermaid
flowchart TB

    subgraph WAN["External"]
        KALI["KALI01 · Kali Linux<br/><i>attacker</i>"]
        NET["Internet"]
        VPN["VPN clients · 10.0.100.0/24<br/><i>remote-access pool</i>"]
    end

    subgraph EDGE["Boundary"]
        OPN["OPNsense<br/><i>firewall · NAT · DHCP · DNS · VPN</i><br/>gateway .1 in every zone"]
    end

    subgraph DMZ["DMZ / Application zone · 10.0.10.0/24"]
        APP01["APP01 · Ubuntu<br/><i>Docker host — web / API / DB containers</i>"]
    end

    subgraph USER["Internal — user zone · 10.0.20.0/24"]
        CLIENT01["CLIENT01 · Windows 11<br/><i>domain-joined employee PC</i>"]
    end

    subgraph SRV["Internal — server / identity zone · 10.0.30.0/24"]
        DC01["DC01 · Windows Server<br/><i>AD DS · Kerberos · LDAP · DNS</i>"]
        IDP01["IDP01<br/><i>OIDC / OAuth2 / SAML / SSO</i>"]
        OPT["Optional modules<br/><i>FILE01 · CA01 · DC02 · MAIL01 · …</i>"]
    end

    NET <-->|WAN / NAT| OPN
    KALI --- NET
    VPN -->|"encrypted tunnel → VPN pool"| OPN
    OPN --- DMZ
    OPN --- USER
    OPN --- SRV

    class KALI,VPN planned
    class APP01,IDP01,OPT planned
    class OPN,CLIENT01,DC01 built

    classDef built fill:#dff5e1,stroke:#2f9e44,color:#111827
    classDef planned fill:#f1f3f5,stroke:#adb5bd,color:#495057,stroke-dasharray:5 5
```

> [!note] Four segments
> Matches the proposal's four zones: internet-facing, DMZ, internal **user**,
> and internal **server / data-center**. Splitting user from server is what
> makes the lateral-movement story (`CLIENT01 → DC01`) legible.

### Target addressing

| Zone | Subnet | Gateway | Key hosts |
| --- | --- | --- | --- |
| DMZ / application | `10.0.10.0/24` | `.1` (OPNsense) | APP01 `.10` |
| Internal — user | `10.0.20.0/24` | `.1` (OPNsense) | CLIENT01 `.20` |
| Internal — server / identity | `10.0.30.0/24` | `.1` (OPNsense) | DC01 `.10`, IDP01 `.20` |
| VPN remote-access pool | `10.0.100.0/24` | OPNsense tunnel endpoint | assigned per client |
| WAN | upstream / NAT | — | OPNsense |

Segmentation is enforced at OPNsense: the VPN pool and the DMZ each reach the
internal zones only through explicit firewall rules, which is what makes
"attacker starts with no internal trust" a real constraint rather than a claim.

Addresses are the target scheme. The lab currently runs one flat
`10.0.0.0/24` on `vm-br0` — see [[Infrastructure]].

## Target stack

Conceptual build stack — the layered vision, not current state. Packer and the
SIEM layer are **not built yet**.

```
Physical machines
└── KVM / libvirt — hypervisor layer                 [built]
    └── Terraform — provisions the VMs               [built]
        └── Packer — will build the base images      [planned; images hand-built today]
            └── Ansible — configures AD, users, misconfigs   [built for AD/client]
                └── AD environment with intentional weaknesses
                    ├── Kerberoastable service accounts
                    ├── Misconfigured ACLs
                    ├── NTLM relay opportunities
                    └── …
                        └── Attack this using a C2 + tooling         [planned]
                            └── Wazuh / SIEM shows what got detected  [planned]
```

See [[Infrastructure]] for the real, current physical + IaC layer.
