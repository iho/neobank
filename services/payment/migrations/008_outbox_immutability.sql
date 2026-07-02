CREATE TABLE IF NOT EXISTS payment.outbox_publications (
    event_id      UUID PRIMARY KEY REFERENCES payment.outbox_events(id),
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO payment.outbox_publications (event_id, published_at)
SELECT id, published_at
FROM payment.outbox_events
WHERE published_at IS NOT NULL
ON CONFLICT (event_id) DO NOTHING;

ALTER TABLE payment.outbox_events DROP COLUMN IF EXISTS published_at;

DROP INDEX IF EXISTS idx_payment_outbox_unpublished;

CREATE INDEX IF NOT EXISTS idx_payment_outbox_publications_published_at
    ON payment.outbox_publications (published_at DESC);

CREATE OR REPLACE FUNCTION payment.outbox_events_immutable()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'payment.outbox_events is append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS outbox_events_no_update ON payment.outbox_events;
CREATE TRIGGER outbox_events_no_update
    BEFORE UPDATE ON payment.outbox_events
    FOR EACH ROW EXECUTE FUNCTION payment.outbox_events_immutable();

DROP TRIGGER IF EXISTS outbox_events_no_delete ON payment.outbox_events;
CREATE TRIGGER outbox_events_no_delete
    BEFORE DELETE ON payment.outbox_events
    FOR EACH ROW EXECUTE FUNCTION payment.outbox_events_immutable();