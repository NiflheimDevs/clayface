# Thesis writing plan — chapter readiness assessment

Assessment date: 2026-09-09. Updated 2026-09-12 after the supervisor's
email guidance. This is a planning document, not thesis text.

Legend: **[CONFIRMED]** = settled by the user or the supervisor.
**[PENDING]** = intended but not certain to land. **[RESOLVED]** = was a
blocker, no longer is. **[OPEN]** = outstanding *work* still to be done.
Every planning *question* has been answered (2026-09-12) — the remaining
`[OPEN]` entries are tasks, not decisions.

## Start here (fresh session)

This file is the complete handoff. Nothing needed from prior chat history.

- **Status: planning is finished; writing has not started.** No thesis
  document exists yet — only the blank template. The next action is to copy
  the template and write **Chapter 2**, then Chapter 3.
- The supervisor's chapter rulings are authoritative — read that section
  before writing anything.
- Every planning decision is recorded below with `[CONFIRMED]`. Do not
  re-litigate them; the user has answered all of them.
- Research for Chapter 3 is **done and verified**. See the source-material
  map at the end of this file for what each research file contains and which
  claims must not be repeated.
- Write in **Persian**, keeping English technical terms in English. Read the
  Mechanics section before choosing Markdown-vs-docx.

## Settled decisions

1. **[CONFIRMED] Framing: the approved proposal.** A *full red team
   simulation — ethical hacking of a corporate network*
   (`docs/redteam_corporate.md`). The "digital twin" reframe is not
   adopted.
2. **[CONFIRMED] Language: Persian**, English technical terms kept in
   English (Active Directory, Kerberos, Terraform, VXLAN…). The واژه‌نامه
   glossary carries the term list. Do not invent Persian equivalents for
   tool or protocol names.
3. **[CONFIRMED] Template**: MSc/PhD cover wording accepted as-is.
4. **[CONFIRMED] Project name: Clayface**, everywhere.
5. **[CONFIRMED] Evaluation quantitative metrics (Ch5):** four —
   reproducibility, VXLAN connectivity, resource footprint, effort-to-extend.
   See chapter 5.

## Supervisor's chapter rulings (email, 2026-09-12) — authoritative

These override the earlier structure of this plan.

- **[CONFIRMED] Ch2 (مفاهيم پایه):** concepts of cyber security, ethical
  hacking, penetration testing, and **modeling / simulation of cyber
  attacks**.
- **[CONFIRMED] Ch3 (كارهاي مرتبط):** introduce **commercial products AND
  open-source projects** briefly, then **compare them against each other**,
  e.g. in a table. This is a survey of *other people's* work.
- **[CONFIRMED] Ch4 + Ch5:** the student's own work and its results.
- **[CONFIRMED] The infrastructure / automation / reproducibility /
  multi-host "hard" work goes in Ch5** ("این موارد سختی را در فصل ۵ بیار").
  It is explicitly *not* excluded — it is a stated part of the results.
- **[CONFIRMED] The comparison of the student's own work vs others goes in
  Ch5**, not Ch3.
- **[CONFIRMED] Design approved:** "قسمت طراحی شبکه را خوب جلو برده‌ای."
- **[CONFIRMED] Attack bar is low:** "بتوانی روی آن سناریو حمله سوار کنی
  کافی هست" — mounting a working attack scenario on the built network is
  enough. GOAD-scale attack breadth is not required.

### What these rulings changed in this plan

- The self-assessment comparison table (Clayface vs others) **moves from
  Ch3 to Ch5.** The two ugliest former Ch3 blockers — the inflated
  self-capability table and the "why not GOAD?" problem — are now Ch5
  material, framed beside the student's own results and infra achievements.
- The "no other project uses a VXLAN overlay" differentiation claim is a
  self-vs-others point → **Ch5**, not Ch3.
- The completed infrastructure work now has a clear home (Ch5). The earlier
  worry that the done work was merely a "side objective" with no chapter is
  resolved by the supervisor directly.
- **Ch3 gains a new gap: commercial products**, which have not been
  researched at all yet (the prior GPT research was open-source only).
- Academic papers are **no longer a hard blocker for Ch3** — the
  supervisor's Ch3 bar is products + projects + comparison table. Papers
  still belong (in Ch2 for concepts; a cited taxonomy strengthens the Ch3
  comparison axes) but do not block the chapter.
- **Ch2 leans security.** Foundational infra concepts (virtualization, IaC,
  overlay networking) are demoted to one short, subordinate section — see
  the Ch2 entry and the open question about it.

## Ch3 scope — selected work to write about

- **[CONFIRMED] Open-source, top 3:** GOAD, Ludus, Splunk Attack Range.
  DetectionLab drops to a one-line mention (unmaintained since 2023-01-01).
  The three hobby repos and the non-existent DEFENSE.LAB are dropped
  entirely (see `ignoreme/gpt-research-corrections.md`).
- **[CONFIRMED] Closed-source, top 2: SimSpace + Cyberbit Range**
  (selected by the GPT research pass, `ignoreme/gpt-research-closedandopen.md`,
  verified 2026-09-12 with two citation fixes below). Chosen against stated
  criteria: environment-first (a built enterprise to operate inside) rather
  than control-validation, network/identity realism, multi-stage
  attack/defence, and on-prem/self-host capability. Both are genuine
  cyber-ranges (SimSpace documents on-prem/hybrid/SaaS; Cyberbit is a
  replicated enterprise range — DCs, DNS, mail, endpoints — though its
  current material leans cloud). The other surveyed names (Immersive,
  Cloud Range, RangeForce, Hack The Box, TryHackMe) are mentioned briefly,
  not given table rows.
- **Ch3 structure the research suggests (adopt it):** 3.x commercial
  cyber-ranges → 3.x BAS platforms (AttackIQ, Cymulate, SafeBreach, Picus),
  explicitly distinguished as *control-validation against production*, NOT
  ranges → 3.x open-source (GOAD, Ludus, Splunk Attack Range) → 3.x position
  of Clayface. The range-vs-BAS distinction is the chapter's cleanest
  taxonomy point.
- **Citation fixes before use:** (1) the Ludus 32 GB/200 GB baseline is
  correct but the research cites the archived `cisagov/Ludus` mirror — cite
  `docs.ludus.cloud` / `gitlab.com/badsectorlabs/ludus` instead; (2) confirm
  or drop the "Cyberbit acquired RangeForce, September 2025" date; (3) global
  RedClay → Clayface.
- Optionally add **Adaz** (christophetd, 429★, AD labs via Terraform +
  Ansible) as a fourth open-source mention — closest tooling match to
  Clayface — if space allows. Not one of the top 3.

## The template

`thesis/CE Thesis Template/CE Thesis Template.docx` — IUST Computer
Engineering, Persian/RTL, bundled IEEE reference style
(`Resources/IEEE Reference Template/IEEE.XSL`), Persian fonts. Chapter
skeleton:

| # | Chapter | Sections |
|---|---|---|
| 1 | مقدمه | 1-1 شرح مسأله · 1-2 انگيزه‌ها · 1-3 مفروضات (اختياري) · 1-4 اهداف · 1-5 ساختار |
| 2 | مفاهيم پایه | 2-1 مقدمه · 2-2 … · 2-3 نتيجه‌گيري |
| 3 | مروري بر كارهاي مرتبط | 3-1 مقدمه · 3-2 عنوان بخش · 3-3 مقايسه كارهاي مرتبط · 3-4 نتيجه‌گيري |
| 4 | روش/طرح پيشنهادي | 4-1 مقدمه · 4-2 متدولوژي · 4-3 كليات طرح · 4-4 جزئيات/اجزاء · 4-5 نتيجه‌گيري |
| 5 | ارزيابي | 5-1 مقدمه · 5-2 روش ارزيابي · 5-3 جزئيات · 5-4 نتايج · 5-5 نتيجه‌گيري |
| 6 | نتيجه‌گيري و کارهاي آينده | 6-1 نتيجه‌گيري · 6-2 کارهاي آينده |

Plus مراجع and واژه‌نامه. No working thesis file exists — only the blank
template. Copy it; never edit the original.

## What is actually built (verified 2026-09-12)

`lab.yaml` `vm_placements` has exactly **`dc01`** (role `dc`, host_a,
10.0.0.10) and **`client01`** (role `client`, host_a, 10.0.0.20). Alpine
VMs commented out; `host_b` commented out (temporary, testing).

Working and reproducible: KVM/QEMU/libvirt multi-host provisioning over
`qemu+ssh`; VXLAN + Linux-bridge L2 overlay; OPNsense edge (DHCP/DNS/NAT/
routing); Terraform driven by the single `lab.yaml` source of truth; two
Ansible dynamic inventories; the Terraform→Ansible `vms` output link;
`deploy.sh` end-to-end; a Windows Server **DC promoted to a forest root**
(`dc.yml`, via `microsoft.ad`); a **Windows 11 client domain-joined**
(`client.yml`); the pinned-MAC → DHCP reservation → DNS A-record chain;
documented base-image builds (`base-image/` with unattend files + README).

**Not built:** APP01 / DMZ Linux host / containers / PostgreSQL; IDP01 /
SSO; DMZ vs internal network segmentation (still one flat L2); an attacker
machine; planted misconfigurations; planted sensitive data; any executed
attack; any telemetry / SIEM.

Note: the supervisor email describes a DMZ Ubuntu host running
containerized apps — that is the **finalized target topology, not the
current state.** Do not describe it as built in Ch4/Ch5.

## Chapter-by-chapter readiness

### فصل 1 — مقدمه · writable now, ~85%, write after Ch3

Proposal §1–2 → 1-1 / 1-2. Attacker model (§5) → 1-3 — now **fixed as
external, no prior foothold** (see the attacker-machine decision under
blockers): entry is through the OPNsense edge into the DMZ. Proposal §3 →
1-4 (main + side objectives). Add the ethics / scope / isolation statement
here (see blockers).

### فصل 2 — مفاهيم پایه · writable now, can complete, ~75%

Safest chapter — pure published knowledge, never invalidated by later
build. Bounded by the supervisor's list:

- Cyber security fundamentals (CIA, threat/vulnerability/risk).
- Ethical hacking and penetration testing; the engagement lifecycle;
  MITRE ATT&CK as the phase/technique vocabulary.
- Active Directory as enterprise identity (Kerberos, LDAP, domain auth,
  SPNs, delegation) — the substrate of the attacks.
- Modeling and simulation of cyber attacks; cyber ranges and security
  testbeds (this is where the survey papers — Yamin 2020/2022, Ukwandu
  2020, Chouliaras 2021, Stamatopoulos 2024 — belong).
- **[CONFIRMED] Include one short subordinate section on enabling
  technologies:** virtualization/hypervisors, IaC, configuration management,
  network overlays — definitions only, pointed forward to Ch4. Target split
  ~85/15 security-to-infra (80/20 acceptable); 75/25 is too infra-heavy for a
  first draft. If the security material cannot fill ~85% when writing, use
  the best split reachable and state in the draft why the infra part could
  not be smaller.

Only real work: the bibliography workflow + prose. Best candidate for full
completion now.

### فصل 3 — مروري بر كارهاي مرتبط · startable now, mostly writable, ~70%

Now a survey of *others*, compared to *each other* — the self-comparison
left for Ch5, which removes the hardest parts.

- Writable now: profiles of the top-3 open-source projects (all verified in
  the corrections file) AND the top-2 commercial products (SimSpace,
  Cyberbit, verified in the closed/open research file), plus the BAS-vs-range
  survey text and the full comparison table — the research supplies every
  dimension.
- Blocking completion: only the citation cleanup above (Ludus source, the
  RangeForce date, RedClay→Clayface) and deciding whether to ground the
  comparison axes in a published cyber-range taxonomy (Yamin/Ukwandu) —
  recommended, strengthens the table, not strictly required.
- Honest positioning note (kept brief here, expanded in Ch5): these are
  large, mature projects; the contribution is not out-competing them on
  attack breadth but building the whole environment from physical hosts up
  as an IaC artifact and running an engagement against it.

### فصل 4 — روش/طرح پيشنهادي · writable now, substantial, not completable, ~60%

The student's own work — design and method. More is built than the earlier
estimate, so more is writable.

Writable now:
- 4-2 methodology: the red-team lifecycle (proposal §6) as the governing
  method, plus the supporting IaC + configuration-management approach and
  the single-source-of-truth rule (`Journal.md` has the real reasoning).
- 4-3 overall design: the **finalized, supervisor-approved topology** —
  three zones (DMZ / internal / VPN), OPNsense edge, the VXLAN overlay
  across hosts, core-plus-optional-modules, infrastructure separated from
  security state. Topology / auth-path / attack-path diagrams already
  exist (the three email images).
- 4-4 details, for what exists: `lab.yaml` schema; Terraform root module +
  per-host provider blocks; the `alpine` / `opnsense` / `domaincontroller`
  / `client` modules; the `edge.host` guard; the two dynamic inventories;
  the pinned-MAC → DHCP → DNS chain; WinRM/NTLM Windows connection; `dc.yml`
  forest promotion; `client.yml` domain join; `deploy.sh`.

Not writable yet (not built): APP01 + containers + PostgreSQL; IDP01/SSO;
DMZ/internal segmentation; the attacker machine; and the **design of the
planted weaknesses** (proposal says they are documented as part of the
design — a substantial section that does not exist yet: per weakness, what
it is, why a real org would have it, which attack it enables, the correct
config). `redclay.yaml` `attack_surface` / `technique_coverage` are the
raw list to build it from.

### فصل 5 — ارزيابي · split; infra half startable now, ~25%, not completable

The supervisor placed three things here: the student's results, the
"hard" infrastructure achievements, and the self-vs-others comparison.

Startable now (the infra half is done):
- The reproducibility / automation story; the multi-host + VXLAN
  achievement; the lightweight claim measured against comparable projects
  (Ludus documents 32 GB RAM per range; GOAD full = 5 Windows VMs).
- The self-vs-others comparison (moved here from Ch3), once Ch3's project
  set is fixed.

Blocked (not done): the **engagement results** — no attack has been run.
- **[CONFIRMED] Evaluation methodology.** Primary: execute a red-team
  scenario against the lab (the supervisor's low bar — one working scenario
  is enough), each step mapped to MITRE ATT&CK, with evidence. Secondary:
  four quantitative infra measures —
  1. **reproducibility** — N clean `deploy.sh` runs all converge; a second
     Ansible run reports 0 changed (idempotence);
  2. **VXLAN connectivity** — every VM reaches every other across both hosts
     (connectivity matrix);
  3. **resource footprint** — full-lab RAM/disk vs Ludus's documented
     32 GB / 200 GB baseline;
  4. **effort-to-extend** — lines of `lab.yaml` to add one VM vs one host.
- **Start capturing evidence now** from every lab session (terminal
  output, screenshots, timings). None of it is recoverable later.

### فصل 6 — نتيجه‌گيري و کارهاي آينده · ~40%

6-2 writable now (`redclay.yaml` `next_major_work`, optional modules,
deferred scenario engine, `docs/TODO.md`). 6-1 blocked on Ch5.

## Blockers

### Resolved since the last assessment
- Windows configuration management now exists (`dc.yml` + `client.yml`,
  `microsoft.ad` / `ansible.windows`, WinRM/NTLM).
- The client workstation and domain join are built.
- The base-image build is now documented and repeatable (`base-image/`
  unattend files + README) rather than an undocumented hand-made qcow2.

### Still open — build work
- **[RESOLVED] Client base = Microsoft ISO** (user-confirmed) — no
  community-image licensing or reproducibility problem. Note: a stray
  `tiny11.iso` still sits in `vm/images/` unused; delete it so it does not
  raise a provenance question later.
- **[OPEN] Packer / image automation.** Base images are still built and
  copied by hand; provisioning does not fetch or build them. Any
  reproducibility claim in Ch5 is only as strong as this step — automate or
  document precisely and state the limitation.
- **[CONFIRMED] Attacker machine = Kali VM on the OPNsense WAN side**
  (external-attacker model). It must break in through the edge, so Ch1's
  attacker model is "external, no prior foothold." **Coupling consequence:**
  an external break-in needs something reachable from the WAN to attack —
  i.e. the DMZ web app (APP01), exposed through an OPNsense port-forward.
  This puts APP01/DMZ on the critical path for the Ch5 engagement; it is no
  longer optional. Neither the attacker VM nor APP01 is built yet.
  *Known fallback lever if the timeline tightens:* placing Kali inside the
  internal L2 instead (assumed-insider foothold) removes the APP01
  dependency and still clears the supervisor's attack bar — but it changes
  Ch1's attacker model, so it is a deliberate decision for the user to make,
  not a silent substitution.
- **[OPEN] Planted data.** Nothing to exfiltrate exists; the proposal's
  impact analysis needs a subject (fake documents, DB contents, shares).
- **[PENDING] SSO / IDP01.** User intends to build it but is not sure it
  will make the timeline. Kept explicitly pending — do NOT hard-drop it to
  future-work yet. Design Ch4/Ch5 so it is additive: if built in time,
  include it; if not, it moves to Ch6 future-work in one line. Nothing in the
  attack bar depends on it.

### Still open — thesis production
- **[OPEN] Ethics / scope / isolation statement** (Ch1) does not exist.
  Concrete issue to resolve: OPNsense NATs outbound, so vulnerable Windows
  VMs have internet access — GOAD's own README warns against this. Decide
  and describe the isolation posture.
- **[OPEN] Figures.** Five diagrams already exist as **mermaid blocks inside
  the Obsidian vault** — see the figure inventory in the source-material map.
  Nothing is lost and nothing needs re-drawing from scratch. Remaining work:
  (1) three of them carry factual errors about what is built and must be
  re-styled before printing; (2) export mermaid → PNG/SVG for the `.docx`;
  (3) split the combined attack-path diagram into one figure per chain for
  Ch5. Keep the four graph types (`redclay.yaml` `graphs`) separate, not
  merged.
- **[OPEN] Bibliography workflow.** Template ships IEEE.XSL for Word; no
  `.bib` / library exists. Pick before writing Ch2. Persian refs = Persian
  digits, right-aligned; English refs = Latin digits, left-aligned.
- **[RESOLVED] Deadline.** User is managing the schedule; not a planning
  input here. Weekly supervisor reports remain the cadence.

## Not a blocker
`host_b` commented out and placements reduced to `dc01` + `client01` is
temporary testing. The multi-host path is what `hosts.yml` configures.
It only matters that **multi-host evidence for Ch4/Ch5 must be captured
from a working two-host run** — re-enable before evidence collection.

## Order of work
1. Ch3 subjects locked (GOAD/Ludus/Splunk + SimSpace/Cyberbit). Do the
   citation cleanup, then Ch3 is writable end-to-end.
2. Write Ch2 (completable now) — security-heavy plus the ~85/15 infra
   section.
3. Write Ch3 open-source profiles + the comparison table.
4. Write Ch1 (framing settled) + the ethics/scope statement.
5. Draft Ch4's built half + the finalized-topology design.
6. Capture evidence for the four fixed metrics from every session.
7. Build attacker machine + planted weaknesses + planted data; run the
   scenario; write Ch5 (infra half can be written before the attack half).
8. Ch6 last.

## Mechanics
`pandoc` and `libreoffice` both installed. Either write chapters as
Markdown and convert with
`pandoc --reference-doc="CE Thesis Template.docx"`, or copy the template to
a working `.docx` and write directly. Persian body text with inline Latin
technical terms is exactly where mixed-direction conversion breaks — test
one paragraph before committing to the Markdown route. Copy the template
first; commit the working file.

`python-docx` is NOT installed (pandoc/libreoffice are the available path).

## Source material map

Where every input lives, and what it is worth.

### Thesis inputs
- `thesis/CE Thesis Template/CE Thesis Template.docx` — the blank IUST CE
  template. **Copy it; never edit in place.** Ships
  `Resources/IEEE Reference Template/IEEE.XSL` for Word citations.
- `docs/redteam_corporate.md` — the approved proposal. Source of truth for
  scope, goals, attacker model and the red-team lifecycle. Feeds Ch1 and
  Ch4-2. Note: IaC is written there as a *side* objective.
- `docs/redclay.yaml` — knowledge base, **intent only**, not read by any
  tooling. `attack_surface` / `technique_coverage` are the raw list for the
  planted-weakness design section; `next_major_work` feeds Ch6-2;
  `scenarios.status` is `conceptual_only`.
- `lab.yaml` — the single source of truth for what actually exists. Check it
  before claiming anything is built.
- `red-clay/Report/Journal.md` — the real reasoning behind design decisions
  (why single-source-of-truth, why the MAC pinning). Best raw material for
  Ch4-2's methodology narrative.
- `red-clay/Architecture/Infrastructure.md`, `.../Identity.md` — vault notes
  on the two design halves. **Both contain mermaid figures** (see the figure
  inventory); `Infrastructure.md` also has the prose for the "placement is a
  scheduling decision, not a network one" argument and the edge-pinning note
  — both good Ch4-3 material.
- `red-clay/Attacker/Attack Paths.md`, `Techniques.md`, `Active Directory.md`,
  `RedTeam.md` — attack-side notes for Ch5's scenario and Ch2's lifecycle
  section.
- `docs/client01-plan.md`, `docs/TODO.md` — remaining build work.
- `red-clay/Overview.md` — the vault's own top-level description of the
  target system, and holder of the primary topology figure.
- `base-image/README.md` — the reproducible base-image procedure plus its
  sysprep/AppX failure modes. Cite for Ch5's reproducibility limits.
- `CLAUDE.md` — authoritative technical detail (data-flow rule, DNS
  topology, WinRM/NTLM troubleshooting, known pain points). The most
  accurate single description of the built system; excellent Ch4-4 source.

### Chapter 3 research (all in `ignoreme/`, which is **gitignored**)
- `ignoreme/gpt-research.md` (1053 lines) — first GPT pass, open-source
  only. Usable but contains errors; **never quote it without checking the
  corrections file.** All its screenshots are expiring `images.openai.com`
  URLs and are unusable.
- `ignoreme/gpt-research-corrections.md` — my verification of the above.
  Lists what is correct (safe to use), what is wrong (DEFENSE.LAB does not
  exist; the AD-Lab URL is dead and its citation used a Fastly mirror; three
  1–2★ hobby repos were over-ranked; the self-capability table claimed
  unbuilt features), and what was missing (7 academic sources with verified
  DOIs, Adaz, BadBlood, a cited taxonomy for the comparison axes).
- `ignoreme/gpt-research-closedandopen.md` (368 lines) — second GPT pass,
  commercial products + the five-way comparison. Good quality: vendor claims
  attributed as claims, no invented pricing, real URLs. **This is the main
  Chapter 3 source.** Its ready-to-use assets: the selection criteria, the
  five-way comparison matrix, the cyber-range-vs-BAS distinction, and the
  Ch3 section structure in its §9.
- **Move these to `docs/` if you want them kept with the thesis** — they are
  currently outside version control.

### Verification caveats carried forward
- The Ludus **32 GB RAM / 200 GB SSD** figure is correct, but the research
  cites the archived `cisagov/Ludus` mirror (archived 2024-03-07, 2 commits,
  redirects to GitLab). Cite `docs.ludus.cloud` or
  `gitlab.com/badsectorlabs/ludus`. Ludus's GitHub repo is itself labelled
  "[GITLAB MIRROR]".
- "Cyberbit acquired RangeForce in **September 2025**" — the acquisition is
  supported (Cyberbit's own page is titled "About RangeForce, a Cyberbit
  company") but **the date is unverified**. Write it without the month, or
  verify first. `cyberbit.com` returned HTTP 403 to automated fetches.
- The VXLAN-uniqueness claim ("no comparable project uses a VXLAN overlay")
  is still **unverified** — Proxmox SDN has a VXLAN zone type, and Ludus
  runs on Proxmox. The narrower honest form survives: *Clayface builds the
  L2 overlay itself on plain Arch Linux KVM hosts with no Proxmox cluster
  layer underneath.* This claim belongs in Ch5, not Ch3.
- Both research files say "RedClay". The project is **Clayface**. Global
  rename when importing any text.
- WebSearch was malfunctioning during verification (returning fabricated
  "no search tool available" responses). Live claims were checked by direct
  fetch instead. If a fresh session needs to re-verify, prefer WebFetch.

### Optional follow-up to the GPT researcher
Not required — the two defects are small enough to fix while writing. Send
this only if you want the model to clean its own citations:

> Two corrections. (1) You cited github.com/cisagov/Ludus for the 32 GB /
> 200 GB baseline — that repo is archived and redirects to
> gitlab.com/badsectorlabs/ludus. Re-cite the current Ludus install docs
> (docs.ludus.cloud) or the GitLab repo, and confirm the 32 GB RAM /
> 200 GB SSD figure from that live source. (2) Give me the exact source and
> date for the claim that Cyberbit acquired RangeForce — a primary press
> release or dated announcement, not a marketing page. If you cannot find a
> dated source, say so and I will drop the date.

### Figures — inventory (all are mermaid blocks in the vault)

The diagrams are **not** image files and not email attachments; they are
```mermaid fences scattered across the Obsidian vault. Five exist. Each is
already styled with a `built` / `planned` class pair (green solid vs grey
dashed), which is exactly the honesty device Ch4 needs — keep that
convention.

| Source | Diagram | Use in thesis |
|---|---|---|
| `red-clay/Overview.md:22` | Full zone topology: External (KALI01, Internet, VPN pool 10.0.100.0/24) → OPNsense boundary → DMZ 10.0.10.0/24 (APP01) / user zone 10.0.20.0/24 (CLIENT01) / server-identity zone 10.0.30.0/24 (DC01, IDP01, optional modules) | **Ch4-3** — the primary topology figure |
| `red-clay/Architecture/Infrastructure.md:14` | Physical layer: host_a / host_b, VM placement, VXLAN 100 over wlan0, bridge vm-br0 | **Ch4-3** and **Ch5** (multi-host result) |
| `red-clay/Architecture/Infrastructure.md:43` | IaC toolchain pipeline: `lab.yaml` → (yamldecode) Terraform + (dynamic inventory) Ansible, Packer → base images, `vms` output → inventory, both → running VMs | **Ch4-2 / 4-4** — this *is* the tooling-pipeline figure; it already exists |
| `red-clay/Architecture/Identity.md:11` | Traditional AD identity (DC01 ← Kerberos/NTLM ← CLIENT01) vs modern app identity (portal → OIDC/SAML → IDP01), with planned federation | **Ch4-4** identity design; also the figure that visualises the `[PENDING]` IDP01 decision |
| `red-clay/Attacker/Attack Paths.md:8` | Two numbered attack chains: (1) stolen creds → VPN tunnel → user zone → CLIENT01 → AD enum/privesc → DC01; (2) public web/API exploit → APP01 → SQLi/creds → PostgreSQL, plus service identity → DC01 | **Ch5** engagement — split into one figure per chain |

**Accuracy fixes required before any of these is printed** — three diagrams
overstate the built state, which is exactly the error the corrections file
flagged in the GPT research:

- `Infrastructure.md:14` styles **client01 and app01 as built on host_b**.
  Reality (`lab.yaml`): `dc01` *and* `client01` are both on **host_a**,
  `host_b` is commented out, and **app01 does not exist**. Re-place and
  re-style before use, or the figure contradicts the text.
- `Overview.md:22` shows **three separate zone subnets** (10.0.10/20/30) and
  a **VPN pool**. Reality is one flat L2 on 10.0.0.0/24 and no VPN. The
  figure is the *target* topology — label it as such (it is the
  supervisor-approved design, so this is legitimate in Ch4-3), and do not
  reuse it as evidence in Ch5.
- `Attack Paths.md:8` chain 1 begins with **stolen VPN credentials**, which
  is a different entry premise from the confirmed attacker model
  (external, no prior foothold, in through the edge). **Chain 2 —
  public web/API exploit against APP01 — is the one matching the confirmed
  Q4 decision**, and is therefore the primary Ch5 scenario. Either reframe
  chain 1's premise or present it as an alternative path not executed.

**Export mechanics:** pandoc will not render mermaid into `.docx` — it
passes the fence through as a code block. `mmdc` (mermaid-cli) is **not
installed**; `node`/`npx` are, so
`npx -y @mermaid-js/mermaid-cli -i in.mmd -o out.svg` is the path, or export
from Obsidian directly. Prefer SVG for the docx. Persian captions with Latin
node labels are the usual RTL trouble spot — check one figure end-to-end
early. Note the diagrams' node labels are in English; that is consistent with
the confirmed language rule (technical terms stay English), so they need no
translation, only Persian captions.
