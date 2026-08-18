package main

import (
	"bytes"
	"context"
	"log"
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

func TestLoggingMiddleware(t *testing.T) {
	called := false
	h := logging(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { called = true; w.WriteHeader(http.StatusNoContent) }))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if !called || w.Code != http.StatusNoContent {
		t.Fatalf("middleware did not serve request: called=%v status=%d", called, w.Code)
	}
}

func TestBrowserCommandSelection(t *testing.T) {
	for _, tc := range []struct {
		os, command string
		args        []string
	}{
		{"darwin", "open", []string{"http://127.0.0.1:8080/"}},
		{"linux", "xdg-open", []string{"http://127.0.0.1:8080/"}},
		{"windows", "rundll32", []string{"url.dll,FileProtocolHandler", "http://127.0.0.1:8080/"}},
	} {
		command, args := browserCommand(tc.os, tc.args[len(tc.args)-1])
		if command != tc.command || len(args) != len(tc.args) {
			t.Fatalf("%s: command=%q args=%v", tc.os, command, args)
		}
		for i := range args {
			if args[i] != tc.args[i] {
				t.Fatalf("%s args=%v", tc.os, args)
			}
		}
	}
}

func TestBrowserFallbackLogsManualURL(t *testing.T) {
	var output bytes.Buffer
	oldWriter, oldBrowser := log.Writer(), openBrowser
	defer func() { log.SetOutput(oldWriter); openBrowser = oldBrowser }()
	log.SetOutput(&output)
	openBrowser = func(string) error { return context.DeadlineExceeded }
	url := "http://127.0.0.1:1234/"
	logBrowserFallback(url, openBrowser(url))
	if !bytes.Contains(output.Bytes(), []byte(url)) || !bytes.Contains(output.Bytes(), []byte("could not open browser")) {
		t.Fatalf("fallback log = %q", output.String())
	}
}

func TestLoggingCapturesRequestFields(t *testing.T) {
	var output bytes.Buffer
	old := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(old)
	h := logging(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) }))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/test?x=1", nil))
	for _, want := range []string{"method=POST", "path=/api/test?x=1", "status=201"} {
		if !bytes.Contains(output.Bytes(), []byte(want)) {
			t.Fatalf("log %q missing %s", output.String(), want)
		}
	}
}
