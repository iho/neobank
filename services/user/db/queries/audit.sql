-- name: InsertAuditLog :exec
INSERT INTO "user".audit_log (id, entity_type, entity_id, action, from_status, to_status, actor, correlation_id, metadata, created_at)
VALUES ($1, $2, $3, $4, sqlc.narg(from_status), sqlc.narg(to_status), $5, sqlc.narg(correlation_id), $6, $7);

-- name: ListAuditLogByEntity :many
SELECT id, entity_type, entity_id, action, COALESCE(from_status, '') AS from_status,
       COALESCE(to_status, '') AS to_status, actor, COALESCE(correlation_id, '') AS correlation_id,
       metadata, created_at
FROM "user".audit_log
WHERE entity_type = $1 AND entity_id = $2
ORDER BY created_at;
