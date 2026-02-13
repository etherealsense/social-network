package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
)

type client struct {
	conn   *websocket.Conn
	userID int32
	chatID int32
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int32]map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int32]map[*client]struct{}),
	}
}

func (h *Hub) Register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[c.chatID] == nil {
		h.clients[c.chatID] = make(map[*client]struct{})
	}
	h.clients[c.chatID][c] = struct{}{}
}

func (h *Hub) Unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[c.chatID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, c.chatID)
		}
	}
}

func (h *Hub) Broadcast(chatID int32, msg MessageResponse) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal broadcast message", "error", err)
		return
	}

	for c := range h.clients[chatID] {
		if err := c.conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			slog.Error("failed to write to websocket", "error", err, "user_id", c.userID)
		}
	}
}
