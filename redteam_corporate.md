# Bachelor Final Project Proposal  
## Full Red Team Simulation: Ethical Hacking of a Corporate Network

### 1. Project Overview

This project aims to simulate a **real-world red team penetration testing engagement** against a **realistically designed corporate network environment**.  
The primary objective is to **study, practice, and demonstrate ethical hacking techniques** by modeling how attackers gain initial access, abuse trust relationships, escalate privileges, move laterally, and ultimately compromise critical assets inside a corporate network.

---

### 2. Motivation

Modern corporate breaches rarely rely on breaking cryptographic protections or bypassing firewalls directly. Instead, they exploit:

- Trust relationships between systems
- Misconfigurations in identity and access control
- Weak authentication practices
- Over-privileged users and services
- Implicit assumptions in internal networks

Understanding **how attackers actually compromise organizations** requires hands-on exposure to the full attack lifecycle. This project provides that exposure by allowing the student to **think and operate from an attacker’s perspective**, while documenting each step.

---

### 3. Project Objectives

The main objectives of this project are:

- To understand how attackers obtain **initial access** to corporate environments
- To analyze **internal network trust models** and how they are abused
- To perform **privilege escalation** on compromised systems
- To demonstrate **lateral movement** across internal hosts
- To evaluate how **Active Directory-based environments** are targeted
- To simulate **data access and exfiltration scenarios**

The side Objectives of this project are:

- To learn IaC 
- To create an automated virtual deployment of envirnoments

---

### 4. Simulated Corporate Environment

The student will design and deploy a **virtualized corporate network** that reflects real enterprise structures.

#### 4.1 Network Structure

The environment will include:

- Internet-facing network segment
- Perimeter / DMZ network
- Internal corporate user network
- Server / data center network

#### 4.2 Core Systems

The simulated corporate infrastructure will include:

- Active Directory Domain Controller
- Domain-joined client machines
- Internal file server
- Database server
- Web application server (internal and/or DMZ)
- Centralized authentication and DNS services

The environment will intentionally contain **realistic misconfigurations and weaknesses**, which will be documented as part of the design.

---

### 5. Attacker Model and Scope

*Not final. I need a bit more information and knowledge to make this decision decisively.*

The project assumes an attacker with:

- No initial internal trust
- External network access only
- No prior knowledge of internal architecture

The attacker will gain access through **realistic entry points**, such as:

- Compromised user credentials
- Abused remote access mechanisms
- Exploitation of exposed services
- Compromised internal endpoints

The project does **not** focus on denial-of-service or destructive attacks. The emphasis is on **stealthy compromise, access expansion, and data exposure**, as seen in real-world breaches.

---

### 6. Methodology (Red Team Lifecycle)

The project follows a standard red team / penetration testing lifecycle:

#### 6.1 Reconnaissance and Enumeration
- External attack surface identification
- Internal network discovery after initial access
- Enumeration of users, services, and trust relationships

#### 6.2 Vulnerability and Misconfiguration Analysis
- Identification of authentication weaknesses
- Analysis of privilege boundaries
- Discovery of excessive permissions and insecure defaults

#### 6.3 Exploitation
- Controlled exploitation of identified weaknesses
- Proof-of-concept demonstrations

#### 6.4 Privilege Escalation
- Escalation from low-privileged access to administrative control
- Analysis of why escalation was possible

#### 6.5 Lateral Movement
- Movement between internal systems
- Abuse of credentials and trust relationships
- Expansion of attacker control within the domain

#### 6.6 Post-Exploitation and Impact Analysis
- Access to sensitive business data
- Evaluation of confidentiality and integrity impact
- Identification of potential persistence mechanisms

---

### 7. Deliverables

The project will produce the following deliverables:

1. **Network Architecture Documentation**
   - Asset inventory
   - Trust relationships

2. **Attack Path Documentation**
   - Step-by-step compromise chain
   - Privilege progression
   - Lateral movement paths

3. **Vulnerability and Risk Analysis**
   - Severity classification
   - Root cause analysis

4. **Penetration Testing Report**
   - Executive summary
   - Technical findings
   - Mitigation recommendations

5. **Final Presentation and Demonstration**

---

### 8. Learning Outcomes

By completing this project, the student will gain:

- Practical understanding of **ethical hacking methodologies**
- Deep insight into **Active Directory and enterprise authentication**
- Experience with **real-world attacker decision-making**
- Ability to analyze and explain **security failures**
- Skills in **technical reporting and professional documentation**

- Deploy highly customizable VMs across multiple machine by **IaC**
- How corporate networks are desighned

