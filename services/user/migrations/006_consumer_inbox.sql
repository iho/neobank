CREATE TABLE IF NOT EXISTS "user".consumer_inbox (
    event_id      UUID PRIMARY KEY,
    event_type    TEXT NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_consumer_inbox_type
    ON "user".consumer_inbox (event_type, processed_at DESC);