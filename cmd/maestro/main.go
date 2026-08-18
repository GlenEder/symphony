package main

import (
	"context"
	"fmt"
	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/config"
	"github.com/gleneder/symphony/internal/conversation"
	"github.com/gleneder/symphony/internal/handler"
	"github.com/gleneder/symphony/internal/llm"
	"github.com/gleneder/symphony/internal/model"
	"github.com/gleneder/symphony/internal/store"
	"github.com/gleneder/symphony/internal/watcher"
	ws "github.com/gleneder/symphony/internal/websocket"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: maestro serve /path/to/codebase")
		os.Exit(2)
	}
	if len(os.Args) > 2 {
		_ = os.Setenv("CODEBASE_PATH", os.Args[2])
	}
	cfg, e := config.Load("")
	if e != nil {
		log.Fatal(e)
	}
	base := "."
	tmpl, e := parseTemplates(base)
	if e != nil {
		log.Fatal(e)
	}
	hub := ws.NewHub(fmt.Sprintf("http://localhost:%d", cfg.Port))
	state := ws.NewAgentState()
	var s *store.PlanStore
	s = store.New(cfg.MaestroPlansDir, func(id string) {
		if p := s.Get(id); p != nil {
			hub.Broadcast(id, model.ToFlatPlan(p, state.GetStatus(id)).JSON())
		}
	})
	if e = s.LoadAll(); e != nil {
		log.Fatal(e)
	}
	poll := watcher.Start(s, cfg.MaestroPlansDir, 500*time.Millisecond)
	defer poll.Close()
	mux := http.NewServeMux()
	cm := conversation.NewManager(s, codebase.New(codebase.Options{}), &llm.Client{BaseURL: cfg.LLMBaseURL, APIKey: cfg.LLMAPIKey, Model: cfg.LLMModel}, cfg.CodebasePath)
	cm.SetExporter(conversation.ExporterFunc(exportApprovedPlan))
	handler.Register(mux, tmpl, s, hub, state, base, cm)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/style.css", http.FileServer(http.Dir("static")))
	mux.Handle("/script.js", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux))
}

func exportApprovedPlan(ctx context.Context, p *model.Plan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("approved plan is required")
	}
	if p == nil || p.Session == nil || !store.ValidPlanID(p.Session.ID) {
		return fmt.Errorf("approved plan identity is required")
	}
	root, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, ".config", "symphony", "work_tickets")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, p.Session.ID+".md"), []byte(renderWorkTicket(p)), 0644)
}

func renderWorkTicket(p *model.Plan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\\n\\n**Status**: pending\\n\\n", p.Title)
	if p.Summary != "" {
		fmt.Fprintf(&b, "%s\\n\\n", p.Summary)
	}
	for _, module := range p.Modules {
		if module.Heading == "" {
			continue
		}
		fmt.Fprintf(&b, "## %s\\n\\n", module.Heading)
		for i, item := range module.Items {
			if module.Type == "criteria" {
				fmt.Fprintf(&b, "- [ ] %s\\n", item.Text)
			} else if module.Type == "steps" {
				fmt.Fprintf(&b, "%d. %s\\n", i+1, item.Text)
			} else {
				fmt.Fprintf(&b, "- %s\\n", item.Text)
			}
		}
		b.WriteString("\\n")
	}
	return b.String()
}

func parseTemplates(base string) (*template.Template, error) {
	t := template.New("").Funcs(template.FuncMap{"timeago": handler.Timeago, "lower": strings.ToLower, "add": func(a, b int) int { return a + b }})
	if _, e := t.ParseFiles(filepath.Join(base, "templates/base.html")); e != nil {
		return nil, e
	}
	e := filepath.WalkDir(filepath.Join(base, "templates/components"), func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && filepath.Ext(path) == ".html" {
			_, e = t.ParseFiles(path)
		}
		return e
	})
	return t, e
}
