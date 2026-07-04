CREATE SCHEMA IF NOT EXISTS kyc;

CREATE TABLE IF NOT EXISTS kyc.applicants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref  TEXT NOT NULL,
    full_name     TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    country_code  CHAR(2) NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    reason        TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_kyc_applicants_status
    ON kyc.applicants (status, created_at);

-- Outbound webhook delivery queue backing pkg/vendorsim.DeliveryStore, for
-- the "kyc.check.completed" verdict webhook.
CREATE TABLE IF NOT EXISTS kyc.webhook_deliveries (
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

CREATE INDEX IF NOT EXISTS idx_kyc_webhook_deliveries_due
    ON kyc.webhook_deliveries (next_attempt_at)
    WHERE delivered_at IS NULL;
