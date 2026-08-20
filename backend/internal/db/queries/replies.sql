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
    u.role AS author_role,
    u.avatar_url AS author_avatar,
    u.color AS author_color
FROM replies r
JOIN users u ON u.id = r.user_id
WHERE r.thread_id = $1
ORDER BY r.created_at ASC;

-- name: CreateReply :one
WITH new_reply AS (
    INSERT INTO replies (thread_id, user_id, body, image_url, reply_to_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT
    nr.id,
    nr.thread_id,
    nr.user_id,
    nr.body,
    nr.image_url,
    nr.reply_to_id,
    nr.created_at,
    u.username AS author_name,
    u.role AS author_role,
    u.avatar_url AS author_avatar,
    u.color AS author_color
FROM new_reply nr
JOIN users u ON u.id = nr.user_id;

-- name: GetReplyByID :one
SELECT id, thread_id, user_id, body, image_url, reply_to_id, created_at
FROM replies
WHERE id = $1;

-- name: DeleteReply :exec
DELETE FROM replies WHERE id = $1;

-- name: ReplyExists :one
SELECT EXISTS(SELECT 1 FROM replies WHERE id = $1 AND thread_id = $2);
