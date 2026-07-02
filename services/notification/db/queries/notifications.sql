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

-- name: CountUnreadNotificationsByUser :one
SELECT COUNT(*)::bigint AS count
FROM notification.notifications
WHERE user_id = $1 AND read = false;

-- name: MarkNotificationRead :one
UPDATE notification.notifications
SET read = true
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, event_type, title, body, read, created_at;

-- name: MarkAllNotificationsRead :execrows
UPDATE notification.notifications
SET read = true
WHERE user_id = $1 AND read = false;