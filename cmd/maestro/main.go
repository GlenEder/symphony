package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gleneder/symphony/internal/config"
	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{}

func newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintln(w, "<h1>Maestro</h1>")
	})
	return mux
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: maestro serve /path/to/codebase")
		os.Exit(2)
	}

	codebasePath := "."
	if len(os.Args) > 2 {
		codebasePath = os.Args[2]
	}
	if abs, err := filepath.Abs(codebasePath); err == nil {
		_ = os.Setenv("CODEBASE_PATH", abs)
	}

	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := os.MkdirAll(cfg.MaestroPlansDir, 0755); err != nil {
		log.Fatalf("create plans directory: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Maestro serving %s on http://localhost%s", cfg.CodebasePath, addr)
	log.Fatal(http.ListenAndServe(addr, newRouter()))
}
