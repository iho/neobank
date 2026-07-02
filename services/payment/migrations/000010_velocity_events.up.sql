CREATE TABLE IF NOT EXISTS payment.velocity_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    amount      NUMERIC(20,8) NOT NULL CHECK (amount >= 0),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_velocity_events_user_recorded
    ON payment.velocity_events (user_id, recorded_at DESC);