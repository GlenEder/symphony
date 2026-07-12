package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FilePoller watches a directory for .toon file changes by polling modtimes.
// Works reliably across all filesystem types including Docker volumes, NFS, and FUSE.
type FilePoller struct {
	stop chan struct{}
}

// StartWatcher starts polling for .toon file changes in dir at the given interval.
// Returns a FilePoller that can be stopped via Close().
func StartWatcher(store *PlanStore, dir string, interval time.Duration) *FilePoller {
	pw := &FilePoller{stop: make(chan struct{})}

	go func() {
		knownMTimes := make(map[string]time.Time)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				entries, err := os.ReadDir(dir)
				if err != nil {
					continue
				}

				// Build set of current files for deletion detection
				current := make(map[string]bool, len(entries))

				for _, e := range entries {
					if e.IsDir() || !strings.HasSuffix(e.Name(), ".toon") {
						continue
					}
					current[e.Name()] = true

					path := filepath.Join(dir, e.Name())
					fi, err := os.Stat(path)
					if err != nil {
						continue
					}
					mtime := fi.ModTime()

					last, seen := knownMTimes[e.Name()]
					if seen && mtime.Equal(last) {
						continue // no change
					}
					knownMTimes[e.Name()] = mtime

					// Skip if the server itself just wrote this file
					if store.IsSelfWrite(e.Name(), mtime) {
						continue
					}

					id := strings.TrimSuffix(e.Name(), ".toon")
					store.loadFile(path)
					if store.onChange != nil {
						store.onChange(id)
					}
				}

				// Detect deleted files
				for name := range knownMTimes {
					if !current[name] {
						id := strings.TrimSuffix(name, ".toon")
						store.RemovePlan(id)
						if store.onChange != nil {
							store.onChange(id)
						}
						delete(knownMTimes, name)
					}
				}

			case <-pw.stop:
				return
			}
		}
	}()

	return pw
}

// Close stops the polling goroutine.
func (pw *FilePoller) Close() error {
	close(pw.stop)
	return nil
}
