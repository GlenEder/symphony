package watcher

import (
	"github.com/gleneder/symphony/internal/store"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FilePoller struct {
	stop chan struct{}
	done chan struct{}
}

func Start(s *store.PlanStore, dir string, interval time.Duration) *FilePoller {
	p := &FilePoller{make(chan struct{}), make(chan struct{})}
	go func() {
		defer close(p.done)
		known := map[string]time.Time{}
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				es, e := os.ReadDir(dir)
				if e != nil {
					continue
				}
				cur := map[string]bool{}
				for _, x := range es {
					if x.IsDir() || !strings.HasSuffix(x.Name(), ".json") {
						continue
					}
					cur[x.Name()] = true
					path := filepath.Join(dir, x.Name())
					i, e := os.Stat(path)
					if e != nil {
						continue
					}
					if t, ok := known[x.Name()]; ok && t.Equal(i.ModTime()) {
						continue
					}
					id := strings.TrimSuffix(x.Name(), ".json")
					if !store.ValidPlanID(id) {
						continue
					}
					known[x.Name()] = i.ModTime()
					if s.IsSelfWrite(x.Name(), i.ModTime()) {
						continue
					}
					if err := s.Reload(path); err != nil {
						// Keep the previous timestamp so a corrected external file is retried.
						delete(known, x.Name())
						continue
					}
					if s.OnChange != nil {
						s.OnChange(id)
					}
				}
				for n := range known {
					if !cur[n] {
						id := strings.TrimSuffix(n, ".json")
						s.Remove(id)
						if s.OnChange != nil {
							s.OnChange(id)
						}
						delete(known, n)
					}
				}
			case <-p.stop:
				return
			}
		}
	}()
	return p
}
func (p *FilePoller) Close() error { close(p.stop); <-p.done; return nil }
