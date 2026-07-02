CREATE TABLE IF NOT EXISTS "user".outbox_publications (
    event_id      UUID PRIMARY KEY REFERENCES "user".outbox_events(id),
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO "user".outbox_publications (event_id, published_at)
SELECT id, published_at
FROM "user".outbox_events
WHERE published_at IS NOT NULL
ON CONFLICT (event_id) DO NOTHING;

ALTER TABLE "user".outbox_events DROP COLUMN IF EXISTS published_at;

DROP INDEX IF EXISTS idx_user_outbox_unpublished;

CREATE INDEX IF NOT EXISTS idx_user_outbox_publications_published_at
    ON "user".outbox_publications (published_at DESC);

CREATE OR REPLACE FUNCTION "user".outbox_events_immutable()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'user.outbox_events is append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS outbox_events_no_update ON "user".outbox_events;
CREATE TRIGGER outbox_events_no_update
    BEFORE UPDATE ON "user".outbox_events
    FOR EACH ROW EXECUTE FUNCTION "user".outbox_events_immutable();

DROP TRIGGER IF EXISTS outbox_events_no_delete ON "user".outbox_events;
CREATE TRIGGER outbox_events_no_delete
    BEFORE DELETE ON "user".outbox_events
    FOR EACH ROW EXECUTE FUNCTION "user".outbox_events_immutable();