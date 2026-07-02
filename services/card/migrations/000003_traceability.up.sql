ALTER TABLE card.outbox_events
    ADD COLUMN IF NOT EXISTS event_version INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS correlation_id TEXT,
    ADD COLUMN IF NOT EXISTS causation_id TEXT;

CREATE TABLE IF NOT EXISTS card.audit_log (
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

CREATE INDEX IF NOT EXISTS idx_card_audit_log_entity
    ON card.audit_log (entity_type, entity_id, created_at);

CREATE TABLE IF NOT EXISTS card.fraud_decisions (
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
    correlation_id    TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_card_fraud_decisions_user
    ON card.fraud_decisions (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS card.reconciliation_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    started_at      TIMESTAMPTZ NOT NULL,
    finished_at     TIMESTAMPTZ,
    checked_count   INT NOT NULL DEFAULT 0,
    break_count     INT NOT NULL DEFAULT 0,
    breaks          JSONB NOT NULL DEFAULT '[]',
    status          TEXT NOT NULL DEFAULT 'running'
);
