package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gleneder/symphony/internal/model"
)

func TestPlanTemplateRendersModuleSpecificFields(t *testing.T) {
	base := filepath.Join("..", "..")
	tmpl, err := parseTemplates(base)
	if err != nil {
		t.Fatal(err)
	}
	page, err := tmpl.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := page.ParseFiles(filepath.Join(base, "templates", "plan.html")); err != nil {
		t.Fatal(err)
	}
	p := &model.Plan{Title: "Module fields", Modules: []model.Module{
		{Type: "criteria", Heading: "Criteria", Items: []model.Item{{Text: "criterion", Answered: true}}},
		{Type: "steps", Heading: "Steps", Items: []model.Item{{Text: "step", Owner: "owner", Status: "done"}}},
		{Type: "assumptions", Heading: "Assumptions", Items: []model.Item{{Text: "assumption"}}},
		{Type: "changes", Heading: "Changes", Items: []model.Item{{Text: "file.go", ChangeType: "modify"}}},
		{Type: "notes", Heading: "Notes", Items: []model.Item{{Text: "note"}}},
		{Type: "risks", Heading: "Risks", Items: []model.Item{{Text: "risk", Severity: "high", Impact: "impact", Mitigation: "mitigation", Owner: "owner"}}},
		{Type: "decision", Heading: "Decisions", Items: []model.Item{{Text: "decision", Options: "option", Rationale: "reason"}}},
		{Type: "questions", Heading: "Questions", Items: []model.Item{{Text: "question", Answered: true, Answer: "answer"}}},
		{Type: "table", Heading: "Table", Columns: []model.Column{{Heading: "Status", Key: "status"}}, Items: []model.Item{{Status: "planned"}}},
	}}
	var out bytes.Buffer
	if err := page.ExecuteTemplate(&out, "base", struct {
		Title int
		Year  int
		Plan  *model.Plan
	}{Year: 2026, Plan: p}); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"criterion", "step", "owner", "done", "assumption", "file.go", "modify", "note", "risk", "impact", "mitigation", "high", "decision", "option", "reason", "question", "answer", "planned"} {
		if !strings.Contains(out.String(), field) {
			t.Errorf("rendered output missing %q", field)
		}
	}
}
