-- name: CreateInboundTransfer :one
INSERT INTO rails.inbound_transfers (account_id, amount, currency, sender_name, reference)
VALUES (@account_id, @amount::numeric, @currency, @sender_name, @reference)
RETURNING id, account_id, amount::text AS amount, currency, sender_name, reference, status, created_at;

-- name: GetInboundTransferByID :one
SELECT id, account_id, amount::text AS amount, currency, sender_name, reference, status, created_at
FROM rails.inbound_transfers
WHERE id = $1;

-- name: ListInboundTransfersInRange :many
SELECT id, account_id, amount::text AS amount, currency, sender_name, reference, status, created_at
FROM rails.inbound_transfers
WHERE created_at >= @from_ts AND created_at < @to_ts
ORDER BY created_at ASC;
