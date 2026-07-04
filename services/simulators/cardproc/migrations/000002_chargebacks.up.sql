ALTER TABLE cardproc.transactions
    ADD COLUMN IF NOT EXISTS expired_at TIMESTAMPTZ;

-- One row per simulated dispute against a captured transaction. Its ID is
-- what the card service tracks as dispute_id.
CREATE TABLE IF NOT EXISTS cardproc.chargebacks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id   UUID NOT NULL REFERENCES cardproc.transactions(id),
    authorization_id TEXT NOT NULL,
    amount           NUMERIC(20,2) NOT NULL CHECK (amount > 0),
    currency         CHAR(3) NOT NULL,
    reason           TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'opened',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cardproc_chargebacks_transaction
    ON cardproc.chargebacks (transaction_id);
