-- name: CreateAuthorization :exec
INSERT INTO card.authorizations (
    id, card_id, user_id, idempotency_key, merchant_name, amount, currency, status
) VALUES (
    @id, @card_id, @user_id, @idempotency_key, sqlc.narg(merchant_name), @amount::numeric, @currency, @status
);

-- name: GetAuthorizationByID :one
SELECT id, card_id, user_id, idempotency_key, COALESCE(merchant_name, '') AS merchant_name,
       amount::text AS amount, currency, status,
       COALESCE(ledger_hold_id, '') AS ledger_hold_id,
       COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       COALESCE(failure_reason, '') AS failure_reason, created_at, captured_at
FROM card.authorizations
WHERE id = $1;

-- name: GetAuthorizationByCardAndIdempotencyKey :one
SELECT id, card_id, user_id, idempotency_key, COALESCE(merchant_name, '') AS merchant_name,
       amount::text AS amount, currency, status,
       COALESCE(ledger_hold_id, '') AS ledger_hold_id,
       COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       COALESCE(failure_reason, '') AS failure_reason, created_at, captured_at
FROM card.authorizations
WHERE card_id = $1 AND idempotency_key = $2;

-- name: ListAuthorizationsByUser :many
SELECT id, card_id, user_id, idempotency_key, COALESCE(merchant_name, '') AS merchant_name,
       amount::text AS amount, currency, status,
       COALESCE(ledger_hold_id, '') AS ledger_hold_id,
       COALESCE(ledger_transfer_id, '') AS ledger_transfer_id,
       COALESCE(failure_reason, '') AS failure_reason, created_at, captured_at
FROM card.authorizations
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: MarkAuthorizationHold :exec
UPDATE card.authorizations
SET ledger_hold_id = $2, status = 'authorized'
WHERE id = $1;

-- name: MarkAuthorizationFailed :exec
UPDATE card.authorizations
SET status = 'declined', failure_reason = $2
WHERE id = $1;

-- name: MarkAuthorizationCaptured :exec
UPDATE card.authorizations
SET status = 'captured', ledger_transfer_id = $2, captured_at = now()
WHERE id = $1;