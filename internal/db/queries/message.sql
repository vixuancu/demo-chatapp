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

-- name: GetRoomMessagesWithCursor :many
-- Cursor-based pagination: Load messages BEFORE a specific message_id
SELECT *
FROM messages
WHERE
    room_id = $1
    AND (sqlc.narg('cursor')::bigint IS NULL OR message_id < sqlc.narg('cursor'))
ORDER BY message_id DESC
LIMIT $2;

-- name: CountRoomMessages :one
SELECT COUNT(*) FROM messages WHERE room_id = $1;