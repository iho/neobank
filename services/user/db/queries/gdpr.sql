-- name: ListKYCSubmissionsByUser :many
SELECT id, kyc_case_id, user_id, document_type, COALESCE(document_number, '') AS document_number,
       provider, COALESCE(provider_reference, '') AS provider_reference, provider_response,
       screening_decision, COALESCE(screening_reason, '') AS screening_reason,
       COALESCE(correlation_id, '') AS correlation_id, created_at
FROM "user".kyc_submissions
WHERE user_id = $1
ORDER BY created_at;

-- name: ListWalletsByUser :many
SELECT id, user_id, currency, ledger_account_id, status
FROM "user".wallets
WHERE user_id = $1
ORDER BY currency;

-- name: CountWalletTransactionsByUser :one
SELECT COUNT(*)::bigint AS count
FROM "user".wallet_transactions
WHERE user_id = $1;

-- name: MaskUserAccount :exec
UPDATE "user".users
SET email = $2, phone = NULL, phone_lookup = NULL, password_hash = $3, status = 'masked', updated_at = now()
WHERE id = $1;

-- name: MaskUserProfile :exec
UPDATE "user".profiles
SET full_name = 'REDACTED', date_of_birth = NULL, date_of_birth_encrypted = NULL, country_code = NULL
WHERE user_id = $1;

-- name: InsertGDPRRequest :exec
INSERT INTO "user".gdpr_requests (id, user_id, request_type, actor, correlation_id)
VALUES ($1, $2, $3, $4, sqlc.narg(correlation_id));