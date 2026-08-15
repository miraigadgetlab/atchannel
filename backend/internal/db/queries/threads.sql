-- name: ListBoards :many
SELECT id, slug, name, description, created_at
FROM boards
ORDER BY created_at ASC;

-- name: GetBoardBySlug :one
SELECT id, slug, name, description, created_at
FROM boards
WHERE slug = $1;

-- name: CreateBoard :one
INSERT INTO boards (slug, name, description)
VALUES ($1, $2, $3)
RETURNING id, slug, name, description, created_at;

-- name: GetBoardThreads :many
SELECT
    t.id,
    t.board_id,
    t.user_id,
    t.title,
    t.body,
    t.image_url,
    t.is_pinned,
    t.is_locked,
    t.bumped_at,
    t.created_at,
    b.slug AS board_slug,
    u.username AS author_name,
    u.role AS author_role,
    rr.reply_count,
    rr.last_reply_at,
    (rr.last_reply_at IS NULL OR t.bumped_at >= rr.last_reply_at) AS bumped
FROM threads t
JOIN boards b ON b.id = t.board_id
JOIN users u ON u.id = t.user_id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS reply_count, COALESCE(MAX(r.created_at), t.created_at)::timestamptz AS last_reply_at
    FROM replies r
    WHERE r.thread_id = t.id
) rr ON true
WHERE t.board_id = $1 ORDER BY t.is_pinned DESC, t.bumped_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBoardThreadCount :one
SELECT COUNT(*) FROM threads WHERE board_id = $1;

-- name: GetThreadByID :one
SELECT
    t.id,
    t.board_id,
    t.user_id,
    t.title,
    t.body,
    t.image_url,
    t.is_pinned,
    t.is_locked,
    t.bumped_at,
    t.created_at,
    b.slug AS board_slug,
    u.username AS author_name,
    u.role AS author_role,
    rr.reply_count,
    rr.last_reply_at,
    (rr.last_reply_at IS NULL OR t.bumped_at >= rr.last_reply_at) AS bumped
FROM threads t
JOIN boards b ON b.id = t.board_id
JOIN users u ON u.id = t.user_id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS reply_count, COALESCE(MAX(r.created_at), t.created_at)::timestamptz AS last_reply_at
    FROM replies r
    WHERE r.thread_id = t.id
) rr ON true
WHERE t.id = $1;

-- name: CreateThread :one
INSERT INTO threads (board_id, user_id, title, body, image_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, board_id, user_id, title, body, image_url, is_pinned, is_locked, bumped_at, created_at;

-- name: TouchThreadBump :exec
UPDATE threads SET bumped_at = now()
WHERE id = $1;

-- name: GetThreadBoardID :one
SELECT board_id FROM threads WHERE id = $1;

-- name: SetThreadPinned :exec
UPDATE threads SET is_pinned = $2 WHERE id = $1;

-- name: SetThreadLocked :exec
UPDATE threads SET is_locked = $2 WHERE id = $1;

-- name: DeleteThread :exec
DELETE FROM threads WHERE id = $1;
