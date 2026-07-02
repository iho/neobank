CREATE TABLE IF NOT EXISTS notification.notification_preferences (
    user_id    UUID PRIMARY KEY,
    transfers  BOOLEAN NOT NULL DEFAULT true,
    cards      BOOLEAN NOT NULL DEFAULT true,
    kyc        BOOLEAN NOT NULL DEFAULT true,
    push       BOOLEAN NOT NULL DEFAULT true,
    email      BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);