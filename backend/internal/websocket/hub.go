package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"syncspace/backend/internal/auth"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	rooms      map[string]*Room
	roomsMu    sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

type Room struct {
	id      string
	clients map[*Client]bool
	mu      sync.RWMutex
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	roomID   string
	userID   int64
	userName string
}

type Message struct {
	Type     string `json:"type"`
	Room     string `json:"room"`
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Data     any    `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.joinRoom(client)
		case client := <-h.unregister:
			h.leaveRoom(client)
		}
	}
}

func (h *Hub) joinRoom(client *Client) {
	h.roomsMu.Lock()
	room, exists := h.rooms[client.roomID]
	if !exists {
		room = &Room{id: client.roomID, clients: make(map[*Client]bool)}
		h.rooms[client.roomID] = room
	}
	h.roomsMu.Unlock()

	room.mu.Lock()
	room.clients[client] = true
	room.mu.Unlock()

	// Notify others that user joined
	h.broadcastToRoom(client.roomID, Message{
		Type:     "user_joined",
		Room:     client.roomID,
		UserID:   client.userID,
		UserName: client.userName,
		Data:     map[string]string{"user_name": client.userName},
	})
}

func (h *Hub) leaveRoom(client *Client) {
	h.roomsMu.RLock()
	room, exists := h.rooms[client.roomID]
	h.roomsMu.RUnlock()

	if exists {
		room.mu.Lock()
		if _, ok := room.clients[client]; ok {
			delete(room.clients, client)
			close(client.send)
		}
		room.mu.Unlock()

		// Notify others that user left
		h.broadcastToRoom(client.roomID, Message{
			Type:     "user_left",
			Room:     client.roomID,
			UserID:   client.userID,
			UserName: client.userName,
			Data:     map[string]string{"user_name": client.userName},
		})
	}
}

func (h *Hub) broadcastToRoom(roomID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.roomsMu.RLock()
	room, exists := h.rooms[roomID]
	h.roomsMu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.clients {
		select {
		case client.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract token from query param
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "missing room", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		roomID:   roomID,
		userID:   claims.UserID,
		userName: claims.Name,
	}

	h.register <- client

	go client.writePump()
	go client.readPump(h)
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msg.UserID = c.userID
		msg.UserName = c.userName
		msg.Room = c.roomID

		// Broadcast to room
		h.broadcastToRoom(c.roomID, msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ParseClassroomRoom(classroomID int64) string {
	return "classroom_" + strconv.FormatInt(classroomID, 10)
}
