package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/llm"
	"github.com/gleneder/symphony/internal/model"
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
	Idle: {Researching}, Researching: {Grilling, Reviewing, Failed}, Grilling: {Generating, Failed},
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
	ID             string        `json:"id"`
	Prompt         string        `json:"prompt"`
	State          State         `json:"state"`
	Context        string        `json:"context,omitempty"`
	DecisionAreas  string        `json:"decision_areas,omitempty"`
	Error          string        `json:"error,omitempty"`
	Question       *model.Prompt `json:"question,omitempty"`
	UpdatedAt      time.Time     `json:"updated_at"`
	QuestionCount  int           `json:"question_count"`
	QuestionKeys   []string      `json:"question_keys,omitempty"`
	TotalQuestions int           `json:"total_questions,omitempty"`
	mu             sync.RWMutex
}

func (s *Session) Snapshot() Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var question *model.Prompt
	if s.Question != nil {
		x := *s.Question
		x.Options = append([]string(nil), s.Question.Options...)
		question = &x
	}
	return Session{ID: s.ID, Prompt: s.Prompt, State: s.State, Context: s.Context, DecisionAreas: s.DecisionAreas, Error: s.Error, Question: question, UpdatedAt: s.UpdatedAt, QuestionCount: s.QuestionCount, QuestionKeys: append([]string(nil), s.QuestionKeys...), TotalQuestions: s.TotalQuestions}
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
	m := &Manager{store: s, explorer: explorer, client: client, root: root, sessions: map[string]*Session{}, subscribers: map[string]map[chan Event]struct{}{}, history: map[string][]Event{}, sequences: map[string]uint64{}, cancels: map[string]context.CancelFunc{}}
	for _, summary := range s.List() {
		if p := s.Get(summary.ID); p != nil && p.Session != nil {
			x := p.Session
			s := &Session{ID: x.ID, Prompt: x.Prompt, State: State(x.State), Context: x.Context, DecisionAreas: x.DecisionAreas, Error: x.Error, Question: x.Question, UpdatedAt: x.UpdatedAt, QuestionCount: x.QuestionCount, QuestionKeys: append([]string(nil), x.QuestionKeys...), TotalQuestions: x.TotalQuestions}
			if s.UpdatedAt.IsZero() {
				s.UpdatedAt = time.Now().UTC()
			}
			if s.QuestionCount == 0 {
				for _, msg := range p.Messages {
					if msg.Prompt != nil {
						s.QuestionCount++
						s.QuestionKeys = append(s.QuestionKeys, msg.Prompt.QuestionKey)
					}
				}
			}
			m.sessions[summary.ID] = s
			if s.State == Researching {
				go m.research(context.Background(), s)
			}
			if s.State == Grilling && s.Question != nil && s.Question.Answered {
				go m.nextQuestion(context.Background(), s, s.Question.Answer)
			}
		}
	}
	return m
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
	if !CanTransition(s.State, next) {
		s.mu.Unlock()
		return fmt.Errorf("invalid transition %s -> %s", s.State, next)
	}
	s.State = next
	s.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()
	return m.persistSession(s)
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
	if p := m.store.Get(id); p != nil {
		p.Session = sessionState(s)
		_ = m.store.Upsert(id, p)
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
	s.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()
	if err := m.persistSession(s); err != nil {
		return err
	}
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

type grillResponse struct {
	Done           bool     `json:"done"`
	QuestionKey    string   `json:"question_key"`
	Question       string   `json:"question"`
	Options        []string `json:"options"`
	AllowCustom    bool     `json:"allow_custom"`
	TotalQuestions int      `json:"total_questions"`
}

const maxQuestions = 12

func (m *Manager) persistSession(s *Session) error {
	p := m.store.Get(s.ID)
	if p == nil {
		return errors.New("plan not found")
	}
	p.Session = sessionState(s)
	return m.store.Upsert(s.ID, p)
}

const maxAnswerLength = 4000

func parseGrill(content string) (grillResponse, error) {
	var out grillResponse
	dec := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return out, fmt.Errorf("invalid grilling response: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return out, errors.New("invalid grilling response: trailing data")
	}
	if out.Done {
		if out.QuestionKey != "" || out.Question != "" || len(out.Options) != 0 || out.AllowCustom || out.TotalQuestions != 0 {
			return out, errors.New("invalid grilling completion")
		}
		return out, nil
	}
	if strings.TrimSpace(out.QuestionKey) == "" || strings.TrimSpace(out.Question) == "" || len(out.Options) == 0 || len(out.Options) > 8 || out.TotalQuestions < 1 || out.TotalQuestions > maxQuestions {
		return out, errors.New("invalid grilling question")
	}
	for i := range out.Options {
		out.Options[i] = strings.TrimSpace(out.Options[i])
		if out.Options[i] == "" || len(out.Options[i]) > 500 {
			return out, errors.New("invalid grilling option")
		}
	}
	return out, nil
}

func (m *Manager) Answer(id, answer string) error {
	answer = strings.TrimSpace(answer)
	if answer == "" || len(answer) > maxAnswerLength {
		return errors.New("answer must be between 1 and 4000 characters")
	}
	m.mu.RLock()
	s := m.sessions[id]
	m.mu.RUnlock()
	if s == nil {
		return errors.New("session not found")
	}
	s.mu.Lock()
	q := s.Question
	if s.State != Grilling || q == nil || q.Answered {
		s.mu.Unlock()
		return errors.New("no unanswered question")
	}
	valid := q.AllowCustom
	for _, option := range q.Options {
		if answer == option {
			valid = true
		}
	}
	if !valid {
		s.mu.Unlock()
		return errors.New("answer must match an option")
	}
	q.Answered = true
	q.Answer = answer
	s.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()
	p := m.store.Get(id)
	if p != nil {
		for i := range p.Messages {
			if p.Messages[i].Prompt != nil && p.Messages[i].Prompt.QuestionKey == q.QuestionKey && !p.Messages[i].Prompt.Answered {
				p.Messages[i].Prompt.Answered = true
				p.Messages[i].Prompt.Answer = answer
				p.Session = sessionState(s)
				_ = m.store.Upsert(id, p)
				break
			}
		}
	}
	m.emit(Event{PlanID: id, State: Grilling, At: time.Now().UTC()})
	m.mu.RLock()
	ctx := context.Background()
	m.mu.RUnlock()
	go m.nextQuestion(ctx, s, answer)
	return nil
}

func (m *Manager) nextQuestion(ctx context.Context, s *Session, answer string) {
	if p := m.store.Get(s.ID); p != nil {
		count := 0
		for _, msg := range p.Messages {
			if msg.Prompt != nil {
				count++
			}
		}
		if s.Snapshot().QuestionCount >= maxQuestions {
			m.completeGrilling(s)
			return
		}
	}
	if m.client == nil {
		m.failGrill(s, errors.New("llm client is unavailable"))
		return
	}
	q := s.Snapshot().Question
	history := fmt.Sprintf("Request:\n%s\nResearch:\n%s\nDecision areas:\n%s\nAnswered question %s: %s", s.Prompt, s.Context, s.DecisionAreas, q.QuestionKey, answer)
	resp, err := m.client.Chat(ctx, []llm.Message{{Role: "system", Content: llm.GrillingSystemPrompt + " Respond ONLY with JSON: {done:boolean,question_key:string,question:string,options:string[],allow_custom:boolean,total_questions:number}. Set done true when no important decisions remain."}, {Role: "user", Content: history}})
	if err != nil {
		m.failGrill(s, err)
		return
	}
	if len(resp.Choices) == 0 {
		m.failGrill(s, errors.New("llm returned no grilling response"))
		return
	}
	out, err := parseGrill(resp.Choices[0].Message.Content)
	if err != nil {
		m.failGrill(s, err)
		return
	}
	if out.Done {
		m.completeGrilling(s)
		return
	}
	if err := m.validateQuestion(s, out); err != nil {
		m.failGrill(s, err)
		return
	}
	m.postQuestion(s, out)
}

func (m *Manager) completeGrilling(s *Session) {
	if err := m.transition(s, Generating); err != nil {
		m.failGrill(s, err)
		return
	}
	s.mu.Lock()
	s.Question = nil
	s.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()
	if p := m.store.Get(s.ID); p != nil {
		p.Session = sessionState(s)
		_ = m.store.Upsert(s.ID, p)
	}
	m.emit(Event{PlanID: s.ID, State: Generating, At: time.Now().UTC()})
}

func (m *Manager) validateQuestion(s *Session, q grillResponse) error {
	x := s.Snapshot()
	if x.QuestionCount >= maxQuestions {
		return errors.New("maximum question count exceeded")
	}
	if q.TotalQuestions < x.QuestionCount+1 || (x.TotalQuestions > 0 && q.TotalQuestions != x.TotalQuestions) {
		return errors.New("invalid grilling progress")
	}
	for _, key := range x.QuestionKeys {
		if key == q.QuestionKey {
			return errors.New("duplicate question key")
		}
	}
	return nil
}

func (m *Manager) postQuestion(s *Session, q grillResponse) {
	p := m.store.Get(s.ID)
	current := 1
	if p != nil {
		for _, msg := range p.Messages {
			if msg.Prompt != nil {
				current++
			}
		}
	}
	prompt := &model.Prompt{QuestionKey: q.QuestionKey, Question: q.Question, Options: q.Options, AllowCustom: q.AllowCustom, TotalQuestions: q.TotalQuestions, Current: current}
	s.mu.Lock()
	s.Question = prompt
	s.QuestionCount++
	s.QuestionKeys = append(s.QuestionKeys, q.QuestionKey)
	s.TotalQuestions = q.TotalQuestions
	s.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()
	if p := m.store.Get(s.ID); p != nil {
		p.Session = sessionState(s)
		_, _ = m.store.AddMessage(s.ID, "agent", q.Question, q.QuestionKey, prompt)
		p = m.store.Get(s.ID)
		p.Session = sessionState(s)
		_ = m.store.Upsert(s.ID, p)
	}
	m.emit(Event{PlanID: s.ID, State: Grilling, At: time.Now().UTC()})
}

func sessionState(s *Session) *model.SessionState {
	x := s.Snapshot()
	return &model.SessionState{ID: x.ID, Prompt: x.Prompt, State: string(x.State), Context: x.Context, DecisionAreas: x.DecisionAreas, Error: x.Error, Question: x.Question, UpdatedAt: x.UpdatedAt, QuestionCount: x.QuestionCount, QuestionKeys: x.QuestionKeys, TotalQuestions: x.TotalQuestions}
}

func (m *Manager) failGrill(s *Session, err error) {
	s.mu.Lock()
	s.Error = err.Error()
	s.State = Failed
	s.UpdatedAt = time.Now().UTC()
	s.mu.Unlock()
	if p := m.store.Get(s.ID); p != nil {
		p.Session = sessionState(s)
		_ = m.store.Upsert(s.ID, p)
	}
	m.emit(Event{PlanID: s.ID, State: Failed, Error: err.Error(), At: time.Now().UTC()})
}

func (m *Manager) research(ctx context.Context, s *Session) {
	defer func() { m.mu.Lock(); delete(m.cancels, s.ID); m.mu.Unlock() }()
	fail := func(err error) {
		s.mu.Lock()
		s.Error = err.Error()
		s.State = Failed
		s.UpdatedAt = time.Now().UTC()
		s.mu.Unlock()
		_ = m.persistSession(s)
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
		resp, e := m.client.Chat(ctx, []llm.Message{{Role: "system", Content: llm.GrillingSystemPrompt + " Respond ONLY with JSON: {done:boolean,question_key:string,question:string,options:string[],allow_custom:boolean,total_questions:number}."}, {Role: "user", Content: s.Prompt + "\n\n" + contextText}})
		if e != nil {
			fail(e)
			return
		}
		if len(resp.Choices) == 0 {
			fail(errors.New("llm returned no grilling response"))
			return
		}
		out, e := parseGrill(resp.Choices[0].Message.Content)
		if e != nil {
			fail(e)
			return
		}
		if out.Done {
			m.completeGrilling(s)
			return
		}
		if e = m.validateQuestion(s, out); e != nil {
			fail(e)
			return
		}
		areas = out.Question
		s.mu.Lock()
		s.DecisionAreas = areas
		s.mu.Unlock()
		if e = m.transition(s, Grilling); e != nil {
			fail(e)
			return
		}
		m.postQuestion(s, out)
		return
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
