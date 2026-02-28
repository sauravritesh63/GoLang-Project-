package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	ws "github.com/sauravritesh63/GoLang-Project-/internal/api/websocket"
)

// TestNewHub_NotNil ensures NewHub returns a non-nil Hub.
func TestNewHub_NotNil(t *testing.T) {
	hub := ws.NewHub()
	if hub == nil {
		t.Fatal("expected non-nil Hub")
	}
}

// TestBroadcast_NoClients verifies that Broadcast is safe to call when no
// WebSocket clients are connected (must not panic).
func TestBroadcast_NoClients(t *testing.T) {
	hub := ws.NewHub()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Broadcast panicked with no clients: %v", r)
		}
	}()
	hub.Broadcast(context.Background(), ws.Event{
		Type:    ws.EventWorkflowStatus,
		Payload: map[string]string{"id": "abc"},
	})
}

// TestBroadcast_CancelledContext verifies that Broadcast respects a cancelled
// context (must not panic).
func TestBroadcast_CancelledContext(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Broadcast panicked with cancelled context: %v", r)
		}
	}()
	hub.Broadcast(ctx, ws.Event{Type: ws.EventTaskStatus, Payload: nil})
}

// dialHub starts an httptest server backed by hub.ServeWS and dials it with
// the gorilla/websocket client, returning the client connection and a cleanup
// function.
func dialHub(t *testing.T, hub *ws.Hub) (*websocket.Conn, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r)
	}))
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial: %v", err)
	}
	return conn, func() {
		conn.Close()
		srv.Close()
	}
}

// TestBroadcast_ClientReceivesEvent verifies that a connected WebSocket client
// receives a broadcast event.
func TestBroadcast_ClientReceivesEvent(t *testing.T) {
	hub := ws.NewHub()
	conn, cleanup := dialHub(t, hub)
	defer cleanup()

	hub.Broadcast(context.Background(), ws.Event{
		Type:    ws.EventWorkerHeartbeat,
		Payload: map[string]string{"worker": "w-1"},
	})

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if !strings.Contains(string(msg), "worker_heartbeat") {
		t.Errorf("expected event type in message, got: %s", msg)
	}
}

// TestEventTypeConstants verifies that the exported EventType constants have
// the expected string values.
func TestEventTypeConstants(t *testing.T) {
	cases := []struct {
		et   ws.EventType
		want string
	}{
		{ws.EventTaskStatus, "task_status"},
		{ws.EventWorkflowStatus, "workflow_status"},
		{ws.EventWorkerHeartbeat, "worker_heartbeat"},
	}
	for _, tc := range cases {
		if string(tc.et) != tc.want {
			t.Errorf("EventType %q: got %q, want %q", tc.want, tc.et, tc.want)
		}
	}
}
