package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// StartWatcher watches the plans directory for .toon file changes and reloads them.
func StartWatcher(store *PlanStore, dir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch the directory itself (not individual files) — fsnotify
	// on dirs reports create/write events for files inside.
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Only care about .toon files
				if !strings.HasSuffix(event.Name, ".toon") {
					continue
				}
				// On write or create, reload the file
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					log.Printf("plan file changed: %s", event.Name)
					store.loadFile(event.Name)
					id := strings.TrimSuffix(filepath.Base(event.Name), ".toon")
					if store.onChange != nil {
						store.onChange(id)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()

	return watcher, nil
}
