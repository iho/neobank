CREATE TABLE IF NOT EXISTS payment.screening_checks (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_type         TEXT NOT NULL,
    subject_user_id    UUID NOT NULL,
    related_user_id    UUID,
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

CREATE INDEX IF NOT EXISTS idx_payment_screening_checks_subject
    ON payment.screening_checks (subject_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_screening_checks_entity
    ON payment.screening_checks (entity_type, entity_id, created_at DESC);