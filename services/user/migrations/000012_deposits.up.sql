CREATE TABLE IF NOT EXISTS "user".deposits (
    id                 UUID PRIMARY KEY,
    user_id            UUID NOT NULL REFERENCES "user".users (id),
    wallet_id          UUID NOT NULL,
    amount             NUMERIC(20, 2) NOT NULL,
    currency           TEXT NOT NULL,
    ledger_transfer_id TEXT,
    status             TEXT NOT NULL,
    idempotency_key    TEXT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at       TIMESTAMPTZ,
    UNIQUE (user_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_user_deposits_user_created
    ON "user".deposits (user_id, created_at DESC);