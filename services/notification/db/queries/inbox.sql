-- name: ConsumerInboxExists :one
SELECT EXISTS(
    SELECT 1 FROM notification.consumer_inbox WHERE event_id = $1
)::boolean;

-- name: InsertConsumerInbox :exec
INSERT INTO notification.consumer_inbox (event_id, event_type, processed_at)
VALUES ($1, $2, $3);