package zinc

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Export all necessary gorilla/websocket types
type (
	// WebSocketConn represents a WebSocket connection.
	WebSocketConn = websocket.Conn

	// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection.
	Upgrader = websocket.Upgrader

	// CloseError represents a close message.
	CloseError = websocket.CloseError
)

// The following constants are exported from github.com/gorilla/websocket for convenience

// TextMessage denotes a text data message. The text message payload is
// interpreted as UTF-8 encoded text data.
const TextMessage = websocket.TextMessage

// BinaryMessage denotes a binary data message.
const BinaryMessage = websocket.BinaryMessage

// CloseMessage denotes a close control message. The optional message
// payload contains a numeric code and text.
const CloseMessage = websocket.CloseMessage

// PingMessage denotes a ping control message. The optional message payload
// is UTF-8 encoded text.
const PingMessage = websocket.PingMessage

// PongMessage denotes a pong control message. The optional message payload
// is UTF-8 encoded text.
const PongMessage = websocket.PongMessage

// Create a package-level variable for default upgrader settings that can be customized
var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

// WebSocketHandler handles WebSocket connections and room management
type WebSocketHandler struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room
	mutex    sync.RWMutex
}

// Room represents a WebSocket room with connected clients
type Room struct {
	clients map[*websocket.Conn]bool
	mutex   sync.RWMutex
}

// Upgrader returns the underlying gorilla/websocket Upgrader
// This allows users to access and customize the upgrader if needed
func (h *WebSocketHandler) Upgrader() *websocket.Upgrader {
	return &h.upgrader
}

// GorillaPkg returns the Gorilla websocket package for direct access
// This function returns nothing but provides a way to import the Gorilla
// WebSocket package in user code through the Zinc framework
func (h *WebSocketHandler) GorillaPkg() {
	// This function intentionally does nothing.
	// It exists solely to allow users to access the websocket package:
	//
	// import (
	//   "github.com/gorilla/websocket" // Direct import
	// )
	//
	// or access through zinc:
	//
	// ws := zinc.WebSocket()
	// conn := ws.Upgrade(...)
	// conn.WriteMessage(zinc.TextMessage, []byte("hello"))
}

// WebSocket creates a new WebSocketHandler with default configuration
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

// WithCustomUpgrader creates a new WebSocketHandler with a custom upgrader
func WithCustomUpgrader(upgrader websocket.Upgrader) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: upgrader,
		rooms:    make(map[string]*Room),
	}
}

// HandleWebSocket handles a WebSocket connection
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

// Upgrade upgrades an HTTP connection to a WebSocket connection
// This is a convenience method to directly use the handler's upgrader
func (h *WebSocketHandler) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	return h.upgrader.Upgrade(w, r, responseHeader)
}

// JoinRoom adds a connection to a room
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

// LeaveRoom removes a connection from a room
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

// BroadcastToRoom sends a message to all connections in a room
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

// BroadcastToRoomWithType sends a message with a specific message type to all connections in a room
func (h *WebSocketHandler) BroadcastToRoomWithType(roomID string, messageType int, message []byte) error {
	h.mutex.RLock()
	room, exists := h.rooms[roomID]
	h.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for conn := range room.clients {
		if err := conn.WriteMessage(messageType, message); err != nil {
			continue
		}
	}

	return nil
}

// GetRoomConnections returns all connections in a room
func (h *WebSocketHandler) GetRoomConnections(roomID string) ([]*websocket.Conn, error) {
	h.mutex.RLock()
	room, exists := h.rooms[roomID]
	h.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	connections := make([]*websocket.Conn, 0, len(room.clients))
	for conn := range room.clients {
		connections = append(connections, conn)
	}

	return connections, nil
}

// RoomExists checks if a room exists
func (h *WebSocketHandler) RoomExists(roomID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, exists := h.rooms[roomID]
	return exists
}

// RoomSize returns the number of connections in a room
func (h *WebSocketHandler) RoomSize(roomID string) (int, error) {
	h.mutex.RLock()
	room, exists := h.rooms[roomID]
	h.mutex.RUnlock()

	if !exists {
		return 0, fmt.Errorf("room %s not found", roomID)
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()
	return len(room.clients), nil
}

// Upgrade upgrades an HTTP connection to a WebSocket connection using the application's WebSocket handler
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

// JoinRoom adds a connection to a room
func (c *Context) JoinRoom(roomID string, conn *websocket.Conn) {
	if app, ok := c.Get("app").(*App); ok && app.wsHandler != nil {
		app.wsHandler.JoinRoom(roomID, conn)
	}
}

// LeaveRoom removes a connection from a room
func (c *Context) LeaveRoom(roomID string, conn *websocket.Conn) {
	if app, ok := c.Get("app").(*App); ok && app.wsHandler != nil {
		app.wsHandler.LeaveRoom(roomID, conn)
	}
}

// BroadcastToRoom sends a message to all connections in a room
func (c *Context) BroadcastToRoom(roomID string, message []byte) error {
	if app, ok := c.Get("app").(*App); ok && app.wsHandler != nil {
		return app.wsHandler.BroadcastToRoom(roomID, message)
	}
	return fmt.Errorf("WebSocket handler not initialized")
}
