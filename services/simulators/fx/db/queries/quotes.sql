-- name: CreateQuote :one
INSERT INTO fx.quotes (from_currency, to_currency, amount, converted_amount, rate, spread_bps, expires_at)
VALUES (@from_currency, @to_currency, @amount::numeric, @converted_amount::numeric, @rate::numeric, @spread_bps, @expires_at)
RETURNING id, from_currency, to_currency, amount::text AS amount, converted_amount::text AS converted_amount,
          rate::text AS rate, spread_bps, status, created_at, expires_at, executed_at;

-- name: GetQuoteByID :one
SELECT id, from_currency, to_currency, amount::text AS amount, converted_amount::text AS converted_amount,
       rate::text AS rate, spread_bps, status, created_at, expires_at, executed_at
FROM fx.quotes
WHERE id = $1;

-- name: MarkQuoteExecuted :one
UPDATE fx.quotes
SET status = 'executed', executed_at = now()
WHERE id = $1
RETURNING id, from_currency, to_currency, amount::text AS amount, converted_amount::text AS converted_amount,
          rate::text AS rate, spread_bps, status, created_at, expires_at, executed_at;
