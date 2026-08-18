package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatRequestAndAuth(t *testing.T) {
	var got completionRequest
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing auth")
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer s.Close()
	out, err := (&Client{BaseURL: s.URL, APIKey: "secret", Model: "test"}).Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil || out.ID != "x" || got.Model != "test" || len(got.Messages) != 1 {
		t.Fatalf("unexpected result: %+v %v", out, err)
	}
}

func TestStream(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"id\":\"x\",\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\ndata: [DONE]\n"))
	}))
	defer s.Close()
	var text strings.Builder
	err := (&Client{BaseURL: s.URL, Model: "test"}).Stream(context.Background(), nil, func(c StreamChunk) error { text.WriteString(c.Choices[0].Delta.Content); return nil })
	if err != nil || text.String() != "hi" {
		t.Fatalf("stream: %q %v", text.String(), err)
	}
}

func TestErrorRedactsKey(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401); _, _ = w.Write([]byte("secret")) }))
	defer s.Close()
	_, err := (&Client{BaseURL: s.URL, APIKey: "secret", Model: "x"}).Chat(context.Background(), nil)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("leaked error: %v", err)
	}
}
