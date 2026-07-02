-- name: CreateTransfer :exec
INSERT INTO payment.transfers (
    id, idempotency_key, type, status, sender_user_id, recipient_user_id, amount, currency, memo
) VALUES (@id, @idempotency_key, @type, @status, @sender_user_id, @recipient_user_id, @amount::numeric, @currency, sqlc.narg(memo));

-- name: GetTransferBySenderAndIdempotencyKey :one
SELECT id, idempotency_key, type, status, sender_user_id, recipient_user_id,
       amount::text AS amount, currency, COALESCE(memo, '') AS memo,
       COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       COALESCE(failure_reason, '') AS failure_reason, created_at, completed_at
FROM payment.transfers
WHERE sender_user_id = $1 AND idempotency_key = $2;

-- name: GetTransferByID :one
SELECT id, idempotency_key, type, status, sender_user_id, recipient_user_id,
       amount::text AS amount, currency, COALESCE(memo, '') AS memo,
       COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       COALESCE(failure_reason, '') AS failure_reason, created_at, completed_at
FROM payment.transfers
WHERE id = $1;

-- name: MarkTransferCompleted :exec
UPDATE payment.transfers
SET status = 'completed', ledger_transfer_id = $2, completed_at = now()
WHERE id = $1;

-- name: MarkTransferFailed :exec
UPDATE payment.transfers
SET status = 'failed', failure_reason = $2, completed_at = now()
WHERE id = $1;

-- name: ListTransfersByUser :many
SELECT id, idempotency_key, type, status, sender_user_id, recipient_user_id,
       amount::text AS amount, currency, COALESCE(memo, '') AS memo,
       COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       COALESCE(failure_reason, '') AS failure_reason, created_at, completed_at
FROM payment.transfers
WHERE (sender_user_id = @user_id OR recipient_user_id = @user_id)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR created_at < sqlc.narg(cursor_created_at)::timestamptz
    OR (created_at = sqlc.narg(cursor_created_at)::timestamptz AND id < sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT @limit_val;