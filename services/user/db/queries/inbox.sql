-- name: ConsumerInboxExists :one
SELECT EXISTS(
    SELECT 1 FROM "user".consumer_inbox WHERE event_id = $1
)::boolean;

-- name: InsertConsumerInbox :exec
INSERT INTO "user".consumer_inbox (event_id, event_type, processed_at)
VALUES ($1, $2, $3);