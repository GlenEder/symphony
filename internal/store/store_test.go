package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gleneder/symphony/internal/model"
)

func TestLoadAllLeftoverTempFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "valid.json"), []byte(`{"title":"Valid","modules":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	tempPath := filepath.Join(dir, ".plan-1-12345.json")
	if err := os.WriteFile(tempPath, []byte(`{"not":"a plan"}`), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(dir, nil)
	if err := s.LoadAll(); err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if s.Get("valid") == nil {
		t.Fatal("valid plan was not loaded")
	}
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Fatalf("leftover temp file still exists: %v", err)
	}
}

func TestLoadAllCorruptNonTempJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(path, []byte(`{"title":`), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(dir, nil)
	if err := s.LoadAll(); err == nil {
		t.Fatal("LoadAll() succeeded for corrupt non-temp JSON")
	}
}

func TestDeletePlanRemovesPersistedFile(t *testing.T) {
	s := New(t.TempDir(), nil)
	if _, err := s.CreatePlan("delete-me", "Delete", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.DeletePlan("delete-me"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.Dir + "/delete-me.json"); !os.IsNotExist(err) {
		t.Fatalf("plan file still exists: %v", err)
	}
	if s.Get("delete-me") != nil {
		t.Fatal("deleted plan remains in memory")
	}
}

func TestPlanRoundTripKeepsAllModuleTypes(t *testing.T) {
	s := New(t.TempDir(), nil)
	p := &model.Plan{Modules: []model.Module{{Type: "criteria"}, {Type: "steps"}, {Type: "assumptions"}, {Type: "changes"}, {Type: "notes"}}}
	if err := s.Upsert("modules", p); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Get("modules").Modules); got != 5 {
		t.Fatalf("got %d modules, want 5", got)
	}
}
