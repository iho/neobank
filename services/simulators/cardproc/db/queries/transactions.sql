-- name: CreateTransaction :one
INSERT INTO cardproc.transactions (card_id, amount, currency, merchant_name, mcc)
VALUES (@card_id, @amount::numeric, @currency, @merchant_name, @mcc)
RETURNING id, card_id, authorization_id, amount::text AS amount, currency, merchant_name, mcc,
          status, reason_code, created_at, captured_at, reversed_at;

-- name: GetTransactionByID :one
SELECT id, card_id, authorization_id, amount::text AS amount, currency, merchant_name, mcc,
       status, reason_code, created_at, captured_at, reversed_at
FROM cardproc.transactions
WHERE id = $1;

-- name: SetTransactionAuthResult :one
UPDATE cardproc.transactions
SET status = @status, authorization_id = @authorization_id, reason_code = @reason_code
WHERE id = @id
RETURNING id, card_id, authorization_id, amount::text AS amount, currency, merchant_name, mcc,
          status, reason_code, created_at, captured_at, reversed_at;

-- name: MarkTransactionCaptured :exec
UPDATE cardproc.transactions SET status = 'captured', captured_at = now() WHERE id = $1;

-- name: MarkTransactionReversed :exec
UPDATE cardproc.transactions SET status = 'reversed', reversed_at = now() WHERE id = $1;
