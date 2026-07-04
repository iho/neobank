CREATE SCHEMA IF NOT EXISTS rails;

CREATE TABLE IF NOT EXISTS rails.accounts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref TEXT NOT NULL,
    currency     CHAR(3) NOT NULL,
    iban         TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (external_ref, currency)
);

CREATE TABLE IF NOT EXISTS rails.inbound_transfers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  UUID NOT NULL REFERENCES rails.accounts(id),
    amount      NUMERIC(20,2) NOT NULL CHECK (amount > 0),
    currency    CHAR(3) NOT NULL,
    sender_name TEXT NOT NULL,
    reference   TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'received',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rails_inbound_transfers_account
    ON rails.inbound_transfers (account_id, created_at);

CREATE INDEX IF NOT EXISTS idx_rails_inbound_transfers_created_at
    ON rails.inbound_transfers (created_at);

-- Outbound webhook delivery queue backing pkg/vendorsim.DeliveryStore.
CREATE TABLE IF NOT EXISTS rails.webhook_deliveries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url             TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT NOT NULL DEFAULT '',
    next_attempt_at TIMESTAMPTZ NOT NULL,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rails_webhook_deliveries_due
    ON rails.webhook_deliveries (next_attempt_at)
    WHERE delivered_at IS NULL;
