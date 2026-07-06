package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

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
		sort.Slice(plans, func(i, j int) bool {
			return plans[i].ID < plans[j].ID
		})
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
		sort.Slice(plans, func(i, j int) bool {
			return plans[i].ID < plans[j].ID
		})
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(plans); err != nil {
			log.Printf("api list error: %v", err)
		}
	})

	// API: get plan, submit response, or set approval
	http.HandleFunc("/api/plan/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/plan/")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodPost {
			// POST /api/plan/{id}/response — submit user response
			if responseID, ok := strings.CutSuffix(id, "/response"); ok {
				var body struct {
					Text string `json:"text"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return
				}
				if err := store.SaveResponse(responseID, body.Text); err != nil {
					log.Printf("save response error: %v", err)
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				plan := store.Get(responseID)
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

			// POST /api/plan/{id}/approve — set approval status
			if approveID, ok := strings.CutSuffix(id, "/approve"); ok {
				var body struct {
					Approved bool `json:"approved"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return
				}
				if err := store.SetApproved(approveID, body.Approved); err != nil {
					log.Printf("set approval error: %v", err)
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				plan := store.Get(approveID)
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
