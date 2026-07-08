package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	plansDir := os.Getenv("PLANS_DIR")
	if plansDir == "" {
		plansDir = "plans"
	}

	// Ensure plans directory exists
	os.MkdirAll(plansDir, 0755)

	tmpl, err := parseTemplates()
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	// Plan store
	hub := NewHub()

	// Agent state tracking (must be before store for onChange closure)
	agentState := NewAgentState()

	var store *PlanStore
	store = NewPlanStore(plansDir, func(id string) {
		if plan := store.Get(id); plan != nil {
			hub.Broadcast(id, toFlatPlan(plan, agentState.GetStatus(id)).JSON())
		}
	})
	if err := store.LoadAll(); err != nil {
		log.Fatalf("load plans: %v", err)
	}

	// Background goroutine: clean stale agent states every 5s
	go func() {
		for {
			time.Sleep(5 * time.Second)
			agentState.GC()
		}
	}()

	// File watcher for live updates
	watcher, err := StartWatcher(store, plansDir)
	if err != nil {
		log.Printf("file watcher: %v (live updates disabled)", err)
	} else {
		defer watcher.Close()
	}

	// Static files
	fsys := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fsys))
	http.Handle("/style.css", fsys)
	http.Handle("/script.js", fsys)

	// Register all routes
	registerRoutes(tmpl, store, hub, agentState)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("Plans directory: %s", plansDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// parseTemplates parses only base.html and component templates
// (not page templates) to create a base template set for cloning.
func parseTemplates() (*template.Template, error) {
	t := template.New("").Funcs(template.FuncMap{
		"lower":   strings.ToLower,
		"timeago": timeago,
		"add":     func(a, b int) int { return a + b },
	})
	// Parse base layout
	if _, err := t.ParseFiles("templates/base.html"); err != nil {
		return nil, err
	}
	// Parse all component templates
	err := filepath.WalkDir("templates/components", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		_, err = t.ParseFiles(path)
		return err
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

// renderPage clones the base template set, parses the page-specific template,
// and renders it using the "base" layout.
func renderPage(w http.ResponseWriter, baseTmpl *template.Template, pagePath string, data any) {
	t, err := baseTmpl.Clone()
	if err != nil {
		log.Printf("clone template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := t.ParseFiles(pagePath); err != nil {
		log.Printf("parse page %s: %v", pagePath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("render error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
