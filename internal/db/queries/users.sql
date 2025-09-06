-- name: CreateUser :one
INSERT INTO
    users (
        user_email,
        user_password,
        user_fullname
    )
VALUES ($1, $2, $3) RETURNING *;