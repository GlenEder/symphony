package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const AgentTimeout = 10 * time.Minute

const (
	StatusListening = "listening"
	StatusThinking  = "thinking"
	StatusOffline   = "offline"
)

// AgentStatus tracks the state of an AI agent per plan.
type AgentStatus struct {
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"-"`
}

// AgentState manages agent statuses across plans.
type AgentState struct {
	mu     sync.RWMutex
	states map[string]*AgentStatus
}

func NewAgentState() *AgentState {
	return &AgentState{
		states: make(map[string]*AgentStatus),
	}
}

// Heartbeat updates the last heartbeat time for a plan's agent.
func (as *AgentState) Heartbeat(planID string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	st, ok := as.states[planID]
	if !ok {
		st = &AgentStatus{Status: StatusListening}
		as.states[planID] = st
	}
	st.LastHeartbeat = time.Now()
	st.Status = StatusListening
}

// SetStatus sets the agent status for a plan.
// Only "offline" is meaningful from external callers;
// "thinking" and "listening" are set internally by the app.
func (as *AgentState) SetStatus(planID, status string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	st, ok := as.states[planID]
	if !ok {
		st = &AgentStatus{Status: status}
		as.states[planID] = st
	}
	st.Status = status
	st.LastHeartbeat = time.Now()
}

// SetOffline sets the agent status to offline for a plan.
func (as *AgentState) SetOffline(planID string) {
	as.SetStatus(planID, StatusOffline)
}

// SetThinking sets the agent status to thinking.
func (as *AgentState) SetThinking(planID string) {
	as.SetStatus(planID, StatusThinking)
}

// SetListening sets the agent status to listening.
func (as *AgentState) SetListening(planID string) {
	as.SetStatus(planID, StatusListening)
}

// GetStatus returns the current agent status for a plan.
func (as *AgentState) GetStatus(planID string) string {
	as.mu.RLock()
	defer as.mu.RUnlock()
	st, ok := as.states[planID]
	if !ok {
		return StatusOffline
	}
	return st.Status
}

// GC transitions stale entries (no heartbeat within AgentTimeout) to offline.
func (as *AgentState) GC() {
	as.mu.Lock()
	defer as.mu.Unlock()
	cutoff := time.Now().Add(-AgentTimeout)
	for _, st := range as.states {
		if st.LastHeartbeat.Before(cutoff) {
			st.Status = StatusOffline
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins
}

// wsConn wraps a WebSocket connection with a write mutex.
// gorilla/websocket does not support concurrent writes on the same connection,
// so all writes must be serialized per connection.
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func newWSConn(conn *websocket.Conn) *wsConn {
	return &wsConn{conn: conn}
}

// Write sends a text message, serialized with the per-connection mutex.
func (wc *wsConn) Write(msg []byte) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	return wc.conn.WriteMessage(websocket.TextMessage, msg)
}

// Close closes the underlying WebSocket connection.
func (wc *wsConn) Close() {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.conn.Close()
}

// Hub manages WebSocket connections per plan ID.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*wsConn]bool // planID -> set of conns
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*wsConn]bool),
	}
}

// Subscribe adds a connection to a plan's update channel.
func (h *Hub) Subscribe(planID string, wc *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[planID] == nil {
		h.clients[planID] = make(map[*wsConn]bool)
	}
	h.clients[planID][wc] = true
}

// Unsubscribe removes a connection.
func (h *Hub) Unsubscribe(planID string, wc *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[planID] != nil {
		delete(h.clients[planID], wc)
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

	for wc := range conns {
		if err := wc.Write(msg); err != nil {
			log.Printf("ws write error: %v", err)
			wc.Close()
			go h.Unsubscribe(planID, wc)
		}
	}
}

// handleWS handles WebSocket upgrade for a plan.
func (h *Hub) handleWS(store *PlanStore, agentState *AgentState) http.HandlerFunc {
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
		wc := newWSConn(conn)

		h.Subscribe(planID, wc)

		// Send the current plan state immediately on connect
		if plan := store.Get(planID); plan != nil {
			fp := toFlatPlan(plan, agentState.GetStatus(planID))
			if err := wc.Write(fp.JSON()); err != nil {
				log.Printf("ws initial write error: %v", err)
				h.Unsubscribe(planID, wc)
				wc.Close()
				return
			}
		}

		// Read loop — detect disconnect
		defer func() {
			h.Unsubscribe(planID, wc)
			wc.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
