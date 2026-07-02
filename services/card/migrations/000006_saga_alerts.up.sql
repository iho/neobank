CREATE TABLE IF NOT EXISTS card.saga_alerts (
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_card_saga_alerts_active
    ON card.saga_alerts (saga_instance_id)
    WHERE alert_status IN ('open', 'investigating');

CREATE INDEX IF NOT EXISTS idx_card_saga_alerts_open
    ON card.saga_alerts (alert_status, last_seen_at DESC);