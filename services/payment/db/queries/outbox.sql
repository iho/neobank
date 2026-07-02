-- name: InsertOutboxEvent :exec
INSERT INTO payment.outbox_events (id, aggregate_type, aggregate_id, event_type, event_version, payload, correlation_id, causation_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, sqlc.narg(correlation_id), sqlc.narg(causation_id), $7);

-- name: FetchUnpublishedOutboxEvents :many
SELECT e.id, e.aggregate_type, e.aggregate_id, e.event_type, e.event_version, e.payload,
       COALESCE(e.correlation_id, '') AS correlation_id,
       COALESCE(e.causation_id, '') AS causation_id,
       e.created_at
FROM payment.outbox_events e
LEFT JOIN payment.outbox_publications p ON p.event_id = e.id
WHERE p.event_id IS NULL
ORDER BY e.created_at
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
INSERT INTO payment.outbox_publications (event_id, published_at)
VALUES ($1, now())
ON CONFLICT (event_id) DO NOTHING;