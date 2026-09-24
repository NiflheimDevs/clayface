# Outstanding credential rotation — portal and identity provider (SECRET)

**Owner:** abed.nad
**Last updated:** 2024-09-05
**Classification:** Secret

This note tracks the credentials that are still shared, still static, or still
stored somewhere they should not be. It is in the portal document store, which
in hindsight is the wrong place for it — see "Where these are recorded" below.

## Outstanding items

### 1. Directory bind credential for IDP01 — overdue

The bind account used by IDP01 to read the directory still holds the shared
lab password rather than its own value, and it has not been rotated since
April. The application stores this one in the `api_keys` table in the portal
database, in clear text, because the deployment reads it from there at boot
rather than from a secret store.

Practical consequence: a portal database dump yields a directory credential.
Whoever reads that table can bind to the domain controller with it. This is the
single highest-value item on the list and it has been carried over for three
cycles.

### 2. Portal service account — rotation not automated

The portal's own account (`svc-app-portal`) is configured from the container
environment and is also written into `api_keys` so the other internal services
can find it. Rotation therefore means redeploying the container, which nobody
wants to do on a Friday, so in practice it does not happen.

### 3. Shared password across the human accounts

Every human account in the directory still uses the same value that it had at
build time, unchanged since the environment was created. This is a known and
accepted deviation for the lab, but it means there is no such thing as a
low-value credential here: one recovered password is every account, and the
same value opens the portal, the domain and the file share.

## Where these are recorded

- Portal and provider credentials: the portal database (`api_keys` table)
- Directory accounts: the domain, managed by the identity playbooks
- Anything with a vendor: the contract itself, in this same document store

## Why this keeps slipping

Two reasons, both boring:

1. There is no secret store — the deployment is Docker Compose on one VM, so
   "configuration" means an `.env` file and "rotation" means a redeploy.
2. The credentials are only used by services, so nothing visibly breaks when
   they go stale, and nothing reminds anyone.

The fix is a real secret store with short-lived credentials. Until then this
note is the only record of what is outstanding, and it is readable by anyone
who can reach the portal document store — which, since the download endpoint
is served by filename, is anyone who can guess a filename.
