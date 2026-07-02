CREATE TABLE IF NOT EXISTS card.authorizations (
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

CREATE INDEX IF NOT EXISTS idx_authorizations_user_created
    ON card.authorizations (user_id, created_at DESC);