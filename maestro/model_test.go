package main

import (
	"strings"
	"testing"
)

const decisionPlan = `{
  "title": "Decision Test",
  "summary": "exercises the decision module type",
  "state": "draft",
  "modules": [
    {
      "type": "decision",
      "heading": "Key Decisions",
      "items": [
        {
          "text": "Use library X for the search layer",
          "options": "library Y — too heavy; library Z — unmaintained",
          "rationale": "X wins on speed and maintenance"
        },
        {
          "text": "Build the indexing pipeline in-house",
          "options": "build in-house; buy SaaS",
          "rationale": "tight latency requirements justify the build cost"
        }
      ]
    }
  ]
}`

func TestDecodePlan_DecisionModule(t *testing.T) {
	plan, err := decodePlan([]byte(decisionPlan))
	if err != nil {
		t.Fatalf("decodePlan: %v", err)
	}
	if len(plan.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(plan.Modules))
	}
	m := plan.Modules[0]
	if m.Type != "decision" {
		t.Errorf("expected module type decision, got %q", m.Type)
	}
	if m.Heading != "Key Decisions" {
		t.Errorf("expected heading 'Key Decisions', got %q", m.Heading)
	}
	if len(m.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(m.Items))
	}
	first := m.Items[0]
	if first.Text != "Use library X for the search layer" {
		t.Errorf("unexpected text: %q", first.Text)
	}
	if !strings.Contains(first.Options, "library Y") {
		t.Errorf("expected options to mention 'library Y', got %q", first.Options)
	}
	if !strings.Contains(first.Rationale, "speed") {
		t.Errorf("expected rationale to mention 'speed', got %q", first.Rationale)
	}
	second := m.Items[1]
	if second.Options == "" || second.Rationale == "" {
		t.Errorf("expected second item to carry options and rationale, got options=%q rationale=%q", second.Options, second.Rationale)
	}
}

func TestFlatPlan_PreservesDecisionFields(t *testing.T) {
	plan := &Plan{
		Title: "Flat Decision Test",
		Modules: []Module{
			{
				Type:    "decision",
				Heading: "Key Decisions",
				Items: []Item{
					{Text: "Pick X", Options: "Y; Z", Rationale: "X is faster"},
				},
			},
		},
	}
	fp := toFlatPlan(plan, "listening")
	if len(fp.Modules) != 1 || len(fp.Modules[0].Items) != 1 {
		t.Fatalf("expected 1 module with 1 item, got %d modules", len(fp.Modules))
	}
	it := fp.Modules[0].Items[0]
	if it.Options != "Y; Z" {
		t.Errorf("flat item lost options: got %q", it.Options)
	}
	if it.Rationale != "X is faster" {
		t.Errorf("flat item lost rationale: got %q", it.Rationale)
	}
}
