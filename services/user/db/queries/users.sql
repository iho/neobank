-- name: CreateUser :exec
INSERT INTO "user".users (id, email, phone, password_hash, status)
VALUES (@id, @email, NULLIF(@phone, ''), @password_hash, @status);

-- name: GetUserByEmail :one
SELECT id, email, COALESCE(phone, '') AS phone, password_hash, status
FROM "user".users
WHERE email = $1;

-- name: GetUserByPhone :one
SELECT id, email, COALESCE(phone, '') AS phone, password_hash, status
FROM "user".users
WHERE phone = @phone;

-- name: GetUserByID :one
SELECT id, email, COALESCE(phone, '') AS phone, password_hash, status
FROM "user".users
WHERE id = $1;

-- name: GetUserProfile :one
SELECT
    u.id,
    u.email,
    COALESCE(u.phone, '') AS phone,
    u.status,
    u.created_at,
    COALESCE(p.full_name, '') AS full_name,
    p.date_of_birth,
    COALESCE(p.country_code, '') AS country_code,
    COALESCE(k.status, 'pending') AS kyc_status
FROM "user".users u
LEFT JOIN "user".profiles p ON p.user_id = u.id
LEFT JOIN LATERAL (
    SELECT status
    FROM "user".kyc_cases
    WHERE user_id = u.id
    ORDER BY submitted_at DESC NULLS LAST, id DESC
    LIMIT 1
) k ON true
WHERE u.id = $1;