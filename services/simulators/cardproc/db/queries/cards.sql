-- name: CreateCard :one
INSERT INTO cardproc.cards (external_ref, cardholder_name, pan_token, last_four, expiry_month, expiry_year)
VALUES (@external_ref, @cardholder_name, @pan_token, @last_four, @expiry_month, @expiry_year)
RETURNING id, external_ref, cardholder_name, pan_token, last_four, expiry_month, expiry_year, status, created_at;

-- name: GetCardByID :one
SELECT id, external_ref, cardholder_name, pan_token, last_four, expiry_month, expiry_year, status, created_at
FROM cardproc.cards
WHERE id = $1;

-- name: CancelCard :exec
UPDATE cardproc.cards SET status = 'cancelled' WHERE id = $1;
