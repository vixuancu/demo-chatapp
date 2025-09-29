package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client đại diện cho một kết nối WebSocket
type Client struct {
	ID       string
	UserUUID uuid.UUID
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[int64]bool // Rooms để lưu trữ các phòng mà client đã tham gia
}

// Message đại diện cho một tin nhắn được gửi qua WebSocket
type Message struct {
	Type      string          `json:"type"`
	RoomID    int64           `json:"room_id"`
	UserUUID  uuid.UUID       `json:"user_uuid"`
	Content   string          `json:"content,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	MessageID *int64          `json:"message_id,omitempty"` // Thêm ID để đảm bảo thứ tự
	Priority  int             `json:"priority,omitempty"`   // Thêm priority để xử lý thứ tự
}

// Configuration for Manager
type ManagerConfig struct {
	ClientBufferSize   int
	RoomQueueSize      int
	BroadcastQueueSize int
	MaxWorkers         int
}

func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		ClientBufferSize:   1024, // Increased from 256
		RoomQueueSize:      1000, // Increased from 50
		BroadcastQueueSize: 1000, // Increased from 100
		MaxWorkers:         10,
	}
}

// Manager quản lý tất cả các kết nối WebSocket và các phòng
type Manager struct {
	// Core data structures - single mutex for all
	mu                sync.RWMutex
	clients           map[string]*Client
	rooms             map[int64]map[string]*Client
	userConnections   map[string]map[int64]*Client // userUUID -> roomID -> client
	roomMessageQueues map[int64]chan Message

	// Channels for async operations
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	cleanup    chan int64 // Room cleanup channel

	// Configuration
	config ManagerConfig

	// Room membership callback function
	roomMembershipCallback RoomMembershipCheckFunc

	// Cleanup tracking
	roomCleanup map[int64]*time.Timer
}

// RoomMembershipCheckFunc callback để kiểm tra quyền phòng
type RoomMembershipCheckFunc func(userUUID uuid.UUID, roomID int64) (bool, error)

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return NewManagerWithConfig(DefaultManagerConfig())
}

// NewManagerWithConfig creates a new WebSocket manager with configuration
func NewManagerWithConfig(config ManagerConfig) *Manager {
	return &Manager{
		clients:           make(map[string]*Client),
		rooms:             make(map[int64]map[string]*Client),
		userConnections:   make(map[string]map[int64]*Client),
		roomMessageQueues: make(map[int64]chan Message),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan Message, config.BroadcastQueueSize),
		cleanup:           make(chan int64, 100),
		config:            config,
		roomCleanup:       make(map[int64]*time.Timer),
	}
}

// Run starts the WebSocket manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()
			log.Printf("Client registered: %s", client.ID)

		case client := <-m.unregister:
			m.removeClientSafely(client)

		case message := <-m.broadcast:
			m.SendToRoom(message)

		case roomID := <-m.cleanup:
			m.cleanupEmptyRoom(roomID)
		}
	}
}

// removeClientSafely removes a client with proper cleanup
func (m *Manager) removeClientSafely(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[client.ID]; !ok {
		return // Client already removed
	}

	userUUID := client.UserUUID.String()

	// Remove from all rooms
	for roomID := range client.Rooms {
		if roomClients, exists := m.rooms[roomID]; exists {
			delete(roomClients, client.ID)

			// Remove from user connections tracking
			if userRooms, userExists := m.userConnections[userUUID]; userExists {
				delete(userRooms, roomID)
				if len(userRooms) == 0 {
					delete(m.userConnections, userUUID)
				}
			}

			// Schedule cleanup if room is empty
			if len(roomClients) == 0 {
				select {
				case m.cleanup <- roomID:
				default:
				}
			}
		}
	}

	// Remove client
	delete(m.clients, client.ID)
	close(client.Send)

	log.Printf("🧹 Client %s unregistered (user: %s)", client.ID, userUUID)
}

// cleanupEmptyRoom cleans up empty room resources
func (m *Manager) cleanupEmptyRoom(roomID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check room is still empty
	if roomClients, exists := m.rooms[roomID]; exists && len(roomClients) == 0 {
		delete(m.rooms, roomID)
		log.Printf("🏠 Room %d is now empty - removed", roomID)

		// Cleanup message queue
		if queue, exists := m.roomMessageQueues[roomID]; exists {
			close(queue)
			delete(m.roomMessageQueues, roomID)
			log.Printf("Cleaned up queue for empty room %d", roomID)
		}

		// Cancel cleanup timer if exists
		if timer, exists := m.roomCleanup[roomID]; exists {
			timer.Stop()
			delete(m.roomCleanup, roomID)
		}
	}
}

func (m *Manager) AddClientToRoom(roomID int64, client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userUUID := client.UserUUID.String()

	// Check if user already has connection in this room
	if m.userConnections[userUUID] == nil {
		m.userConnections[userUUID] = make(map[int64]*Client)
	}

	// Close existing connection for this user in this room
	if existingClient, exists := m.userConnections[userUUID][roomID]; exists {
		log.Printf("🔄 Replacing existing connection for user %s in room %d", userUUID, roomID)

		// Remove old client from room
		if roomClients, roomExists := m.rooms[roomID]; roomExists {
			delete(roomClients, existingClient.ID)
		}

		// Close old connection
		close(existingClient.Send)
		existingClient.Conn.Close()
		delete(m.clients, existingClient.ID)
	}

	// Add new connection
	if m.rooms[roomID] == nil {
		m.rooms[roomID] = make(map[string]*Client)
	}

	m.rooms[roomID][client.ID] = client
	m.userConnections[userUUID][roomID] = client
	client.Rooms[roomID] = true

	log.Printf("✅ Client %s joined room %d (user: %s)", client.ID, roomID, userUUID)
}

// RemoveClientFromRoom removes a client from a specific room
func (m *Manager) RemoveClientFromRoom(roomID int64, client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove room from client's rooms
	delete(client.Rooms, roomID)

	// Remove client from room
	if roomClients, ok := m.rooms[roomID]; ok {
		delete(roomClients, client.ID)

		// If room is empty, schedule cleanup
		if len(roomClients) == 0 {
			select {
			case m.cleanup <- roomID:
			default:
			}
		}
	}

	// Remove from user connections
	userUUID := client.UserUUID.String()
	if userRooms, exists := m.userConnections[userUUID]; exists {
		delete(userRooms, roomID)
		if len(userRooms) == 0 {
			delete(m.userConnections, userUUID)
		}
	}
}

// SendToRoom gửi tin nhắn có thứ tự đến phòng cụ thể
func (m *Manager) SendToRoom(message Message) {
	roomID := message.RoomID

	// Sử dụng queue riêng cho từng phòng để đảm bảo thứ tự
	m.ensureRoomQueue(roomID)

	// Gửi tin nhắn vào queue của phòng
	m.mu.RLock()
	queue, exists := m.roomMessageQueues[roomID]
	m.mu.RUnlock()

	if exists {
		select {
		case queue <- message:
			// Message added to queue successfully
		default:
			// Queue is full, process immediately (fallback)
			m.processBroadcastMessage(message)
		}
	} else {
		// No queue exists, process immediately
		m.processBroadcastMessage(message)
	}
}

// ensureRoomQueue đảm bảo có queue cho phòng
func (m *Manager) ensureRoomQueue(roomID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.roomMessageQueues[roomID]; !exists {
		m.roomMessageQueues[roomID] = make(chan Message, m.config.RoomQueueSize)
		go m.processRoomQueue(roomID)
	}
}

// processRoomQueue xử lý queue tin nhắn của một phòng
func (m *Manager) processRoomQueue(roomID int64) {
	m.mu.RLock()
	queue := m.roomMessageQueues[roomID]
	m.mu.RUnlock()

	for message := range queue {
		m.processBroadcastMessage(message)
	}
}

// processBroadcastMessage xử lý broadcast tin nhắn thực tế
func (m *Manager) processBroadcastMessage(message Message) {
	roomID := message.RoomID

	// Marshal the message once
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Capture a snapshot of clients with proper locking
	clientList := m.GetClientsInRoom(roomID)
	if len(clientList) == 0 {
		log.Printf("No clients in room %d to broadcast to", roomID)
		return
	}

	log.Printf("Broadcasting to %d clients in room %d", len(clientList), roomID)

	// Send to clients without holding lock
	var failedClients []*Client
	for _, client := range clientList {
		select {
		case client.Send <- data:
			// Message sent successfully
		default:
			// Client channel is full or closed
			failedClients = append(failedClients, client)
		}
	}

	// Clean up failed clients
	if len(failedClients) > 0 {
		m.RemoveFailedClients(failedClients)
	}
}

// GetClientsInRoom lấy snapshot của clients trong room
func (m *Manager) GetClientsInRoom(roomID int64) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, exists := m.rooms[roomID]
	if !exists {
		return []*Client{}
	}

	clientList := make([]*Client, 0, len(clients))
	for _, client := range clients {
		clientList = append(clientList, client)
	}

	return clientList
}

// RemoveFailedClients xóa clients bị lỗi
func (m *Manager) RemoveFailedClients(failedClients []*Client) {
	if len(failedClients) == 0 {
		return
	}

	for _, client := range failedClients {
		// Use unregister channel for consistent cleanup
		select {
		case m.unregister <- client:
		default:
			log.Printf("Failed to queue client %s for removal", client.ID)
		}
	}
}

// IsClientInRoom checks if client is in a specific room
func (m *Manager) IsClientInRoom(roomID int64, clientID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if roomClients, exists := m.rooms[roomID]; exists {
		_, inRoom := roomClients[clientID]
		return inRoom
	}
	return false
}

// Register adds a client to the manager
func (m *Manager) Register(client *Client) {
	m.register <- client
}

// Unregister removes a client from the manager
func (m *Manager) Unregister(client *Client) {
	m.unregister <- client
}

// JoinRoom thêm hàm để client join/leave room sau khi đã connect
func (m *Manager) JoinRoom(roomID int64, client *Client) error {
	// Check if user is member of room (call to room service via callback)
	if m.roomMembershipCallback != nil {
		isMember, err := m.roomMembershipCallback(client.UserUUID, roomID)
		if err != nil {
			return fmt.Errorf("failed to check room membership: %w", err)
		}

		if !isMember {
			return fmt.Errorf("user is not a member of room %d", roomID)
		}
	}

	m.AddClientToRoom(roomID, client)

	// Broadcast user joined event
	notification := Message{
		Type:     "user_joined",
		RoomID:   roomID,
		UserUUID: client.UserUUID,
		Data:     []byte(fmt.Sprintf(`{"user_uuid":"%s"}`, client.UserUUID.String())),
	}
	m.SendToRoom(notification)

	return nil
}

func (m *Manager) LeaveRoom(roomID int64, client *Client) {
	m.RemoveClientFromRoom(roomID, client)

	// Broadcast user left event
	notification := Message{
		Type:     "user_left",
		RoomID:   roomID,
		UserUUID: client.UserUUID,
		Data:     []byte(fmt.Sprintf(`{"user_uuid":"%s"}`, client.UserUUID.String())),
	}
	m.SendToRoom(notification)
}

func (m *Manager) SetRoomMembershipCallback(callback RoomMembershipCheckFunc) {
	m.roomMembershipCallback = callback
}

// GetRoomInfo trả về thông tin clients trong phòng một cách an toàn
func (m *Manager) GetRoomInfo(roomID int64) (map[string]*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, exists := m.rooms[roomID]
	if !exists {
		return nil, false
	}

	// Trả về bản sao để tránh thay đổi từ bên ngoài
	clientCopy := make(map[string]*Client, len(clients))
	for id, client := range clients {
		clientCopy[id] = client
	}
	return clientCopy, true
}
