CREATE SCHEMA payment;

CREATE TABLE payment.transfers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key     TEXT NOT NULL,
    type                TEXT NOT NULL DEFAULT 'p2p',
    status              TEXT NOT NULL DEFAULT 'pending',
    sender_user_id      UUID NOT NULL,
    recipient_user_id   UUID NOT NULL,
    amount              NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency            CHAR(3) NOT NULL,
    memo                TEXT,
    ledger_transfer_id  TEXT,
    failure_reason      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    UNIQUE (sender_user_id, idempotency_key)
);

CREATE TABLE payment.outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    event_version   INT NOT NULL DEFAULT 1,
    payload         JSONB NOT NULL,
    correlation_id  TEXT,
    causation_id    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);

CREATE TABLE payment.saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    context         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment.saga_alerts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_instance_id  UUID NOT NULL REFERENCES payment.saga_instances(id),
    saga_type         TEXT NOT NULL,
    idempotency_key   TEXT NOT NULL,
    instance_status   TEXT NOT NULL,
    alert_status      TEXT NOT NULL DEFAULT 'open',
    stuck_since       TIMESTAMPTZ NOT NULL,
    last_seen_at      TIMESTAMPTZ NOT NULL,
    completed_steps   JSONB NOT NULL DEFAULT '{}',
    context           JSONB NOT NULL DEFAULT '{}',
    resolved_by       TEXT,
    notes             TEXT,
    resolved_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payment_saga_alerts_status_check
        CHECK (alert_status IN ('open', 'investigating', 'resolved'))
);

CREATE UNIQUE INDEX idx_payment_saga_alerts_active
    ON payment.saga_alerts (saga_instance_id)
    WHERE alert_status IN ('open', 'investigating');

-- Append-only lifecycle trail: every state transition on a payment entity is
-- inserted here in the same transaction as the mutation itself, so the
-- destructive UPDATEs on payment.transfers never lose history.
CREATE TABLE payment.audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    action          TEXT NOT NULL,
    from_status     TEXT,
    to_status       TEXT,
    actor           TEXT NOT NULL DEFAULT 'system',
    correlation_id  TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_audit_log_entity
    ON payment.audit_log (entity_type, entity_id, created_at);

-- Every fraud evaluation (allow or deny) is recorded, not just acted on, so
-- disputes and regulators can see what rule fired and why.
CREATE TABLE payment.fraud_decisions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type       TEXT NOT NULL,
    entity_id         TEXT NOT NULL,
    user_id           UUID NOT NULL,
    transaction_type  TEXT NOT NULL,
    amount            NUMERIC(20,8) NOT NULL,
    currency          CHAR(3) NOT NULL,
    decision          TEXT NOT NULL,
    reason_code       TEXT NOT NULL,
    risk_score        INT NOT NULL,
    rule_set_version  TEXT NOT NULL DEFAULT 'mvp-1.0.0',
    correlation_id    TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_fraud_decisions_user
    ON payment.fraud_decisions (user_id, created_at DESC);

CREATE TABLE payment.screening_checks (
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

CREATE INDEX idx_payment_screening_checks_subject
    ON payment.screening_checks (subject_user_id, created_at DESC);

-- Records of reconciliation sweeps between payment.transfers and goledger,
-- so "did we check for drift and what did we find" has an auditable answer.
CREATE TABLE payment.reconciliation_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    started_at      TIMESTAMPTZ NOT NULL,
    finished_at     TIMESTAMPTZ,
    checked_count   INT NOT NULL DEFAULT 0,
    break_count     INT NOT NULL DEFAULT 0,
    breaks          JSONB NOT NULL DEFAULT '[]',
    status          TEXT NOT NULL DEFAULT 'running'
);

-- Individual breaks detected during reconciliation runs, tracked to resolution.
CREATE TABLE payment.reconciliation_breaks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID NOT NULL REFERENCES payment.reconciliation_runs(id),
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    reason          TEXT NOT NULL,
    local_status    TEXT,
    ledger_ref      TEXT,
    status          TEXT NOT NULL DEFAULT 'open',
    resolved_by     TEXT,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ,
    CONSTRAINT payment_reconciliation_breaks_status_check
        CHECK (status IN ('open', 'investigated', 'closed'))
);

CREATE UNIQUE INDEX idx_payment_reconciliation_breaks_active
    ON payment.reconciliation_breaks (entity_type, entity_id, reason)
    WHERE status IN ('open', 'investigated');

CREATE INDEX idx_payment_reconciliation_breaks_status
    ON payment.reconciliation_breaks (status, created_at DESC);

-- Post-transaction AML monitoring: every evaluation is recorded; review/report
-- dispositions open cases for compliance workflows (CTR/SAR export).
CREATE TABLE payment.aml_evaluations (
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

CREATE INDEX idx_payment_aml_evaluations_user
    ON payment.aml_evaluations (user_id, created_at DESC);

CREATE INDEX idx_payment_aml_evaluations_entity
    ON payment.aml_evaluations (entity_type, entity_id, created_at DESC);

CREATE TABLE payment.aml_cases (
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

CREATE INDEX idx_payment_aml_cases_status
    ON payment.aml_cases (status, created_at DESC);

CREATE INDEX idx_payment_aml_cases_user
    ON payment.aml_cases (user_id, created_at DESC);
