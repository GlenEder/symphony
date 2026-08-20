package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

const promptPlan = `{
  "title": "Prompt Test",
  "summary": "exercises prompt-bearing agent messages",
  "state": "draft",
  "modules": [
    {
      "type": "questions",
      "heading": "Open Questions",
      "items": [
        {"text": "which database should we use"},
        {"text": "what deployment strategy do you prefer"}
      ]
    }
  ],
  "messages": [
    {
      "role": "agent",
      "text": "Which database engine would you prefer for this project?",
      "prompt": {
        "question_key": "db-choice",
        "options": ["PostgreSQL", "SQLite", "MySQL"],
        "allow_custom": true,
        "total_questions": 3
      }
    },
    {
      "role": "human",
      "text": "PostgreSQL"
    }
  ]
}`

func TestDecodePlan_WithPrompt(t *testing.T) {
	plan, err := decodePlan([]byte(promptPlan))
	if err != nil {
		t.Fatalf("decodePlan: %v", err)
	}
	if len(plan.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(plan.Messages))
	}
	m := plan.Messages[0]
	if m.Role != "agent" {
		t.Errorf("expected agent role, got %q", m.Role)
	}
	if m.Prompt == nil {
		t.Fatal("expected prompt to be non-nil")
	}
	if m.Prompt.QuestionKey != "db-choice" {
		t.Errorf("expected question_key 'db-choice', got %q", m.Prompt.QuestionKey)
	}
	if !m.Prompt.AllowCustom {
		t.Error("expected allow_custom to be true")
	}
	if len(m.Prompt.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(m.Prompt.Options))
	}
	if m.Prompt.TotalQuestions != 3 {
		t.Errorf("expected total_questions=3, got %d", m.Prompt.TotalQuestions)
	}
	if m.Prompt.Answered {
		t.Error("expected answered=false")
	}
	if plan.Messages[1].Role != "human" {
		t.Errorf("expected second message role 'human', got %q", plan.Messages[1].Role)
	}
}

func TestPlan_JSONRoundTrip_WithPrompt(t *testing.T) {
	plan := &Plan{
		Title:   "Prompt Serialize",
		Summary: "Round-trip test",
		State:   "draft",
		Messages: []Message{
			{
				ID:   "msg_abc",
				Role: "agent",
				Text: "Which database?",
				Prompt: &Prompt{
					QuestionKey:    "db-choice",
					Options:        []string{"PostgreSQL", "SQLite"},
					AllowCustom:    true,
					TotalQuestions: 2,
					Answered:       false,
					Recommended:    2,
				},
			},
			{
				ID:   "msg_def",
				Role: "human",
				Text: "PostgreSQL",
			},
		},
	}
	b, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Plan
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Messages) != 2 {
		t.Fatalf("expected 2 messages after round-trip, got %d", len(decoded.Messages))
	}
	m0 := decoded.Messages[0]
	if m0.Prompt == nil {
		t.Fatal("prompt lost during round-trip")
	}
	if m0.Prompt.QuestionKey != "db-choice" {
		t.Errorf("question_key mismatch: got %q", m0.Prompt.QuestionKey)
	}
	if len(m0.Prompt.Options) != 2 {
		t.Errorf("options length mismatch: got %d", len(m0.Prompt.Options))
	}
}

func TestAddMessage_WithPrompt(t *testing.T) {
	dir := t.TempDir()
	// Create a plan file first
	planContent := `{
  "title": "Prompt Msg Test",
  "summary": "testing AddMessage with prompt",
  "state": "draft",
  "modules": [
    {
      "type": "notes",
      "heading": "Notes",
      "items": [
        {"text": "placeholder"}
      ]
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(dir, "promptmsg.json"), []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewPlanStore(dir, nil)
	store.LoadAll()

	plan := store.Get("promptmsg")
	if plan == nil {
		t.Fatal("plan not found after LoadAll")
	}

	prompt := &Prompt{
		QuestionKey:    "env-choice",
		Options:        []string{"staging", "production"},
		AllowCustom:    false,
		TotalQuestions: 1,
	}
	msg, err := store.AddMessage("promptmsg", "agent", "Which environment?", "", prompt)
	if err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	if msg.Prompt == nil {
		t.Fatal("prompt was nil in returned message")
	}
	if msg.Prompt.QuestionKey != "env-choice" {
		t.Errorf("expected question_key 'env-choice', got %q", msg.Prompt.QuestionKey)
	}
	if len(msg.Prompt.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(msg.Prompt.Options))
	}
	if msg.Prompt.TotalQuestions != 1 {
		t.Errorf("expected total_questions=1, got %d", msg.Prompt.TotalQuestions)
	}

	// Verify persisted round-trip — reload from disk
	store2 := NewPlanStore(dir, nil)
	store2.LoadAll()
	plan2 := store2.Get("promptmsg")
	if plan2 == nil {
		t.Fatal("plan not found after reload")
	}
	if len(plan2.Messages) != 1 {
		t.Fatalf("expected 1 message after reload, got %d", len(plan2.Messages))
	}
	m2 := plan2.Messages[0]
	if m2.Prompt == nil {
		t.Fatal("prompt lost during persist round-trip")
	}
	if m2.Prompt.QuestionKey != "env-choice" {
		t.Errorf("question_key mismatch after reload: got %q", m2.Prompt.QuestionKey)
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

func TestFlatPlan_PreservesPromptMessages(t *testing.T) {
	plan := &Plan{
		Title: "Flat Prompt Test",
		Messages: []Message{
			{
				ID:   "msg_abc",
				Role: "agent",
				Text: "Which database?",
				Prompt: &Prompt{
					QuestionKey:    "db-choice",
					Options:        []string{"PostgreSQL", "SQLite"},
					AllowCustom:    true,
					TotalQuestions: 2,
					Answered:       false,
					Recommended:    2,
				},
			},
		},
	}
	fp := toFlatPlan(plan, "listening")
	if len(fp.Messages) != 1 {
		t.Fatalf("expected 1 message in flat plan, got %d", len(fp.Messages))
	}
	m := fp.Messages[0]
	if m.Prompt == nil {
		t.Fatal("prompt lost in FlatPlan conversion")
	}
	if m.Prompt.QuestionKey != "db-choice" {
		t.Errorf("question_key mismatch: got %q", m.Prompt.QuestionKey)
	}
	if len(m.Prompt.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(m.Prompt.Options))
	}
	if !m.Prompt.AllowCustom {
		t.Error("expected allow_custom=true")
	}
	if m.Prompt.TotalQuestions != 2 {
		t.Errorf("expected total_questions=2, got %d", m.Prompt.TotalQuestions)
	}
}
