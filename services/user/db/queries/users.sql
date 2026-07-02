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