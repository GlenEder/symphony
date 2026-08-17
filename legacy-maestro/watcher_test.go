package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTempPlan(t *testing.T, dir, id, content string) string {
	t.Helper()
	path := filepath.Join(dir, id+".json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

const samplePlan = `{
  "title": "Test Plan",
  "summary": "A test plan",
  "modules": [
    {
      "heading": "Criteria",
      "items": [
        {
          "text": "it works"
        }
      ],
      "type": "criteria"
    }
  ]
}`

func TestStore_IsSelfWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	// Load a plan
	writeTempPlan(t, dir, "test", samplePlan)
	store.loadFile(filepath.Join(dir, "test.json"))

	// Initially, no write tracked — IsSelfWrite should return false
	beforeMtime := time.Now()
	if store.IsSelfWrite("test.json", beforeMtime) {
		t.Error("expected false before any persistPlan")
	}

	// After persistPlan, the mtime should be tracked
	if err := store.persistPlan("test"); err != nil {
		t.Fatal(err)
	}

	// Should match the recorded mtime via os.Stat
	fi, err := os.Stat(filepath.Join(dir, "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !store.IsSelfWrite("test.json", fi.ModTime()) {
		t.Error("expected true after persistPlan")
	}

	// A different mtime should return false
	if store.IsSelfWrite("test.json", fi.ModTime().Add(-time.Hour)) {
		t.Error("expected false for different mtime")
	}

	// Unknown file should return false
	if store.IsSelfWrite("nonexistent.json", time.Now()) {
		t.Error("expected false for unknown file")
	}
}

func TestStore_IsSelfWrite_FilenameIdMapping(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	writeTempPlan(t, dir, "demo-plan", samplePlan)
	store.loadFile(filepath.Join(dir, "demo-plan.json"))

	if err := store.persistPlan("demo-plan"); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(filepath.Join(dir, "demo-plan.json"))
	if err != nil {
		t.Fatal(err)
	}

	// TestStore_IsSelfWrite_FilenameIdMapping
	if !store.IsSelfWrite("demo-plan.json", fi.ModTime()) {
		t.Error("IsSelfWrite should match filename with .json extension")
	}
}

func TestPoller_DetectsNewFile(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	// Start poller with a short interval
	poller := StartWatcher(store, dir, 50*time.Millisecond)
	defer poller.Close()

	// No plans initially
	if len(store.List()) != 0 {
		t.Fatal("expected 0 plans initially")
	}

	// Write a plan file
	writeTempPlan(t, dir, "new-plan", samplePlan)

	// Wait for poller to detect it
	deadline := time.After(2 * time.Second)
	for {
		if len(store.List()) > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timeout waiting for poller to detect new file")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	plan := store.Get("new-plan")
	if plan == nil {
		t.Fatal("expected plan to be loaded")
	}
	if plan.Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got %q", plan.Title)
	}
}

func TestPoller_DetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	// Write initial plan and load it
	path := writeTempPlan(t, dir, "change-test", samplePlan)
	store.LoadAll()

	// Track onChange calls (must set before StartWatcher to avoid data race)
	changed := make(chan string, 1)
	store.onChange = func(id string) {
		select {
		case changed <- id:
		default:
		}
	}

	poller := StartWatcher(store, dir, 50*time.Millisecond)
	defer poller.Close()

	// Drain initial detection on first poll tick (redundant load of existing file)
	time.Sleep(150 * time.Millisecond)
	select {
	case <-changed:
	default:
	}

	// Modify the file
	newContent := strings.Replace(samplePlan, "Test Plan", "Updated Plan", 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for onChange
	select {
	case id := <-changed:
		if id != "change-test" {
			t.Errorf("expected change-test, got %q", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for poller to detect file change")
	}

	plan := store.Get("change-test")
	if plan == nil {
		t.Fatal("expected plan to exist")
	}
	if plan.Title != "Updated Plan" {
		t.Errorf("expected title 'Updated Plan', got %q", plan.Title)
	}
}

func TestPoller_DetectsFileDeletion(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	// Write and load a plan
	path := writeTempPlan(t, dir, "delete-me", samplePlan)
	store.LoadAll()

	if store.Get("delete-me") == nil {
		t.Fatal("expected plan to be loaded")
	}

	// Set onChange before starting the poller to avoid data race
	changed := make(chan string, 10)
	store.onChange = func(id string) {
		changed <- id
	}

	poller := StartWatcher(store, dir, 50*time.Millisecond)
	defer poller.Close()

	// Drain initial detection from first poll tick
	time.Sleep(150 * time.Millisecond)
	select {
	case <-changed:
	default:
	}

	// Delete the file
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	// Wait for the plan to be removed
	select {
	case id := <-changed:
		if id != "delete-me" {
			t.Errorf("expected delete-me, got %q", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for poller to detect deletion")
	}

	if store.Get("delete-me") != nil {
		t.Error("expected plan to be removed from store")
	}
}

func TestPoller_SkipsSelfWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	writeTempPlan(t, dir, "self-write", samplePlan)
	store.LoadAll()

	onChangeCount := 0
	store.onChange = func(id string) {
		onChangeCount++
	}

	poller := StartWatcher(store, dir, 50*time.Millisecond)
	defer poller.Close()

	// Perform a server-initiated write (as AddMessage does internally)
	if err := store.persistPlan("self-write"); err != nil {
		t.Fatal(err)
	}

	// onChange is called once by persistPlan's explicit call
	initialCount := onChangeCount

	// Wait for multiple poll cycles
	time.Sleep(300 * time.Millisecond)

	// onChange should NOT have been called again by the poller
	if onChangeCount != initialCount {
		t.Errorf("poller triggered %d extra onChange calls after self-write (expected 0)", onChangeCount-initialCount)
	}
}

func TestPoller_ConfigurableInterval(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	// Use a longer interval to verify it's respected
	poller := StartWatcher(store, dir, 1*time.Second)
	start := time.Now()

	writeTempPlan(t, dir, "slow-plan", samplePlan)

	// Wait for detection (should take ~1s, not <100ms)
	deadline := time.After(3 * time.Second)
	for {
		if store.Get("slow-plan") != nil {
			elapsed := time.Since(start)
			if elapsed < 500*time.Millisecond {
				t.Errorf("poller detected too fast for 1s interval: %v", elapsed)
			}
			break
		}
		select {
		case <-deadline:
			poller.Close()
			t.Fatal("timeout waiting for poller")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	poller.Close()
}

func TestStore_PersistPlanRecordsMtime(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	writeTempPlan(t, dir, "mtime-test", samplePlan)
	store.LoadAll()

	// persist the plan
	if err := store.persistPlan("mtime-test"); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(filepath.Join(dir, "mtime-test.json"))
	if err != nil {
		t.Fatal(err)
	}

	// The recorded mtime should match the actual file mtime
	if !store.IsSelfWrite("mtime-test.json", fi.ModTime()) {
		t.Error("persistPlan should record the file mtime")
	}
}

func TestStore_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store := NewPlanStore(dir, nil)

	if err := store.LoadAll(); err != nil {
		t.Fatal(err)
	}
	if len(store.List()) != 0 {
		t.Error("expected empty list for empty dir")
	}
}

func TestStore_NonExistentDirIsCreated(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent", "subdir")
	store := NewPlanStore(dir, nil)

	if err := store.LoadAll(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("LoadAll should create the directory")
	}
}
