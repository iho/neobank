CREATE SCHEMA IF NOT EXISTS cardproc;

CREATE TABLE IF NOT EXISTS cardproc.cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref    TEXT NOT NULL,
    cardholder_name TEXT NOT NULL,
    pan_token       TEXT NOT NULL,
    last_four       CHAR(4) NOT NULL,
    expiry_month    SMALLINT NOT NULL,
    expiry_year     SMALLINT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per simulated merchant transaction. authorization_id is filled in
-- once the synchronous auth call to the card service returns an approval,
-- so later async webhooks (capture/reversal) can reference the neobank's
-- own authorization ID rather than this simulator's transaction ID.
CREATE TABLE IF NOT EXISTS cardproc.transactions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id          UUID NOT NULL REFERENCES cardproc.cards(id),
    authorization_id TEXT NOT NULL DEFAULT '',
    amount           NUMERIC(20,2) NOT NULL CHECK (amount > 0),
    currency         CHAR(3) NOT NULL,
    merchant_name    TEXT NOT NULL,
    mcc              TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'pending',
    reason_code      TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    captured_at      TIMESTAMPTZ,
    reversed_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_cardproc_transactions_card
    ON cardproc.transactions (card_id, created_at);

-- Outbound webhook delivery queue backing pkg/vendorsim.DeliveryStore, for
-- the async capture/reversal events only (the auth decision is synchronous).
CREATE TABLE IF NOT EXISTS cardproc.webhook_deliveries (
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

CREATE INDEX IF NOT EXISTS idx_cardproc_webhook_deliveries_due
    ON cardproc.webhook_deliveries (next_attempt_at)
    WHERE delivered_at IS NULL;
