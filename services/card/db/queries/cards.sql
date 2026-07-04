-- name: CreateCard :exec
INSERT INTO card.cards (
    id, user_id, wallet_id, processor_ref, pan_token, last_four,
    expiry_month, expiry_year, status, idempotency_key, daily_limit, online_only
) VALUES (
    @id, @user_id, @wallet_id, sqlc.narg(processor_ref), @pan_token, @last_four,
    @expiry_month, @expiry_year, @status, @idempotency_key,
    sqlc.narg(daily_limit)::numeric, COALESCE(sqlc.narg(online_only), false)
);

-- name: GetCardByID :one
SELECT id, user_id, wallet_id, COALESCE(processor_ref, '') AS processor_ref,
       pan_token, last_four, expiry_month, expiry_year, status, idempotency_key,
       COALESCE(daily_limit::text, '') AS daily_limit, online_only, created_at
FROM card.cards
WHERE id = $1;

-- name: GetCardByUserAndIdempotencyKey :one
SELECT id, user_id, wallet_id, COALESCE(processor_ref, '') AS processor_ref,
       pan_token, last_four, expiry_month, expiry_year, status, idempotency_key,
       COALESCE(daily_limit::text, '') AS daily_limit, online_only, created_at
FROM card.cards
WHERE user_id = $1 AND idempotency_key = $2;

-- name: GetCardByProcessorRef :one
SELECT id, user_id, wallet_id, COALESCE(processor_ref, '') AS processor_ref,
       pan_token, last_four, expiry_month, expiry_year, status, idempotency_key,
       COALESCE(daily_limit::text, '') AS daily_limit, online_only, created_at
FROM card.cards
WHERE processor_ref = $1;

-- name: ListCardsByUser :many
SELECT id, user_id, wallet_id, COALESCE(processor_ref, '') AS processor_ref,
       pan_token, last_four, expiry_month, expiry_year, status, idempotency_key,
       COALESCE(daily_limit::text, '') AS daily_limit, online_only, created_at
FROM card.cards
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateCardControls :one
UPDATE card.cards
SET daily_limit = sqlc.narg(daily_limit)::numeric,
    online_only = COALESCE(sqlc.narg(online_only), online_only)
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, wallet_id, COALESCE(processor_ref, '') AS processor_ref,
          pan_token, last_four, expiry_month, expiry_year, status, idempotency_key,
          COALESCE(daily_limit::text, '') AS daily_limit, online_only, created_at;

-- name: UpdateCardStatus :exec
UPDATE card.cards SET status = $2 WHERE id = $1 AND user_id = $3;

-- name: MarkCardCancelled :exec
UPDATE card.cards SET status = 'cancelled' WHERE id = $1;