CREATE TABLE card.disputes (
    id                             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chargeback_id                  TEXT NOT NULL UNIQUE,
    authorization_id               UUID NOT NULL REFERENCES card.authorizations(id),
    card_id                        UUID NOT NULL,
    user_id                        UUID NOT NULL,
    amount                         NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency                       CHAR(3) NOT NULL,
    reason                         TEXT NOT NULL DEFAULT '',
    status                         TEXT NOT NULL DEFAULT 'open',
    provisional_credit_transfer_id TEXT NOT NULL DEFAULT '',
    reversal_transfer_id           TEXT NOT NULL DEFAULT '',
    created_at                     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT card_disputes_status_check CHECK (status IN ('open', 'won', 'lost'))
);

CREATE INDEX idx_card_disputes_authorization
    ON card.disputes (authorization_id);
