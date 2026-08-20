package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gleneder/symphony/internal/store"
)

func TestWatcherIgnoresInvalidFilenames(t *testing.T) {
	dir := t.TempDir()
	changed := make(chan string, 1)
	s := store.New(dir, func(id string) { changed <- id })
	path := filepath.Join(dir, "../bad.json")
	if err := os.WriteFile(path, []byte(`{"title":"Invalid","modules":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	invalid := filepath.Join(dir, "bad..json")
	if err := os.Rename(path, invalid); err != nil {
		t.Fatal(err)
	}

	poller := Start(s, dir, time.Millisecond)
	defer poller.Close()
	time.Sleep(10 * time.Millisecond)
	if s.Get("bad.") != nil {
		t.Fatal("invalid plan entered store")
	}
	if err := os.Remove(invalid); err != nil {
		t.Fatal(err)
	}
	select {
	case id := <-changed:
		t.Fatalf("invalid plan triggered change for %q", id)
	case <-time.After(10 * time.Millisecond):
	}
}

func TestWatcherStaleReloadCannotResurrectDeletedPlan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gone.json")
	if err := os.WriteFile(path, []byte(`{"title":"Gone","modules":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	s := store.New(dir, nil)
	if err := s.LoadAll(); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = s.Reload(path) // stale watcher snapshot
	}()
	go func() {
		defer wg.Done()
		_ = s.DeletePlan("gone")
	}()
	wg.Wait()
	if s.Get("gone") != nil {
		t.Fatal("deleted plan was resurrected in memory")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deleted plan file exists: %v", err)
	}
}
