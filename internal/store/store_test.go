package store

import (
	"os"
	"testing"

	"github.com/gleneder/symphony/internal/model"
)

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
