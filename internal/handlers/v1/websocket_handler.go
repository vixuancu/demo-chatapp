package v1Handler

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/services/v1"
	"chat-app/pkg/auth"
	wsmanager "chat-app/pkg/websocket"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow localhost for development
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" ||
			origin == "http://127.0.0.1:3000" ||
			origin == "" // Allow direct connections
	},
	Subprotocols: []string{"chat"}, // Optional: specify subprotocol
}

type WebSocketHandler struct {
	manager        *wsmanager.Manager
	roomService    services.RoomService
	messageService services.MessageService
	userService    services.UserService
	jwtService     auth.TokenService
}

func NewWebSocketHandler(manager *wsmanager.Manager,
	roomService services.RoomService,
	messageService services.MessageService,
	userService services.UserService,
	jwtService auth.TokenService) *WebSocketHandler {

	return &WebSocketHandler{
		manager:        manager,
		roomService:    roomService,
		messageService: messageService,
		userService:    userService,
		jwtService:     jwtService,
	}
}

// Sửa hàm HandleWebSocket để không yêu cầu room_id ban đầu
func (wh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// Validate token
	claims, err := wh.jwtService.ValidateJWTToken(token)
	if err != nil {
		log.Printf("WebSocket auth failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Convert to UUID
	userID, err := uuid.Parse(claims.UserUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Bỏ yêu cầu room_id khi kết nối
	log.Printf("WebSocket auth success for user: %s", claims.UserUUID)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := &wsmanager.Client{
		ID:       uuid.New().String(),
		UserUUID: userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[int64]bool),
	}

	// Register client
	wh.manager.Register(client)

	// Thiết lập callback để kiểm tra quyền phòng
	wh.manager.SetRoomMembershipCallback(func(userUUID uuid.UUID, roomID int64) (bool, error) {
		return wh.roomService.IsUserMemberOfRoom(context.Background(), userUUID, roomID)
	})

	// Optional: Auto-join room if room_id provided in query param
	roomIDStr := c.Query("room_id")
	if roomIDStr != "" && roomIDStr != "0" {
		roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
		if err == nil {
			err = wh.manager.JoinRoom(roomID, client)
			if err != nil {
				log.Printf("Failed to auto-join room %d: %v", roomID, err)
			} else {
				log.Printf("Client %s auto-joined room %d", client.ID, roomID)
			}
		}
	}

	// Start read/write pumps
	go wh.readPump(client)
	go wh.writePump(client)
}

func (wh *WebSocketHandler) readPump(client *wsmanager.Client) {
	defer func() {
		log.Printf("Client %s disconnecting from readPump", client.ID)
		wh.manager.Unregister(client)
		client.Conn.Close()
	}()

	// Set limits and timeouts
	client.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				log.Printf("Unexpected WebSocket close error for client %s: %v", client.ID, err)
			} else {
				log.Printf("Client %s closed WebSocket connection: %v", client.ID, err)
			}
			break
		}

		log.Printf("Received message from client %s: %s", client.ID, string(message))
		wh.processMessage(message, client)
	}
}

func (wh *WebSocketHandler) writePump(client *wsmanager.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// ✅ Send single message properly
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping: %v", err)
				return
			}
		}
	}
}

func (wh *WebSocketHandler) processMessage(data []byte, client *wsmanager.Client) {
	// Parse message
	var msg wsmanager.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("❌ Error parsing message from client %s: %v", client.ID, err)
		return
	}

	log.Printf("📨 Processing message type '%s' from client %s for room %d", msg.Type, client.ID, msg.RoomID)

	// Process based on message type
	switch msg.Type {
	case "join_room":
		// Process join room request
		roomID := msg.RoomID
		err := wh.manager.JoinRoom(roomID, client)

		response := wsmanager.Message{
			Type:   "room_response",
			RoomID: roomID,
		}

		if err != nil {
			response.Content = fmt.Sprintf("Error joining room: %v", err)
			response.Type = "error"
		} else {
			response.Content = "Successfully joined room"
		}

		if data, err := json.Marshal(response); err == nil {
			client.Send <- data
		}

	case "leave_room":
		// Process leave room request
		roomID := msg.RoomID
		wh.manager.LeaveRoom(roomID, client)

		response := wsmanager.Message{
			Type:    "room_response",
			RoomID:  roomID,
			Content: "Successfully left room",
		}

		if data, err := json.Marshal(response); err == nil {
			client.Send <- data
		}

	case "send_message":
		// Kiểm tra xem client có trong room không
		if !wh.manager.IsClientInRoom(msg.RoomID, client.ID) {
			errMsg := wsmanager.Message{
				Type:    "error",
				Content: "You must join the room first",
			}
			if data, err := json.Marshal(errMsg); err == nil {
				client.Send <- data
			}
			return
		}

		log.Printf("✅ Client %s is in room %d - processing message", client.ID, msg.RoomID)

		// Create and save message to DB
		params := sqlc.CreateMessageParams{
			RoomID:   msg.RoomID,
			UserUuid: client.UserUUID,
			Content:  msg.Content,
		}

		message, err := wh.messageService.CreateMessage(context.Background(), params)
		if err != nil {
			log.Printf("❌ Error saving message to DB: %v", err)
			errMsg := wsmanager.Message{
				Type:    "error",
				Content: "Failed to save message",
			}
			if data, err := json.Marshal(errMsg); err == nil {
				client.Send <- data
			}
			return
		}

		log.Printf("✅ Message saved to DB with ID %d", message.MessageID)

		// Get user info for broadcast
		user, err := wh.userService.GetUserByUUIDWithContext(context.Background(), client.UserUUID.String())
		if err != nil {
			log.Printf("❌ Error getting user info: %v", err)
			return
		}

		// Prepare message data with user info
		messageData := map[string]interface{}{
			"message_id":    message.MessageID,
			"content":       message.Content,
			"user_uuid":     message.UserUuid.String(),
			"user_fullname": user.UserFullname,
			"user_email":    user.UserEmail,
			"created_at":    message.MessageCreatedAt.Format(time.RFC3339),
		}

		dataBytes, _ := json.Marshal(messageData)

		// Broadcast message to room với thứ tự đảm bảo
		broadcastMsg := wsmanager.Message{
			Type:      "new_message",
			RoomID:    message.RoomID,
			UserUUID:  message.UserUuid,
			Content:   message.Content,
			Timestamp: message.MessageCreatedAt.Format(time.RFC3339),
			Data:      dataBytes,
			MessageID: &message.MessageID, // Thêm message_id để đảm bảo thứ tự
			Priority:  1,                  // Priority cao cho tin nhắn thường
		}

		log.Printf("📡 Broadcasting message %d to room %d by user %s", message.MessageID, message.RoomID, user.UserFullname)
		wh.manager.SendToRoom(broadcastMsg)
		log.Printf("✅ Message broadcast completed")

	default:
		log.Printf("❓ Unknown message type: %s", msg.Type)
	}
}
func (wh *WebSocketHandler) GetRoomStatus(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid room ID"})
		return
	}

	// Sử dụng phương thức an toàn để lấy thông tin phòng
	clients, exists := wh.manager.GetRoomInfo(roomID)

	response := gin.H{
		"room_id": roomID,
		"exists":  exists,
	}

	if exists {
		clientInfo := make([]gin.H, 0, len(clients))
		userCount := make(map[string]int) // Count connections per user

		for clientID, client := range clients {
			userUUID := client.UserUUID.String()
			userCount[userUUID]++

			clientInfo = append(clientInfo, gin.H{
				"client_id": clientID,
				"user_uuid": userUUID,
			})
		}

		response["client_count"] = len(clients)
		response["unique_users"] = len(userCount)
		response["clients"] = clientInfo
		response["user_connections"] = userCount
	} else {
		response["client_count"] = 0
		response["unique_users"] = 0
		response["clients"] = []gin.H{}
	}

	c.JSON(200, response)
}
