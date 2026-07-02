-- name: InsertPIIAccessLog :exec
INSERT INTO "user".pii_access_log (id, subject_user_id, resource, actor, correlation_id, metadata, created_at)
VALUES ($1, $2, $3, $4, sqlc.narg(correlation_id), $5, $6);

-- name: ListPIIAccessBySubject :many
SELECT id, subject_user_id, resource, actor, COALESCE(correlation_id, '') AS correlation_id,
       metadata, created_at
FROM "user".pii_access_log
WHERE subject_user_id = $1
ORDER BY created_at;