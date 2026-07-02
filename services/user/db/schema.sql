CREATE SCHEMA "user";

CREATE TABLE "user".users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    phone         TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "user".profiles (
    user_id       UUID PRIMARY KEY REFERENCES "user".users(id),
    full_name     TEXT,
    date_of_birth DATE,
    country_code  CHAR(2)
);

CREATE TABLE "user".kyc_cases (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES "user".users(id),
    status           TEXT NOT NULL DEFAULT 'pending',
    submitted_at     TIMESTAMPTZ,
    decided_at       TIMESTAMPTZ,
    rejection_reason TEXT
);

CREATE TABLE "user".wallets (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES "user".users(id),
    currency          CHAR(3) NOT NULL DEFAULT 'USD',
    ledger_account_id TEXT NOT NULL UNIQUE,
    status            TEXT NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency)
);