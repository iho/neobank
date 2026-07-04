-- One row per neobank-initiated outbound bank transfer, keyed by the rails
-- simulator's payment ID so a redelivered settle/return webhook is a no-op.
CREATE TABLE IF NOT EXISTS payment.bank_transfer_orders (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rails_payment_id   TEXT NOT NULL UNIQUE,
    user_id            UUID NOT NULL,
    amount             NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency           CHAR(3) NOT NULL,
    counterparty_iban  TEXT NOT NULL,
    reference          TEXT NOT NULL DEFAULT '',
    ledger_transfer_id TEXT NOT NULL,
    return_transfer_id TEXT,
    status             TEXT NOT NULL DEFAULT 'processing',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payment_bank_transfer_orders_status_check
        CHECK (status IN ('processing', 'settled', 'returned', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_payment_bank_transfer_orders_user
    ON payment.bank_transfer_orders (user_id, created_at DESC);
