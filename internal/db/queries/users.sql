-- name: CreateUser :one
INSERT INTO
    users (
        user_email,
        user_password,
        user_fullname
    )
VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE user_email = $1;

-- name: GetUserByUUID :one
SELECT * FROM users WHERE user_uuid = $1;

-- name: GetAllUsers :many
SELECT *
FROM users
ORDER BY user_created_at DESC
LIMIT $1
OFFSET
    $2;

-- name: DeleteUser :exec
DELETE FROM users WHERE user_uuid = $1;