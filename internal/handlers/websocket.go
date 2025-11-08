package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/hub"
	"github.com/zhanserikAmangeldi/chat-service/internal/models"
	"github.com/zhanserikAmangeldi/chat-service/internal/queue"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub       *hub.Hub
	publisher *queue.Publisher
}

func NewWebSocketHandler(h *hub.Hub, pub *queue.Publisher) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       h,
		publisher: pub,
	}
}

func (wsh *WebSocketHandler) HandleConnection(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &models.Client{
		ID:       uuid.New().String(),
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
	}

	wsh.hub.Register <- client

	go wsh.writePump(client)
	go wsh.readPump(client)
}

func (wsh *WebSocketHandler) readPump(client *models.Client) {
	defer func() {
		wsh.hub.Unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		wsh.handleMessage(client, &wsMsg)
	}
}

func (wsh *WebSocketHandler) writePump(client *models.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

func (wsh *WebSocketHandler) handleMessage(client *models.Client, wsMsg *models.WSMessage) {
	switch wsMsg.Type {
	case "join":
		wsh.hub.JoinRoom(client, wsMsg.RoomID)

		notification := models.WSMessage{
			Type:    "user_joined",
			RoomID:  wsMsg.RoomID,
			Content: client.Username + " joined the room",
		}
		wsh.hub.BroadcastToRoom(wsMsg.RoomID, notification, client.ID)

	case "leave":
		wsh.hub.LeaveRoom(client, wsMsg.RoomID)

	case "message":
		msg := &models.Message{
			RoomID:    wsMsg.RoomID,
			UserID:    client.UserID,
			Username:  client.Username,
			Content:   wsMsg.Content,
			Type:      "text",
			CreatedAt: time.Now(),
		}

		if err := wsh.publisher.PublishMessage(msg); err != nil {
			log.Printf("Error publishing message: %v", err)
			return
		}

		response := models.WSMessage{
			Type:    "message",
			RoomID:  wsMsg.RoomID,
			Content: wsMsg.Content,
			Data:    msg,
		}
		wsh.hub.BroadcastToRoom(wsMsg.RoomID, response, "")
	}
}
