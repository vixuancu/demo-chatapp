-- name: CreateRoom :one
INSERT INTO
    rooms (
        room_code,
        room_name,
        room_is_direct_chat,
        room_created_by
    )
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: JoinRoom :one
INSERT INTO
    room_members (user_uuid, room_id)
VALUES ($1, $2) RETURNING *;

-- name: GetRoomByID :one
SELECT * FROM rooms WHERE room_id = $1;

-- name: GetRoomByCode :one
SELECT * FROM rooms WHERE room_code = $1;

-- name: ListUserRooms :many
SELECT r.*
FROM rooms r
    JOIN room_members rm ON r.room_id = rm.room_id
WHERE
    rm.user_uuid = $1
ORDER BY r.room_updated_at DESC;

-- name: IsUserMemberOfRoom :one
SELECT EXISTS (
        SELECT 1
        FROM room_members
        WHERE
            user_uuid = $1
            AND room_id = $2
    ) AS is_member;

-- name: GetRoomMembers :many
SELECT u.*
FROM users u
    JOIN room_members rm ON u.user_uuid = rm.user_uuid
WHERE
    rm.room_id = $1;

-- name: GenerateUniqueRoomCode :one
WITH random_code AS (
    SELECT array_to_string(ARRAY(SELECT chr((65 + round(random() * 25))::integer) 
    FROM generate_series(1, 6)), '') AS code
)
SELECT code FROM random_code
WHERE NOT EXISTS (SELECT 1 FROM rooms WHERE room_code = code);