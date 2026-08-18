package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/config"
	"github.com/gleneder/symphony/internal/conversation"
	"github.com/gleneder/symphony/internal/export"
	"github.com/gleneder/symphony/internal/handler"
	"github.com/gleneder/symphony/internal/llm"
	"github.com/gleneder/symphony/internal/model"
	"github.com/gleneder/symphony/internal/store"
	"github.com/gleneder/symphony/internal/watcher"
	ws "github.com/gleneder/symphony/internal/websocket"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var browserCommand = func(goos, url string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{url}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		return "xdg-open", []string{url}
	}
}

var openBrowser = func(url string) error {
	command, args := browserCommand(runtime.GOOS, url)
	return exec.Command(command, args...).Start()
}

func logBrowserFallback(url string, err error) {
	log.Printf("could not open browser: %v; open %s manually", err, url)
}
func newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return h.Hijack()
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.RequestURI(), sw.status, time.Since(started).Round(time.Millisecond))
	})
}

func assetBase() string {
	if base := os.Getenv("MAESTRO_DIR"); base != "" {
		return base
	}
	if executable, err := os.Executable(); err == nil {
		if base := filepath.Dir(executable); fileExists(filepath.Join(base, "templates", "base.html")) {
			return base
		}
	}
	return "."
}

func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }

func run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	return runWithSignals(stop)
}

func runWithSignals(stop <-chan os.Signal) error {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		return fmt.Errorf("usage: maestro serve /path/to/codebase")
	}
	if len(os.Args) > 2 {
		if err := os.Setenv("CODEBASE_PATH", os.Args[2]); err != nil {
			return err
		}
	}
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	base := assetBase()
	tmpl, err := parseTemplates(base)
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(cfg.Address, fmt.Sprintf("%d", cfg.Port))
	origin := fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Address, fmt.Sprintf("%d", cfg.Port)))
	publicURL := origin + "/"
	hub := ws.NewHub(origin)
	state := ws.NewAgentState()
	var s *store.PlanStore
	s = store.New(cfg.MaestroPlansDir, func(id string) {
		if p := s.Get(id); p != nil {
			hub.Broadcast(id, model.ToFlatPlan(p, state.GetStatus(id)).JSON())
		}
	})
	if err = s.LoadAll(); err != nil {
		return err
	}
	poll := watcher.Start(s, cfg.MaestroPlansDir, 500*time.Millisecond)
	defer poll.Close()
	cm := conversation.NewManager(s, codebase.New(codebase.Options{}), &llm.Client{BaseURL: cfg.LLMBaseURL, APIKey: cfg.LLMAPIKey, Model: cfg.LLMModel}, cfg.CodebasePath)
	defer cm.Close()
	cm.SetExporter(export.New(export.Config{TraceabilityURL: cfg.TraceabilityURL}))
	mux := http.NewServeMux()
	handler.Register(mux, tmpl, s, hub, state, base, cm)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(base, "static")))))
	mux.Handle("/style.css", http.FileServer(http.Dir(filepath.Join(base, "static"))))
	mux.Handle("/script.js", http.FileServer(http.Dir(filepath.Join(base, "static"))))
	server := &http.Server{Addr: addr, Handler: logging(mux)}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	origin = fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Address, fmt.Sprintf("%d", actualPort)))
	publicURL = origin + "/"
	hub.SetExpectedOrigin(origin)
	go func() {
		log.Printf("Maestro listening at %s", publicURL)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()
	if err := openBrowser(publicURL); err != nil {
		logBrowserFallback(publicURL, err)
	}
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(2)
	}
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
