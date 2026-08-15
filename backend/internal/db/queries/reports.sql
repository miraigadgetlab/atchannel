-- name: CreateReport :one
INSERT INTO reports (target_type, target_id, reporter_id, reason)
VALUES ($1, $2, $3, $4)
RETURNING id, target_type, target_id, reporter_id, reason, status, created_at;

-- name: GetReportByID :one
SELECT id, target_type, target_id, reporter_id, reason, status, created_at
FROM reports
WHERE id = $1;

-- name: UpdateReportStatus :one
UPDATE reports SET status = $2 WHERE id = $1
RETURNING id, target_type, target_id, reporter_id, reason, status, created_at;

-- name: ListReports :many
SELECT
    r.id,
    r.target_type,
    r.target_id,
    r.reporter_id,
    r.reason,
    r.status,
    r.created_at,
    u.username AS reporter_name,
    COALESCE((SELECT b.slug FROM threads t JOIN boards b ON b.id = t.board_id WHERE r.target_type = 'thread' AND t.id = r.target_id), '')::text AS target_board_slug,
    COALESCE((SELECT t.id FROM threads t WHERE r.target_type = 'thread' AND t.id = r.target_id), (SELECT t.id FROM replies rp JOIN threads t ON t.id = rp.thread_id WHERE r.target_type = 'reply' AND rp.id = r.target_id))::text AS target_thread_id,
    COALESCE((SELECT t.body FROM threads t WHERE r.target_type = 'thread' AND t.id = r.target_id), (SELECT rp.body FROM replies rp WHERE r.target_type = 'reply' AND rp.id = r.target_id))::text AS target_body
FROM reports r
JOIN users u ON u.id = r.reporter_id
WHERE r.status = $1
ORDER BY r.created_at ASC
LIMIT $2 OFFSET $3;
