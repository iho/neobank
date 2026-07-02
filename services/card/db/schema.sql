CREATE SCHEMA card;

CREATE TABLE card.cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    wallet_id       UUID NOT NULL,
    processor_ref   TEXT,
    pan_token       TEXT NOT NULL,
    last_four       CHAR(4) NOT NULL,
    expiry_month    SMALLINT NOT NULL,
    expiry_year     SMALLINT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',
    idempotency_key TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);

CREATE TABLE card.outbox_events (
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

CREATE TABLE card.outbox_publications (
    event_id      UUID PRIMARY KEY REFERENCES card.outbox_events(id),
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_card_outbox_publications_published_at
    ON card.outbox_publications (published_at DESC);

CREATE TABLE card.authorizations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id             UUID NOT NULL REFERENCES card.cards(id),
    user_id             UUID NOT NULL,
    idempotency_key     TEXT NOT NULL,
    merchant_name       TEXT,
    amount              NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency            CHAR(3) NOT NULL DEFAULT 'USD',
    status              TEXT NOT NULL DEFAULT 'authorized',
    ledger_hold_id      TEXT,
    ledger_transfer_id  TEXT,
    failure_reason      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    captured_at         TIMESTAMPTZ,
    UNIQUE (card_id, idempotency_key)
);

CREATE INDEX idx_authorizations_user_created
    ON card.authorizations (user_id, created_at DESC);

CREATE TABLE card.saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'running',
    completed_steps JSONB NOT NULL DEFAULT '{}',
    context         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE card.saga_alerts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_instance_id  UUID NOT NULL REFERENCES card.saga_instances(id),
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
    CONSTRAINT card_saga_alerts_status_check
        CHECK (alert_status IN ('open', 'investigating', 'resolved'))
);

CREATE UNIQUE INDEX idx_card_saga_alerts_active
    ON card.saga_alerts (saga_instance_id)
    WHERE alert_status IN ('open', 'investigating');

-- Append-only lifecycle trail for cards and authorizations.
CREATE TABLE card.audit_log (
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

CREATE INDEX idx_card_audit_log_entity
    ON card.audit_log (entity_type, entity_id, created_at);

-- Every fraud evaluation on a card authorization, allow or deny.
CREATE TABLE card.fraud_decisions (
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

CREATE INDEX idx_card_fraud_decisions_user
    ON card.fraud_decisions (user_id, created_at DESC);

-- Records of reconciliation sweeps between card.authorizations and goledger holds.
CREATE TABLE card.reconciliation_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    started_at      TIMESTAMPTZ NOT NULL,
    finished_at     TIMESTAMPTZ,
    checked_count   INT NOT NULL DEFAULT 0,
    break_count     INT NOT NULL DEFAULT 0,
    breaks          JSONB NOT NULL DEFAULT '[]',
    status          TEXT NOT NULL DEFAULT 'running'
);

CREATE TABLE card.reconciliation_breaks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID NOT NULL REFERENCES card.reconciliation_runs(id),
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
    CONSTRAINT card_reconciliation_breaks_status_check
        CHECK (status IN ('open', 'investigated', 'closed'))
);

CREATE UNIQUE INDEX idx_card_reconciliation_breaks_active
    ON card.reconciliation_breaks (entity_type, entity_id, reason)
    WHERE status IN ('open', 'investigated');

CREATE INDEX idx_card_reconciliation_breaks_status
    ON card.reconciliation_breaks (status, created_at DESC);
