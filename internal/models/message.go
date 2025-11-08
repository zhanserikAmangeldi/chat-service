package models

import (
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	ID        int64     `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"room_id"`
	Content string      `json:"content"`
	Data    interface{} `json:"data,omitempty"`
}

type Client struct {
	ID       string
	UserID   int64
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[string]bool
}
