package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins
}

// Hub manages WebSocket connections per plan ID.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // planID -> set of conns
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

// Subscribe adds a connection to a plan's update channel.
func (h *Hub) Subscribe(planID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[planID] == nil {
		h.clients[planID] = make(map[*websocket.Conn]bool)
	}
	h.clients[planID][conn] = true
	log.Printf("ws: %s subscribed to %s", conn.RemoteAddr(), planID)
}

// Unsubscribe removes a connection.
func (h *Hub) Unsubscribe(planID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[planID] != nil {
		delete(h.clients[planID], conn)
		if len(h.clients[planID]) == 0 {
			delete(h.clients, planID)
		}
	}
}

// Broadcast sends a message to all connections subscribed to a plan.
func (h *Hub) Broadcast(planID string, msg []byte) {
	h.mu.RLock()
	conns := h.clients[planID]
	h.mu.RUnlock()

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("ws write error: %v", err)
			conn.Close()
			go h.Unsubscribe(planID, conn)
		}
	}
}

// handleWS handles WebSocket upgrade for a plan.
func (h *Hub) handleWS(store *PlanStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID := r.PathValue("id")
		if planID == "" {
			http.Error(w, "missing plan id", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return
		}

		h.Subscribe(planID, conn)

		// Send the current plan state immediately on connect
		if plan := store.Get(planID); plan != nil {
			fp := toFlatPlan(plan)
			if err := conn.WriteMessage(websocket.TextMessage, fp.JSON()); err != nil {
				log.Printf("ws initial write error: %v", err)
				h.Unsubscribe(planID, conn)
				conn.Close()
				return
			}
		}

		// Read loop — detect disconnect
		defer func() {
			h.Unsubscribe(planID, conn)
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
