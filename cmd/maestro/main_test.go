package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	newRouter().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", res.Code, http.StatusOK)
	}
	if res.Body.String() != "ok" {
		t.Fatalf("health body = %q, want %q", res.Body.String(), "ok")
	}
}
