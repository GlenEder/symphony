package conversation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/llm"
	"github.com/gleneder/symphony/internal/store"
)

type State string

const (
	Idle        State = "idle"
	Researching State = "researching"
	Grilling    State = "grilling"
	Generating  State = "generating"
	Reviewing   State = "reviewing"
	Approved    State = "approved"
	Exporting   State = "exporting"
	Failed      State = "failed"
)

var validTransitions = map[State][]State{
	Idle: {Researching}, Researching: {Grilling, Reviewing, Failed}, Grilling: {Generating},
	Generating: {Reviewing}, Reviewing: {Approved, Grilling}, Approved: {Exporting}, Exporting: {}, Failed: {Researching},
}

func CanTransition(from, to State) bool {
	for _, candidate := range validTransitions[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

type Event struct {
	PlanID  string    `json:"plan_id"`
	Seq     uint64    `json:"seq"`
	State   State     `json:"state"`
	Error   string    `json:"error,omitempty"`
	Session *Session  `json:"session,omitempty"`
	At      time.Time `json:"at"`
}
type LLM interface {
	Chat(context.Context, []llm.Message) (llm.Response, error)
}
type Researcher interface {
	Summary(string, codebase.SummaryOptions) (string, error)
}

type Session struct {
	ID            string    `json:"id"`
	Prompt        string    `json:"prompt"`
	State         State     `json:"state"`
	Context       string    `json:"context,omitempty"`
	DecisionAreas string    `json:"decision_areas,omitempty"`
	Error         string    `json:"error,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
	mu            sync.RWMutex
}

func (s *Session) Snapshot() Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Session{ID: s.ID, Prompt: s.Prompt, State: s.State, Context: s.Context, DecisionAreas: s.DecisionAreas, Error: s.Error, UpdatedAt: s.UpdatedAt}
}

type Manager struct {
	store       *store.PlanStore
	explorer    Researcher
	client      LLM
	root        string
	mu          sync.RWMutex
	sessions    map[string]*Session
	subscribers map[string]map[chan Event]struct{}
	history     map[string][]Event
	sequences   map[string]uint64
	cancels     map[string]context.CancelFunc
	OnEvent     func(Event)
}

func NewManager(s *store.PlanStore, explorer Researcher, client LLM, root string) *Manager {
	return &Manager{store: s, explorer: explorer, client: client, root: root, sessions: map[string]*Session{}, subscribers: map[string]map[chan Event]struct{}{}, history: map[string][]Event{}, sequences: map[string]uint64{}, cancels: map[string]context.CancelFunc{}}
}
func (m *Manager) Session(id string) *Session {
	m.mu.RLock()
	s := m.sessions[id]
	m.mu.RUnlock()
	if s == nil {
		return nil
	}
	x := s.Snapshot()
	return &x
}
func (m *Manager) Subscribe(id string) (<-chan Event, func()) {
	ch := make(chan Event, 16)
	m.mu.Lock()
	s := m.sessions[id]
	if s == nil {
		m.mu.Unlock()
		return nil, func() {}
	}
	if m.subscribers[id] == nil {
		m.subscribers[id] = map[chan Event]struct{}{}
	}
	m.subscribers[id][ch] = struct{}{}
	snapshot := s.Snapshot()
	seq := m.sequences[id]
	ch <- Event{PlanID: id, Seq: seq, State: snapshot.State, Error: snapshot.Error, Session: &snapshot, At: time.Now().UTC()}
	m.mu.Unlock()
	return ch, func() {
		m.mu.Lock()
		if _, ok := m.subscribers[id][ch]; ok {
			delete(m.subscribers[id], ch)
			close(ch)
		}
		m.mu.Unlock()
	}
}
func (m *Manager) emit(e Event) {
	if e.Session == nil {
		if s := m.Session(e.PlanID); s != nil {
			e.Session = s
		}
	}
	m.mu.Lock()
	m.sequences[e.PlanID]++
	e.Seq = m.sequences[e.PlanID]
	m.history[e.PlanID] = append(m.history[e.PlanID], e)
	if len(m.history[e.PlanID]) > 64 {
		m.history[e.PlanID] = m.history[e.PlanID][len(m.history[e.PlanID])-64:]
	}
	m.mu.Unlock()
	if m.OnEvent != nil {
		m.OnEvent(e)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for ch := range m.subscribers[e.PlanID] {
		select {
		case ch <- e:
		default:
		}
	}
}
func (m *Manager) transition(s *Session, next State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !CanTransition(s.State, next) {
		return fmt.Errorf("invalid transition %s -> %s", s.State, next)
	}
	s.State = next
	s.UpdatedAt = time.Now().UTC()
	return nil
}
func (m *Manager) Start(ctx context.Context, prompt string) (*Session, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, errors.New("prompt is required")
	}
	id := fmt.Sprintf("plan-%d", time.Now().UnixNano())
	s := &Session{ID: id, Prompt: prompt, State: Idle, UpdatedAt: time.Now().UTC()}
	m.mu.Lock()
	m.sessions[id] = s
	m.mu.Unlock()
	if _, err := m.store.CreatePlan(id, "New planning request", prompt); err != nil {
		return nil, err
	}
	if err := m.transition(s, Researching); err != nil {
		return nil, err
	}
	// The request context must not control a session that outlives the HTTP request.
	sessionCtx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[id] = cancel
	m.mu.Unlock()
	m.emit(Event{PlanID: id, State: Researching, At: time.Now().UTC()})
	go m.research(sessionCtx, s)
	x := s.Snapshot()
	return &x, nil
}
func (m *Manager) Retry(id string) error {
	m.mu.RLock()
	s := m.sessions[id]
	m.mu.RUnlock()
	if s == nil {
		return errors.New("session not found")
	}
	s.mu.Lock()
	if s.State != Failed {
		s.mu.Unlock()
		return errors.New("session is not failed")
	}
	s.Error = ""
	s.State = Idle
	s.mu.Unlock()
	m.mu.Lock()
	if cancel := m.cancels[id]; cancel != nil {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancels[id] = cancel
	m.mu.Unlock()
	if err := m.transition(s, Researching); err != nil {
		cancel()
		return err
	}
	m.emit(Event{PlanID: id, State: Researching, At: time.Now().UTC()})
	go m.research(ctx, s)
	return nil
}

func (m *Manager) research(ctx context.Context, s *Session) {
	defer func() { m.mu.Lock(); delete(m.cancels, s.ID); m.mu.Unlock() }()
	fail := func(err error) {
		s.mu.Lock()
		s.Error = err.Error()
		s.State = Failed
		s.UpdatedAt = time.Now().UTC()
		s.mu.Unlock()
		m.emit(Event{PlanID: s.ID, State: Failed, Error: err.Error(), At: time.Now().UTC()})
	}
	if m.explorer == nil {
		fail(errors.New("codebase explorer is unavailable"))
		return
	}
	contextText, err := m.explorer.Summary(m.root, codebase.SummaryOptions{})
	if err != nil {
		fail(err)
		return
	}
	s.mu.Lock()
	s.Context = contextText
	s.mu.Unlock()
	areas := "Research completed. Identify the key decisions needed to make this plan."
	if m.client != nil {
		resp, e := m.client.Chat(ctx, []llm.Message{{Role: "system", Content: "Identify concise decision areas from the codebase context. Do not include secrets or credentials."}, {Role: "user", Content: s.Prompt + "\n\n" + contextText}})
		if e != nil {
			fail(e)
			return
		}
		if len(resp.Choices) > 0 {
			areas = resp.Choices[0].Message.Content
		}
	}
	s.mu.Lock()
	s.DecisionAreas = areas
	s.mu.Unlock()
	if err = m.transition(s, Grilling); err != nil {
		fail(err)
		return
	}
	m.emit(Event{PlanID: s.ID, State: Grilling, At: time.Now().UTC()})
}
