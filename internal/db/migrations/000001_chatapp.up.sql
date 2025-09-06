-- câu lệnh sql thì viết hoa còn lại thì viết thường
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Bảng users: Lưu thông tin người dùng
CREATE TABLE users (
    user_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- UUID làm khóa chính, bảo mật và scale tốt
    user_email VARCHAR(100) UNIQUE NOT NULL, -- Email duy nhất, giới hạn 100 ký tự
    user_password VARCHAR(255) NOT NULL, -- Mật khẩu mã hóa (bcrypt/argon2), giới hạn 255 ký tự
    user_fullname VARCHAR(100) NOT NULL, -- Tên hiển thị của người dùng
    user_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian tạo
    user_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW() -- Thời gian cập nhật gần nhất
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
    room_member_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian tham gia
    room_member_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Thời gian cập nhật
    CONSTRAINT pk_room_members PRIMARY KEY (user_uuid, room_id), -- Composite key tránh trùng lặp
    CONSTRAINT fk_user FOREIGN KEY (user_uuid) REFERENCES users (user_uuid) ON DELETE CASCADE, -- Xóa thành viên nếu người dùng bị xóa
    CONSTRAINT fk_room FOREIGN KEY (room_id) REFERENCES rooms (room_id) ON DELETE CASCADE -- Xóa thành viên nếu phòng bị xóa
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

-- Chỉ mục để tối ưu hóa truy vấn ====== primary key,unique đã có index tự động ======

-- Tìm kiếm các phòng của user (load user's rooms)
CREATE INDEX idx_room_members_user_uuid ON room_members (user_uuid);

