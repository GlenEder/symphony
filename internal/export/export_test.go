package export

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gleneder/symphony/internal/model"
)

func testPlan(state, id string, modules ...model.Module) *model.Plan {
	return &model.Plan{Title: "Ship feature", Summary: "A useful summary.", State: state, Session: &model.SessionState{ID: id}, Modules: modules}
}
func mod(typ, heading string, text ...string) model.Module {
	m := model.Module{Type: typ, Heading: heading}
	for _, x := range text {
		m.Items = append(m.Items, model.Item{Text: x})
	}
	return m
}

func TestStagedExportExactOrderAndFlattening(t *testing.T) {
	dir := t.TempDir()
	p := testPlan("approved", "safe-plan", mod("decision", "Decision", "Use files"), mod("assumptions", "Assumptions", "Go is available"), mod("criteria", "Stage 1: Foundation", "first criterion", "second criterion"), mod("steps", "Stage 1: Foundation", "first step"), mod("risks", "Stage 1: Foundation", "risky"), mod("criteria", "Stage 2: Finish", "last criterion"), mod("steps", "Stage 2: Finish", "last step"))
	if err := New(Config{OutputDir: dir, TraceabilityURL: "http://localhost:8080/plan"}).Export(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "safe-plan-stage-1.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	want := []string{"# Ship feature — Stage 1: Foundation", "**Status**: pending", "**Stage**: 1 of 2", "## Key Decisions", "## Assumptions", "## Acceptance Criteria", "## Implementation Steps", "## Risks", "Generated from [maestro plan](http://localhost:8080/plan/safe-plan)"}
	last := -1
	for _, x := range want {
		n := strings.Index(s, x)
		if n <= last {
			t.Errorf("section %q out of order", x)
		}
		last = n
	}
	if strings.Count(s, "first criterion") != 1 || strings.Contains(s, "last criterion") {
		t.Errorf("wrong stage contents: %s", s)
	}
}

func TestSingleStageAndRejectedDraft(t *testing.T) {
	dir := t.TempDir()
	p := testPlan("approved", "legacy", mod("criteria", "Acceptance Criteria", "works"), mod("steps", "Implementation Steps", "build"))
	if err := New(Config{OutputDir: dir}).Export(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "legacy.md")); err != nil {
		t.Fatal(err)
	}
	p.State = "draft"
	if err := New(Config{OutputDir: dir}).Export(context.Background(), p); err == nil {
		t.Fatal("draft exported")
	}
}

func TestInvalidPlansWriteNothing(t *testing.T) {
	for _, modules := range [][]model.Module{{mod("criteria", "Acceptance Criteria", "bad"), mod("criteria", "Stage 2: Nope", "bad")}, {mod("decision", "Stage 1: Bad", "bad")}, {mod("criteria", "Stage 1: One", "ok"), mod("criteria", "Stage 3: Three", "gap")}} {
		dir := t.TempDir()
		p := testPlan("approved", "plan", modules...)
		if err := New(Config{OutputDir: dir}).Export(context.Background(), p); err == nil {
			t.Fatal("invalid plan exported")
		}
		entries, _ := os.ReadDir(dir)
		if len(entries) != 0 {
			t.Fatalf("partial output: %v", entries)
		}
	}
}

func TestMaliciousIDAndCancellation(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"../escape", ".", "a/b"} {
		if err := New(Config{OutputDir: dir}).Export(context.Background(), testPlan("approved", id)); err == nil {
			t.Errorf("accepted %q", id)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := New(Config{OutputDir: dir}).Export(ctx, testPlan("approved", "cancelled")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}
