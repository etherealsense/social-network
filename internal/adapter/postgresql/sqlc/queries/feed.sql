-- name: GetFeed :many
SELECT
    p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at,
    COALESCE(l.likes_count, 0)::bigint AS likes_count,
    COALESCE(c.comments_count, 0)::bigint AS comments_count,
    (COALESCE(l.likes_count, 0) * 0.5
     + COALESCE(c.comments_count, 0) * 2
     - EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600
    )::float8 AS score
FROM posts p
JOIN follows f ON p.user_id = f.following_id AND f.follower_id = $1
LEFT JOIN (SELECT post_id, COUNT(*) AS likes_count FROM likes GROUP BY post_id) l ON l.post_id = p.id
LEFT JOIN (SELECT post_id, COUNT(*) AS comments_count FROM comments GROUP BY post_id) c ON c.post_id = p.id
ORDER BY score DESC
LIMIT $2 OFFSET $3;
