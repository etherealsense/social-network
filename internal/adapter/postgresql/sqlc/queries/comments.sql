-- name: ListCommentsByPostID :many
SELECT * FROM comments WHERE post_id = $1 ORDER BY created_at ASC;

-- name: CreateComment :one
INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, post_id, user_id, content, created_at, updated_at;

-- name: FindCommentByID :one
SELECT * FROM comments WHERE id = $1;

-- name: UpdateComment :one
UPDATE comments
SET
    content = COALESCE(sqlc.narg('content'), content),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, post_id, user_id, content, created_at, updated_at;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;
