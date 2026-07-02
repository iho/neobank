-- name: CreateReferralInvite :one
INSERT INTO "user".referral_invites (inviter_user_id, invite_code)
VALUES ($1, $2)
RETURNING id, inviter_user_id, invite_code, invitee_user_id, status, created_at, accepted_at;

-- name: ListReferralInvitesByInviter :many
SELECT id, inviter_user_id, invite_code, invitee_user_id, status, created_at, accepted_at
FROM "user".referral_invites
WHERE inviter_user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetReferralInviteByCode :one
SELECT id, inviter_user_id, invite_code, invitee_user_id, status, created_at, accepted_at
FROM "user".referral_invites
WHERE invite_code = $1;

-- name: AcceptReferralInvite :one
UPDATE "user".referral_invites
SET invitee_user_id = $2, status = 'accepted', accepted_at = now()
WHERE invite_code = $1 AND status = 'pending'
RETURNING id, inviter_user_id, invite_code, invitee_user_id, status, created_at, accepted_at;