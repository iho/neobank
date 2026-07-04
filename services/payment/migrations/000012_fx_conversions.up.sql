-- One row per executed FX conversion, keyed by the fx simulator's quote ID
-- so re-executing the same quote (a retried HTTP call) is a no-op rather
-- than a second conversion.
CREATE TABLE IF NOT EXISTS payment.fx_conversions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id               TEXT NOT NULL UNIQUE,
    user_id                UUID NOT NULL,
    from_currency          CHAR(3) NOT NULL,
    to_currency            CHAR(3) NOT NULL,
    amount                 NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    converted_amount       NUMERIC(20,8) NOT NULL CHECK (converted_amount > 0),
    rate                   NUMERIC(20,8) NOT NULL,
    from_ledger_transfer_id TEXT NOT NULL,
    to_ledger_transfer_id   TEXT NOT NULL,
    status                 TEXT NOT NULL DEFAULT 'completed',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payment_fx_conversions_user
    ON payment.fx_conversions (user_id, created_at DESC);
