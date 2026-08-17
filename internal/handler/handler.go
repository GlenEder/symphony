package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gleneder/symphony/internal/model"
	"github.com/gleneder/symphony/internal/store"
	ws "github.com/gleneder/symphony/internal/websocket"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type ListData struct {
	Title string
	Year  int
	Plans []store.PlanSummary
}
type PageData struct {
	Title  string
	Year   int
	Plan   *model.Plan
	PlanID string
}

func Register(m *http.ServeMux, t *template.Template, s *store.PlanStore, h *ws.Hub, a *ws.AgentState, base string) {
	jsonOut := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	m.HandleFunc("GET /plans", func(w http.ResponseWriter, r *http.Request) {
		render(w, t, filepath.Join(base, "templates/plans.html"), ListData{"Plans", time.Now().Year(), s.List()})
	})
	m.HandleFunc("GET /plan/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/plan/")
		if !store.ValidPlanID(id) {
			http.NotFound(w, r)
			return
		}
		p := s.Get(id)
		if p == nil {
			http.NotFound(w, r)
			return
		}
		render(w, t, filepath.Join(base, "templates/plan.html"), PageData{p.Title, time.Now().Year(), p, id})
	})
	m.HandleFunc("GET /api/plans", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, s.List()) })
	m.HandleFunc("POST /api/plans", func(w http.ResponseWriter, r *http.Request) {
		var b struct{ ID, Title, Summary string }
		if json.NewDecoder(r.Body).Decode(&b) != nil || b.ID == "" || b.Title == "" {
			http.Error(w, "id and title are required", 400)
			return
		}
		p, e := s.CreatePlan(b.ID, b.Title, b.Summary)
		if e != nil {
			http.Error(w, e.Error(), 409)
			return
		}
		w.WriteHeader(201)
		jsonOut(w, model.ToFlatPlan(p, a.GetStatus(b.ID)))
	})
	m.HandleFunc("/api/plan/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/plan/")
		parts := strings.Split(path, "/")
		id := parts[0]
		if !store.ValidPlanID(id) {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			if len(parts) == 1 {
				if e := s.DeletePlan(id); e != nil {
					http.Error(w, e.Error(), 404)
					return
				}
				jsonOut(w, map[string]string{"status": "ok"})
				return
			}
			if len(parts) == 3 && parts[1] == "messages" {
				if e := s.DeleteMessage(id, parts[2]); e != nil {
					http.Error(w, e.Error(), 404)
					return
				}
				jsonOut(w, model.ToFlatPlan(s.Get(id), a.GetStatus(id)))
				return
			}
		}
		if r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "messages" {
			var b struct {
				Role, Text, ItemRef string
				Prompt              *model.Prompt
			}
			if json.NewDecoder(r.Body).Decode(&b) != nil || b.Text == "" {
				http.Error(w, "invalid message", 400)
				return
			}
			msg, e := s.AddMessage(id, b.Role, b.Text, b.ItemRef, b.Prompt)
			if e != nil {
				http.Error(w, e.Error(), 404)
				return
			}
			if b.Role == "human" {
				a.SetThinking(id)
			} else {
				a.SetListening(id)
			}
			jsonOut(w, msg)
			return
		}
		if r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "state" {
			var b struct{ State string }
			if json.NewDecoder(r.Body).Decode(&b) != nil || s.SetState(id, b.State) != nil {
				http.Error(w, "invalid state", 400)
				return
			}
			jsonOut(w, model.ToFlatPlan(s.Get(id), a.GetStatus(id)))
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var p model.Plan
			if json.NewDecoder(r.Body).Decode(&p) != nil || s.Upsert(id, &p) != nil {
				http.Error(w, "invalid plan", 400)
				return
			}
			jsonOut(w, model.ToFlatPlan(s.Get(id), a.GetStatus(id)))
			return
		}
		if r.Method == http.MethodPatch {
			var p model.Plan
			if json.NewDecoder(r.Body).Decode(&p) != nil || s.Patch(id, &p) != nil {
				http.Error(w, "invalid patch", 400)
				return
			}
			jsonOut(w, model.ToFlatPlan(s.Get(id), a.GetStatus(id)))
			return
		}
		p := s.Get(id)
		if p == nil {
			http.NotFound(w, r)
			return
		}
		jsonOut(w, model.ToFlatPlan(p, a.GetStatus(id)))
	})
	m.HandleFunc("/ws/plan/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/ws/plan/")
		if !store.ValidPlanID(id) || s.Get(id) == nil {
			http.NotFound(w, r)
			return
		}
		r.SetPathValue("id", id)
		h.Handler(s, a)(w, r)
	})
	m.HandleFunc("POST /api/agent/{id}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		a.Heartbeat(r.PathValue("id"))
		jsonOut(w, map[string]string{"status": "ok"})
	})
	m.HandleFunc("GET /api/agent/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		jsonOut(w, map[string]string{"status": a.GetStatus(r.PathValue("id"))})
	})
	m.HandleFunc("POST /api/agent/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		var b struct{ Status string }
		_ = json.NewDecoder(r.Body).Decode(&b)
		if b.Status == ws.StatusOffline {
			a.SetOffline(r.PathValue("id"))
		}
		jsonOut(w, map[string]string{"status": "ok"})
	})
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/plans", 302) })
}
func render(w http.ResponseWriter, t *template.Template, page string, data any) {
	x, e := t.Clone()
	if e == nil {
		_, e = x.ParseFiles(page)
	}
	if e != nil {
		http.Error(w, "template error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = x.ExecuteTemplate(w, "base", data)
}
func Timeago(ts string) string {
	t, e := time.Parse(time.RFC3339, ts)
	if e != nil {
		return ts
	}
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	}
	return t.Format("Jan 2, 2006")
}
