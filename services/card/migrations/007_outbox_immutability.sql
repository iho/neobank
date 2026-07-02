CREATE TABLE IF NOT EXISTS card.outbox_publications (
    event_id      UUID PRIMARY KEY REFERENCES card.outbox_events(id),
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO card.outbox_publications (event_id, published_at)
SELECT id, published_at
FROM card.outbox_events
WHERE published_at IS NOT NULL
ON CONFLICT (event_id) DO NOTHING;

ALTER TABLE card.outbox_events DROP COLUMN IF EXISTS published_at;

DROP INDEX IF EXISTS idx_card_outbox_unpublished;

CREATE INDEX IF NOT EXISTS idx_card_outbox_publications_published_at
    ON card.outbox_publications (published_at DESC);

CREATE OR REPLACE FUNCTION card.outbox_events_immutable()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'card.outbox_events is append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS outbox_events_no_update ON card.outbox_events;
CREATE TRIGGER outbox_events_no_update
    BEFORE UPDATE ON card.outbox_events
    FOR EACH ROW EXECUTE FUNCTION card.outbox_events_immutable();

DROP TRIGGER IF EXISTS outbox_events_no_delete ON card.outbox_events;
CREATE TRIGGER outbox_events_no_delete
    BEFORE DELETE ON card.outbox_events
    FOR EACH ROW EXECUTE FUNCTION card.outbox_events_immutable();