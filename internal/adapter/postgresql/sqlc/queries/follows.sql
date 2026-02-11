-- name: FollowUser :one
INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) RETURNING id, follower_id, following_id, created_at;

-- name: UnfollowUser :exec
DELETE FROM follows WHERE follower_id = $1 AND following_id = $2;

-- name: ListFollowers :many
SELECT * FROM follows WHERE following_id = $1 ORDER BY created_at DESC;

-- name: ListFollowing :many
SELECT * FROM follows WHERE follower_id = $1 ORDER BY created_at DESC;
