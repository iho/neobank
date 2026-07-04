-- name: CreateOutboundPayment :one
INSERT INTO rails.outbound_payments (account_id, amount, currency, counterparty_iban, reference)
VALUES (@account_id, @amount::numeric, @currency, @counterparty_iban, @reference)
RETURNING id, account_id, amount::text AS amount, currency, counterparty_iban, reference, status, created_at, updated_at;

-- name: GetOutboundPaymentByID :one
SELECT id, account_id, amount::text AS amount, currency, counterparty_iban, reference, status, created_at, updated_at
FROM rails.outbound_payments
WHERE id = $1;

-- name: SetOutboundPaymentStatus :exec
UPDATE rails.outbound_payments
SET status = @status, updated_at = now()
WHERE id = @id;

-- name: ListOutboundPaymentsInRange :many
SELECT id, account_id, amount::text AS amount, currency, counterparty_iban, reference, status, created_at, updated_at
FROM rails.outbound_payments
WHERE created_at >= @from_ts AND created_at < @to_ts
ORDER BY created_at ASC;
