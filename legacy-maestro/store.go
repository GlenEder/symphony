package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// PlanStore holds loaded plans and notifies on changes.
type PlanStore struct {
	mu        sync.RWMutex
	plans     map[string]*Plan
	onChange  func(id string)
	dir       string
	lastWrite map[string]time.Time
}

func NewPlanStore(dir string, onChange func(id string)) *PlanStore {
	return &PlanStore{
		plans:     make(map[string]*Plan),
		onChange:  onChange,
		dir:       dir,
		lastWrite: make(map[string]time.Time),
	}
}

// LoadAll reads all .json files from the plans directory.
func (s *PlanStore) LoadAll() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(s.dir, 0755)
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		s.loadFile(filepath.Join(s.dir, e.Name()))
	}
	return nil
}

func decodePlan(data []byte) (*Plan, error) {
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return &plan, nil
}

// loadFile reads a single .json file and stores it by its basename (without ext).
func (s *PlanStore) loadFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("read plan %s: %v", path, err)
		return
	}
	plan, err := decodePlan(data)
	if err != nil {
		log.Printf("parse plan %s: %v", path, err)
		return
	}
	// Derive UpdatedAt for legacy plans without the field
	if plan.UpdatedAt == "" {
		// Use the most recent message's CreatedAt if available
		if len(plan.Messages) > 0 {
			plan.UpdatedAt = plan.Messages[len(plan.Messages)-1].CreatedAt
		} else {
			// Fall back to file modification time
			if fi, err := os.Stat(path); err == nil {
				plan.UpdatedAt = fi.ModTime().UTC().Format(time.RFC3339)
			}
		}
	}
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	s.mu.Lock()
	s.plans[id] = plan
	s.mu.Unlock()
	log.Printf("loaded plan: %s (%s)", id, plan.Title)
}

// persistPlan writes the in-memory plan to disk as a .json file
// and records the file's mtime so the poller can skip the self-triggered reload.
func (s *PlanStore) persistPlan(id string) error {
	plan, ok := s.plans[id]
	if !ok {
		return fmt.Errorf("plan not found: %s", id)
	}
	plan.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	js, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	path := filepath.Join(s.dir, id+".json")
	if err := os.WriteFile(path, js, 0644); err != nil {
		return err
	}
	// Record the mtime to prevent the poller from re-reading our own write
	if fi, err := os.Stat(path); err == nil {
		s.mu.Lock()
		s.lastWrite[id] = fi.ModTime()
		s.mu.Unlock()
	}
	return nil
}

// AddMessage appends a message to the plan's conversation thread and persists.
func (s *PlanStore) AddMessage(id, role, text, itemRef string, prompt *Prompt) (*Message, error) {
	s.mu.Lock()

	plan, ok := s.plans[id]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("plan not found: %s", id)
	}

	msg := Message{
		ID:        newMsgID(),
		Role:      role,
		Text:      text,
		ItemRef:   itemRef,
		Prompt:    prompt,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	plan.Messages = append(plan.Messages, msg)
	s.mu.Unlock()

	if err := s.persistPlan(id); err != nil {
		return nil, err
	}
	if s.onChange != nil {
		s.onChange(id)
	}
	return &msg, nil
}

// MsgEntry holds a single entry for batch message addition.
type MsgEntry struct {
	Text    string
	ItemRef string
	Prompt  *Prompt
}

// AddMessages appends multiple messages to the plan's conversation thread atomically,
// persists once, and triggers a single broadcast.
func (s *PlanStore) AddMessages(id, role string, entries []MsgEntry) ([]Message, error) {
	s.mu.Lock()

	plan, ok := s.plans[id]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("plan not found: %s", id)
	}

	msgs := make([]Message, 0, len(entries))
	now := time.Now().UTC().Format(time.RFC3339)
	for _, e := range entries {
		if e.Text == "" {
			continue
		}
		msg := Message{
			ID:        newMsgID(),
			Role:      role,
			Text:      e.Text,
			ItemRef:   e.ItemRef,
			Prompt:    e.Prompt,
			CreatedAt: now,
		}
		plan.Messages = append(plan.Messages, msg)
		msgs = append(msgs, msg)
	}
	s.mu.Unlock()

	if len(msgs) == 0 {
		return nil, fmt.Errorf("no non-empty messages to add")
	}

	if err := s.persistPlan(id); err != nil {
		return nil, err
	}
	if s.onChange != nil {
		s.onChange(id)
	}
	return msgs, nil
}

// DeleteMessage removes a message from the plan's conversation thread by ID.
func (s *PlanStore) DeleteMessage(planID, msgID string) error {
	s.mu.Lock()

	plan, ok := s.plans[planID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", planID)
	}

	found := false
	for i, msg := range plan.Messages {
		if msg.ID == msgID {
			plan.Messages = append(plan.Messages[:i], plan.Messages[i+1:]...)
			found = true
			break
		}
	}
	s.mu.Unlock()

	if !found {
		return fmt.Errorf("message not found: %s", msgID)
	}

	if err := s.persistPlan(planID); err != nil {
		return err
	}
	if s.onChange != nil {
		s.onChange(planID)
	}
	return nil
}

// DeletePlan removes a plan from in-memory store and deletes its .json file from disk.
func (s *PlanStore) DeletePlan(id string) error {
	s.mu.Lock()

	if _, ok := s.plans[id]; !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	delete(s.plans, id)
	s.mu.Unlock()

	path := filepath.Join(s.dir, id+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete plan file: %w", err)
	}

	if s.onChange != nil {
		s.onChange(id)
	}
	return nil
}

// CreatePlan creates a new plan with the given title and summary, persists it,
// and returns the created plan. Uses a kebab-case ID derived from the title.
func (s *PlanStore) CreatePlan(id, title, summary string) (*Plan, error) {
	s.mu.Lock()
	if _, ok := s.plans[id]; ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("plan already exists: %s", id)
	}

	plan := &Plan{
		Title:   title,
		Summary: summary,
		State:   "draft",
	}
	s.plans[id] = plan
	s.mu.Unlock()

	if err := s.persistPlan(id); err != nil {
		s.mu.Lock()
		delete(s.plans, id)
		s.mu.Unlock()
		return nil, err
	}
	if s.onChange != nil {
		s.onChange(id)
	}
	return plan, nil
}

// RemovePlan removes a plan from in-memory store without touching disk.
// Used when the file was already deleted externally (e.g., via file watcher).
func (s *PlanStore) RemovePlan(id string) {
	s.mu.Lock()
	delete(s.plans, id)
	s.mu.Unlock()
}

// SetState sets the plan's state. Only "draft" and "approved" are valid.
func (s *PlanStore) SetState(id, state string) error {
	if state != "draft" && state != "approved" {
		return fmt.Errorf("invalid state: %s", state)
	}

	s.mu.Lock()

	plan, ok := s.plans[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	plan.State = state
	s.mu.Unlock()

	if err := s.persistPlan(id); err != nil {
		return err
	}
	if s.onChange != nil {
		s.onChange(id)
	}
	return nil
}

// Get returns a plan by its ID.
func (s *PlanStore) Get(id string) *Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plans[id]
}

// UpsertPlan creates or replaces a plan in the store and persists it to disk.
func (s *PlanStore) UpsertPlan(id string, plan *Plan) error {
	plan.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.mu.Lock()
	s.plans[id] = plan
	s.mu.Unlock()

	if err := s.persistPlan(id); err != nil {
		return err
	}
	if s.onChange != nil {
		s.onChange(id)
	}
	return nil
}

// PatchPlan merges partial updates into an existing plan, preserving messages.
// Only non-zero/non-empty fields in the input are applied.
func (s *PlanStore) PatchPlan(id string, patch *Plan) error {
	s.mu.Lock()
	existing, ok := s.plans[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	if patch.Title != "" {
		existing.Title = patch.Title
	}
	if patch.Summary != "" {
		existing.Summary = patch.Summary
	}
	if patch.State != "" {
		existing.State = patch.State
	}
	if patch.Modules != nil {
		existing.Modules = patch.Modules
	}
	s.mu.Unlock()

	if err := s.persistPlan(id); err != nil {
		return err
	}
	if s.onChange != nil {
		s.onChange(id)
	}
	return nil
}

// List returns all plan IDs and their titles, sorted by most recently updated first.
func (s *PlanStore) List() []PlanSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PlanSummary, 0, len(s.plans))
	for id, p := range s.plans {
		out = append(out, PlanSummary{
			ID:        id,
			Title:     p.Title,
			Summary:   p.Summary,
			UpdatedAt: p.UpdatedAt,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		// Empty UpdatedAt sorts to the end (oldest position)
		if out[i].UpdatedAt == "" {
			return false
		}
		if out[j].UpdatedAt == "" {
			return true
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out
}

// IsSelfWrite checks whether the server itself wrote this file at the given mtime.
// This lets the poller skip redundant reloads after server-initiated writes.
func (s *PlanStore) IsSelfWrite(filename string, mtime time.Time) bool {
	id := strings.TrimSuffix(filename, ".json")
	s.mu.RLock()
	defer s.mu.RUnlock()
	last, ok := s.lastWrite[id]
	return ok && mtime.Equal(last)
}

// PlanSummary is a lightweight view for listing.
type PlanSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
