-- name: CreateBan :one
INSERT INTO bans (user_id, ip, reason, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, ip, reason, expires_at, created_at;

-- name: GetActiveBanForUser :one
SELECT id, user_id, ip, reason, expires_at, created_at
FROM bans
WHERE user_id = $1
  AND (expires_at IS NULL OR expires_at > now())
ORDER BY created_at DESC
LIMIT 1;

-- name: GetActiveBanForIP :one
SELECT id, user_id, ip, reason, expires_at, created_at
FROM bans
WHERE ip = $1
  AND (expires_at IS NULL OR expires_at > now())
ORDER BY created_at DESC
LIMIT 1;

-- name: ListActiveBans :many
SELECT id, user_id, ip, reason, expires_at, created_at
FROM bans
WHERE expires_at IS NULL OR expires_at > now()
ORDER BY created_at DESC;
