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