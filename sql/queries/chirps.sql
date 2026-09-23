-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetChirpById :one
SELECT * From chirps
WHERE id = $1;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;