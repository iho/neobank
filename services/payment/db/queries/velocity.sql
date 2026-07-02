-- name: InsertVelocityEvent :exec
INSERT INTO payment.velocity_events (user_id, amount, recorded_at)
VALUES (@user_id, @amount::numeric, @recorded_at);

-- name: CountVelocityEventsLastHour :one
SELECT COUNT(*)::int AS count
FROM payment.velocity_events
WHERE user_id = @user_id AND recorded_at >= @recorded_at;

-- name: SumVelocityEventsLast24h :one
SELECT COALESCE(SUM(amount), 0)::text AS total
FROM payment.velocity_events
WHERE user_id = @user_id AND recorded_at >= @recorded_at;

-- name: PruneVelocityEventsOlderThan :exec
DELETE FROM payment.velocity_events
WHERE recorded_at < @cutoff;