-- name: ListSavedPayeesByUser :many
SELECT sp.id, sp.user_id, sp.payee_user_id, COALESCE(sp.nickname, '') AS nickname,
       sp.last_used_at, sp.created_at,
       COALESCE(u.email, '') AS payee_email, COALESCE(u.phone, '') AS payee_phone
FROM "user".saved_payees sp
JOIN "user".users u ON u.id = sp.payee_user_id
WHERE sp.user_id = $1
ORDER BY sp.last_used_at DESC
LIMIT $2;

-- name: UpsertSavedPayee :one
INSERT INTO "user".saved_payees (user_id, payee_user_id, nickname, last_used_at)
VALUES ($1, $2, sqlc.narg(nickname), now())
ON CONFLICT (user_id, payee_user_id) DO UPDATE SET
    last_used_at = now(),
    nickname = COALESCE(EXCLUDED.nickname, "user".saved_payees.nickname)
RETURNING id, user_id, payee_user_id, COALESCE(nickname, '') AS nickname, last_used_at, created_at;

-- name: CreateSavedPayee :one
INSERT INTO "user".saved_payees (user_id, payee_user_id, nickname)
VALUES ($1, $2, sqlc.narg(nickname))
ON CONFLICT (user_id, payee_user_id) DO NOTHING
RETURNING id, user_id, payee_user_id, COALESCE(nickname, '') AS nickname, last_used_at, created_at;

-- name: GetSavedPayeeByID :one
SELECT id, user_id, payee_user_id, COALESCE(nickname, '') AS nickname, last_used_at, created_at
FROM "user".saved_payees
WHERE id = $1 AND user_id = $2;

-- name: DeleteSavedPayee :execrows
DELETE FROM "user".saved_payees
WHERE id = $1 AND user_id = $2;