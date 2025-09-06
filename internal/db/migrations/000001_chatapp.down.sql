-- Xóa các trigger
DROP TRIGGER IF EXISTS update_room_members_timestamp ON room_members;

DROP TRIGGER IF EXISTS update_rooms_timestamp ON rooms;

DROP TRIGGER IF EXISTS update_users_timestamp ON users;

-- Xóa các function
DROP FUNCTION IF EXISTS update_room_member_timestamp ();

DROP FUNCTION IF EXISTS update_room_timestamp ();

DROP FUNCTION IF EXISTS update_user_timestamp ();



-- Xóa các chỉ mục

DROP INDEX IF EXISTS idx_room_members_user_uuid;

-- Xóa các bảng theo thứ tự ngược lại (vì có foreign key)
DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS room_members;

DROP TABLE IF EXISTS rooms;

DROP TABLE IF EXISTS users;