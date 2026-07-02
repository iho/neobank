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