-- name: UpsertDeviceToken :one
INSERT INTO "user".device_tokens (user_id, platform, token)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, token) DO UPDATE
SET platform = EXCLUDED.platform, updated_at = now()
RETURNING id, user_id, platform, token, created_at, updated_at;

-- name: DeleteDeviceToken :execrows
DELETE FROM "user".device_tokens
WHERE user_id = $1 AND id = $2;

-- name: ListDeviceTokensByUser :many
SELECT id, user_id, platform, token, created_at, updated_at
FROM "user".device_tokens
WHERE user_id = $1
ORDER BY updated_at DESC;