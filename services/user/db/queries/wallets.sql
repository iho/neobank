-- name: CreateWallet :exec
INSERT INTO "user".wallets (id, user_id, currency, ledger_account_id, status)
VALUES ($1, $2, $3, $4, $5);

-- name: GetWalletByUserAndCurrency :one
SELECT id, user_id, currency, ledger_account_id, status
FROM "user".wallets
WHERE user_id = $1 AND currency = $2;

-- name: DeleteWalletByID :exec
DELETE FROM "user".wallets WHERE id = $1;

-- name: ListWalletsByUserID :many
SELECT id, user_id, currency, ledger_account_id, status
FROM "user".wallets
WHERE user_id = $1
ORDER BY currency;