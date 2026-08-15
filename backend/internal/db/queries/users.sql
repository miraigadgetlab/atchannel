-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, avatar_url, bio, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, password_hash, avatar_url, bio, role, created_at;

-- name: GetUserByID :one
SELECT id, username, email, password_hash, avatar_url, bio, role, created_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, avatar_url, bio, role, created_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, avatar_url, bio, role, created_at
FROM users
WHERE email = $1;

-- name: GetUserPublicByUsername :one
SELECT id, username, avatar_url, bio, role, created_at
FROM users
WHERE username = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET avatar_url = COALESCE($2, avatar_url),
    bio = COALESCE($3, bio)
WHERE id = $1
RETURNING id, username, email, password_hash, avatar_url, bio, role, created_at;

-- name: UserExistsByUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
