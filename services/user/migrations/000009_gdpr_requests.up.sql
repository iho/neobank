CREATE TABLE IF NOT EXISTS "user".gdpr_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES "user".users(id),
    request_type    TEXT NOT NULL,
    actor           TEXT NOT NULL DEFAULT 'system',
    correlation_id  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_gdpr_requests_type_check
        CHECK (request_type IN ('export', 'mask'))
);

CREATE INDEX IF NOT EXISTS idx_user_gdpr_requests_user
    ON "user".gdpr_requests (user_id, created_at DESC);