# Identity

The **authentication / trust graph** — who proves identity to whom. Separate
from the [[Overview]] network map and the [[Attack Paths]] attack graph, per the
project rule that these are different graphs.

Two identity worlds meet here: traditional Windows enterprise identity (AD on
`DC01`) and modern application identity (`IDP01`). Solid = built; dashed =
planned.

```mermaid
flowchart TB

    subgraph TRAD["Traditional enterprise identity"]
        DC01["DC01<br/><i>Active Directory · Kerberos · LDAP</i>"]
        CLIENT01["CLIENT01<br/><i>domain-joined</i>"]
    end

    subgraph MODERN["Modern application identity"]
        IDP01["IDP01<br/><i>OIDC · OAuth2 · SAML · SSO</i>"]
        PORTAL["Web apps on APP01<br/><i>portal · API · admin</i>"]
    end

    CLIENT01 -->|"Kerberos / NTLM domain logon"| DC01
    PORTAL   -->|"OIDC / SAML"| IDP01
    IDP01    -.->|"federation (planned)"| DC01
    APP01SVC["APP01 services"] -.->|"LDAP / Kerberos when useful"| DC01

    class DC01,CLIENT01 built
    class IDP01,PORTAL,APP01SVC planned

    classDef built fill:#dff5e1,stroke:#2f9e44,color:#111827
    classDef planned fill:#f1f3f5,stroke:#adb5bd,color:#495057,stroke-dasharray:5 5
```

## Notes

- **AD is not just a login server.** It carries users, groups, authorization,
  domain policy, and the trust edges that make privilege escalation and lateral
  movement possible — that is why it is core.
- **DNS split**: `clayface.local` is authoritative on `DC01`; everything else
  forwards to OPNsense. The `.clayface` DNS domain (VM discovery) is owned by
  the OPNsense image and is a *separate* thing from the `clayface.local` AD
  forest. See CLAUDE.md for the exact chain.
- **IDP01 ↔ AD federation** is planned, not built — it is what will let web
  apps ride enterprise identity and open the web → IDP → AD path.
