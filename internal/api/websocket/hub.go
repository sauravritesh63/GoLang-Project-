// Package websocket provides a WebSocket hub for broadcasting real-time
// scheduler events (task status, workflow status, worker heartbeat) to all
// connected clients.
package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// EventType labels the kind of real-time event being broadcast.
type EventType string

const (
	// EventTaskStatus is emitted when a task run changes state.
	EventTaskStatus EventType = "task_status"
	// EventWorkflowStatus is emitted when a workflow run changes state.
	EventWorkflowStatus EventType = "workflow_status"
	// EventWorkerHeartbeat is emitted when a worker sends a heartbeat.
	EventWorkerHeartbeat EventType = "worker_heartbeat"
)

// Event is the JSON envelope sent to every connected WebSocket client.
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

var upgrader = websocket.Upgrader{
	// Allow all origins in this development implementation.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub maintains the set of active WebSocket connections and broadcasts events
// to all of them.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

// ServeWS upgrades an HTTP connection to WebSocket, registers the client, and
// blocks until the connection is closed.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.register(conn)
	defer h.unregister(conn)

	// Keep the connection alive; drain incoming messages (we only push, never pull).
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// Broadcast sends event to every currently connected client. Clients that
// have disconnected are silently removed.
func (h *Hub) Broadcast(ctx context.Context, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
				h.unregister(c)
			}
		}
	}
}

func (h *Hub) register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	_ = c.Close()
}
