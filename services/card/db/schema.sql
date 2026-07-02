CREATE SCHEMA card;

CREATE TABLE card.cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    wallet_id       UUID NOT NULL,
    processor_ref   TEXT,
    pan_token       TEXT NOT NULL,
    last_four       CHAR(4) NOT NULL,
    expiry_month    SMALLINT NOT NULL,
    expiry_year     SMALLINT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',
    idempotency_key TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);

CREATE TABLE card.outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);

CREATE TABLE card.authorizations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id             UUID NOT NULL REFERENCES card.cards(id),
    user_id             UUID NOT NULL,
    idempotency_key     TEXT NOT NULL,
    merchant_name       TEXT,
    amount              NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency            CHAR(3) NOT NULL DEFAULT 'USD',
    status              TEXT NOT NULL DEFAULT 'authorized',
    ledger_hold_id      TEXT,
    ledger_transfer_id  TEXT,
    failure_reason      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    captured_at         TIMESTAMPTZ,
    UNIQUE (card_id, idempotency_key)
);

CREATE TABLE card.saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    context         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);