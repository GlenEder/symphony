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

	"github.com/sstraus/toon_go/toon"
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

// LoadAll reads all .toon files from the plans directory.
func (s *PlanStore) LoadAll() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(s.dir, 0755)
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toon") {
			continue
		}
		s.loadFile(filepath.Join(s.dir, e.Name()))
	}
	return nil
}

func decodePlan(data []byte) (*Plan, error) {
	var raw map[string]any
	if err := toon.Unmarshal(data, &raw, &toon.DecodeOptions{Strict: false}); err != nil {
		return nil, fmt.Errorf("toon decode: %w", err)
	}
	js, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}
	var plan Plan
	if err := json.Unmarshal(js, &plan); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return &plan, nil
}

// loadFile reads a single .toon file and stores it by its basename (without ext).
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
	id := strings.TrimSuffix(filepath.Base(path), ".toon")
	s.mu.Lock()
	s.plans[id] = plan
	s.mu.Unlock()
	log.Printf("loaded plan: %s (%s)", id, plan.Title)
}

// persistPlan writes the in-memory plan to disk as a .toon file
// and records the file's mtime so the poller can skip the self-triggered reload.
func (s *PlanStore) persistPlan(id string) error {
	plan, ok := s.plans[id]
	if !ok {
		return fmt.Errorf("plan not found: %s", id)
	}
	plan.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	js, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(js, &raw); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}
	toonBytes, err := toon.Marshal(raw, &toon.EncodeOptions{Indent: 2})
	if err != nil {
		return fmt.Errorf("toon marshal: %w", err)
	}
	path := filepath.Join(s.dir, id+".toon")
	if err := os.WriteFile(path, toonBytes, 0644); err != nil {
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
func (s *PlanStore) AddMessage(id, role, text, itemRef string) (*Message, error) {
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

// DeletePlan removes a plan from in-memory store and deletes its .toon file from disk.
func (s *PlanStore) DeletePlan(id string) error {
	s.mu.Lock()

	if _, ok := s.plans[id]; !ok {
		s.mu.Unlock()
		return fmt.Errorf("plan not found: %s", id)
	}
	delete(s.plans, id)
	s.mu.Unlock()

	path := filepath.Join(s.dir, id+".toon")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete plan file: %w", err)
	}

	if s.onChange != nil {
		s.onChange(id)
	}
	return nil
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
	id := strings.TrimSuffix(filename, ".toon")
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
