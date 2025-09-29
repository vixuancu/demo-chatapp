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

	// Async message processing
	messageQueue chan MessageTask
	workerPool   chan chan MessageTask
	quit         chan bool

	// Caching
	membershipCache *wsmanager.RoomMembershipCache
}

type MessageTask struct {
	Message wsmanager.Message
	Client  *wsmanager.Client
}

type Worker struct {
	ID          int
	Work        chan MessageTask
	WorkerQueue chan chan MessageTask
	QuitChan    chan bool
	Handler     *WebSocketHandler
}

func NewWebSocketHandler(manager *wsmanager.Manager,
	roomService services.RoomService,
	messageService services.MessageService,
	userService services.UserService,
	jwtService auth.TokenService) *WebSocketHandler {

	// Create membership cache with 5 minute TTL
	membershipCache := wsmanager.NewRoomMembershipCache(5 * time.Minute)

	handler := &WebSocketHandler{
		manager:         manager,
		roomService:     roomService,
		messageService:  messageService,
		userService:     userService,
		jwtService:      jwtService,
		messageQueue:    make(chan MessageTask, 1000),    // Buffer 1000 messages
		workerPool:      make(chan chan MessageTask, 10), // 10 workers
		quit:            make(chan bool),
		membershipCache: membershipCache,
	}

	// Start worker pool
	handler.StartWorkerPool(10)

	// Setup cached room membership callback
	originalCallback := func(userUUID uuid.UUID, roomID int64) (bool, error) {
		return roomService.IsUserMemberOfRoom(context.Background(), userUUID, roomID)
	}
	cachedCallback := wsmanager.CachedRoomMembershipCheckFunc(originalCallback, membershipCache)
	manager.SetRoomMembershipCallback(cachedCallback)

	return handler
}

// StartWorkerPool starts the worker pool for async message processing
func (wh *WebSocketHandler) StartWorkerPool(numWorkers int) {
	// Start dispatcher
	go wh.dispatcher()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		worker := Worker{
			ID:          i + 1,
			Work:        make(chan MessageTask),
			WorkerQueue: wh.workerPool,
			QuitChan:    make(chan bool),
			Handler:     wh,
		}
		go worker.Start()
	}
}

// Dispatcher distributes work to available workers
func (wh *WebSocketHandler) dispatcher() {
	for {
		select {
		case work := <-wh.messageQueue:
			// Get an available worker
			go func() {
				worker := <-wh.workerPool
				worker <- work
			}()
		case <-wh.quit:
			return
		}
	}
}

// Worker processes messages asynchronously
func (w *Worker) Start() {
	go func() {
		for {
			// Register worker in the worker queue
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Process the message
				w.Handler.processMessageAsync(work)

			case <-w.QuitChan:
				return
			}
		}
	}()
}

// HandleWebSocket xử lý kết nối WebSocket
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

	// Create new client with configurable buffer size
	client := &wsmanager.Client{
		ID:       uuid.New().String(),
		UserUUID: userID,
		Conn:     conn,
		Send:     make(chan []byte, 1024), // Increased buffer size
		Rooms:    make(map[int64]bool),
	}

	// Register client
	wh.manager.Register(client)

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
		// Parse message quickly and queue for async processing
		var msg wsmanager.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("❌ Error parsing message from client %s: %v", client.ID, err)
			continue
		}

		// Queue for async processing instead of blocking read loop
		select {
		case wh.messageQueue <- MessageTask{Message: msg, Client: client}:
			// Queued successfully
		default:
			// Queue is full, log warning and skip
			log.Printf("⚠️ Message queue full, dropping message from client %s", client.ID)
		}
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

// processMessageAsync processes messages in worker goroutines (non-blocking)
func (wh *WebSocketHandler) processMessageAsync(task MessageTask) {
	msg := task.Message
	client := task.Client

	log.Printf("📨 Processing message type '%s' from client %s for room %d", msg.Type, client.ID, msg.RoomID)

	// Process based on message type
	switch msg.Type {
	case "join_room":
		wh.handleJoinRoom(msg, client)
	case "leave_room":
		wh.handleLeaveRoom(msg, client)
	case "send_message":
		wh.handleSendMessage(msg, client)
	default:
		log.Printf("❓ Unknown message type: %s", msg.Type)
	}
}

// handleJoinRoom processes join room requests
func (wh *WebSocketHandler) handleJoinRoom(msg wsmanager.Message, client *wsmanager.Client) {
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

	wh.sendToClient(client, response)
}

// handleLeaveRoom processes leave room requests
func (wh *WebSocketHandler) handleLeaveRoom(msg wsmanager.Message, client *wsmanager.Client) {
	roomID := msg.RoomID
	wh.manager.LeaveRoom(roomID, client)

	response := wsmanager.Message{
		Type:    "room_response",
		RoomID:  roomID,
		Content: "Successfully left room",
	}

	wh.sendToClient(client, response)
}

// handleSendMessage processes send message requests (async, DB operations)
func (wh *WebSocketHandler) handleSendMessage(msg wsmanager.Message, client *wsmanager.Client) {
	// Quick validation
	if !wh.manager.IsClientInRoom(msg.RoomID, client.ID) {
		errMsg := wsmanager.Message{
			Type:    "error",
			Content: "You must join the room first",
		}
		wh.sendToClient(client, errMsg)
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
		wh.sendToClient(client, errMsg)
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
		MessageID: &message.MessageID,
		Priority:  1,
	}

	log.Printf("📡 Broadcasting message %d to room %d by user %s", message.MessageID, message.RoomID, user.UserFullname)
	wh.manager.SendToRoom(broadcastMsg)
	log.Printf("✅ Message broadcast completed")
}

// sendToClient safely sends message to client with backpressure handling
func (wh *WebSocketHandler) sendToClient(client *wsmanager.Client, msg wsmanager.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Error marshaling message: %v", err)
		return
	}

	select {
	case client.Send <- data:
		// Sent successfully
	default:
		// Channel is full, client is slow - force disconnect
		log.Printf("⚠️ Client %s send buffer full - disconnecting slow client", client.ID)
		wh.manager.Unregister(client)
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
