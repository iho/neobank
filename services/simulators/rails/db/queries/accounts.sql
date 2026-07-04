-- name: CreateAccount :one
INSERT INTO rails.accounts (external_ref, currency, iban)
VALUES (@external_ref, @currency, @iban)
RETURNING id, external_ref, currency, iban, created_at;

-- name: GetAccountByExternalRefAndCurrency :one
SELECT id, external_ref, currency, iban, created_at
FROM rails.accounts
WHERE external_ref = $1 AND currency = $2;

-- name: GetAccountByID :one
SELECT id, external_ref, currency, iban, created_at
FROM rails.accounts
WHERE id = $1;
