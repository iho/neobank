ALTER TABLE "user".outbox_events
    ADD COLUMN IF NOT EXISTS event_version INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS correlation_id TEXT,
    ADD COLUMN IF NOT EXISTS causation_id TEXT;

CREATE TABLE IF NOT EXISTS "user".audit_log (
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

CREATE INDEX IF NOT EXISTS idx_user_audit_log_entity
    ON "user".audit_log (entity_type, entity_id, created_at);
