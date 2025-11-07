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

-- name: LeaveRoom :exec
DELETE FROM room_members WHERE user_uuid = $1 AND room_id = $2;

-- name: GetAllRoomsWithMemberCount :many
SELECT r.*, COUNT(rm.user_uuid) as member_count
FROM rooms r
    LEFT JOIN room_members rm ON r.room_id = rm.room_id
GROUP BY
    r.room_id
ORDER BY r.room_created_at DESC
LIMIT $1
OFFSET
    $2;

-- name: GetTotalRoomsCount :one
SELECT COUNT(*) FROM rooms;

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

-- name: DeleteRoom :exec
DELETE FROM rooms WHERE room_id = $1;

-- name: GenerateUniqueRoomCode :one
WITH random_code AS (
    SELECT array_to_string(ARRAY(SELECT chr((65 + round(random() * 25))::integer) 
    FROM generate_series(1, 6)), '') AS code
)
SELECT code FROM random_code
WHERE NOT EXISTS (SELECT 1 FROM rooms WHERE room_code = code);

-- name: ListUserRoomsWithLastMessage :many
SELECT 
    r.room_id,
    r.room_code,
    r.room_name,
    r.room_is_direct_chat,
    r.room_created_by,
    r.room_created_at,
    r.room_updated_at,
    -- Last message info with COALESCE to handle NULL
    COALESCE(lm.message_id, 0) as last_message_id,
    COALESCE(lm.content, '') as last_message_content,
    COALESCE(lm.message_created_at, r.room_created_at) as last_message_time,
    COALESCE(lm.user_uuid, '00000000-0000-0000-0000-000000000000'::uuid) as last_sender_uuid,
    u.user_fullname as last_sender_name,
    -- Unread count from room_read_status
    COALESCE(rrs.unread_count, 0) as unread_count,
    -- Member count
    (SELECT COUNT(*) FROM room_members WHERE room_id = r.room_id) as member_count
FROM rooms r
INNER JOIN room_members rm ON r.room_id = rm.room_id
LEFT JOIN LATERAL (
    SELECT m.message_id, m.content, m.message_created_at, m.user_uuid
    FROM messages m 
    WHERE m.room_id = r.room_id 
    ORDER BY m.message_created_at DESC 
    LIMIT 1
) lm ON true
LEFT JOIN users u ON lm.user_uuid = u.user_uuid
LEFT JOIN room_read_status rrs ON rrs.user_uuid = $1 AND rrs.room_id = r.room_id
WHERE rm.user_uuid = $1
ORDER BY COALESCE(lm.message_created_at, r.room_created_at) DESC;

-- name: GetUnreadCount :one
SELECT COALESCE(unread_count, 0) as unread_count
FROM room_read_status
WHERE user_uuid = $1 AND room_id = $2;

-- name: IncrementUnreadCount :exec
UPDATE room_read_status
SET unread_count = unread_count + 1
WHERE user_uuid = $1 AND room_id = $2;

-- name: IncrementUnreadCountsForAllMembers :exec
-- Increment unread count for all members except sender
-- Creates record if not exists (UPSERT)
INSERT INTO room_read_status (user_uuid, room_id, unread_count, last_read_at)
SELECT rm.user_uuid, $1, 1, NOW()
FROM room_members rm
WHERE rm.room_id = $1 
  AND rm.user_uuid != $2
ON CONFLICT (user_uuid, room_id)
DO UPDATE SET 
    unread_count = room_read_status.unread_count + 1;

-- name: MarkRoomAsRead :exec
INSERT INTO room_read_status (user_uuid, room_id, unread_count, last_read_message_id, last_read_at)
VALUES ($1, $2, 0, $3, NOW())
ON CONFLICT (user_uuid, room_id) 
DO UPDATE SET 
    unread_count = 0,
    last_read_message_id = EXCLUDED.last_read_message_id,
    last_read_at = NOW();

-- name: GetRoomWithMembers :one
SELECT 
    r.room_id,
    r.room_code,
    r.room_name,
    r.room_is_direct_chat,
    r.room_created_by,
    r.room_created_at,
    r.room_updated_at,
    (SELECT COUNT(*) FROM room_members WHERE room_id = r.room_id) as member_count
FROM rooms r
WHERE r.room_id = $1;