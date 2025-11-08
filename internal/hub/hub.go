package hub

import (
	"encoding/json"
	"github.com/zhanserikAmangeldi/chat-service/internal/models"
	"log"
	"sync"
)

type Hub struct {
	clients    map[string]*models.Client
	rooms      map[string]map[string]*models.Client
	Register   chan *models.Client
	Unregister chan *models.Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Exclude string // Client ID to exclude
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*models.Client),
		rooms:      make(map[string]map[string]*models.Client),
		Register:   make(chan *models.Client),
		Unregister: make(chan *models.Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (User: %s)", client.ID, client.Username)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)

				// Remove from all rooms
				for roomID := range client.Rooms {
					if room, exists := h.rooms[roomID]; exists {
						delete(room, client.ID)
						if len(room) == 0 {
							delete(h.rooms, roomID)
						}
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.ID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[msg.RoomID]
			h.mu.RUnlock()

			for clientID, client := range room {
				if clientID != msg.Exclude {
					select {
					case client.Send <- msg.Message:
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				}
			}
		}
	}
}

func (h *Hub) JoinRoom(client *models.Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[string]*models.Client)
	}
	h.rooms[roomID][client.ID] = client
	client.Rooms[roomID] = true

	log.Printf("Client %s joined room %s", client.Username, roomID)
}

func (h *Hub) LeaveRoom(client *models.Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.rooms[roomID]; exists {
		delete(room, client.ID)
		delete(client.Rooms, roomID)

		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, message interface{}, excludeClientID string) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: data,
		Exclude: excludeClientID,
	}
}

func (h *Hub) GetRoomClients(roomID string) []*models.Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room := h.rooms[roomID]
	clients := make([]*models.Client, 0, len(room))
	for _, client := range room {
		clients = append(clients, client)
	}
	return clients
}
