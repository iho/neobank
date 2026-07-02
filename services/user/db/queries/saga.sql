-- name: GetSagaByIdempotencyKey :one
SELECT id, saga_type, idempotency_key, status, completed_steps, context
FROM "user".saga_instances
WHERE idempotency_key = $1;

-- name: CreateSagaInstance :exec
INSERT INTO "user".saga_instances (id, saga_type, idempotency_key, status, completed_steps, context)
VALUES ($1, $2, $3, $4, '{}', $5);

-- name: UpdateSagaInstance :exec
UPDATE "user".saga_instances
SET status = $2, completed_steps = $3, context = $4, updated_at = now()
WHERE id = $1;