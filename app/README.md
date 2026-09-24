# Clayface Portal

The application tier of the Clayface lab: a small customer portal, a JSON API
and an internal admin panel, built to be attacked.

> **Every vulnerability in this directory is intentional.** This is the
> deliberately vulnerable application of an isolated bachelor-project red team
> lab. The weaknesses are here so that the ethical hacking lifecycle —
> discovery, exploitation, lateral movement, exfiltration — can be practised and
> demonstrated end to end, with each one switchable so that the secure and the
> insecure version of the same code path can be compared. None of it belongs
> anywhere except inside this lab. See `docs/redteam_corporate.md` and
> `docs/ad-identity-design.md` for the scope this sits inside.

`docs/redclay.yaml` describes this VM as APP01: a Linux application host running
Docker and Docker Compose, whose services are a public portal, an API, an
internal admin panel and PostgreSQL. This directory is that service model,
implemented, and it is deliberately composed of containers rather than separate
VMs — the tier exists to be compromised, not to weigh anything down.

## What runs

Three containers, one compose project, one published port range:

| Service | Image | Listens on | Published |
|---|---|---|---|
| `nginx` | `nginx:1.27-alpine` | 443 (TLS), 80 (redirect) | yes |
| `portal` | `clayface/portal:lab` | 8080 (plain HTTP) | no |
| `postgres` | `postgres:16-alpine` | 5432 | no |

The portal is one Go binary serving three route groups on one listener. That was
the design decision for this tier — *one binary, several route groups*, not
microservices — and it is what makes the weakness toggles meaningful: the public
portal, the API and the admin panel are the same process, the same
configuration and the same database handle, so a toggle changes how a request is
handled rather than which service handles it.

TLS terminates at nginx. The compose network between nginx and the portal is
plain HTTP and is a trust boundary: anything that can join it reaches all three
route groups without a certificate.

## Route map

`/healthz` is unauthenticated and exists for the container healthcheck. The
remaining three groups are:

### `/portal/*` — the public portal

| Method | Path | Authentication | Toggle that changes it |
|---|---|---|---|
| GET | `/portal/` | none | — |
| GET | `/portal/login` | none | — |
| POST | `/portal/login` | none | `WEAK_SQLI_LOGIN` |
| GET | `/portal/logout` | none | — |
| GET | `/portal/profile` | session | — |
| GET | `/portal/search?q=` | none | `WEAK_SQLI_SEARCH` |
| GET | `/portal/download?file=` | none | `WEAK_PATH_TRAVERSAL_DOWNLOAD` |

### `/api/*` — the JSON API

| Method | Path | Authentication | Toggle that changes it |
|---|---|---|---|
| POST | `/api/login` | none | `WEAK_SQLI_LOGIN` |
| GET | `/api/profile` | session | — |
| GET | `/api/orders` | session | — |
| GET | `/api/orders/{id}` | none (hardened: session) | `WEAK_IDOR_DOCUMENTS` |
| GET | `/api/invoices/{id}` | none (hardened: session) | `WEAK_IDOR_DOCUMENTS` |
| GET | `/api/search?q=` | none | `WEAK_SQLI_SEARCH` |
| GET | `/api/keys` | session | `WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB` (data, not access) |
| GET | `/api/fetch?url=` | none | `WEAK_SSRF_URL_FETCH` |
| GET | `/api/debug/config` | none | `WEAK_HARDCODED_DB_CREDENTIALS` |

### `/admin/*` — the internal admin panel

| Method | Path | Authentication | Toggle that changes it |
|---|---|---|---|
| GET | `/admin/` | admin | `WEAK_BROKEN_ADMIN_AUTHZ` |
| GET | `/admin/users` | admin | `WEAK_BROKEN_ADMIN_AUTHZ` |
| GET | `/admin/audit` | admin | `WEAK_BROKEN_ADMIN_AUTHZ` |
| GET | `/admin/diagnostics?host=` | admin | `WEAK_COMMAND_INJECTION_DIAGNOSTICS` |
| GET | `/admin/config` | admin | `WEAK_HARDCODED_DB_CREDENTIALS` |

### `/debug/*` — mounted only while the verbose-errors toggle is on

| Method | Path |
|---|---|
| GET | `/debug/pprof/`, `/debug/pprof/cmdline`, `/debug/pprof/profile`, `/debug/pprof/symbol`, `/debug/pprof/trace` |
| GET | `/debug/panic` — triggers a panic, to demonstrate how one is rendered |

The rule the two "authentication" columns follow: the customer-facing surface
(the portal pages, the two search endpoints, the URL fetch) is reachable
without a session, because that is the surface an attacker lands on first; the
endpoints that return *one account's* data require a session. The single-order
endpoints are the exception while `WEAK_IDOR_DOCUMENTS` is on — an
authorization check that is missing cannot require the session it would have
used, so the weak path serves any row to any caller.

## Weakness toggles

Each toggle is an environment variable read at start-up. Each defaults to `true`
when unset, so the lab is vulnerable unless somebody has said otherwise. Set one
to `false` in `.env` to run the hardened version of that path — the same binary,
the same routes, a genuinely safe implementation.

| Toggle | Where it manifests | What the weakness is | What `false` does |
|---|---|---|---|
| `WEAK_SQLI_LOGIN` | `POST /portal/login`, `POST /api/login` | The credential lookup is built by string concatenation, so `' OR '1'='1' -- ` as the username logs the caller in as the first user in the table. | Binds the username and password as parameters. |
| `WEAK_SQLI_SEARCH` | `GET /portal/search`, `GET /api/search` | The term is concatenated into `... WHERE title LIKE '%<q>%'`, so a `UNION SELECT` dumps `api_keys` — the credential table. | Binds the term as a parameter, where it is searched for literally. |
| `WEAK_COMMAND_INJECTION_DIAGNOSTICS` | `GET /admin/diagnostics?host=` | The host is interpolated into a command run through `sh -c`, so `; id` runs a second command and returns its output. | Runs `ping` directly with no shell, after matching the host against a hostname pattern. |
| `WEAK_IDOR_DOCUMENTS` | `GET /api/orders/{id}`, `GET /api/invoices/{id}` | The row is selected by primary key alone. No ownership check, and no session needed. | Requires a session and adds the ownership predicate; somebody else's row is a 403. |
| `WEAK_BROKEN_ADMIN_AUTHZ` | every `/admin/*` route | The header `X-Clayface-Role: admin` is accepted as proof of the admin role. | Ignores the header and requires the admin role in the session. |
| `WEAK_PATH_TRAVERSAL_DOWNLOAD` | `GET /portal/download?file=` | The name is joined onto the documents directory uncleaned, so `../../etc/passwd` escapes it. | Rejects separators and `..`, then confirms the resolved path is still inside the directory. |
| `WEAK_HARDCODED_DB_CREDENTIALS` | `GET /admin/config`, `GET /api/debug/config`, and the binary itself | The database password also exists as a literal in the Go source, used when `DB_PASSWORD` is unset, and the config endpoints print it and say where it came from. | Redacts the value, and `DB_PASSWORD` becomes mandatory — there is no fallback at all. |
| `WEAK_FORGEABLE_SESSION_TOKEN` | the session cookie | The cookie is `base64url("username\|role\|issuedAt")` with no signature, so a client can write `adm-hermione\|admin\|<now>` and be an administrator. | HMAC-SHA256 over the payload, keyed on `SESSION_SECRET`, compared in constant time; the cookie is `HttpOnly`, `Secure` and `SameSite=Lax`. |
| `WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB` | the `api_keys` table itself | A row holds the plaintext `SERVICE_ACCOUNT_USER` / `SERVICE_ACCOUNT_PASSWORD`, labelled as a directory bind credential. | The row keeps a bcrypt-shaped placeholder instead, so the table yields nothing usable. |
| `WEAK_SSRF_URL_FETCH` | `GET /api/fetch?url=` | Any URL is fetched server-side and the body returned, so the portal proxies for the attacker into the compose network and the lab LAN. | Only `https`, and no loopback, private, link-local or unspecified address literal — including across redirects. |
| `WEAK_VERBOSE_ERRORS_DEBUG_ENDPOINT` | panics, error responses, `/debug/pprof/*`, and the audit log's client field | `net/http/pprof` is mounted unauthenticated, a panic renders the stack trace and the effective configuration, database errors carry their SQL, and the leftmost `X-Forwarded-For` value is trusted as the client address. | No pprof, no `/debug/panic`, generic error pages, and the client address comes from the connection. |

`GET /healthz` reports readiness and nothing else. It is never weakness-gated,
because a healthcheck that depends on a toggle is a healthcheck that lies.

## Configuration

All configuration is environment variables. Compose reads them from `.env` in
this directory (`/opt/clayface/app/.env` on the VM), which Ansible generates at
deploy time; nothing in this repository carries a real credential.

| Variable | Default | Purpose |
|---|---|---|
| `DB_HOST` | `postgres` | Database host — the compose service name. |
| `DB_PORT` | `5432` | Database port. |
| `DB_NAME` | `clayface` | Database name. |
| `DB_USER` | `clayface_app` | Database role. |
| `DB_PASSWORD` | — | Required (see below). |
| `SESSION_SECRET` | — | Required when `WEAK_FORGEABLE_SESSION_TOKEN=false`, ignored otherwise. |
| `SERVICE_ACCOUNT_USER` | `svc-app-portal` | The portal's own service account, as stored in `api_keys`. |
| `SERVICE_ACCOUNT_PASSWORD` | — | Required; seeded into `api_keys` while the service-account toggle is on. |
| `DATA_DIR` | `/data` | Document store root; the served directory is `$DATA_DIR/documents`. |
| `LISTEN_ADDR` | `:8080` | The portal's listener. The healthcheck assumes 8080. |

`DB_PASSWORD`, `SESSION_SECRET` and `SERVICE_ACCOUNT_PASSWORD` have no fallback
in `docker-compose.yml` — Compose refuses to start the stack without them
(`:?` in the compose file). That is on purpose: a default there would be a
hardcoded credential in a file that gets committed, and the one place this lab
allows a fallback password to exist is behind `WEAK_HARDCODED_DB_CREDENTIALS`,
where it is the finding.

## Seed data

On first boot, if the `users` table is absent, the portal applies
`internal/seed/schema.sql` then `internal/seed/seed.sql`, both compiled into the
binary with `//go:embed`. If `$DATA_DIR/documents` is empty it also writes five
sample documents. Neither happens twice: the schema is only applied to an empty
database, and the document directory is only populated when it is empty, so a
restart never overwrites what a student has changed.

The data is built so that a database dump is worth something and so that the
lab's two halves connect. The identities are the AD identities from
`lab.yaml` and `ansible/playbooks/ad.yml`:

| Username | Role | Password | Note |
|---|---|---|---|
| `ron.weas` | user | `Passw0rd!Lab` | owns customers 1–3 |
| `bob.sing` | user | `Passw0rd!Lab` | owns customers 4–5 |
| `abed.nad` | user | `Passw0rd!Lab` | owns customer 6 |
| `adm-hermione` | admin | `Passw0rd!Lab` | the only account that can use `/admin/*` when the header bypass is off |
| `svc-app-portal` | service | `Svc-Portal!2024` | the portal's own account, rewritten from the environment at every boot |
| `svc-idp-ldap` | service | `Passw0rd!Lab` | the AD directory bind account — the pivot |
| `svc-monitoring` | service | `Svc-Monitor!2024` | a read-only scrape credential |

`Passw0rd!Lab` is the lab's shared user password (`LAB_USER_PASS`, defaulted in
`deploy.env.example`), which is the deviation recorded in
`docs/ad-identity-design.md` section 14: every human account shares one value.
Here it is shared on purpose, and it is what turns two exercises into one chain:
dump the portal database, find `svc-idp-ldap`'s bind credential in `api_keys` in
clear text, and bind to LDAP on `dc01` with it. The passwords are stored as a
bare SHA-256 digest — unsalted, uniterated, and crackable — which is the
baseline the login weakness lands on.

`internal/seed/documents/` holds the on-disk half: an HR memo, a vendor
contract, a credentials-rotation note, an access policy and a customer export.
The rotation note is the narrative hinge — it names the accounts that still
share a password and says, in as many words, that the bind credential is stored
in the `api_keys` table of this database.

## Building and running

On the lab VM, Ansible does all of this. By hand, from this directory:

```bash
# 1. The secrets. Compose will not start without these three.
cat > .env <<'EOF'
DB_PASSWORD=change-me-local-only
SESSION_SECRET=change-me-local-only
SERVICE_ACCOUNT_PASSWORD=change-me-local-only
EOF

# 2. A certificate. Self-signed is fine here and it is deliberately not
#    committed: a certificate in the repository is a published private key.
mkdir -p certs
openssl req -x509 -newkey rsa:2048 -nodes -days 365 \
  -keyout certs/privkey.pem -out certs/fullchain.pem \
  -subj "/CN=app01.clayface"

# 3. Build the application image.
docker build -t clayface/portal:lab .

# 4. The database image has to be present locally too, because both services
#    are pinned to pull_policy: never.
docker pull postgres:16-alpine
docker pull nginx:1.27-alpine

# 5. Start it.
docker compose up -d
docker compose ps
curl -k https://localhost/healthz
```

On the VM there is no route to a registry, so the images are built and pulled on
the Ansible control node and shipped together:

```bash
docker build -t clayface/portal:lab app/
docker pull postgres:16-alpine
docker pull nginx:1.27-alpine
docker save clayface/portal:lab postgres:16-alpine nginx:1.27-alpine \
  | gzip > clayface-app-images.tar.gz
# copy to the VM, then: gunzip -c clayface-app-images.tar.gz | docker load
```

The compose project name is `clayface-app`, so the volumes are
`clayface-app_postgres_data` and `clayface-app_portal_data`.

## Where the weaknesses live

The deliberately vulnerable code paths are marked so they can be found in a
review and cited in a thesis. Every one of them is wrapped in a comment block
that opens with the toggle's name:

```go
// --- DELIBERATE WEAKNESS: WEAK_SQLI_LOGIN ---
```

```bash
grep -rn "DELIBERATE WEAKNESS" internal/
```

Each block names the toggle, says what the weakness demonstrates, and — for the
toggles that have one — describes what the hardened branch does instead. The
two postures are always the same function with one branch differing, never two
separate implementations, so that "what was fixed" is answerable by reading a
few adjacent lines.

## Known limitations and deliberate simplifications

Recorded so that they read as choices rather than oversights.

- **`WEAK_HARDCODED_DB_CREDENTIALS` and the literal.** The requirement for the
  hardened posture is "no literal in the source", and that cannot be literally
  true of a file that has to contain the literal in order to be the weakness.
  What the toggle actually does is remove the *use*: with it off, `Load` does
  not fall back to the constant and `Validate` refuses to start without
  `DB_PASSWORD`, and both config endpoints redact. The constant is still in the
  compiled binary either way — which is the honest form of this finding, since
  that is exactly why hardcoded credentials cannot be rotated by configuration
  alone.
- **The session signature has no expiry.** The HMAC proves the payload was
  issued by this server; it does not prove it was issued recently. A nonce, a
  max-age or a server-side session store is the next step and is not
  implemented.
- **The SSRF guard validates literals.** A hostname that resolves into the
  private range still passes, because the check never resolves anything. Closing
  that needs resolution-time checks or an egress policy, which is a network
  control rather than an application one.
- **The SQL seed splitter is simple.** `internal/seed/seed.go` splits the SQL
  files on end-of-line semicolons and understands full-line comments only. It
  would need replacing before these files could contain functions, dollar-quoted
  bodies or trailing comments on a statement line.
- **Password hashing is SHA-256.** Fast, unsalted, and wrong. It is the baseline
  the login toggle lands on rather than a numbered finding of its own, and
  fixing it would not fix the injection.
- **The application has no rate limiting, no CSRF token and no lockout.** Also
  baseline for this tier, and all three would slow the lab down more than they
  would teach.
