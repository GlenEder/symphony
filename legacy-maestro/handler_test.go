package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// setUpTestHandler creates a store, loads a test plan, and returns a mux
// with the POST /api/plan/{id}/messages route registered.
func setUpTestHandler(t *testing.T) (*PlanStore, *AgentState, *http.ServeMux) {
	t.Helper()

	dir := t.TempDir()
	planContent := `{
  "title": "Handler Test",
  "summary": "handler-level prompt test",
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
	if err := os.WriteFile(filepath.Join(dir, "handler-test.json"), []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	store := NewPlanStore(dir, nil)
	if err := store.LoadAll(); err != nil {
		t.Fatal(err)
	}

	agentState := NewAgentState()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/plan/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		planID := r.PathValue("id")
		if planID == "" {
			http.NotFound(w, r)
			return
		}

		// Try batch format first
		var batchBody struct {
			Messages []struct {
				Role    string  `json:"role"`
				Text    string  `json:"text"`
				ItemRef string  `json:"item_ref,omitempty"`
				Prompt  *Prompt `json:"prompt,omitempty"`
			} `json:"messages"`
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(bodyBytes, &batchBody); err == nil && len(batchBody.Messages) > 1 {
			role := batchBody.Messages[0].Role
			if role != "agent" && role != "human" {
				http.Error(w, "role must be 'agent' or 'human'", http.StatusBadRequest)
				return
			}
			entries := make([]MsgEntry, 0, len(batchBody.Messages))
			for _, m := range batchBody.Messages {
				if m.Role != role {
					http.Error(w, "all messages in a batch must have the same role", http.StatusBadRequest)
					return
				}
				if m.Text == "" {
					http.Error(w, "text is required for all messages", http.StatusBadRequest)
					return
				}
				if m.Prompt != nil && m.Prompt.Recommended > len(m.Prompt.Options) {
					http.Error(w, "recommended index is out of bounds", http.StatusBadRequest)
					return
				}
				entries = append(entries, MsgEntry{Text: m.Text, ItemRef: m.ItemRef, Prompt: m.Prompt})
			}
			msgs, err := store.AddMessages(planID, role, entries)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if role == "human" {
				agentState.SetThinking(planID)
			} else if role == "agent" {
				agentState.SetListening(planID)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(msgs)
			return
		}

		// Single message
		var single struct {
			Role    string  `json:"role"`
			Text    string  `json:"text"`
			ItemRef string  `json:"item_ref,omitempty"`
			Prompt  *Prompt `json:"prompt,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &single); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if single.Role != "agent" && single.Role != "human" {
			http.Error(w, "role must be 'agent' or 'human'", http.StatusBadRequest)
			return
		}
		if single.Text == "" {
			http.Error(w, "text is required", http.StatusBadRequest)
			return
		}
		if single.Prompt != nil && single.Prompt.Recommended > len(single.Prompt.Options) {
			http.Error(w, "recommended index is out of bounds", http.StatusBadRequest)
			return
		}
		msg, err := store.AddMessage(planID, single.Role, single.Text, single.ItemRef, single.Prompt)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if single.Role == "human" {
			agentState.SetThinking(planID)
		} else if single.Role == "agent" {
			agentState.SetListening(planID)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(msg)
	})

	return store, agentState, mux
}

func TestHandler_PostSingleMessage_WithPrompt(t *testing.T) {
	store, _, mux := setUpTestHandler(t)

	body := map[string]interface{}{
		"role": "agent",
		"text": "Which database?",
		"prompt": map[string]interface{}{
			"question_key":    "db-choice",
			"options":         []string{"PostgreSQL", "SQLite"},
			"allow_custom":    true,
			"total_questions": 3,
			"recommended":     1,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/plan/handler-test/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var msg Message
	if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if msg.Prompt == nil {
		t.Fatal("prompt was nil in response")
	}
	if msg.Prompt.QuestionKey != "db-choice" {
		t.Errorf("expected question_key 'db-choice', got %q", msg.Prompt.QuestionKey)
	}
	if len(msg.Prompt.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(msg.Prompt.Options))
	}

	// Verify persisted round-trip via the store
	plan := store.Get("handler-test")
	if plan == nil || len(plan.Messages) != 1 {
		t.Fatal("message not persisted")
	}
	if plan.Messages[0].Prompt == nil {
		t.Fatal("prompt lost in persisted message")
	}
}

func TestHandler_PostMessage_WithInvalidRecommended(t *testing.T) {
	_, _, mux := setUpTestHandler(t)

	body := map[string]interface{}{
		"role": "agent",
		"text": "Which database?",
		"prompt": map[string]interface{}{
			"question_key":    "db-choice",
			"options":         []string{"PostgreSQL", "SQLite"},
			"allow_custom":    true,
			"total_questions": 3,
			"recommended":     99,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/plan/handler-test/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-bounds recommended, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_PostBatchMessages_WithPrompt(t *testing.T) {
	store, _, mux := setUpTestHandler(t)

	body := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "agent",
				"text": "Question 1",
				"prompt": map[string]interface{}{
					"question_key":    "q1",
					"options":         []string{"A", "B"},
					"total_questions": 2,
					"recommended":     1,
				},
			},
			{
				"role": "agent",
				"text": "Question 2",
				"prompt": map[string]interface{}{
					"question_key":    "q2",
					"options":         []string{"C", "D"},
					"total_questions": 2,
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/plan/handler-test/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var msgs []Message
	if err := json.Unmarshal(rec.Body.Bytes(), &msgs); err != nil {
		t.Fatalf("failed to unmarshal batch response: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Prompt == nil || msgs[0].Prompt.QuestionKey != "q1" {
		t.Errorf("batch[0] prompt mismatch: %+v", msgs[0].Prompt)
	}
	if msgs[1].Prompt == nil || msgs[1].Prompt.QuestionKey != "q2" {
		t.Errorf("batch[1] prompt mismatch: %+v", msgs[1].Prompt)
	}

	// Verify persisted
	plan := store.Get("handler-test")
	if plan == nil || len(plan.Messages) != 2 {
		t.Fatalf("expected 2 persisted messages, got %d", len(plan.Messages))
	}
}

func TestHandler_PostMessage_WithoutPrompt(t *testing.T) {
	// Verify the handler still works without prompt field
	_, _, mux := setUpTestHandler(t)

	body := map[string]interface{}{
		"role": "human",
		"text": "I think PostgreSQL is fine",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/plan/handler-test/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var msg Message
	if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if msg.Prompt != nil {
		t.Error("expected nil prompt when not provided")
	}
}
