-- seed.sql — the demo data the portal starts life with.
--
-- Applied once, at first boot, immediately after schema.sql. It exists to give
-- the lab something worth stealing: real-looking customers, orders, invoices,
-- documents and API keys, so that an SQL injection dump, an IDOR walk or a
-- path traversal returns material a student can actually interpret instead of
-- placeholder rows.
--
-- Two deliberate properties, and both matter for the thesis:
--
--   1. The identities are the lab's AD identities. ron.weas, bob.sing,
--      abed.nad and adm-hermione are the same four humans that ad.yml creates
--      in the clayface.local forest, and svc-idp-ldap is the same directory
--      bind account. The passwords here are the lab's shared user password
--      (LAB_USER_PASS, Passw0rd!Lab), which is the documented deviation in
--      docs/ad-identity-design.md section 14 — one password for every human —
--      reused on purpose here. That reuse is what turns "dump the portal
--      database" and "authenticate to the domain controller" into one chain
--      rather than two unrelated exercises. The svc-idp-ldap row in api_keys
--      holds that same password in clear text, which is the pivot: recover it
--      from the database, then bind to LDAP on dc01 with it.
--
--   2. Passwords are stored as a bare SHA-256 hex digest, unsalted and
--      uniterated. That is wrong on purpose — it is a fast hash, so a dump is
--      crackable in seconds — but it is not one of the numbered weakness
--      toggles: it is the baseline the login weakness lands on. The toggles
--      change how the lookup is *built*, not how the digest is stored.
--      (Digests below are sha256 of the plaintext named in the comment.)
--
-- PII here is invented. Card numbers are the published test values
-- (4111 1111 1111 1111 and friends) and IBANs are the standard example
-- numbers, so nothing in this file is a real person's data.
--
-- The splitter in seed.go reads this file statement by statement and treats a
-- line whose first non-space characters are "--" as a comment, so keep every
-- comment on its own line. It also splits on a line ending in ";", so do not
-- put a semicolon inside a string literal.

-- Lab users. Passwords: ron.weas / bob.sing / abed.nad / svc-idp-ldap all use
-- Passw0rd!Lab (the lab's shared AD password), adm-hermione uses the same, and
-- the two service accounts have their own. svc-app-portal's credential is not
-- authoritative here: seed.go rewrites that row at every boot from
-- SERVICE_ACCOUNT_PASSWORD.
INSERT INTO users (username, password_hash, role, full_name, email, department)
VALUES
    ('ron.weas', 'd29876cf860fed2758d5a2af60f3f5be3051fef3f72b215dd921e863d159c51d', 'user', 'Ron Weasley', 'ron.weas@clayface.local', 'Sales'),
    ('bob.sing', 'd29876cf860fed2758d5a2af60f3f5be3051fef3f72b215dd921e863d159c51d', 'user', 'Bob Singer', 'bob.sing@clayface.local', 'Engineering'),
    ('abed.nad', 'd29876cf860fed2758d5a2af60f3f5be3051fef3f72b215dd921e863d159c51d', 'user', 'Abed Nadir', 'abed.nad@clayface.local', 'IT'),
    ('adm-hermione', 'd29876cf860fed2758d5a2af60f3f5be3051fef3f72b215dd921e863d159c51d', 'admin', 'Hermione Granger', 'adm-hermione@clayface.local', 'IT'),
    ('svc-app-portal', '0f6593122c39763db231fb9847ad292f14454f9cbc5816e7c9cfdb5b39595e4d', 'service', 'Portal Service Account', 'portal-svc@clayface.local', 'Platform'),
    ('svc-idp-ldap', 'd29876cf860fed2758d5a2af60f3f5be3051fef3f72b215dd921e863d159c51d', 'service', 'IDP01 Directory Bind', 'idp-svc@clayface.local', 'Platform'),
    ('svc-monitoring', '9a0261e8eab1b72afaf48893624dc03e4b407dd1106a67427bad78bbc0eef741', 'service', 'Monitoring Read-Only', 'monitoring@clayface.local', 'Platform');

-- Customers, each owned by the portal user who manages the account. The
-- ownership links are what the IDOR toggle is measured against: ron.weas
-- owns customers 1 to 3, bob.sing owns 4 and 5, abed.nad owns 6.
INSERT INTO customers (user_id, company, contact_name, email, phone, vat_number, address_line, city, country, iban, card_number)
VALUES
    (1, 'Northwind Logistics BV', 'Petra Jansen', 'petra.jansen@northwind.example', '+31 20 555 0142', 'NL812345678B01', 'Havenstraat 44', 'Amsterdam', 'NL', 'NL91ABNA0417164300', '4111111111111111'),
    (1, 'Initech Systems GmbH', 'Dieter Krause', 'd.krause@initech.example', '+49 30 5550 1877', 'DE123456789', 'Friedrichstrasse 12', 'Berlin', 'DE', 'DE89370400440532013000', '5555555555554444'),
    (1, 'Globex Trading SRL', 'Ioana Popescu', 'ioana.popescu@globex.example', '+40 21 555 0193', 'RO12345678', 'Calea Victoriei 155', 'Bucharest', 'RO', 'RO49AAAA1B31007593840000', '4012888888881881'),
    (2, 'Umbrella Facilities Ltd', 'Grace Okonkwo', 'grace.okonkwo@umbrella.example', '+44 20 7946 0321', 'GB123456789', '18 Kingsway', 'London', 'GB', 'GB29NWBK60161331926819', '4111111111111111'),
    (2, 'Vandelay Industries Inc', 'Art Vandelay', 'art.vandelay@vandelay.example', '+1 212 555 0173', 'US987654321', '350 Fifth Avenue', 'New York', 'US', 'GB33BUKB20201555555555', '4222222222222222'),
    (3, 'Stark Components AB', 'Elin Sandberg', 'elin.sandberg@stark.example', '+46 8 555 0110', 'SE556677889901', 'Sveavagen 9', 'Stockholm', 'SE', 'SE3550000000054910000003', '5105105105105100');

-- Orders spread across the customer book, in the various states a real order
-- table accumulates. References are sequential so that a dump reads as a
-- coherent business rather than a row generator.
INSERT INTO orders (customer_id, order_ref, description, amount_cents, currency, status, created_at)
VALUES
    (1, 'CF-2024-0001', 'Pallet handling contract, Q1', 1840000, 'EUR', 'fulfilled', '2024-01-15 09:12:00+00'),
    (1, 'CF-2024-0007', 'Warehouse slot reservation, extension', 465000, 'EUR', 'fulfilled', '2024-03-04 11:40:00+00'),
    (1, 'CF-2024-0022', 'Cold chain audit support', 219900, 'EUR', 'open', '2024-05-21 14:05:00+00'),
    (2, 'CF-2024-0009', 'Internal tooling licence renewal', 742500, 'EUR', 'fulfilled', '2024-02-02 08:30:00+00'),
    (2, 'CF-2024-0031', 'On-call support retainer, H2', 960000, 'EUR', 'open', '2024-06-30 16:20:00+00'),
    (3, 'CF-2024-0014', 'Bulk commodity brokerage fee', 3312500, 'EUR', 'fulfilled', '2024-02-19 10:00:00+00'),
    (3, 'CF-2024-0033', 'Customs paperwork handling', 128750, 'EUR', 'cancelled', '2024-07-08 12:45:00+00'),
    (4, 'CF-2024-0018', 'Facilities maintenance, annual', 1450000, 'GBP', 'fulfilled', '2024-03-27 09:55:00+00'),
    (4, 'CF-2024-0040', 'Emergency callout cover', 385000, 'GBP', 'open', '2024-08-12 07:15:00+00'),
    (5, 'CF-2024-0026', 'Import/export consultancy', 2200000, 'USD', 'fulfilled', '2024-04-30 15:35:00+00'),
    (5, 'CF-2024-0044', 'Late payment penalty reassessment', 75000, 'USD', 'disputed', '2024-08-29 13:10:00+00'),
    (6, 'CF-2024-0037', 'Component sourcing, batch 12', 5175000, 'SEK', 'open', '2024-07-19 10:25:00+00'),
    (6, 'CF-2024-0048', 'Reverse logistics setup', 890000, 'SEK', 'open', '2024-09-02 09:00:00+00');

-- Invoices, one per fulfilled or open order. The unpaid ones are the ones a
-- business actually cares about, which is why several carry a dunning note.
INSERT INTO invoices (order_id, invoice_ref, amount_cents, currency, issued_on, due_on, paid, notes)
VALUES
    (1, 'INV-2024-0101', 1840000, 'EUR', '2024-01-31', '2024-03-01', true, 'Paid 2024-02-14.'),
    (2, 'INV-2024-0118', 465000, 'EUR', '2024-03-15', '2024-04-14', true, 'Paid 2024-04-02.'),
    (3, 'INV-2024-0163', 219900, 'EUR', '2024-05-31', '2024-06-30', false, 'Reminder sent 2024-07-04.'),
    (4, 'INV-2024-0109', 742500, 'EUR', '2024-02-15', '2024-03-16', true, 'Paid 2024-03-01.'),
    (5, 'INV-2024-0187', 960000, 'EUR', '2024-06-30', '2024-07-30', false, 'Payment plan requested.'),
    (6, 'INV-2024-0125', 3312500, 'EUR', '2024-02-29', '2024-03-30', true, 'Paid 2024-03-22 in two parts.'),
    (8, 'INV-2024-0142', 1450000, 'GBP', '2024-03-31', '2024-04-30', true, 'Paid 2024-04-18.'),
    (9, 'INV-2024-0198', 385000, 'GBP', '2024-08-15', '2024-09-14', false, 'Chased by phone 2024-09-16.'),
    (10, 'INV-2024-0156', 2200000, 'USD', '2024-04-30', '2024-05-30', true, 'Paid 2024-05-11, FX variance written off.'),
    (11, 'INV-2024-0203', 75000, 'USD', '2024-08-31', '2024-09-30', false, 'Disputed. Do not chase until resolved.'),
    (12, 'INV-2024-0191', 5175000, 'SEK', '2024-07-31', '2024-08-30', false, 'Split across three purchase orders.');

-- The document register. Every row here has a matching file on disk under
-- DATA_DIR/documents, because the download endpoint serves the file while the
-- search endpoint searches this table's title and body.
INSERT INTO documents (title, filename, classification, owner_id, body)
VALUES
    ('HR memo - September salary review', 'hr-memo-2024-09-salary-review.md', 'confidential', 4, 'Salary review cycle for September 2024. Band adjustments for Sales and Engineering. Contains individual figures and must not be circulated outside the management group.'),
    ('Vendor contract - Northwind Logistics', 'vendor-contract-northwind-logistics.md', 'confidential', 1, 'Master services agreement with Northwind Logistics BV covering pallet handling, cold chain and warehouse slot reservation. Includes payment terms, bank details and the termination clause.'),
    ('Credentials rotation note', 'credentials-rotation-note.md', 'secret', 4, 'Outstanding credential rotation work for the portal and the identity provider. Names the accounts that still share a password and where the bind credential is stored.'),
    ('Datacenter access policy', 'datacenter-access-policy.md', 'internal', 4, 'Physical and logical access rules for the server room and for the hypervisor hosts. Lists who holds keys and who may approve an access request.'),
    ('Customer export 2024 Q2', 'customer-export-2024-Q2.txt', 'confidential', 1, 'Quarterly export of the customer book for the finance reconciliation. One line per customer with contact, VAT number and payment details.');

-- API keys held by the internal services. The svc-idp-ldap row is the one that
-- matters for the lab: it is the directory bind credential, stored in clear
-- text in a table that a UNION SELECT can reach.
INSERT INTO api_keys (key_name, service_account, api_key, credential_type, notes)
VALUES
    ('idp01-ldap-bind', 'svc-idp-ldap', 'Passw0rd!Lab', 'ldap-bind', 'Directory bind credential used by IDP01. Last rotation 2024-04-02, next rotation overdue.'),
    ('monitoring-readonly', 'svc-monitoring', 'Svc-Monitor!2024', 'api-key', 'Read-only Prometheus scrape credential for the portal metrics endpoint.'),
    ('portal-internal-api', 'svc-app-portal', 'PLACEHOLDER-NOT-A-CREDENTIAL', 'application-account', 'Credential for the portal''s own service account. Rewritten at every boot from SERVICE_ACCOUNT_PASSWORD.');

-- A little history, so the audit view is not empty on a fresh deployment and so
-- the login-failure noise a student generates has something to sit next to.
INSERT INTO audit_log (occurred_at, username, action, detail, client_ip)
VALUES
    ('2024-09-16 08:02:11+00', 'ron.weas', 'login.success', 'portal session opened', '10.0.0.20'),
    ('2024-09-16 08:19:47+00', 'ron.weas', 'document.download', 'hr-memo-2024-09-salary-review.md', '10.0.0.20'),
    ('2024-09-16 11:41:03+00', 'bob.sing', 'login.success', 'portal session opened', '10.0.0.20'),
    ('2024-09-17 09:03:52+00', 'abed.nad', 'admin.users.view', 'user list rendered', '10.0.0.20'),
    ('2024-09-17 22:58:19+00', 'Administrator', 'login.failed', 'unknown user', '10.0.0.1'),
    ('2024-09-18 07:31:26+00', 'svc-monitoring', 'api.keys.list', 'read-only scrape', '10.0.0.30');
