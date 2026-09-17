# Attack Paths

Attacker movement through Clayface — the **attack graph**, kept separate from
the [[Overview]] network map and the [[Identity]] trust graph on purpose. Every
chain starts at the attacker (`KALI01`) and is numbered; dashed = not built /
exercised yet.

```mermaid
flowchart LR

    KALI["KALI01<br/><i>attacker</i>"]
    VPNP["VPN pool<br/>10.0.100.0/24"]

    subgraph LAB["Clayface"]
        OPN["OPNsense"]
        APP01["APP01<br/>web / API / DB"]
        CLIENT01["CLIENT01"]
        DC01["DC01 · AD"]
        PG["PostgreSQL<br/><i>(container on APP01)</i>"]
    end

    KALI -->|"1 · stolen creds"| VPNP
    VPNP -->|"1 · tunnel"| OPN
    OPN  -->|"1 · reaches user zone"| CLIENT01
    CLIENT01 -->|"1 · AD enum → privesc"| DC01

    KALI -->|"2 · public web / API exploit"| OPN
    OPN  -->|"2 · published HTTPS"| APP01
    APP01 -->|"2 · SQLi / creds"| PG
    APP01 -.->|"2 · service identity"| DC01

    class KALI,VPNP attacker
    classDef attacker fill:#ffe3e3,stroke:#e03131,color:#111827
```

## Chains

1. **Credential → VPN → workstation → AD**
   Stolen employee creds get a VPN foothold, land on `CLIENT01`, then AD
   enumeration and privilege escalation against `DC01`. Primary path.
2. **Public web → APP01 → database / AD**
   Exploit a published web/API container on `APP01`, pivot to `PostgreSQL`
   for stored credentials, and where a service identity exists, on toward AD.

Both paths converge on **AD as the crown jewel** — the point the whole lab is
built to demonstrate.

## Not yet exercised

`FILE01` SMB/NTLM relay, `CA01` AD CS abuse, `DC02` replication attacks, and
web → `IDP01` → AD are planned once those optional modules exist. See the
optional-modules list in the knowledge base.
