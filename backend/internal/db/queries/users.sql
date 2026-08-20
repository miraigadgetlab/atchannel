-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, avatar_url, bio, role, color)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, username, email, password_hash, avatar_url, bio, role, color, created_at;

-- name: GetUserByID :one
SELECT id, username, email, password_hash, avatar_url, bio, role, color, created_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, avatar_url, bio, role, color, created_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, avatar_url, bio, role, color, created_at
FROM users
WHERE email = $1;

-- name: GetUserPublicByUsername :one
SELECT id, username, avatar_url, bio, role, color, created_at
FROM users
WHERE username = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET avatar_url = CASE WHEN $2::text = '' THEN avatar_url ELSE $2::text END,
    bio = CASE WHEN $3::text = '' THEN bio ELSE $3::text END,
    color = CASE WHEN $4::text = '' THEN color ELSE $4::text END
WHERE id = $1
RETURNING id, username, email, password_hash, avatar_url, bio, role, color, created_at;

-- name: UserExistsByUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
