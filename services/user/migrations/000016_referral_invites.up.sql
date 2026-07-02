CREATE TABLE IF NOT EXISTS "user".referral_invites (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inviter_user_id UUID NOT NULL REFERENCES "user".users (id),
    invite_code     TEXT NOT NULL UNIQUE,
    invitee_user_id UUID REFERENCES "user".users (id),
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_referral_invites_inviter
    ON "user".referral_invites (inviter_user_id, created_at DESC);