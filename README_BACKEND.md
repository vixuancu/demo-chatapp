# Chat App Backend - Realtime WebSocket Chat

Một backend chat realtime được xây dựng với Go, Gin, WebSocket, và PostgreSQL. Hỗ trợ đa phòng, đa người dùng với tin nhắn thời gian thực.

## ✨ Tính năng

### 🔥 Realtime Chat

- **WebSocket Connection**: Kết nối thời gian thực an toàn với JWT authentication
- **Multi-Room Support**: Hỗ trợ nhiều phòng chat đồng thời
- **Message Ordering**: Đảm bảo thứ tự tin nhắn khi nhiều người gửi cùng lúc
- **Concurrency Safe**: An toàn với nhiều kết nối đồng thời, tránh race conditions
- **Auto-Reconnect**: Xử lý kết nối bị đứt và tự động kết nối lại

### 👥 User Management

- **JWT Authentication**: Bảo mật với JSON Web Token
- **User Roles**: Hỗ trợ phân quyền Admin/Member
- **User Registration/Login**: Đăng ký và đăng nhập người dùng

### 🏠 Room Management

- **Create Rooms**: Tạo phòng chat với mã phòng 6 ký tự
- **Join/Leave**: Tham gia và rời phòng qua mã phòng hoặc room ID
- **Room Members**: Quản lý thành viên trong phòng
- **Direct Chat**: Hỗ trợ chat 1-1 và chat nhóm

### 💬 Message Features

- **Persistent Messages**: Lưu trữ tin nhắn trong database
- **Message History**: Lấy lịch sử tin nhắn với pagination
- **User Info**: Tin nhắn kèm thông tin người gửi
- **Message Timestamps**: Thời gian gửi chính xác

## 🏗️ Kiến trúc

### Clean Architecture

```
├── cmd/api/                    # Application entry point
├── internal/
│   ├── app/                    # Module registration
│   ├── handlers/v1/            # HTTP & WebSocket handlers
│   ├── services/v1/            # Business logic
│   ├── repository/             # Data access layer
│   ├── middleware/             # Authentication, CORS
│   ├── routes/                 # Route registration
│   └── dto/v1/                 # Data transfer objects
├── pkg/
│   ├── websocket/              # WebSocket manager
│   ├── auth/                   # JWT service
│   └── cache/                  # Redis cache
```

### WebSocket Architecture

```
Client 1 ─┐
Client 2 ─┼── WebSocket Manager ── Message Queue ── Database
Client N ─┘                     ├── Room 1 Queue
                                ├── Room 2 Queue
                                └── Room N Queue
```

## 🚀 API Endpoints

### Authentication

```http
POST /api/v1/auth/register     # Đăng ký người dùng
POST /api/v1/auth/login        # Đăng nhập
POST /api/v1/auth/logout       # Đăng xuất
```

### Rooms

```http
GET    /api/v1/rooms                    # Lấy danh sách phòng của user
POST   /api/v1/rooms                    # Tạo phòng mới
POST   /api/v1/rooms/join-by-code       # Tham gia phòng bằng mã
POST   /api/v1/rooms/{roomID}/join      # Tham gia phòng bằng ID
POST   /api/v1/rooms/{roomID}/leave     # Rời phòng
GET    /api/v1/rooms/{roomID}/members   # Lấy danh sách thành viên
```

### Messages

```http
GET  /api/v1/rooms/{roomID}/messages    # Lấy lịch sử tin nhắn
POST /api/v1/rooms/{roomID}/messages    # Gửi tin nhắn (REST)
```

### WebSocket

```http
WS /api/v1/chat/ws?token={JWT_TOKEN}&room_id={ROOM_ID}
```

### WebSocket Status

```http
GET /api/v1/chat/rooms/{roomID}/status  # Trạng thái phòng realtime
```

## 📡 WebSocket Messages

### Client → Server

#### Join Room

```json
{
  "type": "join_room",
  "room_id": 1
}
```

#### Leave Room

```json
{
  "type": "leave_room",
  "room_id": 1
}
```

#### Send Message

```json
{
  "type": "send_message",
  "room_id": 1,
  "content": "Hello everyone!"
}
```

### Server → Client

#### New Message

```json
{
  "type": "new_message",
  "room_id": 1,
  "user_uuid": "uuid-here",
  "content": "Hello everyone!",
  "timestamp": "2023-09-28T10:30:00Z",
  "message_id": 123,
  "data": {
    "message_id": 123,
    "content": "Hello everyone!",
    "user_uuid": "uuid-here",
    "user_fullname": "John Doe",
    "user_email": "john@example.com",
    "created_at": "2023-09-28T10:30:00Z"
  }
}
```

#### User Joined/Left

```json
{
  "type": "user_joined",
  "room_id": 1,
  "user_uuid": "uuid-here",
  "data": {
    "user_uuid": "uuid-here"
  }
}
```

#### Error Messages

```json
{
  "type": "error",
  "content": "You must join the room first"
}
```

## 🛠️ Setup & Installation

### Prerequisites

- Go 1.19+
- PostgreSQL 12+
- Redis (optional)
- Node.js (for testing)

### 1. Clone Repository

```bash
git clone <repository-url>
cd chat-app_server
```

### 2. Install Dependencies

```bash
go mod tidy
npm install ws  # For WebSocket testing
```

### 3. Environment Setup

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=chatapp

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRY_HOURS=24

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Server
PORT=8080
```

### 4. Database Setup

```bash
# Create database
createdb chatapp

# Run migrations
make migrate-up
# Or manually:
migrate -path internal/db/migrations -database "postgres://user:password@localhost/chatapp?sslmode=disable" up
```

### 5. Generate SQLC (if needed)

```bash
sqlc generate
```

### 6. Run Server

```bash
# Development
go run cmd/api/main.go

# Or using Makefile
make run

# Build and run
make build
./bin/server
```

## 🧪 Testing

### 1. Test REST APIs

```bash
# Run comprehensive API tests
./test_api.sh

# Test specific endpoints
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"user_email":"test@example.com","user_password":"password123","user_fullname":"Test User"}'
```

### 2. Test WebSocket

```bash
# Run WebSocket tests (requires tokens from API test)
node test_websocket_enhanced.js

# Or use the existing test
node test_websocket.js
```

### 3. Manual WebSocket Testing

```javascript
// In browser console or Node.js
const ws = new WebSocket(
  "ws://localhost:8080/api/v1/chat/ws?token=YOUR_JWT_TOKEN"
);

ws.onopen = () => {
  console.log("Connected!");

  // Join room
  ws.send(
    JSON.stringify({
      type: "join_room",
      room_id: 1,
    })
  );

  // Send message
  ws.send(
    JSON.stringify({
      type: "send_message",
      room_id: 1,
      content: "Hello World!",
    })
  );
};

ws.onmessage = (event) => {
  console.log("Received:", JSON.parse(event.data));
};
```

## 🏆 Key Features Implementation

### 1. Message Ordering (Thứ tự tin nhắn)

- **Per-Room Queues**: Mỗi phòng có queue riêng để xử lý tin nhắn theo thứ tự
- **Message ID**: Tin nhắn có ID auto-increment đảm bảo thứ tự
- **Priority System**: Hệ thống priority cho các loại tin nhắn khác nhau

### 2. Concurrency Safety

- **Channel-based Communication**: Sử dụng Go channels để tránh race conditions
- **Mutex Locking**: RWMutex để bảo vệ shared data structures
- **Goroutine Safety**: Mỗi client có goroutine riêng cho read/write

### 3. Connection Management

- **Unique Connection**: Một user chỉ có một kết nối active per room
- **Auto-cleanup**: Tự động dọn dẹp kết nối bị lỗi
- **Graceful Shutdown**: Xử lý graceful khi client disconnect

### 4. Room Isolation

- **Separate Queues**: Mỗi phòng có message queue độc lập
- **User Tracking**: Theo dõi user trong từng phòng riêng biệt
- **Cross-room Protection**: Tin nhắn chỉ được gửi đến đúng phòng

## 📊 Database Schema

### Users

```sql
users (
  user_uuid UUID PRIMARY KEY,
  user_email VARCHAR(100) UNIQUE,
  user_password VARCHAR(255),
  user_fullname VARCHAR(100),
  user_role VARCHAR(20) DEFAULT 'Member',
  user_created_at TIMESTAMPTZ,
  user_updated_at TIMESTAMPTZ
)
```

### Rooms

```sql
rooms (
  room_id BIGSERIAL PRIMARY KEY,
  room_code VARCHAR(6) UNIQUE,
  room_name VARCHAR(255),
  room_is_direct_chat BOOLEAN DEFAULT FALSE,
  room_created_by UUID,
  room_created_at TIMESTAMPTZ,
  room_updated_at TIMESTAMPTZ
)
```

### Room Members

```sql
room_members (
  user_uuid UUID,
  room_id BIGINT,
  member_role VARCHAR(20) DEFAULT 'Member',
  room_member_created_at TIMESTAMPTZ,
  room_member_updated_at TIMESTAMPTZ,
  PRIMARY KEY (user_uuid, room_id)
)
```

### Messages

```sql
messages (
  message_id BIGSERIAL PRIMARY KEY,
  room_id BIGINT,
  user_uuid UUID,
  content TEXT,
  message_created_at TIMESTAMPTZ
)
```

## 🔧 Development Tools

### Makefile Commands

```bash
make run          # Run development server
make build        # Build binary
make test         # Run tests
make migrate-up   # Run database migrations
make migrate-down # Rollback migrations
make sqlc         # Generate SQLC code
```

### Docker Support

```bash
# Start dependencies
docker-compose up -d postgres redis

# Run full stack
docker-compose up
```

## 🚨 Important Notes

### Security

- **JWT Authentication**: Tất cả endpoints đều yêu cầu authentication
- **CORS Configuration**: Cấu hình CORS cho development và production
- **SQL Injection Protection**: Sử dụng SQLC để tránh SQL injection
- **Input Validation**: Validate input ở tầng handler

### Performance

- **Connection Pooling**: Database connection pooling
- **Message Queuing**: Per-room queuing để tránh bottleneck
- **Efficient Queries**: Optimized SQL queries với indexes
- **Memory Management**: Proper cleanup của resources

### Scalability

- **Horizontal Scaling**: Có thể scale bằng load balancer
- **Redis Support**: Sẵn sàng cho Redis clustering
- **Database Sharding**: Cấu trúc DB hỗ trợ sharding nếu cần

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## 📝 License

MIT License - xem file LICENSE để biết thêm chi tiết.

---

**Backend này đã sẵn sàng cho production với đầy đủ tính năng của một ứng dụng chat realtime!** 🚀
