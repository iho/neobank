CREATE TABLE IF NOT EXISTS payment.aml_evaluations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type       TEXT NOT NULL,
    entity_id         TEXT NOT NULL,
    user_id           UUID NOT NULL,
    transaction_type  TEXT NOT NULL,
    amount            NUMERIC(20,8) NOT NULL,
    currency          CHAR(3) NOT NULL,
    disposition       TEXT NOT NULL,
    reason_code       TEXT NOT NULL,
    risk_score        INT NOT NULL,
    rule_set_version  TEXT NOT NULL DEFAULT 'mvp-1.0.0',
    correlation_id    TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payment_aml_evaluations_user
    ON payment.aml_evaluations (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_aml_evaluations_entity
    ON payment.aml_evaluations (entity_type, entity_id, created_at DESC);

CREATE TABLE IF NOT EXISTS payment.aml_cases (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    evaluation_id     UUID NOT NULL REFERENCES payment.aml_evaluations(id),
    user_id           UUID NOT NULL,
    entity_type       TEXT NOT NULL,
    entity_id         TEXT NOT NULL,
    case_type         TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'open',
    reason_code       TEXT NOT NULL,
    filing_reference  TEXT,
    correlation_id    TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at         TIMESTAMPTZ,
    CONSTRAINT payment_aml_cases_status_check
        CHECK (status IN ('open', 'filed', 'closed')),
    CONSTRAINT payment_aml_cases_type_check
        CHECK (case_type IN ('ctr', 'sar', 'review'))
);

CREATE INDEX IF NOT EXISTS idx_payment_aml_cases_status
    ON payment.aml_cases (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_aml_cases_user
    ON payment.aml_cases (user_id, created_at DESC);