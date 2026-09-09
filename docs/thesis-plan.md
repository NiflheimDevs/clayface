# Thesis writing plan — chapter readiness assessment

Assessment date: 2026-09-09. Updated 2026-09-10 with settled decisions.
This is a planning document, not thesis text.

## Settled decisions

1. **Framing: the approved proposal.** The thesis is a *full red team
   simulation — ethical hacking of a corporate network*, per
   `docs/redteam_corporate.md`. The "lightweight distributed enterprise
   digital twin" reframe suggested by the comparative research is **not**
   adopted. Consequences are worked through below — they are significant.
2. **Language: Persian**, with English technical terminology kept in
   English rather than translated. Latin terms stay Latin inline
   (Active Directory, Kerberos, Terraform, VXLAN, …); the واژه‌نامه
   glossary at the back carries the term list. Do not invent Persian
   equivalents for tool names or protocol names.
3. **Template**: the MSc/PhD cover wording is accepted as-is.
4. **Project name: Clayface**, everywhere. Anything still saying "RedClay"
   or "red clay" gets renamed when it is used in the thesis.
5. Evaluation methodology — see the dedicated section below; the decision
   was not understood the first time, so it is restated properly there.

## What the framing decision changes

This is the most important consequence of decision 1, and it should be
read before anything else in this document.

Under the proposal framing, the thesis is judged on **the red-team
lifecycle**: reconnaissance, initial access, privilege escalation, lateral
movement, post-exploitation and data exposure (proposal §6), producing the
deliverables in proposal §7 — network architecture documentation, attack
path documentation, vulnerability and risk analysis, and a penetration
test report.

The proposal itself lists infrastructure-as-code as a **side objective**
("To learn IaC", "To create an automated virtual deployment of
environments"). So the work completed so far — multi-host KVM, VXLAN
overlay, OPNsense edge, Terraform+Ansible automation, the single source of
truth — is, under this framing, the *supporting* half of the thesis, not
the headline result.

Plainly: **the primary deliverable of the chosen framing is currently at
zero percent.** There is no Active Directory, no domain-joined client, no
application or database, no planted misconfiguration, no attacker machine,
no executed attack, no evidence. The secondary deliverable is well
advanced. The plan below is ordered around closing that gap.

This is not an argument to switch framings — it is the scope the proposal
was approved on, and the corrections file already shows the alternative
framing rests on a differentiation claim that is itself unverified. It
just means the build work between now and a defensible chapter 5 is
larger than it looks.

## The template

`thesis/CE Thesis Template/CE Thesis Template.docx` is the IUST Computer
Engineering template, in Persian, right-to-left, with a bundled IEEE
reference style (`Resources/IEEE Reference Template/IEEE.XSL`) and Persian
fonts (B Nazanin, B Titr). Its chapter skeleton is:

| # | Chapter | Sections |
|---|---|---|
| 1 | مقدمه | 1-1 شرح مسأله · 1-2 انگيزه‌هاي پژوهش · 1-3 مفروضات پژوهش (اختياري) · 1-4 اهداف پژوهش · 1-5 ساختار پايان‌نامه |
| 2 | مفاهيم پایه | 2-1 مقدمه · 2-2 … · 2-3 نتيجه‌گيري |
| 3 | مروري بر كارهاي مرتبط | 3-1 مقدمه · 3-2 عنوان بخش · 3-3 مقايسه كارهاي مرتبط · 3-4 نتيجه‌گيري |
| 4 | روش/فن/طرح پيشنهادي | 4-1 مقدمه · 4-2 معرفي روش يا متدولوژي · 4-3 كليات طرح · 4-4 جزئيات/اجزاء · 4-5 نتيجه‌گيري |
| 5 | ارزيابي روش/فن/طرح پيشنهادي | 5-1 مقدمه · 5-2 روش ارزيابي و مدل‌سازي · 5-3 جزئيات استفاده از روش · 5-4 نتايج ارزيابي · 5-5 نتيجه‌گيري |
| 6 | نتيجه‌گيري و کارهاي آينده | 6-1 نتيجه‌گيري · 6-2 کارهاي آينده |

Plus مراجع (references) and واژه‌نامه (glossary).

No working thesis file exists yet — only the blank template. Do not edit
the template in place; copy it.

## Source material available

| Source | What it gives |
|---|---|
| `docs/redteam_corporate.md` | The approved proposal — now the governing document. Motivation, objectives, attacker model, red-team lifecycle, deliverables, learning outcomes. Feeds chapters 1, 4, 5. |
| `docs/redclay.yaml` | Intent: goals, design principles, target topology, identity architecture, module policy, decisions **with reasons**, constraints, current state. Feeds chapters 1, 4, 6. Its attack-path and technique-coverage sections are the seed of chapter 5's scenario design. |
| `red-clay/Report/Journal.md` | Chronological reasoning: why multi-host, why a single source of truth, why the knowledge-base split. The hardest material to reconstruct later — use it for chapter 4's methodology narrative. |
| `red-clay/Attacker/*.md` | Notes on techniques, Active Directory, red teaming. Raw material for chapters 2 and 5. |
| `red-clay/Overview.md` | The end-goal stack diagram (Packer → Terraform → Ansible → vulnerable AD → C2 → SIEM). Useful as the chapter 4 target-architecture figure, but note it describes intent, not current state. |
| `lab.yaml`, `terraform/`, `ansible/`, `deploy.sh` | Ground truth of what is implemented. Chapter 4 must be written against these, not against `redclay.yaml`. |
| `docs/TODO.md` | Known unbuilt work. Feeds chapter 6. |
| `ignoreme/gpt-research.md` | Comparative project research — read its corrections file first. |
| `ignoreme/gpt-research-corrections.md` | Verification of the above: what is right, wrong, and missing. |

## Chapter-by-chapter readiness

### فصل 1 — مقدمه · writable now, ~85%

Everything needed exists, and the framing decision unblocks it. 1-1 (شرح
مسأله) and 1-2 (انگيزه‌ها) come from the proposal's §1–2 plus the journal's
framing of the two challenges (real simulation, dynamic environment). 1-3
(مفروضات) maps onto the proposal's attacker model — which the proposal
itself marks *"Not final"*, so either settle it or state it explicitly as a
research assumption — plus `redclay.yaml`'s constraints. 1-4 (اهداف) is the
proposal's §3, main and side objectives, stated in that order of
importance. 1-5 is written last.

Write this after chapter 3, so the positioning against existing work is
already argued.

### فصل 2 — مفاهيم پایه · writable now, ~70%

Depends on published knowledge, not on implementation progress, so it can
never be invalidated by later build work. Safest place to start writing.
Candidate sections, all of which the project genuinely uses:

- Virtualization: KVM, QEMU, libvirt; type-1 vs type-2 and why it matters
  for a lab.
- Infrastructure as Code: declarative vs imperative; Terraform's model
  (providers, modules, state); configuration management with Ansible
  (inventories, playbooks, idempotence).
- Network virtualization: Linux bridges, VXLAN and L2 overlays (RFC 7348),
  why an overlay is needed to span physical hosts.
- Enterprise identity: Active Directory, Kerberos, LDAP, domain
  authentication, service principal names, delegation.
- Perimeter services: firewall/router, NAT, DHCP, DNS, the DMZ concept.
- Offensive security: the red-team / penetration-testing lifecycle and
  MITRE ATT&CK as the vocabulary for phases and techniques. Under the
  chosen framing this section carries more weight than the others — give
  it room.
- Cyber ranges and security testbeds, per the published surveys.

Missing: citations. Every one of these needs a source and the project has
no bibliography yet. Building the reference list *is* the work in this
chapter; the prose is the easy part.

### فصل 3 — مروري بر كارهاي مرتبط · ~60%, not writable yet

Blocking gaps, detailed in `ignoreme/gpt-research-corrections.md`:

1. **No academic sources.** Seven relevant peer-reviewed works were
   verified with DOIs and listed in that file. Must be added.
2. **DEFENSE.LAB cannot be found**; three other listed projects are 1–2
   star personal repositories presented as peer work. Fix the tier list.
3. **The comparison table credits Clayface with capabilities it does not
   have.** Must be rebuilt as implemented-versus-planned.
4. **The "no comparable project uses a VXLAN overlay" claim is
   unverified.** Proxmox SDN has a VXLAN zone type and Ludus runs on
   Proxmox. Check before claiming uniqueness.

The framing decision adds a fifth item, and it is the awkward one:

5. **Under the proposal framing, the comparison axis moves onto GOAD's home
   ground.** A red-team-simulation thesis is naturally compared on attack
   surface, attack-path depth and technique coverage — exactly where GOAD
   (5 VMs, 2 forests, 3 domains, mature ADCS/Kerberos/ACL paths, 8,300
   stars) and AD-Security-Lab are far ahead. The chapter must answer "why
   not just use GOAD?" honestly. The defensible answer is not "Clayface has
   more vulnerabilities" — it is that the thesis builds the *entire*
   environment, from physical hosts and network fabric upward, as the
   object of study, and then conducts and documents a full engagement
   against something it designed. GOAD hands you a pre-built target; here
   the design of the target, including *why* each weakness exists, is part
   of the contribution. Say that plainly and do not overclaim.

Effort to unblock: a few hours of reading for items 1–2, an afternoon of
documentation checking for item 4, and a paragraph of honest positioning
for item 5.

### فصل 4 — روش/فن/طرح پيشنهادي · ~50%

Writable now, restricted to what is built:

- 4-2 (methodology): the red-team lifecycle as the methodology the work
  follows (proposal §6) — under this framing that is the primary answer to
  "معرفي روش يا متدولوژي". Then the supporting methodology: IaC plus
  configuration management, the single-source-of-truth rule, why facts live
  in `lab.yaml` and derived state flows through Terraform outputs. The
  journal documents the real reasoning that led there.
- 4-3 (كليات): physical hosts → VXLAN/bridge overlay → OPNsense edge → VMs;
  core-plus-optional-modules architecture; separation of infrastructure
  from security state; the target corporate topology from proposal §4.
- 4-4 (جزئيات), for the parts that exist: the `lab.yaml` schema, Terraform
  root module and per-host provider blocks, the `alpine` / `opnsense` /
  `domaincontroller` modules, the `edge.host` guard, the two Ansible
  dynamic inventories, the `vms` output link, `deploy.sh` orchestration.

Not writable yet, because not built: Active Directory configuration
(the `domaincontroller` module provisions a VM from a hand-made base image
but nothing promotes or configures a domain), CLIENT01 and domain join,
APP01 and its containerized services, PostgreSQL, IDP01, DMZ/internal
segmentation, the attacker machine, planted misconfigurations, planted
sensitive data, and the documented attack paths.

Under the proposal framing this chapter also owns **the design of the
intentional weaknesses** — the proposal says they "will be documented as
part of the design". That is a substantial section that does not exist in
any form yet: for each planted weakness, what it is, why a real
organization would plausibly have it, which attack it enables, and what
the correct configuration would be. `redclay.yaml`'s `attack_surface` and
`technique_coverage` sections are the raw list to build it from.

Realistically: chapter 4 drafts to about half its final length now and
completes as the Active Directory phase lands.

### فصل 5 — ارزيابي · ~10%, effectively blocked

See the evaluation section below — the methodology decision is still open,
and nothing has been measured or executed.

### فصل 6 — نتيجه‌گيري و کارهاي آينده · ~40%

6-2 (کارهاي آينده) is writable now from `redclay.yaml`'s
`next_major_work`, the optional-module catalogue, the deferred scenario
engine, and `docs/TODO.md` (network segmentation, hardcoded network facts,
the three-host VXLAN mesh limitation). 6-1 is blocked on chapter 5.

## What "decision 5" actually asks — evaluation methodology

The earlier wording was unclear. This is **not** about comparing Clayface
against other projects — that is chapter 3, and there the answer "compare
fairly on all sides" is right.

Chapter 5 (ارزيابي روش/فن/طرح پيشنهادي) asks a different question:

> **How do you prove that what you built actually works and is worth
> something — and by what measure?**

You cannot answer "it works because I built it". The chapter needs a
stated method, applied, producing results. The decision is *which method*,
because it determines what you must instrument and record while building,
and retroactive measurement is both painful and less credible.

Under the proposal framing, the primary evaluation is the **engagement
itself**. The natural structure is:

- **5-2 (روش ارزيابي)** — the evaluation method is the execution of a
  full red-team engagement against the environment, following the
  lifecycle in proposal §6, with each step mapped to a MITRE ATT&CK
  technique. State the rules of engagement, the attacker's starting
  position, and what counts as success at each phase.
- **5-3 (جزئيات)** — the engagement as executed: reconnaissance, initial
  access, privilege escalation, lateral movement, post-exploitation, with
  commands, tool output and screenshots as evidence.
- **5-4 (نتايج)** — the results: which attack paths were completed, which
  planted weaknesses were reachable and which were not, the privilege
  progression, the root-cause analysis per finding, and severity
  classification. This is where proposal §7's "vulnerability and risk
  analysis" and penetration-test report land.

Secondary, quantitative evaluation of the infrastructure side objective —
pick two or three, not all:

- **Reproducibility**: run `deploy.sh` from clean N times; does it converge
  to an identical environment? Ansible idempotence (second run reports zero
  changed tasks).
- **Resource footprint**: RAM/CPU/disk of the deployed lab, against the
  published baselines of comparable projects (Ludus documents 32 GB RAM per
  range; GOAD full is 5 Windows VMs). This is the quantitative backing for
  any "lightweight" claim, and the comparison numbers are already
  documented.
- **Deployment time**: wall-clock from zero to a running environment.
- **Extensibility**: configuration effort to add a VM or a host, in lines
  changed — one `vm_placements` entry versus one host entry plus a provider
  block plus per-module blocks. This quantifies a known architectural wart
  honestly.
- **Functional validation**: connectivity matrix across the VXLAN overlay,
  DHCP/DNS resolution through OPNsense, VM-to-VM reachability across
  physical hosts.

**What to decide now:** confirm the engagement is the primary evaluation,
and pick which two or three quantitative measures accompany it. Everything
else in chapter 5 follows from that.

**What to start doing now, regardless:** capture evidence as you go.
Terminal output, screenshots, timings, `deploy.sh` logs. Every command run
against the lab from here on is potential chapter 5 material, and none of
it can be recovered later.

## Other blockers

Found while checking the repository against the thesis requirements. These
are not writing problems; they are build problems that will stop the
thesis if left until late.

### 1. The Windows base image is hand-made and outside the repository

`terraform/modules/domaincontroller/main.tf` builds DC01 from a backing
store at a hardcoded absolute path:
`/var/lib/libvirt/images/ws-ad-base.qcow2`. That image is not in the
repository, is not fetched by any tooling, is not described anywhere, and
its path is not in `lab.yaml` (which also violates the project's own
single-source-of-truth rule — the bridge name `vm-br0` and the storage
pool `default` are hardcoded in that module too).

Every comparable project solves this with **Packer** — GOAD, Ludus,
Range-as-Code and DetectionLab all build templates from ISOs
automatically. `red-clay/Overview.md` already lists Packer in the intended
stack, and `docs/redclay.yaml` does not.

Thesis impact: any reproducibility claim in chapter 5 is only as strong as
this step. "Reproducible except for one hand-built Windows image that only
exists on my machine" is a weak claim. Either automate it with Packer, or
document the image build procedure precisely enough that someone else could
repeat it, and state the limitation openly.

### 2. There is no configuration management for Windows at all

`ansible/roles/` is empty, and nothing in `ansible/` mentions WinRM,
`ansible_connection`, or Windows in any form. The three playbooks configure
hypervisors, the edge, and VM startup — all Linux, all over SSH.

So the entire Active Directory phase — domain promotion, users, groups,
ACLs, service accounts, the planted misconfigurations, the client domain
join — has no automation path yet. This is the single largest piece of
unbuilt work, and it is on the critical path for chapters 4 and 5 both.
Decide early whether it will be Ansible over WinRM, unattended
`autounattend.xml` plus scripts, or manual configuration with documented
steps. If it ends up manual, say so in the thesis rather than implying
automation that does not exist.

### 3. `tiny11.iso` is not a defensible base image for a thesis

`vm/images/` contains `tiny11.iso` (5.6 GB) — a community-modified,
debloated Windows 11 build, not a Microsoft-distributed image. Using it
makes results non-reproducible for anyone else and raises a licensing
question in a document that gets formally defended.

Microsoft distributes 180-day evaluation ISOs for Windows Server and
90-day enterprise evaluation images for Windows client, and this is what
GOAD uses and documents. Switch to those, and state the evaluation-license
constraint as a limitation.

### 4. There is no attacker machine in the topology

`docs/redclay.yaml` lists the core as OPNsense, DC01, CLIENT01, APP01,
IDP01 — no attacker node. Every comparable project has one (Range-as-Code
and DetectionLab ship a Kali VM; the proposal's §5 attacker model assumes
"external network access only").

Under the chosen framing this is a required component, not an optional
one: chapter 5 is an engagement, and it has to be conducted from
somewhere. Decide whether the attacker is a VM inside the lab on the WAN
side of OPNsense, or the physical host reaching in — the choice determines
what "external, no prior knowledge" actually means, and it must match what
chapter 1's assumptions section claims.

### 5. There is nothing to exfiltrate

The proposal's objectives include "data access and exfiltration scenarios"
and impact analysis on "sensitive business data". No planted data exists
anywhere in the design — not in `redclay.yaml`, not in the code. Fake
sensitive documents, database contents, credential files and share
contents need to be part of the environment design (chapter 4) for the
impact analysis in chapter 5 to have a subject.

### 6. No ethics, scope and isolation statement exists

A thesis that documents working attacks needs an explicit statement of
authorization, scope and containment — that the environment is
purpose-built, isolated, owned by the author, and that no technique was
exercised outside it. This belongs in chapter 1 (or a dedicated short
section) and does not exist yet.

There is also a concrete containment question to answer honestly: OPNsense
provides NAT to the outside, so deliberately vulnerable Windows machines
have outbound internet access. GOAD's own README warns to keep such
environments off the internet. Decide the isolation posture, implement it,
and describe it.

### 7. Figures do not exist

The vault has one Excalidraw drawing. The thesis needs, at minimum: the
physical/host topology, the network topology with the VXLAN overlay, the
logical corporate network with zones, the tooling pipeline
(Terraform/Ansible/Packer flow), and one attack-path diagram per
documented chain. `redclay.yaml`'s `graphs` section already argues these
are four *different* graphs that should not be merged into one picture —
follow that advice.

Note also that every screenshot in `ignoreme/gpt-research.md` is an
`images.openai.com` URL that will expire. Nothing from there is reusable.

### 8. No bibliography or reference workflow

The template ships an IEEE XSL style for Word's citation manager. There is
no `.bib`, no Zotero library, no reference list anywhere in the repository,
and chapters 2 and 3 are both bibliography-heavy. Pick the workflow before
writing chapter 2, not after — and note the template's own instruction
that Persian references are numbered with Persian digits and right-aligned
while English references use Latin digits and are left-aligned.

### 9. Unknown: the deadline

Nothing in the repository records a submission date. The build work still
needed for the primary deliverable (items 1, 2, 4, 5 above, plus the whole
AD phase) is measured in weeks. Fix the date and work backwards from it.

## Not a blocker

`lab.yaml` currently has `host_b` commented out and `vm_placements` reduced
to `dc01` — this is temporary while testing, not a regression. The
multi-host path is still what `ansible/playbooks/hosts.yml` configures.
It only matters in that **any multi-host measurement or screenshot for
chapters 4 and 5 has to be captured from a working two-host run**, so
re-enable it before evidence collection, not after.

## Suggested order of work

1. Fix the Related Work research: add academic sources, drop/demote the
   weak repos, settle the VXLAN question, write the honest
   "why not GOAD" position.
2. Decide the evaluation methodology (the section above) and start
   capturing evidence from every lab session.
3. Write chapter 2 — safest, never invalidated by later implementation.
4. Write chapter 3 on the corrected research.
5. Write chapter 1, framing now settled; add the ethics/scope statement.
6. Resolve blockers 1–5 as build work: base image strategy, Windows
   configuration management, attacker machine, planted data.
7. Draft chapter 4 for the implemented half; add the weakness-design
   section as the AD phase lands.
8. Run the engagement, collect evidence, then write chapter 5.
9. Chapter 6 last.

## Mechanics

`pandoc` and `libreoffice` are both installed. Two viable workflows:

- Write chapters as Markdown under `thesis/`, convert with
  `pandoc --reference-doc="CE Thesis Template.docx"` to inherit the styles.
  Good for version control and for iterating with an agent; RTL, Persian
  heading numbering, and mixed Persian/Latin runs all need checking after
  conversion — and mixed-direction text is exactly where automated
  conversion tends to break.
- Copy the template to a working `.docx` and write directly in Word or
  LibreOffice. Safest for formatting compliance, worst for diffs.

Given decision 2 (Persian body text with Latin technical terms inline), the
bidirectional-text handling is worth testing on a single paragraph before
committing to the Markdown-plus-pandoc route.

Either way: copy the template first, never edit the original, and commit
the working file so drafts are tracked.
