CREATE TABLE IF NOT EXISTS rails.outbound_payments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id         UUID NOT NULL REFERENCES rails.accounts(id),
    amount             NUMERIC(20,2) NOT NULL CHECK (amount > 0),
    currency           CHAR(3) NOT NULL,
    counterparty_iban  TEXT NOT NULL,
    reference          TEXT NOT NULL DEFAULT '',
    status             TEXT NOT NULL DEFAULT 'accepted',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rails_outbound_payments_account
    ON rails.outbound_payments (account_id, created_at);

CREATE INDEX IF NOT EXISTS idx_rails_outbound_payments_created_at
    ON rails.outbound_payments (created_at);
