CREATE SCHEMA IF NOT EXISTS "user";

CREATE TABLE IF NOT EXISTS "user".users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    phone         TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "user".profiles (
    user_id       UUID PRIMARY KEY REFERENCES "user".users(id),
    full_name     TEXT,
    date_of_birth DATE,
    country_code  CHAR(2)
);

CREATE TABLE IF NOT EXISTS "user".kyc_cases (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES "user".users(id),
    status           TEXT NOT NULL DEFAULT 'pending',
    submitted_at     TIMESTAMPTZ,
    decided_at       TIMESTAMPTZ,
    rejection_reason TEXT
);

CREATE TABLE IF NOT EXISTS "user".wallets (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES "user".users(id),
    currency          CHAR(3) NOT NULL DEFAULT 'USD',
    ledger_account_id TEXT NOT NULL UNIQUE,
    status            TEXT NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency)
);

CREATE TABLE IF NOT EXISTS "user".outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
    ON "user".outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS "user".saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    current_step    TEXT NOT NULL DEFAULT '',
    context         JSONB NOT NULL DEFAULT '{}',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);