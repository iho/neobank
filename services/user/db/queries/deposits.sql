-- name: GetDepositByUserAndIdempotencyKey :one
SELECT id, user_id, wallet_id, amount::text, currency, COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       status, idempotency_key, created_at, completed_at
FROM "user".deposits
WHERE user_id = $1 AND idempotency_key = $2;

-- name: InsertDeposit :exec
INSERT INTO "user".deposits (
    id, user_id, wallet_id, amount, currency, ledger_transfer_id, status, idempotency_key, completed_at
) VALUES ($1, $2, $3, $4, $5, sqlc.narg(ledger_transfer_id), $6, $7, sqlc.narg(completed_at));

-- name: UpdatePasswordHash :exec
UPDATE "user".users
SET password_hash = $2, updated_at = now()
WHERE id = $1;