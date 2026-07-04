-- name: CreateApplicant :one
INSERT INTO kyc.applicants (external_ref, full_name, date_of_birth, country_code, status, reason)
VALUES (@external_ref, @full_name, @date_of_birth, @country_code, @status, @reason)
RETURNING id, external_ref, full_name, date_of_birth, country_code, status, reason, created_at, decided_at;

-- name: GetApplicantByID :one
SELECT id, external_ref, full_name, date_of_birth, country_code, status, reason, created_at, decided_at
FROM kyc.applicants
WHERE id = $1;

-- name: ResolveApplicant :one
UPDATE kyc.applicants
SET status = @status, reason = @reason, decided_at = now()
WHERE id = @id
RETURNING id, external_ref, full_name, date_of_birth, country_code, status, reason, created_at, decided_at;
