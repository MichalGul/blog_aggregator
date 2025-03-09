-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT users.id, users.created_at, users.updated_at, users.name from users where users.name=$1; 

-- name: GetUsers :many
SELECT * FROM users;

-- name: DeleteUsers :exec
DELETE from users;