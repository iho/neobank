-- name: EnqueueDelivery :exec
INSERT INTO cardproc.webhook_deliveries (id, url, event_type, payload, next_attempt_at)
VALUES (@id, @url, @event_type, @payload, @next_attempt_at);

-- name: ClaimDueDeliveries :many
SELECT id, url, event_type, payload, attempts, last_error, next_attempt_at, delivered_at, created_at
FROM cardproc.webhook_deliveries
WHERE delivered_at IS NULL AND next_attempt_at <= @now
ORDER BY next_attempt_at ASC
LIMIT @limit_val
FOR UPDATE SKIP LOCKED;

-- name: MarkDeliveryDelivered :exec
UPDATE cardproc.webhook_deliveries
SET delivered_at = @delivered_at, attempts = attempts + 1
WHERE id = @id;

-- name: MarkDeliveryFailed :exec
UPDATE cardproc.webhook_deliveries
SET attempts = attempts + 1, next_attempt_at = @next_attempt_at, last_error = @last_error
WHERE id = @id;

-- name: ListDeliveries :many
SELECT id, url, event_type, payload, attempts, last_error, next_attempt_at, delivered_at, created_at
FROM cardproc.webhook_deliveries
ORDER BY created_at DESC
LIMIT @limit_val;

-- name: GetDelivery :one
SELECT id, url, event_type, payload, attempts, last_error, next_attempt_at, delivered_at, created_at
FROM cardproc.webhook_deliveries
WHERE id = $1;
