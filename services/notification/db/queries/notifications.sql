-- name: InsertNotification :exec
INSERT INTO notification.notifications (id, user_id, event_id, event_type, title, body)
VALUES (@id, @user_id, @event_id, @event_type, @title, @body)
ON CONFLICT (event_id, user_id) DO NOTHING;

-- name: ListNotificationsByUser :many
SELECT id, user_id, event_type, title, body, read, created_at
FROM notification.notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;