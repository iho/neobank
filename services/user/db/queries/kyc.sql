-- name: UpsertProfile :exec
INSERT INTO "user".profiles (user_id, full_name, date_of_birth, country_code)
VALUES (@user_id, @full_name, @date_of_birth, @country_code)
ON CONFLICT (user_id) DO UPDATE
SET full_name = EXCLUDED.full_name,
    date_of_birth = EXCLUDED.date_of_birth,
    country_code = EXCLUDED.country_code;

-- name: CreateKYCCase :one
INSERT INTO "user".kyc_cases (id, user_id, status, submitted_at)
VALUES (@id, @user_id, @status, now())
RETURNING id, user_id, status, COALESCE(rejection_reason, '') AS rejection_reason;

-- name: GetLatestKYCCaseByUser :one
SELECT id, user_id, status, COALESCE(rejection_reason, '') AS rejection_reason
FROM "user".kyc_cases
WHERE user_id = @user_id
ORDER BY submitted_at DESC NULLS LAST, id DESC
LIMIT 1;

-- name: ApproveKYCCase :exec
UPDATE "user".kyc_cases
SET status = 'approved', decided_at = now(), decided_by = @decided_by
WHERE id = @id;

-- name: RejectKYCCase :exec
UPDATE "user".kyc_cases
SET status = 'rejected', decided_at = now(), rejection_reason = @rejection_reason, decided_by = @decided_by
WHERE id = @id;