-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, family_id, token_hash, revoked, replaced_by, expires_at, created_at;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, family_id, token_hash, revoked, replaced_by, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked = true, replaced_by = $2
WHERE id = $1;

-- name: RevokeFamily :exec
UPDATE refresh_tokens
SET revoked = true
WHERE family_id = $1;

-- name: GetActiveTokenByFamily :one
SELECT id, user_id, family_id, token_hash, revoked, replaced_by, expires_at, created_at
FROM refresh_tokens
WHERE family_id = $1 AND revoked = false
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < now();
