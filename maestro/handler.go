package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// timeago returns a human-readable relative time string for an RFC3339 timestamp.
func timeago(ts string) string {
	if ts == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		n := int(d.Minutes())
		if n == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", n)
	case d < 24*time.Hour:
		n := int(d.Hours())
		if n == 1 {
			return "1 hr ago"
		}
		return fmt.Sprintf("%d hr ago", n)
	case d < 30*24*time.Hour:
		n := int(d.Hours() / 24)
		if n == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", n)
	default:
		return t.Format("Jan 2, 2006")
	}
}

// PlanListData is passed to the plan list template.
type PlanListData struct {
	Title string
	Year  int
	Plans []PlanSummary
}

// PlanPageData is passed to the plan detail template.
type PlanPageData struct {
	Title  string
	Year   int
	Plan   *Plan
	PlanID string
}

func registerRoutes(baseTmpl *template.Template, store *PlanStore, hub *Hub) {
	// Plan listing
	http.HandleFunc("/plans", func(w http.ResponseWriter, r *http.Request) {
		plans := store.List()
		data := PlanListData{
			Title: "Plans",
			Year:  2026,
			Plans: plans,
		}
		renderPage(w, baseTmpl, "templates/plans.html", data)
	})

	// Plan detail page
	http.HandleFunc("/plan/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/plan/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		plan := store.Get(id)
		if plan == nil {
			http.NotFound(w, r)
			return
		}
		data := PlanPageData{
			Title:  plan.Title,
			Year:   2026,
			Plan:   plan,
			PlanID: id,
		}
		renderPage(w, baseTmpl, "templates/plan.html", data)
	})

	// API: list plans
	http.HandleFunc("/api/plans", func(w http.ResponseWriter, r *http.Request) {
		plans := store.List()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(plans); err != nil {
			log.Printf("api list error: %v", err)
		}
	})

	// API: plan operations (get, add message, set state)
	http.HandleFunc("/api/plan/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/plan/")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodDelete {
			// DELETE /api/plan/{id}/messages/{msgId} — delete a message from the thread
			parts := strings.SplitN(id, "/", 3)
			if len(parts) == 3 && parts[1] == "messages" {
				if err := store.DeleteMessage(parts[0], parts[2]); err != nil {
					log.Printf("delete message error: %v", err)
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				plan := store.Get(parts[0])
				if plan == nil {
					http.NotFound(w, r)
					return
				}
				fp := toFlatPlan(plan)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(fp.JSON())
				return
			}
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodPost {
			// POST /api/plan/{id}/messages — add a message to the conversation thread
			if msgID, ok := strings.CutSuffix(id, "/messages"); ok {
				var body struct {
					Role    string `json:"role"`
					Text    string `json:"text"`
					ItemRef string `json:"item_ref,omitempty"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return
				}
				if body.Role != "agent" && body.Role != "human" {
					http.Error(w, "role must be 'agent' or 'human'", http.StatusBadRequest)
					return
				}
				if body.Text == "" {
					http.Error(w, "text is required", http.StatusBadRequest)
					return
				}
				msg, err := store.AddMessage(msgID, body.Role, body.Text, body.ItemRef)
				if err != nil {
					log.Printf("add message error: %v", err)
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(msg)
				return
			}

			// POST /api/plan/{id}/state — set plan state (only human can approve)
			if stateID, ok := strings.CutSuffix(id, "/state"); ok {
				var body struct {
					State string `json:"state"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return
				}
				if err := store.SetState(stateID, body.State); err != nil {
					log.Printf("set state error: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				plan := store.Get(stateID)
				if plan == nil {
					http.NotFound(w, r)
					return
				}
				fp := toFlatPlan(plan)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(fp.JSON())
				return
			}

			http.NotFound(w, r)
			return
		}

		// GET /api/plan/{id} — return the full plan as JSON
		plan := store.Get(id)
		if plan == nil {
			http.NotFound(w, r)
			return
		}
		fp := toFlatPlan(plan)
		w.Header().Set("Content-Type", "application/json")
		w.Write(fp.JSON())
	})

	// WebSocket for plan updates
	http.HandleFunc("/ws/plan/", hub.handleWS(store))

	// Redirect root to /plans
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/plans", http.StatusFound)
	})
}
