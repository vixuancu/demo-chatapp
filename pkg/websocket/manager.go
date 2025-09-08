package websocket

import (
	"encoding/json"
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
}

// Manager quản lý tất cả các kết nối WebSocket và các phòng
type Manager struct {
	// Clients map client ID to Client
	clients map[string]*Client

	// Rooms map room ID to all clients in that room
	rooms map[int64]map[string]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Send message to specific room
	broadcast chan Message

	// Mutex for concurrent access to rooms and clients maps
	mutex sync.RWMutex
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		rooms:      make(map[int64]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 100),
		mutex:      sync.RWMutex{},
	}
}

// Run starts the WebSocket manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client.ID] = client
			m.mutex.Unlock()
			log.Printf("Client registered: %s", client.ID)

		case client := <-m.unregister:
			if _, ok := m.clients[client.ID]; ok {
				m.mutex.Lock()
				// Remove from rooms
				for roomID := range client.Rooms {
					if _, ok := m.rooms[roomID]; ok {
						delete(m.rooms[roomID], client.ID)
						// If room is empty, remove it
						if len(m.rooms[roomID]) == 0 {
							delete(m.rooms, roomID)
						}
					}
				}
				// Remove client
				delete(m.clients, client.ID)
				close(client.Send)
				m.mutex.Unlock()
				log.Printf("Client unregistered: %s", client.ID)
			}

		case message := <-m.broadcast:
			roomID := message.RoomID
			m.mutex.RLock()
			if clients, ok := m.rooms[roomID]; ok {
				// Marshal the message
				data, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling message: %v", err)
					m.mutex.RUnlock()
					continue
				}

				// Send to all clients in the room
				for _, client := range clients {
					select {
					case client.Send <- data:
						// Message sent successfully
					default:
						close(client.Send)
						m.mutex.RUnlock()
						m.mutex.Lock()
						delete(clients, client.ID)
						delete(m.clients, client.ID)
						m.mutex.Unlock()
						m.mutex.RLock()
					}
				}
			}
			m.mutex.RUnlock()
		}
	}
}

// AddClientToRoom adds a client to a specific room
func (m *Manager) AddClientToRoom(roomID int64, client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Add room to client's rooms
	client.Rooms[roomID] = true

	// Add client to room
	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[string]*Client)
	}
	m.rooms[roomID][client.ID] = client
}

// RemoveClientFromRoom removes a client from a specific room
func (m *Manager) RemoveClientFromRoom(roomID int64, client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Remove room from client's rooms
	delete(client.Rooms, roomID)

	// Remove client from room
	if _, ok := m.rooms[roomID]; ok {
		delete(m.rooms[roomID], client.ID)
		// If room is empty, remove it
		if len(m.rooms[roomID]) == 0 {
			delete(m.rooms, roomID)
		}
	}
}

// SendToRoom broadcasts a message to all clients in a room
func (m *Manager) SendToRoom(message Message) {
	m.broadcast <- message
}

// Register adds a client to the manager
func (m *Manager) Register(client *Client) {
	m.register <- client
}

// Unregister removes a client from the manager
func (m *Manager) Unregister(client *Client) {
	m.unregister <- client
}
