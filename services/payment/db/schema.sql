CREATE SCHEMA payment;

CREATE TABLE payment.transfers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key     TEXT NOT NULL,
    type                TEXT NOT NULL DEFAULT 'p2p',
    status              TEXT NOT NULL DEFAULT 'pending',
    sender_user_id      UUID NOT NULL,
    recipient_user_id   UUID NOT NULL,
    amount              NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency            CHAR(3) NOT NULL,
    memo                TEXT,
    ledger_transfer_id  TEXT,
    failure_reason      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    UNIQUE (sender_user_id, idempotency_key)
);

CREATE TABLE payment.outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);

CREATE TABLE payment.saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    context         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);