ALTER TABLE "user".kyc_cases
    ADD COLUMN IF NOT EXISTS decided_by TEXT;

CREATE TABLE IF NOT EXISTS "user".kyc_submissions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kyc_case_id        UUID NOT NULL REFERENCES "user".kyc_cases(id),
    user_id            UUID NOT NULL REFERENCES "user".users(id),
    document_type      TEXT,
    document_number    TEXT,
    provider           TEXT NOT NULL,
    provider_reference TEXT,
    provider_response  JSONB NOT NULL DEFAULT '{}',
    screening_decision TEXT NOT NULL,
    screening_reason   TEXT,
    correlation_id     TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_kyc_submissions_case
    ON "user".kyc_submissions (kyc_case_id, created_at DESC);

CREATE TABLE IF NOT EXISTS "user".screening_checks (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_type         TEXT NOT NULL,
    subject_user_id    UUID NOT NULL,
    entity_type        TEXT NOT NULL,
    entity_id          TEXT NOT NULL,
    decision           TEXT NOT NULL,
    reason_code        TEXT NOT NULL,
    provider           TEXT NOT NULL,
    provider_reference TEXT,
    raw_response       JSONB NOT NULL DEFAULT '{}',
    correlation_id     TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_screening_checks_subject
    ON "user".screening_checks (subject_user_id, created_at DESC);