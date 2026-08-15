-- name: ListThreadReplies :many
SELECT
    r.id,
    r.thread_id,
    r.user_id,
    r.body,
    r.image_url,
    r.reply_to_id,
    r.created_at,
    u.username AS author_name,
    u.role AS author_role
FROM replies r
JOIN users u ON u.id = r.user_id
WHERE r.thread_id = $1
ORDER BY r.created_at ASC;

-- name: CreateReply :one
INSERT INTO replies (thread_id, user_id, body, image_url, reply_to_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, thread_id, user_id, body, image_url, reply_to_id, created_at;

-- name: GetReplyByID :one
SELECT id, thread_id, user_id, body, image_url, reply_to_id, created_at
FROM replies
WHERE id = $1;

-- name: DeleteReply :exec
DELETE FROM replies WHERE id = $1;

-- name: ReplyExists :one
SELECT EXISTS(SELECT 1 FROM replies WHERE id = $1 AND thread_id = $2);
