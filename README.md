# 💬 Chat Application - Backend Server

> Real-time chat application backend built with Go, WebSocket, PostgreSQL, and Redis

---

## 📋 Giới thiệu

**Chat Application Backend** là một hệ thống chat real-time được xây dựng với kiến trúc clean architecture, hỗ trợ nhiều người dùng tham gia các phòng chat, gửi tin nhắn tức thời qua WebSocket, và quản lý phòng chat một cách linh hoạt.

### ✨ Đặc điểm nổi bật

- **Real-time Communication**: Sử dụng WebSocket để giao tiếp tức thời
- **Scalable Architecture**: Thiết kế modular với clean architecture pattern
- **High Performance**: Tối ưu hóa với Redis cache và connection pooling
- **Secure**: JWT authentication, password hashing với bcrypt
- **Production Ready**: Containerized với Docker, có health checks và graceful shutdown

---

## 🛠 Công nghệ sử dụng

### Core Technologies

| Công nghệ      | Phiên bản         | Mục đích                     |
| -------------- | ----------------- | ---------------------------- |
| **Go**         | 1.24.5            | Backend language             |
| **Gin**        | 1.10.1            | Web framework                |
| **PostgreSQL** | 17.5              | Primary database             |
| **Redis**      | 8.0               | Caching & session management |
| **WebSocket**  | gorilla/websocket | Real-time communication      |

### Libraries & Tools

#### Authentication & Security

- `golang-jwt/jwt/v5` - JWT token generation & validation
- `golang.org/x/crypto` - Password hashing (bcrypt)

#### Database

- `jackc/pgx/v5` - PostgreSQL driver với connection pooling
- `sqlc` - Type-safe SQL code generation

#### Validation & Utils

- `go-playground/validator/v10` - Request validation
- `google/uuid` - UUID generation
- `joho/godotenv` - Environment variables management

#### DevOps

- **Docker** - Containerization
- **Docker Compose** - Multi-container orchestration
- **golang-migrate** - Database migrations
- **Makefile** - Task automation

---

## ⚙️ Kiến trúc hệ thống

```
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── app/             # Application modules initialization
│   ├── config/          # Configuration management
│   ├── db/              # Database connection & migrations
│   │   ├── migrations/  # SQL migration files
│   │   ├── queries/     # SQL queries for sqlc
│   │   └── sqlc/        # Generated type-safe Go code
│   ├── dto/             # Data Transfer Objects
│   ├── handlers/        # HTTP & WebSocket handlers
│   ├── middleware/      # Middleware (auth, cors, etc.)
│   ├── repository/      # Data access layer
│   ├── routes/          # Route definitions
│   ├── services/        # Business logic layer
│   ├── utils/           # Helper functions
│   └── validation/      # Custom validators
├── pkg/
│   ├── auth/            # JWT service
│   ├── cache/           # Redis cache service
│   └── websocket/       # WebSocket manager & cache
└── system/
    └── redis/           # Redis configuration
```

### Design Patterns

- **Clean Architecture**: Tách biệt business logic và infrastructure
- **Repository Pattern**: Abstract data access layer
- **Dependency Injection**: Loose coupling giữa các module
- **Factory Pattern**: Khởi tạo objects thông qua factory functions

---

## 🚀 Tính năng chính

### 1. 🔐 Authentication & Authorization

- **User Registration**: Đăng ký tài khoản với email và password
- **User Login**: Xác thực và nhận JWT token
- **JWT Authentication**: Bảo vệ các API endpoints
- **Role-Based Access Control**: Phân quyền User và Admin
- **Session Management**: Logout và invalidate tokens

**Endpoints:**

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
GET    /api/v1/auth/me
```

---

### 2. 💬 Real-time Chat với WebSocket

- **WebSocket Connection**: Kết nối persistent cho real-time communication
- **Room-based Messaging**: Gửi tin nhắn theo phòng chat
- **Join/Leave Room**: Tham gia và rời khỏi phòng động
- **Typing Indicators**: Hiển thị trạng thái đang nhập
- **Message Broadcasting**: Broadcast tin nhắn đến tất cả members trong room
- **Connection Management**: Auto-reconnect và graceful disconnection
- **Message Queue**: Đảm bảo thứ tự tin nhắn với room-specific queues
- **Worker Pool**: Xử lý tin nhắn async với worker pool pattern

**WebSocket Events:**

```javascript
// Client → Server
{ type: "join_room", room_id: 123 }
{ type: "leave_room", room_id: 123 }
{ type: "send_message", room_id: 123, content: "Hello" }
{ type: "typing_start", room_id: 123 }
{ type: "typing_stop", room_id: 123 }

// Server → Client
{ type: "message", room_id: 123, content: "...", user_uuid: "..." }
{ type: "user_joined", room_id: 123, user_uuid: "..." }
{ type: "user_left", room_id: 123, user_uuid: "..." }
{ type: "typing", room_id: 123, user_uuid: "..." }
```

**WebSocket Connection:**

```
WS /api/v1/chat/ws?token={JWT_TOKEN}&room_id={ROOM_ID}
```

---

### 3. 🏠 Room Management

- **Create Room**: Tạo phòng chat mới với mã phòng tự động
- **Join Room by Code**: Tham gia phòng bằng mã 6 ký tự
- **Join Room by ID**: Tham gia phòng trực tiếp bằng ID
- **Leave Room**: Rời khỏi phòng chat
- **List User Rooms**: Xem danh sách phòng của user
- **Get Room Members**: Xem danh sách thành viên trong phòng
- **Direct Chat**: Hỗ trợ chat 1-1
- **Mark as Read**: Đánh dấu đã đọc tin nhắn trong phòng
- **Unread Count**: Đếm số tin nhắn chưa đọc


---

### 7. ⚡ Performance Optimization

#### Redis Caching

- **JWT Token Cache**: Cache token validation (planning)
- **Room Membership Cache**: Cache membership checks (TTL 5 phút)
  - Giảm 90-99% database queries cho membership validation
  - In-memory map với auto-cleanup

#### Database Optimization

- **Connection Pooling**: PostgreSQL connection pool với pgx
- **Prepared Statements**: Type-safe queries với sqlc
- **Indexes**: Tối ưu queries cho room_members, messages

#### WebSocket Optimization

- **Room Message Queues**: Đảm bảo thứ tự tin nhắn per room
- **Worker Pool**: Async message processing
- **Client Buffering**: Buffered channels (1024 messages)
- **Graceful Cleanup**: Auto-remove empty rooms

---

## 📦 Cài đặt và Chạy

### Prerequisites

- Go 1.24.5+
- Docker & Docker Compose
- PostgreSQL 17.5+
- Redis 8.0+
- Make (optional)

### 1. Clone repository

```bash
git clone https://github.com/vixuancu/demo-chatapp.git
cd chat-app_server
```

### 2. Cấu hình môi trường

Tạo file `.env`:

```bash
# Server
SERVER_PORT=8081
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=chat-app
DB_SSLMODE=disable

# Redis
REDIS_ADDR=localhost:6379
REDIS_USER=
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# JWT
JWT_SECRET=your_super_secret_jwt_key_here
```

### 3. Chạy với Docker Compose (Recommended)

```bash
# Build và chạy tất cả services
docker-compose up -d --build

# Xem logs
docker-compose logs -f chat-api

# Dừng services
docker-compose down
```

### 4. Chạy Local Development

```bash
# Start database & redis
make up-container

# Run migrations
make migrate-up

# Generate sqlc code
make sqlc

# Run server
make server
```

---

## 📝 Database Migrations

```bash
# Tạo migration mới
make migrate-create NAME=add_new_feature

# Apply migrations
make migrate-up

# Rollback 1 migration
make migrate-down

# Rollback n migrations
make migrate-down-n N=2

# Force version
make migrate-force VERSION=1

# Go to specific version
make migrate-goto VERSION=3
```

---

## 🧪 API Testing

### Health Check

```bash
curl http://localhost:8081/api/v1/chat/health
```

### Register User

```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "fullname": "John Doe"
  }'
```

### Login

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### WebSocket Connection (với JavaScript)

```javascript
const token = "your_jwt_token";
const ws = new WebSocket(`ws://localhost:8081/api/v1/chat/ws?token=${token}`);

ws.onopen = () => {
  // Join room
  ws.send(
    JSON.stringify({
      type: "join_room",
      room_id: 123,
    })
  );

  // Send message
  ws.send(
    JSON.stringify({
      type: "send_message",
      room_id: 123,
      content: "Hello World!",
    })
  );
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
};
```

---

## 🔧 Makefile Commands

```bash
make up-container         # Start Docker containers
make remove-container     # Remove Docker containers
make stop-container       # Stop containers
make restart-container    # Restart containers

make sqlc                 # Generate sqlc code
make migrate-create       # Create new migration
make migrate-up           # Apply migrations
make migrate-down         # Rollback migration

make server               # Run server locally
make build                # Build binary
```

---


## 👨‍💻 Author

**Vi Xuan Cu**

- GitHub: [@vixuancu](https://github.com/vixuancu)
- Repository: [demo-chatapp](https://github.com/vixuancu/demo-chatapp)

---

## 📄 License

This project is open source and available under the MIT License.

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📞 Support

If you have any questions or need help, please open an issue on GitHub.

---

**⭐ If you find this project useful, please give it a star!**
