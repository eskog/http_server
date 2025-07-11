-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DropAllChirps :exec
TRUNCATE TABLE chirps CASCADE;


-- name: GetAllChirps :many
SELECT id, created_at, updated_at, body, user_id FROM chirps 
ORDER BY created_at ASC;

-- name: GetAllChirpsByAuthor :many
SELECT * FROM chirps
where user_id = $1
ORDER BY created_at ASC;

-- name: GetSingleChirp :one
SELECT * FROM chirps
where id = $1;

-- name: DeleteSingleChirp :exec
DELETE FROM chirps
WHERE id = $1;