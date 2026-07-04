-- name: UpsertProfile :exec
INSERT INTO "user".profiles (user_id, full_name, date_of_birth, date_of_birth_encrypted, country_code)
VALUES (@user_id, @full_name, sqlc.narg(date_of_birth), sqlc.narg(date_of_birth_encrypted), @country_code)
ON CONFLICT (user_id) DO UPDATE
SET full_name = EXCLUDED.full_name,
    date_of_birth = EXCLUDED.date_of_birth,
    date_of_birth_encrypted = EXCLUDED.date_of_birth_encrypted,
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

-- name: MarkKYCCaseManualReview :exec
UPDATE "user".kyc_cases
SET status = 'manual_review'
WHERE id = @id;

-- name: SetKYCCaseVendorApplicant :exec
UPDATE "user".kyc_cases
SET vendor_applicant_id = @vendor_applicant_id
WHERE id = @id;

-- name: GetKYCCaseByID :one
SELECT id, user_id, status, COALESCE(rejection_reason, '') AS rejection_reason,
       COALESCE(vendor_applicant_id, '') AS vendor_applicant_id
FROM "user".kyc_cases
WHERE id = @id;

-- name: GetKYCCaseByVendorApplicant :one
SELECT id, user_id, status, COALESCE(rejection_reason, '') AS rejection_reason,
       COALESCE(vendor_applicant_id, '') AS vendor_applicant_id
FROM "user".kyc_cases
WHERE vendor_applicant_id = @vendor_applicant_id;