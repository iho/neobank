CREATE SCHEMA IF NOT EXISTS notification;

CREATE TABLE notification.notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    event_id    UUID NOT NULL,
    event_type  TEXT NOT NULL,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL,
    read        BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, user_id)
);

CREATE INDEX idx_notifications_user_created
    ON notification.notifications (user_id, created_at DESC);

-- Event-level dedup for at-least-once Kafka/HTTP delivery.
CREATE TABLE notification.consumer_inbox (
    event_id      UUID PRIMARY KEY,
    event_type    TEXT NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_consumer_inbox_type
    ON notification.consumer_inbox (event_type, processed_at DESC);