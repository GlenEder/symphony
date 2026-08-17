package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gleneder/symphony/internal/model"
)

type PlanSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
type MsgEntry struct {
	Text    string
	ItemRef string
	Prompt  *model.Prompt
}
type PlanStore struct {
	mu        sync.RWMutex
	plans     map[string]*model.Plan
	OnChange  func(string)
	Dir       string
	lastWrite map[string]time.Time
	persistMu sync.Mutex
}

var validID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func ValidPlanID(id string) bool { return id != "" && validID.MatchString(id) }
func validateID(id string) error {
	if !ValidPlanID(id) {
		return fmt.Errorf("invalid plan id: %q", id)
	}
	return nil
}
func New(dir string, onChange func(string)) *PlanStore {
	return &PlanStore{plans: map[string]*model.Plan{}, OnChange: onChange, Dir: dir, lastWrite: map[string]time.Time{}}
}

func (s *PlanStore) LoadAll() error {
	es, e := os.ReadDir(s.Dir)
	if e != nil {
		if os.IsNotExist(e) {
			return os.MkdirAll(s.Dir, 0755)
		}
		return e
	}
	for _, x := range es {
		if !x.IsDir() && strings.HasSuffix(x.Name(), ".json") {
			if err := s.load(filepath.Join(s.Dir, x.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *PlanStore) load(path string) error {
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	if filepath.Base(path) != id+".json" {
		return fmt.Errorf("invalid plan path: %s", path)
	}
	if err := validateID(id); err != nil {
		return err
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	var p model.Plan
	if e = json.Unmarshal(b, &p); e != nil {
		return fmt.Errorf("reload %s: %w", id, e)
	}
	if p.UpdatedAt == "" {
		if len(p.Messages) > 0 {
			p.UpdatedAt = p.Messages[len(p.Messages)-1].CreatedAt
		} else if i, err := os.Stat(path); err == nil {
			p.UpdatedAt = i.ModTime().UTC().Format(time.RFC3339)
		}
	}
	s.mu.Lock()
	s.plans[id] = &p
	s.mu.Unlock()
	return nil
}
func (s *PlanStore) persist(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	s.mu.Lock()
	p, ok := s.plans[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	p.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	b, e := json.MarshalIndent(p, "", "  ")
	s.mu.Unlock()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(s.Dir, 0755); e != nil {
		return e
	}
	path := filepath.Join(s.Dir, id+".json")
	f, e := os.CreateTemp(s.Dir, "."+id+"-*.json")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, e = f.Write(b); e == nil {
		e = f.Sync()
	}
	if closeErr := f.Close(); e == nil {
		e = closeErr
	}
	if e != nil {
		return e
	}
	if e = os.Chmod(tmp, 0644); e != nil {
		return e
	}
	if e = os.Rename(tmp, path); e != nil {
		return e
	}
	if i, err := os.Stat(path); err == nil {
		s.mu.Lock()
		s.lastWrite[id] = i.ModTime()
		s.mu.Unlock()
	}
	return nil
}
func (s *PlanStore) changed(id string) {
	if s.OnChange != nil {
		s.OnChange(id)
	}
}
func clone(p *model.Plan) *model.Plan {
	if p == nil {
		return nil
	}
	b, _ := json.Marshal(p)
	var c model.Plan
	_ = json.Unmarshal(b, &c)
	return &c
}
func (s *PlanStore) Get(id string) *model.Plan {
	if validateID(id) != nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.plans[id])
}
func (s *PlanStore) List() []PlanSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]PlanSummary, 0, len(s.plans))
	for id, p := range s.plans {
		o = append(o, PlanSummary{id, p.Title, p.Summary, p.UpdatedAt})
	}
	sort.Slice(o, func(i, j int) bool { return o[i].UpdatedAt > o[j].UpdatedAt })
	return o
}
func (s *PlanStore) CreatePlan(id, title, summary string) (*model.Plan, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	if _, ok := s.plans[id]; ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("plan already exists: %s", id)
	}
	s.plans[id] = &model.Plan{Title: title, Summary: summary, State: "draft"}
	s.mu.Unlock()
	if e := s.persist(id); e != nil {
		return nil, e
	}
	s.changed(id)
	return s.Get(id), nil
}
func (s *PlanStore) Upsert(id string, p *model.Plan) error {
	if err := validateID(id); err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("plan is required")
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	s.plans[id] = clone(p)
	s.mu.Unlock()
	if e := s.persist(id); e != nil {
		return e
	}
	s.changed(id)
	return nil
}
func (s *PlanStore) Patch(id string, p *model.Plan) error {
	if err := validateID(id); err != nil {
		return err
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	x := s.plans[id]
	if x == nil {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	if p.Title != "" {
		x.Title = p.Title
	}
	if p.Summary != "" {
		x.Summary = p.Summary
	}
	if p.State != "" {
		x.State = p.State
	}
	if p.Modules != nil {
		x.Modules = clone(&model.Plan{Modules: p.Modules}).Modules
	}
	s.mu.Unlock()
	if e := s.persist(id); e != nil {
		return e
	}
	s.changed(id)
	return nil
}
func (s *PlanStore) SetState(id, state string) error {
	if state != "draft" && state != "approved" {
		return fmt.Errorf("invalid state: %s", state)
	}
	return s.Patch(id, &model.Plan{State: state})
}
func (s *PlanStore) AddMessage(id, role, text, ref string, prompt *model.Prompt) (*model.Message, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	p := s.plans[id]
	if p == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("plan not found: %s", id)
	}
	m := model.Message{ID: model.NewMessageID(), Role: role, Text: text, ItemRef: ref, Prompt: clonePrompt(prompt), CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	p.Messages = append(p.Messages, m)
	s.mu.Unlock()
	if e := s.persist(id); e != nil {
		return nil, e
	}
	s.changed(id)
	return &m, nil
}
func clonePrompt(p *model.Prompt) *model.Prompt {
	if p == nil {
		return nil
	}
	x := *p
	x.Options = append([]string(nil), p.Options...)
	return &x
}
func (s *PlanStore) AddMessages(id, role string, es []MsgEntry) ([]model.Message, error) {
	var out []model.Message
	for _, e := range es {
		if e.Text != "" {
			m, err := s.AddMessage(id, role, e.Text, e.ItemRef, e.Prompt)
			if err != nil {
				return nil, err
			}
			out = append(out, *m)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no non-empty messages to add")
	}
	return out, nil
}
func (s *PlanStore) DeleteMessage(id, mid string) error {
	if err := validateID(id); err != nil {
		return err
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	p := s.plans[id]
	if p == nil {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	for i, m := range p.Messages {
		if m.ID == mid {
			p.Messages = append(p.Messages[:i], p.Messages[i+1:]...)
			s.mu.Unlock()
			if e := s.persist(id); e != nil {
				return e
			}
			s.changed(id)
			return nil
		}
	}
	s.mu.Unlock()
	return fmt.Errorf("message not found: %s", mid)
}
func (s *PlanStore) DeletePlan(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	if _, ok := s.plans[id]; !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	delete(s.plans, id)
	delete(s.lastWrite, id)
	s.mu.Unlock()
	if e := os.Remove(filepath.Join(s.Dir, id+".json")); e != nil && !os.IsNotExist(e) {
		return e
	}
	s.changed(id)
	return nil
}
func (s *PlanStore) Remove(id string) {
	if validateID(id) != nil {
		return
	}
	s.mu.Lock()
	delete(s.plans, id)
	s.mu.Unlock()
}
func (s *PlanStore) IsSelfWrite(name string, t time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	x, ok := s.lastWrite[strings.TrimSuffix(name, ".json")]
	return ok && x.Equal(t)
}
func (s *PlanStore) Reload(path string) error { return s.load(path) }
