-- name: CreateMessage :one
INSERT INTO messages (chat_id, sender_id, content, created_at, is_read)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, chat_id, sender_id, content, created_at, is_read;

-- name: ListMessagesByChatID :many
SELECT id, chat_id, sender_id, content, created_at, is_read
FROM messages
WHERE chat_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
