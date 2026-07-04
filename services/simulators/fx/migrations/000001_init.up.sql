CREATE SCHEMA IF NOT EXISTS fx;

-- Every quote a caller requested, whether or not it was ever executed —
-- this is the audit trail an FX vendor would show a regulator.
CREATE TABLE IF NOT EXISTS fx.quotes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency     CHAR(3) NOT NULL,
    to_currency       CHAR(3) NOT NULL,
    amount            NUMERIC(20,2) NOT NULL CHECK (amount > 0),
    converted_amount  NUMERIC(20,2) NOT NULL CHECK (converted_amount > 0),
    rate              NUMERIC(20,8) NOT NULL,
    spread_bps        INT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'pending',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at        TIMESTAMPTZ NOT NULL,
    executed_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fx_quotes_status
    ON fx.quotes (status, created_at);
