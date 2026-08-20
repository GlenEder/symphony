package handler

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/conversation"
	"github.com/gleneder/symphony/internal/store"
	ws "github.com/gleneder/symphony/internal/websocket"
)

func testHandler(t *testing.T) (*http.ServeMux, *store.PlanStore) {
	t.Helper()
	s := store.New(t.TempDir(), nil)
	if _, err := s.CreatePlan("plan-1", "Plan", "Summary"); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	Register(mux, template.New("test"), s, ws.NewHub(), ws.NewAgentState(), ".")
	return mux, s
}

func TestSessionRetryRouteIsReachable(t *testing.T) {
	s := store.New(t.TempDir(), nil)
	m := conversation.NewManager(s, failingResearcher{}, nil, ".")
	defer m.Close()
	started, err := m.Start(context.Background(), "prompt")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := m.Session(started.ID); got != nil && got.State == conversation.Failed {
			break
		}
		time.Sleep(time.Millisecond)
	}
	mux := http.NewServeMux()
	Register(mux, template.New("test"), s, ws.NewHub(), ws.NewAgentState(), ".", m)
	r := httptest.NewRequest(http.MethodPost, "/api/session/"+started.ID+"/retry", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	get := httptest.NewRequest(http.MethodGet, "/api/session/"+started.ID+"/retry", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, get)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want 405", w.Code)
	}
}

type failingResearcher struct{}

func (failingResearcher) Summary(string, codebase.SummaryOptions) (string, error) {
	return "", errors.New("failed")
}

func TestLegacyJSONEndpointsRejectMalformedPayloads(t *testing.T) {
	cases := []struct {
		name, method, path, body string
	}{
		{"create unknown", http.MethodPost, "/api/plans", `{"id":"x","title":"X","extra":true}`},
		{"create trailing", http.MethodPost, "/api/plans", `{"id":"x","title":"X"}{}`},
		{"message unknown", http.MethodPost, "/api/plan/plan-1/messages", `{"text":"x","extra":true}`},
		{"state trailing", http.MethodPost, "/api/plan/plan-1/state", `{"state":"draft"} {}`},
		{"upsert unknown", http.MethodPut, "/api/plan/plan-1", `{"title":"x","extra":true}`},
		{"patch trailing", http.MethodPatch, "/api/plan/plan-1", `{"summary":"x"}{}`},
		{"status unknown", http.MethodPost, "/api/agent/plan-1/status", `{"status":"offline","extra":true}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux, _ := testHandler(t)
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code < 400 || w.Code >= 500 {
				t.Fatalf("status = %d, want 4xx", w.Code)
			}
		})
	}
}

func TestJSONEndpointsRejectOversizedPayloads(t *testing.T) {
	mux, _ := testHandler(t)
	body := bytes.Repeat([]byte("x"), 6000)
	r := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(append([]byte(`{"id":"x","title":"`), append(body, []byte(`"}`)...)...)))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code < 400 || w.Code >= 500 {
		t.Fatalf("status = %d, want 4xx", w.Code)
	}
	_ = os.Remove(filepath.Join(t.TempDir(), "unused"))
}
