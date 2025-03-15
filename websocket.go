package zinc

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room
	mutex    sync.RWMutex
}

type Room struct {
	clients map[*websocket.Conn]bool
	mutex   sync.RWMutex
}

func WebSocket() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			EnableCompression: true,
		},
		rooms: make(map[string]*Room),
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *Context) error {
	c.Response.Header().Set("Sec-WebSocket-Version", "13")
	c.Response.Header().Set("Connection", "Upgrade")
	c.Response.Header().Set("Upgrade", "websocket")

	conn, err := h.upgrader.Upgrade(c.Response, c.Request, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade error: %v", err)
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if err := conn.WriteMessage(messageType, message); err != nil {
			break
		}
	}

	return nil
}

func (h *WebSocketHandler) JoinRoom(roomID string, conn *websocket.Conn) {
	h.mutex.Lock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = &Room{
			clients: make(map[*websocket.Conn]bool),
		}
	}
	h.mutex.Unlock()

	room := h.rooms[roomID]
	room.mutex.Lock()
	room.clients[conn] = true
	room.mutex.Unlock()
}

func (h *WebSocketHandler) LeaveRoom(roomID string, conn *websocket.Conn) {
	h.mutex.RLock()
	room, exists := h.rooms[roomID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	room.mutex.Lock()
	delete(room.clients, conn)
	room.mutex.Unlock()

	// Clean up empty rooms
	if len(room.clients) == 0 {
		h.mutex.Lock()
		delete(h.rooms, roomID)
		h.mutex.Unlock()
	}
}

func (h *WebSocketHandler) BroadcastToRoom(roomID string, message []byte) error {
	h.mutex.RLock()
	room, exists := h.rooms[roomID]
	h.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for conn := range room.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			continue
		}
	}

	return nil
}

func (c *Context) Upgrade() (*websocket.Conn, error) {
	if app, ok := c.Get("app").(*App); ok {
		if app.wsHandler == nil {
			return nil, fmt.Errorf("WebSocket handler not initialized")
		}
		c.Response.Header().Set("Sec-WebSocket-Version", "13")
		c.Response.Header().Set("Connection", "Upgrade")
		c.Response.Header().Set("Upgrade", "websocket")

		return app.wsHandler.upgrader.Upgrade(c.Response, c.Request, nil)
	}
	return nil, fmt.Errorf("app context not found")
}

func (c *Context) JoinRoom(roomID string, conn *websocket.Conn) {
	if app, ok := c.Get("app").(*App); ok && app.wsHandler != nil {
		app.wsHandler.JoinRoom(roomID, conn)
	}
}

func (c *Context) LeaveRoom(roomID string, conn *websocket.Conn) {
	if app, ok := c.Get("app").(*App); ok && app.wsHandler != nil {
		app.wsHandler.LeaveRoom(roomID, conn)
	}
}

func (c *Context) BroadcastToRoom(roomID string, message []byte) error {
	if app, ok := c.Get("app").(*App); ok && app.wsHandler != nil {
		return app.wsHandler.BroadcastToRoom(roomID, message)
	}
	return fmt.Errorf("WebSocket handler not initialized")
}
