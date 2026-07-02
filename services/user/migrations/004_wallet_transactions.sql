CREATE TABLE IF NOT EXISTS "user".wallet_transactions (
    user_id         UUID NOT NULL REFERENCES "user".users(id),
    id              TEXT NOT NULL,
    source_event_id UUID NOT NULL,
    tx_type         TEXT NOT NULL,
    amount          TEXT NOT NULL,
    currency        TEXT NOT NULL,
    direction       TEXT NOT NULL,
    status          TEXT NOT NULL,
    counterparty    TEXT,
    memo            TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, id),
    CONSTRAINT user_wallet_tx_type_check
        CHECK (tx_type IN ('p2p_out', 'p2p_in', 'card_hold', 'card_purchase')),
    CONSTRAINT user_wallet_tx_direction_check
        CHECK (direction IN ('debit', 'credit'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_wallet_tx_event_dedup
    ON "user".wallet_transactions (user_id, source_event_id);

CREATE INDEX IF NOT EXISTS idx_user_wallet_tx_user_created
    ON "user".wallet_transactions (user_id, created_at DESC);