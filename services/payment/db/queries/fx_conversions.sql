-- name: CreateFXConversion :one
INSERT INTO payment.fx_conversions (
    quote_id, user_id, from_currency, to_currency, amount, converted_amount, rate,
    from_ledger_transfer_id, to_ledger_transfer_id
) VALUES (
    @quote_id, @user_id, @from_currency, @to_currency, @amount::numeric, @converted_amount::numeric, @rate::numeric,
    @from_ledger_transfer_id, @to_ledger_transfer_id
)
RETURNING id, quote_id, user_id, from_currency, to_currency, amount::text AS amount,
          converted_amount::text AS converted_amount, rate::text AS rate,
          from_ledger_transfer_id, to_ledger_transfer_id, status, created_at;

-- name: GetFXConversionByQuoteID :one
SELECT id, quote_id, user_id, from_currency, to_currency, amount::text AS amount,
       converted_amount::text AS converted_amount, rate::text AS rate,
       from_ledger_transfer_id, to_ledger_transfer_id, status, created_at
FROM payment.fx_conversions
WHERE quote_id = $1;
