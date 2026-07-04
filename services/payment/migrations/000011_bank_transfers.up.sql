-- Local mirror of the virtual IBAN issued by the rails simulator (or, later,
-- a real payment rail) for a user's wallet in a given currency.
CREATE TABLE IF NOT EXISTS payment.bank_accounts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL,
    currency          CHAR(3) NOT NULL,
    rails_account_id  TEXT NOT NULL,
    iban              TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency)
);

-- One row per inbound rails transfer processed, keyed by the simulator's
-- transfer ID so a redelivered/duplicate webhook is a no-op rather than a
-- second credit.
CREATE TABLE IF NOT EXISTS payment.bank_transfers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rails_transfer_id   TEXT NOT NULL UNIQUE,
    user_id             UUID NOT NULL,
    amount              NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency            CHAR(3) NOT NULL,
    sender_name         TEXT NOT NULL,
    reference           TEXT NOT NULL DEFAULT '',
    ledger_transfer_id  TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'completed',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payment_bank_transfers_user
    ON payment.bank_transfers (user_id, created_at DESC);
