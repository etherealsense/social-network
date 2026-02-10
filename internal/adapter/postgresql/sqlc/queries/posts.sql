-- name: ListPostsByUserID :many
SELECT * FROM posts WHERE user_id = $1;

-- name: CreatePost :one
INSERT INTO posts (user_id, title, content) VALUES ($1, $2, $3) RETURNING id, user_id, title, content, created_at, updated_at;

-- name: FindPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: UpdatePost :one
UPDATE posts 
SET 
    title = COALESCE(sqlc.narg('title'), title),
    content = COALESCE(sqlc.narg('content'), content),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, user_id, title, content, created_at, updated_at;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;
