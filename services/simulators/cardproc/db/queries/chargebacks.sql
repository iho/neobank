-- name: CreateChargeback :one
INSERT INTO cardproc.chargebacks (transaction_id, authorization_id, amount, currency, reason)
VALUES (@transaction_id, @authorization_id, @amount::numeric, @currency, @reason)
RETURNING id, transaction_id, authorization_id, amount::text AS amount, currency, reason,
          status, created_at, updated_at;

-- name: GetChargebackByID :one
SELECT id, transaction_id, authorization_id, amount::text AS amount, currency, reason,
       status, created_at, updated_at
FROM cardproc.chargebacks
WHERE id = $1;

-- name: SetChargebackStatus :one
UPDATE cardproc.chargebacks
SET status = @status, updated_at = now()
WHERE id = @id
RETURNING id, transaction_id, authorization_id, amount::text AS amount, currency, reason,
          status, created_at, updated_at;
