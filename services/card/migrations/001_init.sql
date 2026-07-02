CREATE SCHEMA IF NOT EXISTS card;

CREATE TABLE IF NOT EXISTS card.cards (
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

CREATE TABLE IF NOT EXISTS card.outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_card_outbox_unpublished
    ON card.outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS card.saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    context         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);