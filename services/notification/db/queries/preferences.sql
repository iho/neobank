-- name: GetNotificationPreferences :one
SELECT user_id, transfers, cards, kyc, push, email, updated_at
FROM notification.notification_preferences
WHERE user_id = $1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification.notification_preferences (user_id, transfers, cards, kyc, push, email)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id) DO UPDATE
SET transfers = EXCLUDED.transfers,
    cards = EXCLUDED.cards,
    kyc = EXCLUDED.kyc,
    push = EXCLUDED.push,
    email = EXCLUDED.email,
    updated_at = now()
RETURNING user_id, transfers, cards, kyc, push, email, updated_at;