
```
Physical machines
└── Proxmox (or raw KVM) — hypervisor layer
    └── Terraform — provisions the VMs
        └── Packer — built the base images Terraform uses
            └── Ansible — configures AD, users, misconfigs
                └── AD environment with intentional weaknesses
                    ├── Kerberoastable service accounts
                    ├── Misconfigured ACLs
                    ├── NTLM relay opportunities
                    └── ...
                        └── You attack this using a C2 + tooling
                            └── Wazuh/SIEM shows what got detected
```