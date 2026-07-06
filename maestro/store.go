package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sstraus/toon_go/toon"
)

// PlanStore holds loaded plans and notifies on changes.
type PlanStore struct {
	mu       sync.RWMutex
	plans    map[string]*Plan
	onChange func(id string)
	dir      string
}

func NewPlanStore(dir string, onChange func(id string)) *PlanStore {
	return &PlanStore{
		plans:    make(map[string]*Plan),
		onChange: onChange,
		dir:      dir,
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
	id := strings.TrimSuffix(filepath.Base(path), ".toon")
	s.mu.Lock()
	s.plans[id] = plan
	s.mu.Unlock()
	log.Printf("loaded plan: %s (%s)", id, plan.Title)
}

// SaveResponse updates a plan's Response field and persists the change to disk.
func (s *PlanStore) SaveResponse(id, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan, ok := s.plans[id]
	if !ok {
		return fmt.Errorf("plan not found: %s", id)
	}

	plan.Response = text

	// Round-trip through JSON to get a plain map[string]any for TOON encoding
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
	return os.WriteFile(path, toonBytes, 0644)
}

// SetApproved updates a plan's Approved status and persists the change to disk.
func (s *PlanStore) SetApproved(id string, approved bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan, ok := s.plans[id]
	if !ok {
		return fmt.Errorf("plan not found: %s", id)
	}

	plan.Approved = approved

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
	return os.WriteFile(path, toonBytes, 0644)
}

// Get returns a plan by its ID.
func (s *PlanStore) Get(id string) *Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plans[id]
}

// List returns all plan IDs and their titles.
func (s *PlanStore) List() []PlanSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PlanSummary, 0, len(s.plans))
	for id, p := range s.plans {
		out = append(out, PlanSummary{
			ID:      id,
			Title:   p.Title,
			Summary: p.Summary,
		})
	}
	return out
}

// PlanSummary is a lightweight view for listing.
type PlanSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
