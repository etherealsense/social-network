-- name: CreateChatParticipant :exec
INSERT INTO chat_participants (chat_id, user_id, joined_at) VALUES ($1, $2, $3);

-- name: GetChatParticipantByChatIDAndUserID :one
SELECT * FROM chat_participants WHERE chat_id = $1 AND user_id = $2;

-- name: ListChatParticipantsByChatID :many
SELECT * FROM chat_participants WHERE chat_id = $1 ORDER BY joined_at LIMIT $2 OFFSET $3;

-- name: DeleteChatParticipant :exec
DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2;
