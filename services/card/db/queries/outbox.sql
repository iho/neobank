-- name: InsertOutboxEvent :exec
INSERT INTO card.outbox_events (id, aggregate_type, aggregate_id, event_type, payload, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FetchUnpublishedOutboxEvents :many
SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at, published_at
FROM card.outbox_events
WHERE published_at IS NULL
ORDER BY created_at
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE card.outbox_events SET published_at = now() WHERE id = $1;