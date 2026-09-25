# Clayface — Application Tier Design (APP01)

Design for the deliberately vulnerable application hosts of the Clayface lab:
what runs on `app01`, why it is shaped that way, which weaknesses are planted
and how each one is switched, and how the tier connects to the AD identity
layer that already exists.

Status: **built.** `terraform/modules/app` provisions the VM, `lab.yaml` places
it, `ansible/playbooks/app.yml` deploys the stack and
`ansible/playbooks/app_validate.yml` asserts the result. Nothing in this
document has been exercised against a running lab yet — see section 16 and
`docs/app01-verification-pending.md` for the exact list of what is untested.

---

## 1. Scope and non-goals

**In scope.** The `app01` VM: its Terraform module and placement; the
containerized stack that runs on it (nginx, the Go portal, PostgreSQL); the
image build and shipping pipeline; the weakness-toggle model; the portal's
configuration surface; the `svc-app-portal` service identity; and the
validation playbook.

**Out of scope.** IDP01 and SSO. The portal authenticates against its own
`users` table; it does not federate, and no OpenID/OAuth/SAML surface exists.
The attacker machine. Network segmentation (section 3). Anything on the
Windows side — the portal has no dependency on `dc01` at deploy time.

**Explicitly rejected.**

- **Domain-joining `app01`.** The tier's identity decision was "service account
  in configuration and in the database, no domain join". A joined Linux host
  would have been a second, different identity story layered on the first, and
  it would have made the portal's own login weakness compete with Kerberos for
  the reader's attention. The pivot this tier is for — recover a credential from
  the database, then use it *against* the domain — only reads as a pivot if the
  two halves are separate systems.
- **Separate services for the API and the admin panel.** One binary, three
  route groups. See section 4.
- **A Redis container.** `docs/redclay.yaml`'s `example_services` list includes
  one; it is intent, not this cycle. Nothing in the planted weakness set needs a
  cache, and a service that exists only to be enumerated is weight without a
  lesson.

---

## 2. Baseline facts this design respects

- **The base image is `/var/lib/libvirt/images/ubuntu24.04-base`.** It has no
  `.qcow2` extension, so its format can never be inferred from the filename and
  the module declares `qcow2` explicitly. It is a **BIOS** install: the module
  is `type_machine = "q35"` with no UEFI loader, matching the domain controller
  module and unlike the client module. Booting it under UEFI finds no ESP.
- **It is a hand-built image, not something this repository produces.** Docker
  and the docker compose v2 plugin are expected to be present in it. `app.yml`
  asserts both and installs nothing — a gap fails with a message naming the gap
  rather than being repaired (section 13.2).
- **Its virtual size is 25 GiB.** The overlay declares 40 GiB, which satisfies
  the house rule that an overlay's capacity must be at least its backing file's
  virtual size. The extra room is for the image tarball, the container layers
  and the PostgreSQL data directory.
- **DNS domain `clayface`**, served by OPNsense. `app01.clayface` resolves from
  the guest's first boot because the VM is pinned (section 3.2).
- **It sits in the DMZ.** `10.0.10.10` on `vm-dmz0`, behind a default-deny
  boundary on OPNsense that allows exactly five things (LDAP to dc01, DNS,
  DHCP, and internal users in on 443) and logs everything else. This is the
  most consequential fact in the whole document and section 3 is about it.
  The boundary is specified in `docs/network-design.md`.

---

## 3. Where APP01 sits in the lab

### 3.1 APP01 is a DMZ host

`docs/redclay.yaml` and `red-clay/Overview.md` describe a four-zone target
topology with APP01 in a DMZ at `10.0.10.0/24`. As of 2026-09-25 **that is
what runs**, for this one zone: APP01 is at `10.0.10.10` on `vm-dmz0`, with
`dc01` (`10.0.0.10`) and `client01` (`10.0.0.20`) on the internal LAN
(`vm-lan0`), and OPNsense routing and filtering between them.

The boundary is **default-deny**, and the allowances are the deliberate part
(see `docs/network-design.md` for the full policy):

| Direction | Allowed |
|---|---|
| DMZ → dc01 | tcp 389, tcp 636 — the documented Chapter 5 pivot |
| DMZ → OPNsense | tcp/udp 53, udp 67 |
| LAN → APP01 | tcp 443 — internal users browsing the portal |
| everything else | denied, the DMZ→LAN direction logged |

So the honest consequence is now the *opposite* of what this section used to
say: **APP01 being compromised is a genuinely weaker position than an internal
host being compromised**, and crossing from it to the domain requires the
deliberate LDAP allowance. A Chapter 5 narrative may now treat "the DMZ web app
fell" as a distinct step from "an internal host fell" — and the fact that the
pivot exists only because of one narrow, logged rule is itself part of the
story rather than a gap in it.

The segmentation work is closed in `docs/TODO.md`, which keeps the history:
this tier was deliberately built *before* segmentation, and the paragraph
above is what that debt cost while it was outstanding.

**What the boundary is not.** The hypervisor is multi-homed into every zone by
construction — it holds a bridge for each segment and the control node has an
address on each — so a compromised *hypervisor* is outside this model
entirely. The boundary constrains APP01, the guest, and nothing else. That is
a provisioning fact, not a claim about the design, and `docs/network-design.md`
states it at length so no reader mistakes the diagram for a hypervisor
boundary.

### 3.2 How it gets its address

The same pinned-MAC chain as `dc01` and `client01`, and for the same reason:
dnsmasq's `regdhcp=1` registers whatever hostname the guest supplies, and a
hostname is not something Terraform knows before the VM exists.

```
terraform:  mac = 52:54:00:$(sha256("app01")[0:6])   ->  52:54:00:54:e9:bb
            (+ the guest cannot change a MAC)

lab.yaml:   app01: { module: app, host: host_a, network: dmz, ip: 10.0.10.10 }

opnsense.yml:  one dnsmasq host row = DHCP reservation + DNS A record
```

An Ubuntu guest does keep the hostname baked into its image, so unlike the
Windows VMs this chain is not strictly required for name resolution. It is done
anyway: the lab has one rule for every VM, and a reservation is what gives
`app01` a static address rather than a pool lease.

---

## 4. Container topology

```
                    host_a · KVM
  ┌──────────────────────────────────────────────────────────────┐
  │  app01 · 10.0.10.10 · Ubuntu 24.04 · BIOS · 4 GiB · 2 vCPU   │
  │                                                              │
  │   nginx:1.27-alpine        ← the only published ports        │
  │     :443 TLS (self-signed) ──┐   :80 redirects to :443       │
  │                              │                               │
  │   clayface/portal:lab    ◄───┘   Go binary, :8080            │
  │     /portal/*  /api/*  /admin/*  (one binary, 3 groups)      │
  │                              │                               │
  │   postgres:16-alpine     ◄───┘   :5432, compose-internal     │
  │     clayface · users, orders, invoices, documents, api_keys  │
  └──────────────────────────────────────────────────────────────┘
        compose network: only nginx joins the host's ports
```

Three services, one compose file, one `.env`.

### 4.1 Why nginx in front

Two reasons, and the first is the one that decided it.

**TLS is most of the point.** The attack surfaces this tier is meant to
demonstrate — SQL injection in a login form, a `UNION SELECT` through a search
box, a forged session cookie — are all more legible over HTTPS than over
plaintext HTTP, because that is what a real external service looks like and
because it puts a second, independent component between the attacker and the
application. A reader has to reason about what the proxy forwards, not just
what the app does.

**It gives the tier a single ingress.** Only nginx publishes host ports. The
portal and the database stay on the compose network and are unreachable from
the lab LAN, so `nmap app01` finds 80 and 443 and nothing else. That is a real
constraint on the scenario rather than a decoration: an attacker has to go
through the application, which is where the planted weaknesses are.

The certificate is **generated on the VM** by `app.yml` (openssl, self-signed,
CN and SAN = the VM's address, 10-year validity) and never committed. A
self-signed certificate in the repository would be a certificate whose private
key is public, which is worse than no certificate at all because it looks like
one. Generation is guarded by `creates:`, so a re-run cannot rotate the lab's
TLS identity underneath a scenario.

A consequence worth stating: **`app_validate.yml` and any manual testing must
use `-k` / `--insecure`**, because the certificate is self-signed by design.

### 4.2 Why one Go binary with three route groups

The alternative was three services — a public portal, a JSON API, an internal
admin panel — which is closer to how a company would actually build it. It was
rejected for a reason specific to this lab.

The tier's teaching unit is **the same request path behaving two ways**. With
the weakness toggle on, `POST /portal/login` concatenates a string into SQL;
with it off, the identical handler binds a parameter. That comparison is only
crisp if nothing else about the request differs. Across three services there
would be three login implementations, three session mechanisms and three
configs, and "the hardened one is on port 8081" would be doing work that
"`WEAK_SQLI_LOGIN=false`" does better.

The three groups are therefore route prefixes on one mux, and the deployment
topology is deliberately *not* part of the lesson. The trade-off is real:
nothing here demonstrates service-to-service trust, and the split would need
building before that could be a scenario.

### 4.3 Ports

| Port | Published by | Purpose |
|---|---|---|
| 443 | nginx | TLS. Everything real happens here. |
| 80 | nginx | 301 to 443. Exists so that a browser or a scanner typing the bare hostname lands somewhere sane. |
| 8080 | — | The portal's listener, compose-internal only. |
| 5432 | — | PostgreSQL, compose-internal only. |

---

## 5. The image pipeline

The question this had to answer: how do container images reach a VM whose base
image has Docker installed but no images, in a way that stays reproducible and
does not require the lab to reach a registry.

**Decision: build on the control node, ship one tarball.**

```
control node (host_a)                            app01
─────────────────────                            ─────
git checkout app/ ──► docker build ──► clayface/portal:lab ─┐
                                                                ├─► docker save ──► ssh ──► docker load ──► docker compose up
docker pull postgres:16-alpine ──────────────────────────────┘                              (pull_policy: never)
docker pull nginx:1.27-alpine ───────────────────────────────┘
```

Points that matter:

- **The portal image is stamped with a source digest.** The digest of the
  tracked files under `app/` is written into an image label, so a second run
  with unchanged sources does not rebuild. The same digest is what
  `app_validate.yml` reads to report what it is validating against.
- **Third-party tags are pulled on the control node, never on the VM.**
  `postgres:16-alpine` and `nginx:1.27-alpine` are pulled only when absent
  locally, so a deploy cannot silently move a database or proxy version.
- **`pull_policy: never` in the compose file.** The VM is forbidden from
  fetching an image. If the tarball does not carry a tag the compose file
  needs, the stack fails loudly instead of quietly reaching the internet. All
  three tags — portal, postgres, nginx — have to be in the save list; the nginx
  one was the easiest to forget and is the reason the save list is written out
  in full in `app.yml` rather than derived.
- **The VM's docker daemon is the control node's problem.** `app.yml` verifies
  the daemon is reachable and refuses to start it. This is a deliberate
  boundary: starting a daemon is a host-level change with security
  consequences, and a deploy script that does it silently is a deploy script
  that can enable a root-equivalent service without being asked.

---

## 6. The configuration surface

Everything the portal knows comes from environment variables, and everything in
the environment comes from one file that Ansible renders:
`/opt/clayface/app/.env` on the VM, mode `0640`, owned by `clayface`.

A config file read from disk was the alternative. It was rejected because the
compose stack is the deployment mechanism: `.env` is what compose already reads,
so a config file would have meant a bind mount, a parser, and a second place for
the same values to be wrong.

The `.env` is also where the weakness model lands (section 7): the toggles are
environment variables like everything else.

Credential sourcing, in one place because it is easy to get wrong:

| Value | Source | Default |
|---|---|---|
| `DB_PASSWORD` | `LAB_APP_DB_PASS` | `Clayface_DB_2024!` — deliberately the same string as the portal's compiled-in fallback (section 8.7) |
| `SESSION_SECRET` | `LAB_APP_SESSION_SECRET` | `clayface-lab-session-secret` |
| `SERVICE_ACCOUNT_USER` | `LAB_SVC_APP_USER`, else `lab.yaml` `app.service_account` | `svc-app-portal` |
| `SERVICE_ACCOUNT_PASSWORD` | `LAB_SVC_APP_PASS`, else `LAB_USER_PASS` | `Passw0rd!Lab` |

No credential is written into `lab.yaml`. That file holds no secrets at all, and
it is the file most likely to be read or shared.

---

## 7. The weakness model

### 7.1 Baseline versus toggle

The single most important design decision in this tier: **a weakness is a
toggle, and flipping it off must produce a genuinely safe implementation of the
same route — not a disabled route, not a 403, not a removed endpoint.**

That constraint is what makes the tier useful for the thesis. The comparison
"vulnerable lab vs hardened lab" is only evidence if the two are the same
system. A portal with its search endpoint switched off has not been hardened;
it has been reduced. `app/README.md` records, per toggle, what `false` does,
and every one of those entries is a real implementation: parameter binding,
`exec.Command` without a shell, an ownership predicate, a resolved-path check,
HMAC with a constant-time comparison, an address-literal filter.

Both postures ship in the same binary and the same image. Nothing is rebuilt,
nothing is branched at deploy time except one `.env` line per toggle.

### 7.2 The toggle list lives in `lab.yaml`

```yaml
app:
  service_account: svc-app-portal
  weaknesses:
    sqli_login: true
    ...
```

`lab.yaml` is the lab's single source of truth, so the toggles live there next
to the VM placement. One map is three things at once:

1. the **declaration** — what this tier is supposed to be vulnerable to;
2. the **environment contract** — `app.yml` renders each key as
   `WEAK_<KEY>` uppercased, and the Go config layer reads exactly those names;
3. the **assertion list** — `app_validate.yml` asserts each key both ways.

Because the map is the contract, adding a twelfth weakness means editing
`lab.yaml`, the handler, and the validator's `app_weakness_keys` list — and
`app.yml` fails loudly if the two lists disagree in either direction (a key
`lab.yaml` does not mention, or a key it mentions that nothing asserts).
`WEAK_<KEY>` uppercased is a mechanical mapping rather than a hand-kept table.

### 7.3 All eleven default to `true`

Unlike `ad.weaknesses` — which is entirely `false` and read by nothing yet —
this map defaults to `true` and is fully wired. The lab exists to be attacked;
a portal with every weakness off is only useful as the baseline half of a
comparison. Hardening is the explicit act.

The Go layer reflects the same posture: an unset or unparseable `WEAK_*`
variable is treated as **enabled**, so a typo leaves the lab vulnerable and
obvious rather than quietly safe.

### 7.4 The eleven

Grouped the way `lab.yaml` groups them. Full per-toggle detail, including the
exact payload and what the hardened path does, is in `app/README.md`; this table
is the design-level summary of *why each one is here*.

| Key | Class | Why this one is in the lab |
|---|---|---|
| `sqli_login` | Injection | The canonical first finding. A login form that concatenates is the most common real-world SQL injection there is, and it hands over a session immediately, which gives the rest of the route table something to be attacked with. |
| `sqli_search` | Injection | A search box is where injection survives longest in practice, because it is read-only and nobody thinks of it as a trust boundary. Here it is the tier's exfiltration primitive: a `UNION SELECT` dumps `api_keys`. |
| `command_injection_diagnostics` | Injection | The admin panel's diagnostics runner. It exists because the tier otherwise has no OS-level injection, and an RCE-flavoured finding is what a Ch5 narrative needs to move from "read data" to "run something on the host". |
| `idor_documents` | Broken access control | The order and invoice endpoints select by primary key alone. It is the cleanest possible IDOR — no encoding, no bypass, just change the number — and it is the one weakness where the *hardened* path also adds authentication, which is worth seeing. |
| `broken_admin_authz` | Broken access control | `X-Clayface-Role: admin` is trusted. One header opens the whole panel: user list, audit log, diagnostics, config. It is the tier's "authorization decided by the client" lesson and it is deliberately implemented in one wrapper so the whole route group flips together. |
| `path_traversal_download` | Broken access control | `?file=../../etc/passwd`. Common, immediately legible, and it gives the reader something outside the application's own data model. |
| `hardcoded_db_credentials` | Credential exposure | The database password also exists as a literal in the Go source, and both config endpoints print the live password and say where it came from. Two independent disclosure paths to one credential. |
| `forgeable_session_token` | Credential exposure | The cookie is `base64url("username\|role\|issuedAt")` with no signature, so `adm-hermione\|admin\|<now>` is a valid admin session. It is the tier's privilege-escalation step that needs no injection at all, and it demonstrates why a server must not trust a value it did not sign. |
| `service_account_credential_in_db` | Credential exposure | The pivot. A business table holds `svc-app-portal`'s plaintext password next to API keys that are also plaintext. This is the weakness that makes the tier part of the AD story rather than a standalone web exercise (section 9). |
| `ssrf_url_fetch` | Server-side extra | `?url=` is fetched server-side and the body returned, so the portal proxies into the compose network and the lab LAN. It is what turns an application finding into an internal network position. |
| `verbose_errors_debug_endpoint` | Server-side extra | Four things behind one toggle: unauthenticated `net/http/pprof`, a panic that renders the stack trace *and the effective configuration including the database password*, SQL text in database errors, and a trusted leftmost `X-Forwarded-For`. Each is a real information-disclosure class. |

### 7.5 One deliberate coupling

`WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT` gates four separate behaviours (pprof
mounting, panic rendering, error verbosity, forwarded-for trust). That is a
compromise: as weaknesses they are independent, and a stricter design would
give each its own key.

They are grouped because they share a single premise — *the application tells
the caller more than it should* — and because four keys would have quadrupled a
part of the validator that is already the longest section without adding a
distinct lesson. The grouping is visible: `app/README.md` lists all four
manifestations under the one toggle, and the validator asserts the four
separately, so a reader who wants them split can see exactly what to split.

---

## 8. Per-weakness notes that are not obvious from the code

### 8.1 `sqli_login` logs you in as the *first* user, not as an admin

The payload `' OR '1'='1' -- ` makes the query return `LIMIT 1` of an unfiltered
scan, which is row 1 — `ron.weas`, an ordinary user. That is not a bug in the
weakness; it is what makes it a better exercise, because the injected session is
low-privileged and the reader has to find the escalation (the forged token, or
the admin header) rather than being handed the panel.

### 8.2 The password hash is SHA-256, and that is the baseline

`users.password_hash` holds a bare, unsalted, uniterated SHA-256 digest. That is
wrong and it is deliberate, but it is **not one of the eleven toggles**: it is
the ground the login weakness stands on. It is what makes a dumped `users` table
crackable in seconds, which is what makes the seed data worth stealing.

The toggles change how the lookup is *built*, not how the digest is *stored*.
Fixing the hash would not fix the injection, and pretending otherwise would
teach the wrong lesson.

### 8.3 The seeded digests are real

`seed.sql` stores the actual SHA-256 of the password it documents, and the
comments say so. Two consequences: the file is honest about what it contains,
and a reader can verify the hash-to-password relationship themselves rather than
taking it on faith.

### 8.4 The seed identities are the AD identities, on purpose

`ron.weas`, `bob.sing`, `abed.nad`, `adm-hermione` and `svc-idp-ldap` are the
same five principals `ad.yml` creates in `clayface.local`. Their portal
passwords are the lab's shared user password (`LAB_USER_PASS`,
`Passw0rd!Lab`) — the documented deviation in `docs/ad-identity-design.md`
section 14, one password for every human, reused here.

The sixth principal is the exception: `svc-app-portal` is the one account whose
password comes from its own variable (`LAB_SVC_APP_PASS`,
`docs/ad-identity-design.md` section 4.2) rather than the shared one, and the
portal's own service-credential row uses that value instead. The reasoning is
in section 9.

**That reuse is the whole chain.** Dump the portal database, crack a fast hash,
get `Passw0rd!Lab`, and it is also a working domain credential. Without the
reuse the tier would be two unrelated exercises instead of one attack path.

### 8.5 The `svc-idp-ldap` row in `api_keys` is the LDAP pivot

`api_keys` holds `('idp01-ldap-bind', 'svc-idp-ldap', 'Passw0rd!Lab',
'ldap-bind', ...)` — the directory bind credential, in clear text, in a business
table, with a note saying its rotation is overdue. Recover it from a database
dump or a `UNION SELECT` and it binds to LDAP on `dc01`. It is the second
credential in the same table that reaches the domain, and it is reachable
through the *search* weakness rather than the service-account one.

### 8.6 Seed material is written once and never re-synced

The schema is applied only to an empty database, and `$DATA_DIR/documents` is
populated only when it is empty. Both live in a named volume that survives
redeployment.

This is not laziness — it is the requirement. A real deployment would have real
documents in it, and overwriting them would destroy the thing a student is
supposed to exfiltrate. The embedded set is a first-boot convenience and never a
sync. The same reasoning protects `api_keys` drift.

### 8.7 The database password default is not a free choice

`app.yml` renders `DB_PASSWORD` from `LAB_APP_DB_PASS`, defaulting to
`Clayface_DB_2024!`. That string is *also* the `hardcodedDBPassword` constant
compiled into the portal.

The coincidence is the finding. The toggle is documented as "anyone who obtains
the image — or the source — obtains a working database password without ever
touching the environment", and that sentence is only true while the deployed
password and the literal are the same value. If they drift, the weakness
degrades into "the binary contains a password that does not work", which is a
decoration. **Change one, change both** — the comment is repeated in
`app.yml`, `app_validate.yml`, `deploy.env.example` and `config.go` so that
whoever changes it is told.

### 8.8 Verbose errors leak the database password on purpose

`recoverPanic` prints the effective configuration — database host, user and
password — into the response body when the toggle is on. That is a lot to hand
over, and it is the point: a stack trace in production is an information
disclosure, and here the disclosure is bounded, known and measured.

---

## 9. Identity: `svc-app-portal`

This is the tier's connection to the AD layer, and it is deliberately
**one-directional and thin**.

The portal is not domain-joined and does not authenticate anyone against AD.
What it has is a credential: `svc-app-portal`, an AD account created by
`ad.yml`, which the portal carries in its configuration and (under
`service_account_credential_in_db`) plants in plaintext in its own database.

The chain the tier exists to make possible:

```
sqli_search / sqli_login  ──►  dump api_keys
                                      │
                        svc-app-portal's plaintext password
                                      │
                                      ▼
                              authenticate to clayface.local
```

Three properties make that chain work, and each was a decision:

1. **`svc-app-portal` has no group memberships.** `lab.yaml` gives it
   `groups: []`. Its reach is the whole of its own rights, which keeps the
   finding honest — a reader who recovers it has gained exactly an authenticated
   low-privilege domain identity, not a shortcut to Domain Admin.
2. **The planted copy and the real password are the same string.** `ad.yml`
   reads the account's own `password_env:` (`LAB_SVC_APP_PASS`) so the account
   can rotate without touching the human accounts, and `app.yml` reads the same
   variable for the value it plants. If they diverged, the pivot would silently
   stop working — which is the worst kind of lab failure, because it looks like
   a mistake the student made.
3. **No domain join.** See section 1. Keeping the two systems separate is what
   makes the recovered credential a *pivot* rather than an artifact of the
   portal's own authentication.

Note the correction this forces in the older design document:
`docs/ad-identity-design.md` section 1 states that APP01 does not consume AD and
that `svc-idp-ldap` is the lab's only AD service account. Both sentences were
true of that cycle's design and are no longer true. They have been corrected in
place.

---

## 10. Attack-path relevance

This tier is Ch5 chain 2: **public web/API exploit → APP01 → SQLi/credentials →
PostgreSQL, plus service identity → DC01.** It is the chain that matches the
confirmed attacker model (external, no prior foothold, in through the edge),
which is why it is the primary scenario rather than the VPN-credentials chain.

Mapping the tier's surface onto that chain:

| Chain step | What the tier provides |
|---|---|
| Initial access | `sqli_login` or `sqli_search` through nginx on 443 |
| Execution / host access | `command_injection_diagnostics` (admin panel, reachable via `broken_admin_authz` alone) |
| Credential access | `api_keys` via `sqli_search`; `users` table via a dump; the DB password via `verbose_errors_debug_endpoint` |
| Privilege escalation (in-app) | `forgeable_session_token`, or `broken_admin_authz` with one header |
| Collection | `idor_documents` walks orders and invoices; the five seed documents; `path_traversal_download` reaches outside |
| Lateral movement (network) | `ssrf_url_fetch` from inside the compose network and the lab LAN |
| Lateral movement (identity) | `service_account_credential_in_db` → `svc-app-portal` → `clayface.local` |

**What is missing from this chain, and must not be implied:**

- No boundary was crossed. Section 3.1.
- No attack has been run. Everything above is *affordance* — a surface that
  exists — not a result. The thesis's engagement chapter has no results yet.
- Nothing detects any of it. There is no SIEM, no Wazuh, no log shipping. The
  portal writes an `audit_log` table and a container log line per request, and
  both are written by the system being attacked; an attacker who can write to
  the database can edit the first one. The container log is the marginally
  better record, and neither reaches anywhere an operator would see it.

---

## 11. Baseline versus weakness, as a thesis artifact

The tier supports a comparison the rest of the lab cannot yet: **the same
system, in two postures, with an assertion playbook that proves which posture it
is in.**

`app_validate.yml` runs one section per toggle. Each section asserts the
weakness is *present* when the toggle is on and *absent* when it is off, using
only external HTTP requests against the running stack. It mutates nothing,
restarts nothing, and takes no side effect beyond its own requests — the
database rows it reads are rows the application wrote.

The shape of an assertion is the same everywhere: establish a baseline ("start
with nothing leaked"), attempt the attack, record what came back, then assert
the recorded result against what `lab.yaml` declares. Many of them are
deliberately *positive* assertions of the vulnerability — "assert the injected
login authenticates" — because a lab whose only test is "the attack failed"
cannot tell a hardened system from a broken one.

The tally at the end reports three things: the toggles as `lab.yaml` declares
them, which ones were asserted, and **which ones `lab.yaml` declares that
nothing above asserted**. That last list is the honest one, and it is designed
to be empty.

---

## 12. Data flow

Nothing in this tier invents a fact that lives elsewhere.

```
lab.yaml
  vm_placements.app01 ────────────► terraform/locals.tf ──► module "app_host_a"
  app.service_account ────────────► lab_inventory.py ──┐
  app.weaknesses ─────────────────► lab_inventory.py ──┤
                                                        ▼
                        ansible/playbooks/app.yml  (vars: lab_app)
                                                        │
                    ┌───────────────────────────────────┼──────────────────┐
                    ▼                                   ▼                  ▼
              .env (WEAK_*)                   compose file (verbatim)   nginx.conf (verbatim)
                    │
                    └──► app/internal/config/config.go (env var names as constants)

terraform vms output ──► inventory/terraform_vms.py ──► group vms_app
                                                          │
                                          app.yml / app_validate.yml / opnsense.yml (vms_pinned)
```

Derived state flows the house way: Terraform owns "which VMs exist", so `vms_app`
comes from the `vms` output and not from a list in `lab.yaml`. Static facts flow
the house way too: the address, the service account name and the toggle map are
all in `lab.yaml` and are read, never restated.

The one mirrored constant is `LAB_DOMAIN` (default `clayface`), for the reason
`CLAUDE.md` records — it is owned by the OPNsense image and `lab.yaml` cannot
change what that image baked in.

---

## 13. Automation plan

### 13.1 Files

| Path | What it is |
|---|---|
| `terraform/modules/app/` | The VM module: `libvirt_volume` (40 GiB overlay on the base) + `libvirt_domain` (BIOS, 4 GiB, 2 vCPU, virtio, pinned MAC, no graphics device) |
| `terraform/main.tf` | `module "app_host_a"` / `"app_host_b"`, instantiated per host from `vm_placements` |
| `terraform/locals.tf` | `module_os` gained `app = "linux"` |
| `ansible/playbooks/app.yml` | Build, ship, deploy. Two plays. |
| `ansible/playbooks/app_validate.yml` | Read-only assertions, one section per toggle |
| `ansible/inventory/terraform_vms.py` | `vms_app` group, the `ssh` + `clayface` connection vars, `LAB_APP_SSH_PASS` |
| `ansible/inventory/lab_inventory.py` | Distributes `lab_app` as a host var |
| `app/` | The Go portal, the Dockerfile, the compose file, the nginx config, the seed, `README.md` |
| `deploy.sh` | `app.yml` then `app_validate.yml`, after `client.yml` and before `ad_validate.yml` |

### 13.2 What `app.yml` refuses to do

- It does not install, upgrade or repair anything on the VM. Docker and the
  compose v2 plugin are part of the base image. The play asserts them, the
  sudo access, and docker socket membership, and **fails with a message naming
  the gap**. Repairing a base image from a deploy script would hide the fact
  that the image is not what it is supposed to be.
- It does not start the docker daemon on the control node.
- It does not publish anything but nginx's ports.
- It does not generate a certificate a second time.

### 13.3 Idempotency

Every mutating step is guarded, and the guards are the interesting part:

| Step | Guard |
|---|---|
| Portal image build | source digest label; unchanged sources do not rebuild |
| Third-party pulls | only when the tag is absent locally |
| Tarball | repacked only when an image moved or the tarball is gone |
| Copy to VM | `copy` by checksum, so an unchanged tarball is not re-sent |
| `docker load` | only when the tarball changed or a tag is missing (load always rewrites the tag, so it cannot be left unguarded) |
| Compose file / nginx.conf | `copy` by checksum |
| TLS keypair | `creates:` on the private key |
| `.env` | rendered from the same inputs, so an unchanged posture produces an unchanged file |
| `docker compose up` | Compose's own reconcile |

`app_validate.yml` is read-only by construction and asserts nothing about
idempotency, because asserting it would mean running `app.yml` twice.

### 13.4 Order in `deploy.sh`

```
... client.yml
    app.yml            ← needs a running docker daemon on the control node
    app_validate.yml
    ad_validate.yml    ← runs last: it asserts workstation state
```

`app.yml` runs after `client.yml` only because of where it sits in the file;
it has no dependency on the domain. It runs *before* `app_validate.yml`
obviously, and both run before `ad_validate.yml` because the AD validator
asserts workstation state that only exists after the client joins.

Two ordering facts worth knowing:

- **`app.yml` must not be run with `--ask-become-pass`.** The `app` group's
  sudo password comes from the inventory (`ansible_become_password`, derived
  from `LAB_APP_SSH_PASS`), and the `-K` prompt would override it with the
  control node's password. `deploy.sh` passes `-K` only to the two
  `lab_inventory`-only playbooks (`hosts.yml`, `edge.yml`), so this is already
  correct — but a manual `app.yml` run has to respect it. The comment is in
  `inventory/terraform_vms.py`.
- **`LAB_SVC_APP_PASS` must be set before the first `ad.yml` run, or set
  consistently after.** If it changes between `ad.yml` and `app.yml`, the
  planted credential and the real account disagree and the pivot stops working.

---

## 14. Validation

`app_validate.yml` is the tier's proof. Its contract:

- **Read-only.** No `docker` mutation, no restart, no deploy. Every request is
  an HTTP request against the stack over TLS.
- **Keyed to `lab.yaml`.** It reads `app.weaknesses` and asserts each key in
  the direction the map declares, so the same playbook validates the vulnerable
  and the hardened posture without being edited.
- **Positive and negative.** Both "the attack works" and "the attack fails" are
  assertions, not just the latter.
- **External only.** It uses the same vantage point an attacker would: over the
  network, through nginx, with `--insecure` for the self-signed certificate. It
  never reads the container's filesystem or the `.env`, so it cannot pass by
  inspecting state an attacker could not see.
- **Credentialed where it needs to be.** It logs in as a low-privileged lab user
  so that the session-dependent weaknesses are tested from the position they are
  meant to be exploited from.

What it does **not** do: it does not prove the AD pivot works. Recovering
`svc-app-portal`'s credential is asserted (the database holds it); authenticating
to `clayface.local` with it is not, because that belongs to the AD validator's
side of the contract. That step is untested — see section 16.

---

## 15. Documented deviations and known limitations

Recorded so they read as choices rather than oversights. `app/README.md`
carries the application-level list; these are the design-level ones.

- **No network segmentation *inside* the internal zone.** Closed for the DMZ
  on 2026-09-25 (section 3.1): APP01 is behind a default-deny boundary now.
  What remains unbuilt is the internal split itself — `dc01`, `client01` and
  IDP01 still share one flat segment, so the boundary this document describes
  is the DMZ's edge and not an internal tiering. `docs/network-design.md` has
  the zones that do exist.
- **The hardened posture keeps the literal in the binary.** "No hardcoded
  credential" cannot be literally true of a file that must contain the literal
  in order to be the weakness. What the toggle removes is the *use* — no
  fallback, `DB_PASSWORD` mandatory, config endpoints redacting. The constant
  stays in the compiled image either way, which is the honest form of the
  finding: it is exactly why a hardcoded credential cannot be rotated by
  configuration alone.
- **The session signature has no expiry.** The HMAC proves the payload was
  issued by this server; it does not prove it was issued recently. No nonce, no
  max-age, no server-side store.
- **The SSRF guard validates literals.** A hostname that resolves into a private
  range still passes, because the check never resolves anything. Closing that
  needs a resolution-time check or an egress policy, which is a network control
  rather than an application one — and that is arguably the right place for it.
- **The SQL seed splitter is simple.** `internal/seed/seed.go` splits on
  end-of-line semicolons and understands full-line comments only. It must be
  replaced before the seed files can contain functions, dollar-quoted bodies, or
  a trailing comment on a statement line.
- **No rate limiting, no CSRF token, no lockout.** Deliberate at this stage;
  each of them would be a weakness-toggle candidate in a later cycle.
- **Four behaviours share one toggle.** Section 7.5.
- **The validator's credential assertions trust the seed.** It reads the
  database through the application rather than connecting to PostgreSQL
  directly, because connecting directly would require the database password to
  be reachable from the control node — which is a second place for the same
  secret to live. The trade-off is that a validator bug and an application bug
  look the same.

---

## 16. What is not verified

Everything above was written against a lab that was, at the time, **down**: no
segments built, no VMs running, no docker daemon on the control node. The
offline
checks that did run — Terraform `validate` and `fmt`, `gofmt`, `go vet`,
`go build`, YAML and JSON parsing of every touched file, `--syntax-check` on
the three playbooks — prove the artifacts are internally consistent. They prove
nothing about the system.

Note that the *network* half of this section has since moved: the DMZ segment,
the boundary and the DHCP/DNS the tier needs are built and asserted by
`ansible/playbooks/opnsense_dmz.yml` (`docs/network-design.md` §10). The tier's
own behaviour is still what this section is about.

The specific untested items, with the exact command to check each and what to do
when it fails, are in **`docs/app01-verification-pending.md`**. That file is the
worklist for the first real deploy; this one is the design.

Do not cite any behaviour of this tier as demonstrated until those items are
closed and `app_validate.yml` has run green against a live stack.

---

## 17. Open questions

- **Should the portal be exposed to the WAN?** `opnsense.yml` implements a
  destination-NAT port forward, gated on `LAB_WAN_EXPOSE_APP` and off by
  default. It is **unverified** (see `docs/app01-verification-pending.md`). It
  is needed for the external-attacker model the proposal commits to, and it is a
  real change to the edge firewall in the meantime.
- **When does the *internal* segmentation land?** The DMZ boundary is built
  (section 3.1, `docs/network-design.md`). The internal zone is still flat, so
  a foothold on `dc01` reaches `client01` without crossing anything. This
  tier's own contribution to the "no initial trust" claim is now the boundary
  it sits behind, which is a real one. `docs/TODO.md` records what remains.
- **Does the tier need a second route group with a different trust story?**
  Section 4.2 argues against splitting the binary. The cost is that no scenario
  demonstrates service-to-service trust; if one is wanted, that is a second
  service, not a second route group.
- **Should `svc-app-portal` gain a permission that makes the pivot pay off?**
  Today it recovers a low-privilege domain identity, which is honest but short.
  Giving it a right that reaches something — a share, an SPN, a delegation —
  would lengthen the chain at the cost of a second planted misconfiguration.
  This is the weakness-wave decision, not this cycle's.
