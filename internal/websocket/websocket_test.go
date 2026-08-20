package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gleneder/symphony/internal/store"
)

func TestOriginAllowedRequiresExactLocalOrigin(t *testing.T) {
	cases := []struct {
		name   string
		origin string
		want   bool
	}{
		{"allowed", "http://localhost:8080", true},
		{"https", "https://localhost:8080", false},
		{"alternate loopback", "http://127.0.0.1:8080", false},
		{"wrong port", "http://localhost:9090", false},
		{"missing", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ws/plan/demo", nil)
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			if got := originAllowed(r, "http://localhost:8080"); got != tc.want {
				t.Fatalf("originAllowed(%q) = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}

func TestHandlerRejectsNonexistentPlanBeforeUpgrade(t *testing.T) {
	s := store.New(t.TempDir(), nil)
	h := NewHub("http://localhost:8080")
	r := httptest.NewRequest(http.MethodGet, "/ws/plan/missing", nil)
	r.SetPathValue("id", "missing")
	r.Header.Set("Origin", "http://localhost:8080")
	w := httptest.NewRecorder()
	h.Handler(s, NewAgentState()).ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
