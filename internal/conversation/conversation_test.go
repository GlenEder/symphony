package conversation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/store"
)

type testResearcher struct{ err error }

func (r testResearcher) Summary(string, codebase.SummaryOptions) (string, error) {
	return "context", r.err
}

func TestStartUsesIndependentContextAndPersistsFailure(t *testing.T) {
	m := NewManager(store.New(t.TempDir(), nil), testResearcher{err: errors.New("research failed")}, nil, ".")
	requestCtx, cancel := context.WithCancel(context.Background())
	s, err := m.Start(requestCtx, "make a plan")
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := m.Session(s.ID); got != nil && got.State == Failed {
			if got.Error != "research failed" {
				t.Fatalf("error = %q", got.Error)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("session did not reach failed state")
}

func TestSubscribeSendsCurrentSnapshot(t *testing.T) {
	m := NewManager(store.New(t.TempDir(), nil), testResearcher{}, nil, ".")
	s, err := m.Start(context.Background(), "prompt")
	if err != nil {
		t.Fatal(err)
	}
	ch, done := m.Subscribe(s.ID)
	defer done()
	select {
	case event := <-ch:
		if event.Session == nil || event.Session.ID != s.ID {
			t.Fatalf("snapshot = %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("no snapshot")
	}
}
