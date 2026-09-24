-- schema.sql — the Clayface Portal application schema.
--
-- Applied once, at first boot, by internal/seed. The tables are a plain
-- business schema: users of the public portal, the customers they serve, the
-- orders and invoices between them, the documents the portal publishes, the
-- API keys other internal services authenticate with, and an audit trail.
--
-- Two things about this file are deliberate and worth saying out loud:
--
--   * The column set is chosen so that the search weakness has somewhere to
--     go. The public search selects (id, title, body) from documents, and
--     api_keys carries (id, service_account, api_key) — three columns of
--     compatible types. That is what makes a UNION SELECT able to append the
--     credential table to a document search, which is the whole point of
--     WEAK_SQLI_SEARCH. It is not an accident of layout.
--
--   * api_keys.service_account is UNIQUE, so the seed can upsert the
--     application service-account row and the WEAK_SERVICE_ACCOUNT_CREDENTIAL_IN_DB
--     toggle can rewrite exactly that one row without touching the others.
--
-- The seed splitter in seed.go reads this file statement by statement and
-- treats a line whose first non-space characters are "--" as a comment, so a
-- full-line comment anywhere is fine but a trailing comment after a statement
-- on the same line is not. Keep comments on their own lines.

CREATE TABLE users (
    id            serial PRIMARY KEY,
    username      text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL DEFAULT 'user',
    full_name     text,
    email         text,
    department    text,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE customers (
    id           serial PRIMARY KEY,
    user_id      integer REFERENCES users (id),
    company      text NOT NULL,
    contact_name text,
    email        text,
    phone        text,
    vat_number   text,
    address_line text,
    city         text,
    country      text,
    iban         text,
    card_number  text,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id           serial PRIMARY KEY,
    customer_id  integer NOT NULL REFERENCES customers (id),
    order_ref    text NOT NULL,
    description  text,
    amount_cents integer NOT NULL DEFAULT 0,
    currency     text NOT NULL DEFAULT 'EUR',
    status       text NOT NULL DEFAULT 'open',
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE invoices (
    id           serial PRIMARY KEY,
    order_id     integer NOT NULL REFERENCES orders (id),
    invoice_ref  text NOT NULL,
    amount_cents integer NOT NULL DEFAULT 0,
    currency     text NOT NULL DEFAULT 'EUR',
    issued_on    date NOT NULL,
    due_on       date NOT NULL,
    paid         boolean NOT NULL DEFAULT false,
    notes        text
);

CREATE TABLE documents (
    id             serial PRIMARY KEY,
    title          text NOT NULL,
    filename       text NOT NULL,
    classification text NOT NULL DEFAULT 'internal',
    owner_id       integer REFERENCES users (id),
    body           text,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
    id               serial PRIMARY KEY,
    key_name         text NOT NULL,
    service_account  text NOT NULL UNIQUE,
    api_key          text NOT NULL,
    credential_type  text NOT NULL DEFAULT 'api-key',
    notes            text,
    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_log (
    id          serial PRIMARY KEY,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    username    text,
    action      text NOT NULL,
    detail      text,
    client_ip   text
);

CREATE INDEX orders_customer_id_idx ON orders (customer_id);

CREATE INDEX invoices_order_id_idx ON invoices (order_id);

CREATE INDEX documents_filename_idx ON documents (filename);

CREATE INDEX audit_log_occurred_at_idx ON audit_log (occurred_at);
