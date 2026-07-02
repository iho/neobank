-- name: InsertWalletTransaction :exec
INSERT INTO "user".wallet_transactions (
    user_id, id, source_event_id, tx_type, amount, currency, direction, status,
    counterparty, memo, created_at, updated_at
) VALUES (
    @user_id, @id, @source_event_id, @tx_type, @amount, @currency, @direction, @status,
    @counterparty, @memo, @created_at, @created_at
)
ON CONFLICT (user_id, id) DO NOTHING;

-- name: UpsertWalletTransactionCapture :exec
INSERT INTO "user".wallet_transactions (
    user_id, id, source_event_id, tx_type, amount, currency, direction, status,
    counterparty, created_at, updated_at
) VALUES (
    @user_id, @id, @source_event_id, @tx_type, @amount, @currency, 'debit', @status,
    @counterparty, @created_at, @created_at
)
ON CONFLICT (user_id, id) DO UPDATE SET
    tx_type = EXCLUDED.tx_type,
    amount = EXCLUDED.amount,
    currency = EXCLUDED.currency,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.created_at;

-- name: ListWalletTransactionsByUser :many
SELECT id, tx_type, amount, currency, direction, status, counterparty, memo, created_at
FROM "user".wallet_transactions
WHERE user_id = @user_id
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR created_at < sqlc.narg(cursor_created_at)::timestamptz
    OR (created_at = sqlc.narg(cursor_created_at)::timestamptz AND id < sqlc.narg(cursor_id))
  )
ORDER BY created_at DESC, id DESC
LIMIT @limit_val;