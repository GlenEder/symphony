package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gleneder/symphony/internal/model"
)

func TestWelcomeTemplateUsesExplicitStatusAndElementReferences(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "templates", "welcome.html"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	for _, want := range []string{"statusEl.textContent", "Researching your codebase…", "Preparing next question…", "Grilling complete. Preparing the plan…", "document.getElementById('status')"} {
		if !strings.Contains(script, want) {
			t.Errorf("welcome template missing %q", want)
		}
	}
	for _, forbidden := range []string{"status.textContent", "context.textContent", "prompt.value", "progress.textContent", "question.textContent", "options.replaceChildren", "custom.value"} {
		if strings.Contains(script, forbidden) {
			t.Errorf("welcome template uses bare element reference %q", forbidden)
		}
	}
}

func renderPlanTemplate(t *testing.T, p *model.Plan) string {
	t.Helper()
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
	var out bytes.Buffer
	if err := page.ExecuteTemplate(&out, "base", struct {
		Title int
		Year  int
		Plan  *model.Plan
	}{Year: 2026, Plan: p}); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestPlanTemplateRendersModuleSpecificFields(t *testing.T) {
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
	out := renderPlanTemplate(t, p)
	for _, field := range []string{"criterion", "step", "owner", "done", "assumption", "file.go", "modify", "note", "risk", "impact", "mitigation", "high", "decision", "option", "reason", "question", "answer", "planned"} {
		if !strings.Contains(out, field) {
			t.Errorf("rendered output missing %q", field)
		}
	}
}

func TestPlanTemplateGatesReviewControlsAndShowsExportFailure(t *testing.T) {
	out := renderPlanTemplate(t, &model.Plan{Title: "Approved", Session: &model.SessionState{State: "approved", ExportStatus: "failed", ExportError: "ticket export failed"}})
	for _, want := range []string{"This plan is read-only", "Export failed: ticket export failed", "WebSocket", "/ws/plan/", "render(JSON.parse(e.data))", "renderModules", "drafts()", "review_state", "export_status", "pagehide", "beforeunload", "reconnectDelay", "Math.min(d*2,30000)"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output missing %q", want)
		}
	}
	for _, forbidden := range []string{"id=\"approve-plan\"", "id=\"general-feedback\"", "class=\"item-feedback\"", "location.reload", "onmessage=()=>{pending=true;scheduleRefresh()}"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("read-only output contains %q", forbidden)
		}
	}
}

func TestPlanTemplateRendersExportStatuses(t *testing.T) {
	cases := []struct {
		name   string
		state  string
		status string
		want   string
	}{
		{name: "successful", state: "approved", status: "succeeded", want: "Export completed successfully."},
		{name: "pending", state: "exporting", status: "pending", want: "Export status: pending"},
		{name: "failed", state: "failed", status: "failed", want: "Export failed: failed export"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := renderPlanTemplate(t, &model.Plan{Title: "Export", Session: &model.SessionState{State: tc.state, ExportStatus: tc.status, ExportError: "failed export"}})
			if !strings.Contains(out, tc.want) {
				t.Errorf("rendered output missing %q", tc.want)
			}
		})
	}
}
