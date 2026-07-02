CREATE TABLE IF NOT EXISTS "user".device_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES "user".users (id),
    platform   TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    token      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, token)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user
    ON "user".device_tokens (user_id);