CREATE TABLE IF NOT EXISTS "user".saved_payees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    payee_user_id   UUID NOT NULL,
    nickname        TEXT,
    last_used_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, payee_user_id)
);

CREATE INDEX IF NOT EXISTS idx_saved_payees_user_last_used
    ON "user".saved_payees (user_id, last_used_at DESC);