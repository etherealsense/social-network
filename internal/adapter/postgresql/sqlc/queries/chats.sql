-- name: CreateChat :one
INSERT INTO chats (created_at) VALUES ($1) RETURNING id, created_at;

-- name: GetChat :one
SELECT * FROM chats WHERE id = $1;

-- name: GetChatByTwoUsers :one
SELECT c.id, c.created_at FROM chats c
JOIN chat_participants cp1 ON cp1.chat_id = c.id AND cp1.user_id = $1
JOIN chat_participants cp2 ON cp2.chat_id = c.id AND cp2.user_id = $2;

-- name: ListChatsByUserID :many
SELECT c.id, c.created_at FROM chats c
JOIN chat_participants cp ON cp.chat_id = c.id
WHERE cp.user_id = $1
ORDER BY c.created_at DESC;

-- name: DeleteChat :exec
DELETE FROM chats WHERE id = $1;
