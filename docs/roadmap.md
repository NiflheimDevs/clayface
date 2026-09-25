# Clayface — roadmap to a finished thesis

Ordered list of the phases still to come, from the current state to a
defended thesis. Written 2026-09-24 after reviewing the uncommitted
application-tier changes.

## Scope decisions locked (2026-09-24)

The maximal path was chosen on all four open forks:

- **IDP01 / SSO:** build it (not deferred to future work).
- **Attacker model:** external, entering through the OPNsense WAN via the
  APP01 port-forward.
- **Network segmentation:** build the DMZ (a real boundary between APP01 and
  the internal zone), not just document its absence.
- **Attack scenario depth:** the chain runs through to Domain Admin, which
  requires the AD privilege-escalation weaknesses to be planted and
  exercised, not left as the `false` toggles they are today.

## Two tracks that interleave

- **Build / lab track (Phases 0–7)** — bring the lab up, finish building it,
  run the engagement.
- **Thesis track (Phases A–G)** — the writing.

They overlap deliberately. Chapters 2 and 3 (Phases B and C) have no lab
dependency and should be written while the build proceeds. Chapter 5 (Phase
F) is the exception: it cannot be written until the engagement in Phase 6
has produced evidence, so Phases 1→6 sit on the critical path to the
heaviest results chapter.

**Standing rule: capture evidence from Phase 1 onward and never stop.**
Terminal output, screenshots, timings, and packet captures are not
recoverable after the fact.

---

## Build / lab track

### Phase 0 — Baseline the working tree

- Review the uncommitted changes — done (APP01 tier, the `svc-app-portal`
  service account, the WAN port-forward, the docs and vault rewire).
- Decide the `app/vendor/` policy — commit it (~170k lines of vendored Go
  dependencies). Reason: an offline `go build -mod=vendor` is the
  self-contained-reproducibility claim Chapter 5 leans on. The tree was
  verified clean offline (`gofmt`, `go vet`, `go build`).
- Commit in logical chunks: (a) the Terraform `app` module + the `lab.yaml`
  placement, (b) the `app/` portal source + vendor, (c)
  `app.yml`/`app_validate.yml` + the opnsense port-forward, (d) the `ad.yml`
  service account, (e) the docs, (f) the vault notes.
- Tidy: delete the stray, unused `vm/images/tiny11.iso` before it raises a
  provenance question in Chapter 5.

### Phase 1 — Bring the lab up and verify what is already written

Everything built so far is unproven — the lab was down when it was written.
The authoritative worklist is `docs/app01-verification-pending.md`.

- Re-enable `dc01` and `client01` in `lab.yaml` — they are currently
  commented out, so the lab as committed builds only `app01`. The domain
  needs both back.
- Host prep: `systemctl enable --now sshd docker`; recreate the segment
  bridges (`playbooks/hosts.yml`); point the control node's resolver at the
  lab DNS (`resolvectl dns vm-lan0 10.0.0.1`, and the same for `vm-dmz0` with
  `10.0.10.1`). Note that `hosts.yml` sets this at runtime only — it is lost
  on reboot, so re-run hosts.yml after one. It is not made persistent with
  `nmcli`: the bridges are created with `ip link add ... type bridge`, so
  NetworkManager reports them `connected (externally)` and does not manage
  them. `hosts.yml` uses `resolvectl` for that reason.
- APP01 base-image hazards, all before anything boots: back up
  `ubuntu24.04-base`; `chown` it to `libvirt-qemu`; and
  `virsh undefine ubuntu-base --nvram`. This last one is the highest-risk
  item — the live builder domain points straight at the read-only base
  image, and booting it corrupts every overlay stacked on top, silently.
- Run `terraform plan` (the first time ever for the `app` module), read it
  carefully — it must not propose replacing `dc01` or `client01` — then
  `apply`.
- Verify the guest contract: SSH to `clayface@10.0.10.10`; confirm the docker
  CLI, the compose v2 plugin, docker-group membership, and sudo are all
  present in the image.
- Run `app.yml` (the first Docker build ever), then `app_validate.yml` — it
  must pass green with all eleven weakness toggles `true`.
- Run `dc.yml` and `client.yml` to re-confirm the forest promotion and the
  domain join still work.
- Set `LAB_SVC_APP_PASS` before the first `ad.yml` run. `ad.yml` uses
  `update_password: on_create`, so it will not reset the account later, but
  the database copy is rewritten every boot — a mismatch makes the pivot
  silently useless.
- Run `ad.yml` (never executed): OUs, users, groups, and `svc-app-portal`.
- Functional checks: TLS answers on 443; 80 redirects to 443; only 80 and
  443 are open (`nmap`); the seed produced 7 tables, 7 users, 3 API-key
  rows, and 5 documents.
- Outcome: a known-good single-host lab with APP01, DC01, and CLIENT01 all
  real. Strike items off the verification-pending file with dates, then
  retire it.

### Phase 2 — Multi-host and DMZ segmentation

**The DMZ half is done** (2026-09-25, `docs/network-plan.md` Part B); the
design is `docs/network-design.md`. What landed: the `dmz` segment
(`10.0.10.0/24`, bridge `vm-dmz0`, VXLAN 101), a third NIC on OPNsense, APP01
moved to `10.0.10.10`, and `ansible/playbooks/opnsense_dmz.yml` building the
boundary — five filter rules from `networks.dmz.allow` in lab.yaml, two logged
denies, a dnsmasq range so APP01 gets a lease at all, and Unbound on the new
leg. A pivot from APP01 to dc01 now crosses a filtered, logged boundary, which
closes the `docs/TODO.md` caveat that a DMZ compromise used to be equivalent to
an internal foothold.

Outstanding in this phase:

- Re-enable `host_b` in `lab.yaml` plus its provider block and per-module
  blocks in `terraform/main.tf`, and bring up the VXLAN single-peer mesh. The
  DMZ's VXLAN 101 is declared and will follow automatically — the segment/VNI
  pairing is data now, not a second code path.
- Record the connectivity matrix — every VM reaching every other VM across
  both hosts, **and per zone**, since the DMZ makes "reachable" a two-part
  answer. This is a Chapter 5 metric; capture it as you verify it.
- Split the internal zone itself (`10.0.20.0/24` / `10.0.30.0/24`, the
  user/server split). DC01, CLIENT01 and IDP01 currently share one flat
  segment, which is honest but is not the topology the proposal draws.

### Phase 3 — External attacker entry (WAN)

- Verify the WAN port-forward — the `d_nat` API calls in `opnsense.yml` were
  written from documentation, not against a live system. Fix the rule-body
  field names against a real rule (read `/conf/config.xml` for the names
  OPNsense itself writes); confirm the rule is idempotent (it is matched by
  its `descr` marker); confirm it does nothing when `LAB_WAN_EXPOSE_APP` is
  unset.
- Build the Kali attacker VM on the WAN side. Prefer a Terraform module over
  a hand-made VM, for consistency and to keep the reproducibility metric
  honest. Place it on the OPNsense WAN leg, outside `vm-lan0`.
- Prove the entry: Kali reaches APP01 through the WAN port-forward, and
  nothing else internal is reachable from the WAN.

### Phase 4 — Plant and wire the remaining weaknesses (AD privesc to DA)

**Done** (2026-09-25). All seven `ad.weaknesses` toggles in `lab.yaml` are wired
and asserted in both postures. The design is `docs/ad-identity-design.md`
section 13 — its normal-to-vulnerable table is the specification, and the
toggles now implement it — and the Chapter 4 section is
`docs/planted-weaknesses.md` (what each weakness is, the attack it enables, its
detection signal, and which route in the chain leans on it). The seven are split
across **three** playbooks, not the one this bullet named: `ad.yml` owns W1, W3,
W5, W6 and W7, `client.yml` owns W2 (a local group membership, so it has no
domain-side owner) and `ad_gpo.yml` owns W4. Each play asserts the toggles it
implements and `ad_validate.yml` asserts all seven, in both postures: with every
toggle on the validator is green, with every toggle off it is green again, so
"off" is a real configuration change that restores the baseline rather than a
skipped task. On is the lab's default posture, matching `app.weaknesses`. The
exfiltration subject exists too: `data.shares` in `lab.yaml` (the Finance share
on DC01, the Projects share on CLIENT01), planted by `lab_data.yml`.

Outstanding in this phase:

- W3's second gate is deliberately not opened. W3 is the LAPS read right on
  `OU=Workstations`, and the LAPS GPO has a second gate: the decryption
  principal, which stays `GG-IT-Admins` rather than being widened to a broader
  group. Opening one gate is enough for the finding, and opening both would make
  it indistinguishable from a broken LAPS deployment.
- W5 is wired and asserted but is not load-bearing. Unconstrained delegation on
  `svc-idp-ldap` is a real finding with a real detection, but the Chapter 5
  chain reaches Domain Admin through W1 and W6, from a LAN foothold; W5 is
  offered as an adjacent path, and `docs/planted-weaknesses.md` says so rather
  than implying the chain needs it.
- Two documentation sharpenings deferred from the Phase 4 review, both in
  `docs/planted-weaknesses.md`: W2's row and the local-privilege paragraph say
  local admin is what makes the planted shares reachable, when the shipped read
  grant is to `GG-Employees` (so the accurate statement is that DC01's Finance
  share is network-readable by any employee credential and CLIENT01's Projects
  share is a local-session target); and W3's read right is redundant with W2 for
  reaching local admin, its independent value being the `lapsadmin` credential
  that survives W2 being turned off. Wording only — the code and the ACLs are
  what they should be.

### Phase 5 — IDP01 / SSO build

- Decide the IdP product (Keycloak vs Authentik) at the start of the phase.
  The AD-side contract already exists in `docs/ad-identity-design.md`: AD is
  the upstream user store via `svc-idp-ldap`, with no protocol-level
  federation.
- Provision IDP01: a Terraform module + a `lab.yaml` placement with an `ip:`;
  a playbook to deploy the IdP; the LDAP bind to AD.
- Integrate the portal → OIDC/SAML → IDP01 → AD. The portal today
  authenticates against its own `users` table (out of scope in
  `app01-design.md`); this adds the federated path.
- Update the `Identity.md` figure to show federation as built, not planned.

### Phase 6 — Run the engagement and capture all evidence

- Execute the attack chain to Domain Admin: Kali → WAN → APP01 (one SQLi
  toggle) → dump the database → recover the bind credential from `api_keys`
  and bind LDAP on `dc01` → **a LAN foothold** → enumerate AD → escalate via
  the Phase 4 weaknesses → Domain Admin → exfiltrate the planted data.
- Plan and document the LAN foothold the escalation runs from. The DMZ
  boundary permits DMZ → `dc01` tcp 389/636, DNS and DHCP, and nothing else, so
  no rule carries the attack out of the DMZ into the LAN; `app01` has one NIC
  and no second path. Phase 4 stated this gap and deliberately did not close
  it, because closing it means widening the boundary — a design change, not a
  fix. The intended route is a compromised employee workstation, which is what
  W2 and W3 supply, and it needs to be a planned, evidenced step like any
  other rather than an implied one.
- Map every step to a MITRE ATT&CK technique ID.
- Capture the four infrastructure metrics:
  1. reproducibility — N clean `deploy.sh` runs all converge, and a second
     Ansible run reports 0 changed (idempotence);
  2. VXLAN connectivity — the full cross-host matrix;
  3. resource footprint — full-lab RAM/disk against Ludus's documented
     32 GB / 200 GB baseline;
  4. effort-to-extend — lines of `lab.yaml` to add one VM versus one host.
- Baseline versus weakness — re-run with the toggles `false` to demonstrate
  the safe code paths. This is the comparison artifact, not a demo of an off
  switch.

### Phase 7 — Reproducibility hardening (Packer)

- Automate the base-image builds with Packer, or document the manual build
  precisely and state the limitation. The Chapter 5 reproducibility claim is
  only as strong as this step: the minimum is a precise description, and
  automation strengthens the claim. Can run in parallel with the writing.

---

## Thesis track

Language: Persian, with English technical terms kept in English. Template:
`thesis/CE Thesis Template/CE Thesis Template.docx`, copied to a working
file — never edit the original.

### Phase A — Production setup

- Fill the working `thesis/Clayface-Thesis.docx` cover with the real title,
  student name, supervisor, and year (`1394` is a placeholder).
- Decide the authoring route — Markdown + pandoc versus editing the `.docx`
  directly. Test one mixed Persian/Latin RTL paragraph first; mixed-direction
  text is exactly where the pandoc conversion breaks.
- Bibliography workflow — pick a `.bib`/Zotero library plus the bundled
  `IEEE.XSL`. Persian references use Persian digits, right-aligned; English
  references use Latin digits, left-aligned.
- Move `ignoreme/*.md` (the Chapter 3 research) into `docs/` if it should be
  kept under version control — it is currently gitignored.

### Phase B — Chapter 2, مفاهيم پایه (completable now, no lab dependency)

- Security fundamentals (CIA; threat / vulnerability / risk); ethical hacking
  and penetration testing; the engagement lifecycle; MITRE ATT&CK as the
  technique vocabulary; Active Directory as enterprise identity (Kerberos,
  LDAP, SPNs, delegation); modeling and simulation of cyber attacks; cyber
  ranges and testbeds (the survey papers — Yamin 2020/2022, Ukwandu 2020,
  Chouliaras 2021, Stamatopoulos 2024).
- One short (~15%) subordinate section on enabling technologies —
  virtualization, IaC, configuration management, network overlays. Target an
  85/15 security-to-infra split.

### Phase C — Chapter 3, مروري بر كارهاي مرتبط (a survey of others)

- Citation cleanup first: cite Ludus from `docs.ludus.cloud` or the GitLab
  repo, not the archived mirror; verify or drop the RangeForce acquisition
  month; rename RedClay → Clayface everywhere the research is imported.
- Profiles: the open-source top three (GOAD, Ludus, Splunk Attack Range),
  plus a one-line DetectionLab mention and optionally Adaz; the commercial
  top two (SimSpace, Cyberbit); the BAS-versus-range distinction (AttackIQ,
  Cymulate, SafeBreach, Picus).
- A comparison table, with axes grounded in a published cyber-range taxonomy
  (Yamin / Ukwandu). The self-versus-others comparison stays in Chapter 5.

### Phase D — Chapter 1, مقدمه (after Chapter 3)

- 1-1 / 1-2 from the proposal; 1-3 the attacker model, now fixed as external
  with no prior foothold (matching the built entry point); 1-4 objectives;
  1-5 structure.
- The ethics / scope / isolation statement. Resolve the concrete issue that
  OPNsense NATs outbound, so the vulnerable Windows VMs have internet access
  (GOAD's own README warns against this) — decide and describe the isolation
  posture.

### Phase E — Chapter 4, روش/طرح پيشنهادي (built half writable now)

- 4-2 methodology — the red-team lifecycle as the governing method, plus the
  IaC and configuration-management approach and the single-source-of-truth
  rule (`red-clay/Report/Journal.md` holds the real reasoning).
- 4-3 overall design — the finalized, supervisor-approved topology, now
  actually built with a real DMZ.
- 4-4 details — the `lab.yaml` schema; the Terraform root module and the
  `alpine`/`opnsense`/`domaincontroller`/`client`/`app`/`idp` modules; the
  `edge.host` guard; the two dynamic inventories; the pinned-MAC → DHCP → DNS
  chain; the WinRM/NTLM Windows connection; `dc.yml`, `client.yml`, `ad.yml`,
  `app.yml`, and `deploy.sh`; APP01's container topology, weakness-toggle
  table, and the `svc-app-portal` contract; and the planted-weakness design
  section produced in Phase 4.
- Figures — fix the three inaccurate mermaid diagrams
  (`Infrastructure.md:14`, `Overview.md:22`, `Attack Paths.md:8`, which
  overstate the built state), export mermaid → SVG
  (`npx @mermaid-js/mermaid-cli`), and add Persian captions with English node
  labels.

### Phase F — Chapter 5, ارزيابي (needs Phase 6 evidence)

- 5-2 method — the executed scenario plus the four infrastructure metrics.
- 5-3 / 5-4 — the engagement walkthrough to Domain Admin (per-MITRE-step,
  with evidence); the metric results; the baseline-versus-weakness
  comparison; the self-versus-others comparison (moved here from Chapter 3);
  and the multi-host + VXLAN result with the lightweight-versus-Ludus claim
  in its honest form (the L2 overlay is self-built on plain Arch KVM hosts
  with no Proxmox layer underneath).
- 5-5 conclusion.

### Phase G — Chapter 6, front and back matter, defense

- 6-1 conclusion (after Chapter 5); 6-2 future work (IDP extensions if only
  partially built, the deferred scenario engine, `docs/TODO.md`, and
  `redclay.yaml` `next_major_work`).
- The glossary (واژه‌نامه); finalize the references; place the figures; a full
  RTL/Latin pass.
- Assemble, proofread, and export the final PDF. Keep the weekly supervisor
  reports throughout, and prepare for the defense.
