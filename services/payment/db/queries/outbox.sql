-- name: InsertOutboxEvent :exec
INSERT INTO payment.outbox_events (id, aggregate_type, aggregate_id, event_type, event_version, payload, correlation_id, causation_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, sqlc.narg(correlation_id), sqlc.narg(causation_id), $7);

-- name: FetchUnpublishedOutboxEvents :many
SELECT id, aggregate_type, aggregate_id, event_type, event_version, payload,
       COALESCE(correlation_id, '') AS correlation_id,
       COALESCE(causation_id, '') AS causation_id,
       created_at, published_at
FROM payment.outbox_events
WHERE published_at IS NULL
ORDER BY created_at
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE payment.outbox_events SET published_at = now() WHERE id = $1;
