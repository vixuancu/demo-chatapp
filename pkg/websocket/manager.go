package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

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

// Manager quản lý tất cả các kết nối WebSocket và các phòng
type Manager struct {
	// Clients map client ID to Client
	clients map[string]*Client

	// Rooms map room ID to all clients in that room
	Rooms map[int64]map[string]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Send message to specific room
	broadcast chan Message

	// Mutex for concurrent access to rooms and clients maps
	Mutex sync.RWMutex

	userConnections map[string]map[int64]*Client // userUUID -> roomID -> client

	// Room membership callback function
	roomMembershipCallback RoomMembershipCheckFunc

	// Message ordering channels per room
	roomMessageQueues map[int64]chan Message
	roomQueueMutex    sync.RWMutex
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		clients:           make(map[string]*Client),
		Rooms:             make(map[int64]map[string]*Client),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan Message, 100),
		Mutex:             sync.RWMutex{},
		userConnections:   make(map[string]map[int64]*Client),
		roomMessageQueues: make(map[int64]chan Message),
		roomQueueMutex:    sync.RWMutex{},
	}
}

// Run starts the WebSocket manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.Mutex.Lock()
			m.clients[client.ID] = client
			m.Mutex.Unlock()
			log.Printf("Client registered: %s", client.ID)

		case client := <-m.unregister:
			m.Mutex.Lock()
			if _, ok := m.clients[client.ID]; ok {
				// Remove from rooms
				for roomID := range client.Rooms {
					if roomClients, exists := m.Rooms[roomID]; exists {
						delete(roomClients, client.ID)
						// If room is empty, remove it and cleanup queue
						if len(roomClients) == 0 {
							delete(m.Rooms, roomID)
							m.CleanupRoom(roomID)
						}
					}
				}
				// Remove client
				delete(m.clients, client.ID)
				close(client.Send)
			}
			m.Mutex.Unlock()
			log.Printf("Client unregistered: %s", client.ID)

		case message := <-m.broadcast:
			m.SendToRoom(message)
		}
	}
}

func (m *Manager) AddClientToRoom(roomID int64, client *Client) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	userUUID := client.UserUUID.String()

	// ✅ Check if user already has connection in this room
	if m.userConnections[userUUID] == nil {
		m.userConnections[userUUID] = make(map[int64]*Client)
	}

	// ✅ Close existing connection for this user in this room
	if existingClient, exists := m.userConnections[userUUID][roomID]; exists {
		log.Printf("🔄 Replacing existing connection for user %s in room %d", userUUID, roomID)

		// Remove old client from room
		if roomClients, roomExists := m.Rooms[roomID]; roomExists {
			delete(roomClients, existingClient.ID)
		}

		// Close old connection
		close(existingClient.Send)
		existingClient.Conn.Close()
		delete(m.clients, existingClient.ID)
	}

	// Add new connection
	if m.Rooms[roomID] == nil {
		m.Rooms[roomID] = make(map[string]*Client)
	}

	m.Rooms[roomID][client.ID] = client
	m.userConnections[userUUID][roomID] = client
	client.Rooms[roomID] = true

	log.Printf("✅ Client %s joined room %d (user: %s)", client.ID, roomID, userUUID)
}

// RemoveClientFromRoom removes a client from a specific room
func (m *Manager) RemoveClientFromRoom(roomID int64, client *Client) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	// Remove room from client's rooms
	delete(client.Rooms, roomID)

	// Remove client from room
	if _, ok := m.Rooms[roomID]; ok {
		delete(m.Rooms[roomID], client.ID)
		// If room is empty, remove it
		if len(m.Rooms[roomID]) == 0 {
			delete(m.Rooms, roomID)
		}
	}
}

// SendToRoom gửi tin nhắn có thứ tự đến phòng cụ thể
func (m *Manager) SendToRoom(message Message) {
	roomID := message.RoomID

	// Sử dụng queue riêng cho từng phòng để đảm bảo thứ tự
	m.ensureRoomQueue(roomID)

	// Gửi tin nhắn vào queue của phòng
	m.roomQueueMutex.RLock()
	queue, exists := m.roomMessageQueues[roomID]
	m.roomQueueMutex.RUnlock()

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
	m.roomQueueMutex.Lock()
	defer m.roomQueueMutex.Unlock()

	if _, exists := m.roomMessageQueues[roomID]; !exists {
		m.roomMessageQueues[roomID] = make(chan Message, 50)
		go m.processRoomQueue(roomID)
	}
}

// processRoomQueue xử lý queue tin nhắn của một phòng
func (m *Manager) processRoomQueue(roomID int64) {
	m.roomQueueMutex.RLock()
	queue := m.roomMessageQueues[roomID]
	m.roomQueueMutex.RUnlock()

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

// Thêm hàm để lấy snapshot của clients trong room
func (m *Manager) GetClientsInRoom(roomID int64) []*Client {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	clients, exists := m.Rooms[roomID]
	if !exists {
		return []*Client{}
	}

	clientList := make([]*Client, 0, len(clients))
	for _, client := range clients {
		clientList = append(clientList, client)
	}

	return clientList
}

// Thêm hàm để xóa clients bị lỗi
func (m *Manager) RemoveFailedClients(failedClients []*Client) {
	if len(failedClients) == 0 {
		return
	}

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, client := range failedClients {
		// Remove from all rooms
		for roomID := range client.Rooms {
			if roomClients, exists := m.Rooms[roomID]; exists {
				delete(roomClients, client.ID)
				log.Printf("Removed client %s from room %d", client.ID, roomID)

				// If room is empty, remove it
				if len(roomClients) == 0 {
					delete(m.Rooms, roomID)
					log.Printf("Room %d is now empty - removed", roomID)
				}
			}
		}

		// Remove from clients map
		delete(m.clients, client.ID)
		close(client.Send)
		log.Printf("Removed failed client %s", client.ID)
	}
}

// IsClientInRoom checks if client is in a specific room
func (m *Manager) IsClientInRoom(roomID int64, clientID string) bool {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	if roomClients, exists := m.Rooms[roomID]; exists {
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
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if _, ok := m.clients[client.ID]; ok {
		userUUID := client.UserUUID.String()

		// Remove from all rooms
		for roomID := range client.Rooms {
			if roomClients, exists := m.Rooms[roomID]; exists {
				delete(roomClients, client.ID)

				// Remove from user connections tracking
				if userRooms, userExists := m.userConnections[userUUID]; userExists {
					delete(userRooms, roomID)
					if len(userRooms) == 0 {
						delete(m.userConnections, userUUID)
					}
				}

				// Remove empty rooms
				if len(roomClients) == 0 {
					delete(m.Rooms, roomID)
					log.Printf("🏠 Room %d is now empty - removed", roomID)
				}
			}
		}

		// Remove client
		delete(m.clients, client.ID)
		close(client.Send)

		log.Printf("🧹 Client %s unregistered (user: %s)", client.ID, userUUID)
	}
}

// Thêm hàm để client join/leave room sau khi đã connect
func (m *Manager) JoinRoom(roomID int64, client *Client) error {
	// Check if user is member of room (call to room service via callback)
	isMember, err := m.roomMembershipCallback(client.UserUUID, roomID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}

	if !isMember {
		return fmt.Errorf("user is not a member of room %d", roomID)
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

// RoomMembershipCheckFunc callback để kiểm tra quyền phòng
type RoomMembershipCheckFunc func(userUUID uuid.UUID, roomID int64) (bool, error)

func (m *Manager) SetRoomMembershipCallback(callback RoomMembershipCheckFunc) {
	m.roomMembershipCallback = callback
}

// GetRoomInfo trả về thông tin clients trong phòng một cách an toàn
func (m *Manager) GetRoomInfo(roomID int64) (map[string]*Client, bool) {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	clients, exists := m.Rooms[roomID]
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

// CleanupRoom dọn dẹp queue khi phòng trống
func (m *Manager) CleanupRoom(roomID int64) {
	m.roomQueueMutex.Lock()
	defer m.roomQueueMutex.Unlock()

	if queue, exists := m.roomMessageQueues[roomID]; exists {
		close(queue)
		delete(m.roomMessageQueues, roomID)
		log.Printf("Cleaned up queue for empty room %d", roomID)
	}
}
