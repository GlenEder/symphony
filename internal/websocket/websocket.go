package websocket

import (
	"github.com/gleneder/symphony/internal/model"
	"github.com/gleneder/symphony/internal/store"
	gws "github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const AgentTimeout = 10 * time.Minute
const (
	StatusListening = "listening"
	StatusThinking  = "thinking"
	StatusOffline   = "offline"
)

type AgentState struct {
	mu     sync.RWMutex
	states map[string]*status
}
type status struct {
	Status        string
	LastHeartbeat time.Time
}

func NewAgentState() *AgentState          { return &AgentState{states: map[string]*status{}} }
func (a *AgentState) Heartbeat(id string) { a.SetStatus(id, StatusListening) }
func (a *AgentState) SetStatus(id, v string) {
	a.mu.Lock()
	a.states[id] = &status{v, time.Now()}
	a.mu.Unlock()
}
func (a *AgentState) SetOffline(id string)   { a.SetStatus(id, StatusOffline) }
func (a *AgentState) SetThinking(id string)  { a.SetStatus(id, StatusThinking) }
func (a *AgentState) SetListening(id string) { a.SetStatus(id, StatusListening) }
func (a *AgentState) GetStatus(id string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if x := a.states[id]; x != nil {
		return x.Status
	}
	return StatusOffline
}
func (a *AgentState) GC() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, x := range a.states {
		if time.Since(x.LastHeartbeat) > AgentTimeout {
			x.Status = StatusOffline
		}
	}
}

type conn struct {
	c  *gws.Conn
	mu sync.Mutex
}

func (x *conn) write(b []byte) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.c.WriteMessage(gws.TextMessage, b)
}
func (x *conn) close() { x.c.Close() }

type Hub struct {
	mu             sync.RWMutex
	clients        map[string]map[*conn]bool
	expectedOrigin string
}

func (h *Hub) ExpectedOrigin() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.expectedOrigin
}

func (h *Hub) SetExpectedOrigin(origin string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.expectedOrigin = origin
}

func (h *Hub) OriginAllowed(r *http.Request) bool {
	return originAllowed(r, h.ExpectedOrigin())
}

func NewHub(expectedOrigin ...string) *Hub {
	origin := "http://localhost:8080"
	if len(expectedOrigin) > 0 && expectedOrigin[0] != "" {
		origin = expectedOrigin[0]
	}
	return &Hub{clients: map[string]map[*conn]bool{}, expectedOrigin: origin}
}
func (h *Hub) Broadcast(id string, b []byte) {
	h.mu.RLock()
	cs := make([]*conn, 0, len(h.clients[id]))
	for c := range h.clients[id] {
		cs = append(cs, c)
	}
	h.mu.RUnlock()
	for _, c := range cs {
		if e := c.write(b); e != nil {
			log.Print(e)
			c.close()
			h.Unsubscribe(id, c)
		}
	}
}
func (h *Hub) Subscribe(id string, c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[id] == nil {
		h.clients[id] = map[*conn]bool{}
	}
	h.clients[id][c] = true
}
func (h *Hub) Unsubscribe(id string, c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[id], c)
}

func originAllowed(r *http.Request, expectedOrigin string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	expected, err := url.Parse(expectedOrigin)
	if err != nil {
		return false
	}
	actual, err := url.Parse(origin)
	if err != nil || actual.Scheme != expected.Scheme || actual.Host != expected.Host {
		return false
	}
	return actual.User == nil && actual.Path == "" && actual.RawQuery == "" && actual.Fragment == ""
}

func newUpgrader(expectedOrigin string) gws.Upgrader {
	return gws.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return originAllowed(r, expectedOrigin)
	}}
}

func (h *Hub) Handler(s *store.PlanStore, a *AgentState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !store.ValidPlanID(id) || s.Get(id) == nil {
			http.NotFound(w, r)
			return
		}
		hubOrigin := h.ExpectedOrigin()
		upgrader := newUpgrader(hubOrigin)
		c, e := upgrader.Upgrade(w, r, nil)
		if e != nil {
			return
		}
		x := &conn{c: c}
		h.Subscribe(id, x)
		defer func() { h.Unsubscribe(id, x); x.close() }()
		if p := s.Get(id); p != nil {
			_ = x.write(model.ToFlatPlan(p, a.GetStatus(id)).JSON())
		}
		for {
			if _, _, e = c.ReadMessage(); e != nil {
				return
			}
		}
	}
}
