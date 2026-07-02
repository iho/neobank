-- Append-only trail for reads of customer PII (profile, KYC, wallet lookups).
CREATE TABLE IF NOT EXISTS "user".pii_access_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_user_id UUID NOT NULL,
    resource        TEXT NOT NULL,
    actor           TEXT NOT NULL DEFAULT 'system',
    correlation_id  TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_pii_access_log_subject
    ON "user".pii_access_log (subject_user_id, created_at);

CREATE INDEX IF NOT EXISTS idx_user_pii_access_log_actor
    ON "user".pii_access_log (actor, created_at);