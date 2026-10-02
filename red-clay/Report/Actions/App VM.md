Three components deployed in docker containers:
1. Vulnerable service
2. Nginx
3. Database

## Vulnerable service

A go backend service for serving service and data to employees or users. I made it do a multi job just for the sake of simplicity. couldn't see a real and good reason for writing multiple services.

One binary, three route groups: `/portal/*` for the human-facing pages, `/api/*` for the JSON (JavaScript Object Notation) endpoints, `/admin/*` for the internal panel. The route split is a prefix on one mux, not three deployments.

Every weakness is a toggle in `lab.yaml` under `app.weaknesses`. All eleven default to `true`. Flipping one to `false` gives a safe implementation of the same route, not a disabled route. That is what makes the vulnerable-vs-hardened comparison a real comparison. Each vulnerable branch is marked in the source with a `DELIBERATE WEAKNESS: <TOGGLE>` comment, so `grep -rn "DELIBERATE WEAKNESS" internal/` finds all of them.

The service vulnerabilities are:
1. sqli in login
2. sqli in search (a UNION SELECT dumps the api_keys table)
3. command injection in the admin diagnostics page
4. IDOR on documents and invoices (select by id, nothing else)
5. broken admin authorization (the `X-Clayface-Role: admin` header is trusted)
6. path traversal in the file download
7. hardcoded database credential in the source
8. forgeable session token (base64 of user|role, no signature)
9. service account credential stored in plaintext in the database
10. SSRF in the url-fetch endpoint
11. verbose errors and an unauthenticated debug endpoint

The ninth one is the important one for the thesis. It plants `svc-app-portal`'s password next to the API keys. Recover it from a dump and it is a working domain credential on `clayface.local`. That is the whole app-to-AD (Active Directory) link, since the VM (virtual machine) is not domain-joined.

### `sqli_login` — injection in the login form

sqli is short for SQL injection, and SQL (Structured Query Language) is the language the database is queried in. `POST /portal/login` and `POST /api/login` build the credential lookup by string concatenation, so the username is never a value, it is part of the query. The payload `' OR '1'='1' -- ` makes the query return the first row of an unfiltered scan, and the caller is logged in as that user.

Row one is `ron.weas`, an ordinary user, not an admin. That is on purpose. The injection hands over a session, not the panel, so there is still an escalation to find.

Hardened: the username and password are bound as parameters, so the payload is only ever a string to compare against.

### `sqli_search` — injection in the search box

`GET /portal/search` and `GET /api/search` concatenate the term into `WHERE title LIKE '%<q>%'`. A `UNION SELECT` appends a second query to the first, so the caller reads any table the database user can read. In practice that is `api_keys`, which is the exfiltration primitive of the whole tier.

Hardened: the term is bound as a parameter and searched for literally, so the injection characters are just characters.

### `command_injection_diagnostics` — the admin panel's runner

`GET /admin/diagnostics?host=` interpolates the host into a command run through `sh -c`, so `; id` runs a second command and returns its output. This is the tier's code execution: the container runs as an unprivileged user with almost nothing in it besides the app and busybox, so it is a shell in the container and not on the VM. That boundary is deliberate, it is what makes the finding bounded.

Hardened: `ping` runs directly with no shell, after the host is matched against a hostname pattern.

### `idor_documents` — no ownership check

IDOR is insecure direct object reference. `GET /api/orders/{id}` and `GET /api/invoices/{id}` select the row by primary key alone. No ownership check, and no session needed either. Change the number and you read somebody else's order.

Hardened: a session is required and the ownership predicate is added, so another user's row is a 403.

### `broken_admin_authz` — the client decides its own role

Every `/admin/*` route accepts the header `X-Clayface-Role: admin` as proof of the admin role. One header, added by the caller, opens the whole panel: user list, audit log, diagnostics, config. It is implemented in one wrapper, so the entire route group flips together.

Hardened: the header is ignored and the admin role has to be in the session.

### `path_traversal_download` — out of the documents directory

`GET /portal/download?file=` joins the name onto the documents directory without cleaning it, so `../../etc/passwd` escapes the directory and is served. This one is worth having because it reaches data outside the application's own model, which the other findings do not.

Hardened: separators and `..` are rejected, and the resolved path is confirmed to still be inside the documents directory.

### `hardcoded_db_credentials` — the password is in the source

The database password also exists as a literal in the Go source, used as a fallback when `DB_PASSWORD` is unset. On top of that, `GET /admin/config` and `GET /api/debug/config` print the live password and say where it came from. Two independent ways to the same credential. Anyone who obtains the image, or the source, has a working database password without touching the environment.

Hardened: the value is redacted and `DB_PASSWORD` becomes mandatory, so there is no fallback at all. The literal is still in the compiled binary, because the file has to contain it to be the weakness. What the toggle removes is the use, not the string. That is the honest form of the finding: it is exactly why a hardcoded credential cannot be rotated by configuration.

### `forgeable_session_token` — unsigned cookie

The session cookie is `base64url("username|role|issuedAt")` with no signature. Nothing stops a client from writing their own, so a cookie saying `adm-hermione|admin|<now>` is an administrator session. No injection needed, no password needed. It shows why a server must not trust a value it did not sign.

Hardened: HMAC-SHA256 (hash-based message authentication code, using the Secure Hash Algorithm at 256 bits) over the payload, keyed on `SESSION_SECRET` and compared in constant time. The cookie also becomes `HttpOnly`, `Secure` and `SameSite=Lax`. Known gap: even signed, the payload carries no expiry, so a captured cookie stays valid.

### `service_account_credential_in_db` — the pivot

A row in `api_keys` holds the plaintext `SERVICE_ACCOUNT_USER` and `SERVICE_ACCOUNT_PASSWORD`, labelled as a directory bind credential, with a note saying its rotation is overdue. That account is `svc-app-portal`, a real account in `clayface.local`. This is the weakness that makes the tier part of the AD story instead of a standalone web exercise.

A second row holds `svc-idp-ldap`'s password in the same table, reachable through the search weakness rather than through this one. That is the credential the DMZ boundary lets out, as a bind on `dc01`.

Hardened: the row keeps a bcrypt-shaped placeholder, so the table yields nothing usable.

### `ssrf_url_fetch` — the server fetches for you

SSRF is server-side request forgery. `GET /api/fetch?url=` fetches the URL (Uniform Resource Locator) server-side and returns the body, so the portal makes requests on the attacker's behalf. It reaches the compose network for real, which is where the database and the portal's own admin routes live.

It does not reach the lab LAN. The DMZ boundary from the newer network work denies that direction and logs the attempt, so the SSRF stops at the firewall. An earlier sentence in `docs/app01-design.md` says it proxies into the LAN; that predates the boundary and is no longer true.

Hardened: only `https`, and no loopback, private, link-local or unspecified address literal, checked across redirects too. Known gap: the check reads literals only, so a hostname that resolves into a private range still passes.

### `verbose_errors_debug_endpoint` — four leaks in one toggle

Four separate behaviours sit behind this one toggle.

1. `net/http/pprof`, Go's profiling endpoints, are mounted unauthenticated under `/debug/pprof/*`.
2. A panic renders the stack trace and the effective configuration, database password included.
3. Database errors carry their SQL, so a failed query shows the schema.
4. The leftmost value of `X-Forwarded-For` is trusted as the client address, and nginx keeps whatever the client sent in front of the real address. So the audit log's client field is spoofable.

Hardened: no pprof, no panic page, generic error text, and the client address comes from the connection instead of the header.

### Detection

There is one answer for all eleven, and it is a weak one. The portal writes an `audit_log` table row and a container log line per request, and both are written by the system being attacked. An attacker who can write to the database can edit the first one. There is no SIEM (security information and event management), no log shipping, nothing that reaches an operator. The container log is the better of the two records, and it is still inside the container that just got compromised.

## Nginx

Exposes the service. Nothing special. HTTPS only. let's not make it too easy. even though most services that are served for only inside network, don't need it.

HTTPS is HTTP (Hypertext Transfer Protocol) over TLS (Transport Layer Security), so the traffic is encrypted and there is a certificate to get wrong. That is the reason it is here: the findings of this tier are more legible over a real external service than over plaintext, and the proxy adds a component the attacker has to reason about.

It is also the only thing that publishes host ports. The portal and the database stay on the compose network. So `nmap` finds 80 and 443 and nothing else, and every reachable route goes through the application.

The certificate is self-signed and generated on the VM. A certificate in the repo would be a published private key.

## Database

The database is postgres version 16 alpine. I didn't make the version dynamic
(changeable) from out single source of truth (lab.yaml) because the changes in config file placements and volume handling between versions were breaking changes. The cost of handling those breaking changes actually sort of handicapped the user deploying the lab so i decided to leave it be. If someone wants to use a different version, just change the docker compose!

It holds `users`, `orders`, `invoices`, `documents` and `api_keys`. Passwords are stored as a bare SHA-256 (Secure Hash Algorithm, 256-bit) digest, no salt and no iteration. That one is the baseline, not a toggle. It is what makes a dumped `users` table crackable in seconds, which is what makes the seed worth stealing.

The schema and the seed documents are written once, on first boot, into a named volume. They are never re-synced. Overwriting them would destroy the data an actor is supposed to steal.

## Deployment

Ansible builds the images on the control node and ships them as one tarball. The VM has no route to a registry, so `pull_policy` is `never` and a missing image fails loudly.
Everything is offline. f internet problems!

The VM sits at `10.0.10.10` in the DMZ (demilitarized zone). Only LAN (local area network) users on 443 get in, plus DNS (Domain Name System) and DHCP (Dynamic Host Configuration Protocol), and one LDAP (Lightweight Directory Access Protocol) bind out to `dc01`. The identity side of this is in [[Domain Users]], the VM side in [[Architecture/Infrastructure]], and the long version in `docs/app01-design.md`.
