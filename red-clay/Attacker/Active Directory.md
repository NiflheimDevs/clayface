# Active Directory
**Active Directory is a directory service built by Microsoft** that manages authentication and authorization across an entire corporate network. Think of it as a central database that answers two questions: *who are you* and *what are you allowed to do*.

In a corporate network, instead of every machine managing its own users and passwords, there's a **Domain Controller (DC)** — a server running AD that is the single source of truth. When you log into any machine joined to that domain, your credentials are verified against the DC, not the local machine.

The DC stores:
- All user accounts and their password hashes
- All computer accounts
- Groups and their memberships
- Policies (what users/machines can and can't do)
- **Trust relationships** between everything

Every machine that is "domain-joined" trusts the DC completely. This central trust model is both what makes corporate networks manageable and what makes them vulnerable — if you can abuse the DC's trust, you can abuse the entire network.

---
# Authentication Protocols — Kerberos and NTLM

These two are the backbone of AD authentication. Every attack I mentioned exploits one of them.

## NTLM (NT LAN Manager)

NTLM is the older of the two. When you authenticate with NTLM, your password is **never sent over the network**. Instead:

1. Client says "I want to authenticate"
2. Server sends a random **challenge** (a nonce)
3. Client takes your **NT hash** (the hashed version of your password) and encrypts the challenge with it, sending back the result
4. Server (or DC) does the same encryption and compares results

The critical thing: **the NT hash is functionally equivalent to your password**. If I steal the hash, I don't need to crack it — I can just use the hash directly to authenticate. That's **Pass-the-Hash**.

## Kerberos

Kerberos is the newer, preferred protocol. It works on a **ticket** system and involves three parties: the client, the service they want to access, and a **Key Distribution Center (KDC)** — which in Windows, runs on the DC.

The flow:
1. You log in → your client sends your password hash to the KDC and gets back a **TGT (Ticket Granting Ticket)** — an encrypted blob that proves "this person authenticated"
2. When you want to access a specific service (say, a file server), your client presents the TGT to the KDC and asks for a **TGS (Ticket Granting Service) ticket** for that specific service
3. The KDC issues a TGS encrypted with the **service account's password hash**
4. You present the TGS to the service — the service decrypts it using its own password hash to verify you

The important detail: the TGS is encrypted with the **service's password hash**, not yours. The KDC sends it to you, but you theoretically can't read it — only the service can. But you *can* take that encrypted blob offline and try to crack it. That's **Kerberoasting**.

---


---

## 4. C2 Frameworks

**C2 = Command and Control.** It's the infrastructure an attacker uses to maintain persistent access to compromised machines and issue commands to them.

When you compromise a machine, you want a way to interact with it that:
- Survives reboots
- Communicates back to you over normal-looking traffic (HTTP/S, DNS)
- Can be operated across many compromised machines simultaneously
- Has built-in tooling for post-exploitation (dumping hashes, lateral movement, etc.)

A C2 framework provides all of this. The attacker runs a **C2 server** (the "team server"), and on each compromised machine runs an **agent/implant** (a small program that calls home to the server and executes commands).

- **Cobalt Strike** — the industry standard, commercial, very feature-rich, heavily used by both red teams and real threat actors. Extremely expensive
- **Sliver** — open source, written in Go actually, actively maintained, good alternative to Cobalt Strike
- **Havoc** — newer open source option, modern features
- **Brute Ratel** — commercial, designed specifically to evade EDR

Without a C2, your "red team simulation" is just running individual tools in sequence. A real attacker operates through a persistent channel — the C2 is what makes it a campaign instead of a series of isolated steps.

---

## 5. Defensive Technologies

### EDR (Endpoint Detection and Response)

Software running on each machine (endpoint) that monitors behavior at a deep level — syscalls, process injection, memory access patterns, network connections. Unlike old antivirus (which just matched file signatures), EDR watches what programs actually *do* at runtime.

Examples: CrowdStrike Falcon, SentinelOne, Microsoft Defender for Endpoint.

### SIEM (Security Information and Event Management)

A centralized system that collects logs from all machines and services in the network, normalizes them, correlates events, and generates alerts. It's the blue team's window into what's happening across the entire environment.

Example: Splunk, Elastic SIEM, Microsoft Sentinel.

### Wazuh

An open source SIEM / host-based intrusion detection system. Very good fit for a lab environment because it's free, well-documented, and integrates with the Elastic stack for visualization.

### Windows Event Forwarding (WEF)

A native Windows feature where machines forward their event log entries to a central collection server. It's a lightweight alternative to deploying full SIEM agents everywhere. In a lab, you'd forward Security logs (4624 logon events, 4688 process creation, etc.) to a central Windows server, then ingest those into something like Wazuh.
