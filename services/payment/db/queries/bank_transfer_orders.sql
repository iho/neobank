-- name: CreateBankTransferOrder :one
INSERT INTO payment.bank_transfer_orders (
    rails_payment_id, user_id, amount, currency, counterparty_iban, reference, ledger_transfer_id
) VALUES (
    @rails_payment_id, @user_id, @amount::numeric, @currency, @counterparty_iban, @reference, @ledger_transfer_id
)
RETURNING id, rails_payment_id, user_id, amount::text AS amount, currency, counterparty_iban, reference,
          ledger_transfer_id, COALESCE(return_transfer_id, '') AS return_transfer_id, status, created_at, updated_at;

-- name: GetBankTransferOrderByRailsPaymentID :one
SELECT id, rails_payment_id, user_id, amount::text AS amount, currency, counterparty_iban, reference,
       ledger_transfer_id, COALESCE(return_transfer_id, '') AS return_transfer_id, status, created_at, updated_at
FROM payment.bank_transfer_orders
WHERE rails_payment_id = $1;

-- name: MarkBankTransferOrderSettled :exec
UPDATE payment.bank_transfer_orders
SET status = 'settled', updated_at = now()
WHERE rails_payment_id = $1 AND status = 'processing';

-- name: MarkBankTransferOrderReturned :exec
UPDATE payment.bank_transfer_orders
SET status = @status, return_transfer_id = @return_transfer_id, updated_at = now()
WHERE rails_payment_id = @rails_payment_id;
