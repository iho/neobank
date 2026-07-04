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
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment.outbox_publications (
    event_id      UUID PRIMARY KEY REFERENCES payment.outbox_events(id),
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_outbox_publications_published_at
    ON payment.outbox_publications (published_at DESC);

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

CREATE TABLE payment.velocity_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    amount      NUMERIC(20,8) NOT NULL CHECK (amount >= 0),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_velocity_events_user_recorded
    ON payment.velocity_events (user_id, recorded_at DESC);

-- Local mirror of the virtual IBAN issued by the rails simulator (or, later,
-- a real payment rail) for a user's wallet in a given currency.
CREATE TABLE payment.bank_accounts (
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
CREATE TABLE payment.bank_transfers (
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

CREATE INDEX idx_payment_bank_transfers_user
    ON payment.bank_transfers (user_id, created_at DESC);

-- One row per executed FX conversion, keyed by the fx simulator's quote ID
-- so re-executing the same quote (a retried HTTP call) is a no-op rather
-- than a second conversion.
CREATE TABLE payment.fx_conversions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id               TEXT NOT NULL UNIQUE,
    user_id                UUID NOT NULL,
    from_currency          CHAR(3) NOT NULL,
    to_currency            CHAR(3) NOT NULL,
    amount                 NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    converted_amount       NUMERIC(20,8) NOT NULL CHECK (converted_amount > 0),
    rate                   NUMERIC(20,8) NOT NULL,
    from_ledger_transfer_id TEXT NOT NULL,
    to_ledger_transfer_id   TEXT NOT NULL,
    status                 TEXT NOT NULL DEFAULT 'completed',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_fx_conversions_user
    ON payment.fx_conversions (user_id, created_at DESC);

-- One row per neobank-initiated outbound bank transfer, keyed by the rails
-- simulator's payment ID so a redelivered settle/return webhook is a no-op.
CREATE TABLE payment.bank_transfer_orders (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rails_payment_id   TEXT NOT NULL UNIQUE,
    user_id            UUID NOT NULL,
    amount             NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency           CHAR(3) NOT NULL,
    counterparty_iban  TEXT NOT NULL,
    reference          TEXT NOT NULL DEFAULT '',
    ledger_transfer_id TEXT NOT NULL,
    return_transfer_id TEXT,
    status             TEXT NOT NULL DEFAULT 'processing',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payment_bank_transfer_orders_status_check
        CHECK (status IN ('processing', 'settled', 'returned', 'failed'))
);

CREATE INDEX idx_payment_bank_transfer_orders_user
    ON payment.bank_transfer_orders (user_id, created_at DESC);
