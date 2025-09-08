package v1Handler

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/services/v1"
	"chat-app/internal/utils"
	wsmanager "chat-app/pkg/websocket"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// For development, allow all origins
		return true
	},
}

type WebSocketHandler struct {
	manager        *wsmanager.Manager
	roomService    services.RoomService
	messageService services.MessageService
	userService    services.UserService
}

func NewWebSocketHandler(manager *wsmanager.Manager,
	roomService services.RoomService,
	messageService services.MessageService,
	userService services.UserService) *WebSocketHandler {

	return &WebSocketHandler{
		manager:        manager,
		roomService:    roomService,
		messageService: messageService,
		userService:    userService,
	}
}

func (wh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user UUID from authentication (assuming JWT auth)
	userUUID, exists := c.Get("userUUID")
	if !exists {
		utils.ResponseError(c, utils.NewError("unauthorized", utils.ErrorCodeUnauthorized))
		return
	}

	// Convert to UUID
	userID, err := uuid.Parse(userUUID.(string))
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest))
		return
	}

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

	// Start read/write pumps
	go wh.readPump(client)
	go wh.writePump(client)
}

func (wh *WebSocketHandler) readPump(client *wsmanager.Client) {
	defer func() {
		wh.manager.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Process message
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

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (wh *WebSocketHandler) processMessage(data []byte, client *wsmanager.Client) {
	// Parse message
	var msg wsmanager.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Process based on message type
	switch msg.Type {
	case "join_room":
		// Join a room
		var payload struct {
			RoomID int64 `json:"room_id"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("Invalid join room payload: %v", err)
			return
		}

		// Check if user is member of the room
		isMember, err := wh.roomService.IsUserMemberOfRoom(context.Background(), client.UserUUID, payload.RoomID)
		if err != nil || !isMember {
			// Send error response
			errMsg := wsmanager.Message{
				Type:    "error",
				Content: "You are not a member of this room",
			}
			data, _ := json.Marshal(errMsg)
			client.Send <- data
			return
		}

		// Add client to room
		wh.manager.AddClientToRoom(payload.RoomID, client)

		// Send confirmation
		response := wsmanager.Message{
			Type:   "room_joined",
			RoomID: payload.RoomID,
		}
		responseData, _ := json.Marshal(response)
		client.Send <- responseData

	case "leave_room":
		var payload struct {
			RoomID int64 `json:"room_id"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("Invalid leave room payload: %v", err)
			return
		}

		// Remove client from room
		wh.manager.RemoveClientFromRoom(payload.RoomID, client)

	case "send_message":
		// Create and save message to DB
		params := sqlc.CreateMessageParams{
			RoomID:   msg.RoomID,
			UserUuid: client.UserUUID,
			Content:  msg.Content,
		}

		message, err := wh.messageService.CreateMessage(context.Background(), params)
		if err != nil {
			log.Printf("Error saving message: %v", err)
			return
		}

		// Broadcast message to room
		broadcastMsg := wsmanager.Message{
			Type:      "new_message",
			RoomID:    message.RoomID,
			UserUUID:  message.UserUuid,
			Content:   message.Content,
			Timestamp: message.MessageCreatedAt.Format(time.RFC3339),
		}

		wh.manager.SendToRoom(broadcastMsg)
	}
}
