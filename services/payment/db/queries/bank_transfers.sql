-- name: CreateBankAccount :one
INSERT INTO payment.bank_accounts (user_id, currency, rails_account_id, iban)
VALUES (@user_id, @currency, @rails_account_id, @iban)
RETURNING id, user_id, currency, rails_account_id, iban, created_at;

-- name: GetBankAccountByUserAndCurrency :one
SELECT id, user_id, currency, rails_account_id, iban, created_at
FROM payment.bank_accounts
WHERE user_id = $1 AND currency = $2;

-- name: CreateBankTransfer :one
INSERT INTO payment.bank_transfers (
    rails_transfer_id, user_id, amount, currency, sender_name, reference, ledger_transfer_id
) VALUES (
    @rails_transfer_id, @user_id, @amount::numeric, @currency, @sender_name, @reference, @ledger_transfer_id
)
RETURNING id, rails_transfer_id, user_id, amount::text AS amount, currency, sender_name, reference,
          ledger_transfer_id, status, created_at;

-- name: GetBankTransferByRailsTransferID :one
SELECT id, rails_transfer_id, user_id, amount::text AS amount, currency, sender_name, reference,
       ledger_transfer_id, status, created_at
FROM payment.bank_transfers
WHERE rails_transfer_id = $1;
