-- câu lệnh sql thì viết hoa còn lại thì viết thường
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Bảng users: Lưu thông tin người dùng
CREATE TABLE users (
    user_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid (), -- UUID làm khóa chính, bảo mật và scale tốt
    user_email VARCHAR(100) UNIQUE NOT NULL, -- Email duy nhất, giới hạn 100 ký tự
    user_password VARCHAR(255) NOT NULL, -- Mật khẩu mã hóa (bcrypt/argon2), giới hạn 255 ký tự
    user_fullname VARCHAR(100) NOT NULL, -- Tên hiển thị của người dùng
    user_role VARCHAR(20) NOT NULL DEFAULT 'Member', -- THÊM: Phân quyền người dùng
    user_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian tạo
    user_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian cập nhật gần nhất
    CONSTRAINT chk_user_role CHECK (
        user_role IN ('ppppp                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               ,', 'Member')
    ) -- Chỉ cho phép 2 role
);

-- Bảng rooms: Lưu thông tin phòng chat (chat nhóm hoặc 1-1)
CREATE TABLE rooms (
    room_id BIGSERIAL PRIMARY KEY, -- BIGSERIAL làm khóa chính, tối ưu join và kích thước nhỏ
    room_code VARCHAR(6) UNIQUE NOT NULL, -- Mã phòng 6 ký tự để tham gia
    room_name VARCHAR(255), -- Tên phòng, NULL cho chat 1-1
    room_is_direct_chat BOOLEAN NOT NULL DEFAULT FALSE, -- Phân biệt chat 1-1 và chat nhóm
    room_created_by UUID, -- Người tạo phòng, tham chiếu đến user_uuid
    room_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian tạo
    room_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian cập nhật
    CONSTRAINT fk_created_by FOREIGN KEY (room_created_by) REFERENCES users (user_uuid) ON DELETE SET NULL, -- Khóa ngoại đến users
    CONSTRAINT chk_room_code CHECK (
        room_code ~ '^[A-Za-z0-9]{6}$'
    ) -- Mã phòng chỉ chứa 6 chữ cái/số
);

-- Bảng room_members: Lưu thành viên của phòng chat
CREATE TABLE room_members (
    user_uuid UUID NOT NULL, -- Người dùng, tham chiếu đến user_uuid
    room_id BIGINT NOT NULL, -- Phòng chat, tham chiếu đến room_id
    member_role VARCHAR(20) NOT NULL DEFAULT 'Member', -- THÊM: Vai trò trong phòng
    room_member_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian tham gia
    room_member_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian cập nhật
    CONSTRAINT pk_room_members PRIMARY KEY (user_uuid, room_id), -- Composite key tránh trùng lặp
    CONSTRAINT fk_user FOREIGN KEY (user_uuid) REFERENCES users (user_uuid) ON DELETE CASCADE, -- Xóa thành viên nếu người dùng bị xóa
    CONSTRAINT fk_room FOREIGN KEY (room_id) REFERENCES rooms (room_id) ON DELETE CASCADE, -- Xóa thành viên nếu phòng bị xóa
    CONSTRAINT chk_member_role CHECK (
        member_role IN ('Owner', 'Admin', 'Member')
    ) -- Vai trò trong phòng
);

-- Bảng messages: Lưu tin nhắn trong phòng chat
CREATE TABLE messages (
    message_id BIGSERIAL PRIMARY KEY, -- BIGSERIAL làm khóa chính, tối ưu insert và phân trang
    room_id BIGINT NOT NULL, -- Phòng chat chứa tin nhắn
    user_uuid UUID NOT NULL, -- Người gửi tin nhắn
    content TEXT NOT NULL, -- Nội dung tin nhắn, giới hạn 2000 ký tự
    message_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian gửi
    CONSTRAINT fk_room_msg FOREIGN KEY (room_id) REFERENCES rooms (room_id) ON DELETE CASCADE, -- Xóa tin nhắn nếu phòng bị xóa
    CONSTRAINT fk_user_msg FOREIGN KEY (user_uuid) REFERENCES users (user_uuid) ON DELETE CASCADE, -- Xóa tin nhắn nếu người dùng bị xóa
    CONSTRAINT chk_message_length CHECK (LENGTH(content) <= 2000) -- Giới hạn độ dài tin nhắn
);

-- Trigger để tự động cập nhật user updated_at
CREATE OR REPLACE FUNCTION update_user_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.user_updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_timestamp
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_user_timestamp();

-- Trigger để tự động cập nhật room updated_at
CREATE OR REPLACE FUNCTION update_room_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.room_updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_rooms_timestamp
BEFORE UPDATE ON rooms
FOR EACH ROW
EXECUTE FUNCTION update_room_timestamp();

-- Trigger để tự động cập nhật room_member updated_at
CREATE OR REPLACE FUNCTION update_room_member_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.room_member_updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_room_members_timestamp
BEFORE UPDATE ON room_members
FOR EACH ROW
EXECUTE FUNCTION update_room_member_timestamp();

-- THÊM: Trigger để cập nhật room_updated_at khi có tin nhắn mới
CREATE OR REPLACE FUNCTION update_room_timestamp_on_new_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE rooms
    SET room_updated_at = NOW()
    WHERE room_id = NEW.room_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_room_on_new_message
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION update_room_timestamp_on_new_message();

-- Bảng room_read_status: Theo dõi unread count của từng user trong từng phòng
CREATE TABLE room_read_status (
    user_uuid UUID NOT NULL, -- Người dùng
    room_id BIGINT NOT NULL, -- Phòng chat
    unread_count INT NOT NULL DEFAULT 0, -- Số tin nhắn chưa đọc
    last_read_message_id BIGINT, -- ID tin nhắn cuối cùng đã đọc (NULL nếu chưa đọc gì)
    last_read_at TIMESTAMPTZ, -- Thời gian đọc tin nhắn cuối cùng
    CONSTRAINT pk_room_read_status PRIMARY KEY (user_uuid, room_id),
    CONSTRAINT fk_user_read_status FOREIGN KEY (user_uuid) REFERENCES users (user_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_room_read_status FOREIGN KEY (room_id) REFERENCES rooms (room_id) ON DELETE CASCADE,
    CONSTRAINT fk_last_read_message FOREIGN KEY (last_read_message_id) REFERENCES messages (message_id) ON DELETE SET NULL,
    CONSTRAINT chk_unread_count CHECK (unread_count >= 0) -- Không cho phép số âm
);

-- Trigger: Tự động tạo room_read_status khi user join room (unread_count = 0)
CREATE OR REPLACE FUNCTION create_room_read_status_on_join()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO room_read_status (user_uuid, room_id, unread_count, last_read_at)
    VALUES (NEW.user_uuid, NEW.room_id, 0, NOW())
    ON CONFLICT (user_uuid, room_id) DO NOTHING; -- Tránh duplicate nếu đã tồn tại
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_room_read_status
AFTER INSERT ON room_members
FOR EACH ROW
EXECUTE FUNCTION create_room_read_status_on_join();

-- Chỉ mục để tối ưu hóa truy vấn ====== primary key,unique đã có index tự động ======

-- Tìm kiếm các phòng của user (load user's rooms)
CREATE INDEX idx_room_members_user_uuid ON room_members (user_uuid);

-- Tìm kiếm room_read_status của user (check unread count)
CREATE INDEX idx_room_read_status_user ON room_read_status (user_uuid);

-- Tìm kiếm room_read_status theo room (increment unread cho tất cả members)
CREATE INDEX idx_room_read_status_room ON room_read_status (room_id);