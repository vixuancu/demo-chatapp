-- name: CreateMessage :one
INSERT INTO
    messages (room_id, user_uuid, content)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetRoomMessages :many
SELECT *
FROM messages
WHERE
    room_id = $1
ORDER BY message_created_at DESC
LIMIT $2
OFFSET
    $3;

-- name: CountRoomMessages :one
SELECT COUNT(*) FROM messages WHERE room_id = $1;