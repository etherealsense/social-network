-- name: LikePost :one
INSERT INTO likes (user_id, post_id) VALUES ($1, $2) RETURNING id, user_id, post_id, created_at;

-- name: UnlikePost :exec
DELETE FROM likes WHERE user_id = $1 AND post_id = $2;

-- name: ListLikesByPostID :many
SELECT * FROM likes WHERE post_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountLikesByPostID :one
SELECT COUNT(*) FROM likes WHERE post_id = $1;
