-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUserFromEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DropAllUsers :exec
TRUNCATE TABLE users CASCADE;

-- name: UpdateUserEmailPassword :one
UPDATE users
SET email = $2,
    hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: UpgradeUser :exec
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1;