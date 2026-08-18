package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/conversation"
	"github.com/gleneder/symphony/internal/export"
	"github.com/gleneder/symphony/internal/handler"
	"github.com/gleneder/symphony/internal/llm"
	"github.com/gleneder/symphony/internal/model"
	"github.com/gleneder/symphony/internal/store"
	"github.com/gleneder/symphony/internal/websocket"
)

func TestHTTPWorkflowWithLocalOpenAIProvider(t *testing.T) {
	codebaseDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(codebaseDir, "main.go"), []byte("package example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []llm.Message `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		prompt := request.Messages[len(request.Messages)-1].Content
		content := `{"done":false,"question_key":"scope","question":"Which scope?","options":["small","large"],"allow_custom":true,"total_questions":1}`
		if strings.Contains(request.Messages[0].Content, "Required types") {
			content = integrationPlanJSON()
		} else if strings.Contains(prompt, "Answered question") {
			content = `{"done":true,"question_key":"","question":"","options":[],"allow_custom":false,"total_questions":0}`
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":%s}}]}`, strconvQuote(content))
	}))
	defer provider.Close()

	plansDir, ticketsDir := t.TempDir(), t.TempDir()
	s := store.New(plansDir, nil)
	m := conversation.NewManager(s, codebase.New(codebase.Options{}), &llm.Client{BaseURL: provider.URL, Model: "local", HTTPClient: provider.Client()}, codebaseDir)
	defer m.Close()
	m.SetExporter(export.New(export.Config{OutputDir: ticketsDir}))
	mux := http.NewServeMux()
	base := assetBase()
	if !fileExists(filepath.Join(base, "templates", "base.html")) {
		wd, _ := os.Getwd()
		base = filepath.Dir(filepath.Dir(wd))
	}
	tmpl, err := parseTemplates(base)
	if err != nil {
		t.Fatal(err)
	}
	handler.Register(mux, tmpl, s, websocket.NewHub("http://example"), websocket.NewAgentState(), base, m)
	server := httptest.NewServer(mux)
	defer server.Close()

	var started struct{ ID string }
	postJSON(t, server.Client(), server.URL+"/api/session/start", `{"prompt":"plan this codebase"}`, &started)
	if started.ID == "" {
		t.Fatal("session ID missing")
	}
	ch, done := m.Subscribe(started.ID)
	defer done()
	waitForState(t, ch, conversation.Grilling)
	var session model.SessionState
	postJSON(t, server.Client(), server.URL+"/api/session/"+started.ID+"/answer", `{"answer":"small"}`, &session)
	if session.ID != started.ID {
		t.Fatalf("answer session = %#v", session)
	}
	waitForState(t, ch, conversation.Reviewing)
	postJSON(t, server.Client(), server.URL+"/api/session/"+started.ID+"/approve", `{}`, &session)
	if session.State != string(conversation.Approved) || session.ExportStatus != "succeeded" {
		t.Fatalf("approval session = %#v", session)
	}
	data, err := os.ReadFile(filepath.Join(ticketsDir, started.ID+"-stage-1.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("Stage 1")) {
		t.Fatalf("ticket = %q", data)
	}
}

func TestRunWithSignalsShutsDownAllServices(t *testing.T) {
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(signal.String(), func(t *testing.T) {
			t.Setenv("PORT", "18080")
			t.Setenv("MAESTRO_PLANS_DIR", t.TempDir())
			wd, _ := os.Getwd()
			t.Setenv("MAESTRO_DIR", filepath.Dir(filepath.Dir(wd)))
			oldArgs, oldBrowser := os.Args, openBrowser
			defer func() { os.Args, openBrowser = oldArgs, oldBrowser }()
			os.Args = []string{"maestro", "serve"}
			stop := make(chan os.Signal, 1)
			opened := make(chan string, 1)
			openBrowser = func(url string) error { opened <- url; stop <- signal; return nil }
			if err := runWithSignals(stop); err != nil {
				t.Fatal(err)
			}
			select {
			case url := <-opened:
				if !strings.HasPrefix(url, "http://127.0.0.1:") {
					t.Fatal(url)
				}
			default:
				t.Fatal("browser was not opened")
			}
		})
	}
}

func postJSON(t *testing.T, client *http.Client, url, body string, out any) {
	t.Helper()
	resp, err := client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("POST %s: %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}
func waitForState(t *testing.T, ch <-chan conversation.Event, state conversation.State) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for {
		select {
		case event := <-ch:
			if event.State == state {
				return
			}
			if event.State == conversation.Failed {
				t.Fatalf("workflow failed: %s", event.Error)
			}
		case <-ctx.Done():
			t.Fatalf("did not reach %s", state)
		}
	}
}
func strconvQuote(s string) string { b, _ := json.Marshal(s); return string(b) }
func integrationPlanJSON() string {
	modules := []model.Module{{Type: "decision", Heading: "Decisions", Items: []model.Item{{Text: "Choose scope"}}}, {Type: "assumptions", Heading: "Assumptions", Items: []model.Item{{Text: "Local provider"}}}, {Type: "changes", Heading: "Changes", Items: []model.Item{{Text: "Update code"}}}, {Type: "notes", Heading: "Notes", Items: []model.Item{{Text: "Review output"}}}}
	for stage := 1; stage <= 2; stage++ {
		for _, typ := range []string{"criteria", "steps", "risks"} {
			modules = append(modules, model.Module{Type: typ, Heading: fmt.Sprintf("Stage %d: Implementation", stage), Items: []model.Item{{Text: "Complete " + typ}}})
		}
	}
	b, _ := json.Marshal(struct {
		Title   string         `json:"title"`
		Summary string         `json:"summary"`
		Modules []model.Module `json:"modules"`
	}{"Integration plan", "A local integration plan", modules})
	return string(b)
}
