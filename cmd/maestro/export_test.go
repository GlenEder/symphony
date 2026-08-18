package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gleneder/symphony/internal/export"
	"github.com/gleneder/symphony/internal/model"
)

func TestExportApprovedPlanWritesDeterministicTicket(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := &model.Plan{
		Title:   "Stage 6",
		State:   "approved",
		Summary: "Finish the boundary.",
		Session: &model.SessionState{ID: "plan-stage-6"},
		Modules: []model.Module{{Type: "criteria", Heading: "Acceptance Criteria", Items: []model.Item{{Text: "Exports safely"}}}},
	}
	if err := export.New(export.Config{}).Export(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "symphony", "work_tickets", "plan-stage-6.md")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# Stage 6", "**Status**: pending", "- [ ] Exports safely"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("ticket missing %q", want)
		}
	}
}
