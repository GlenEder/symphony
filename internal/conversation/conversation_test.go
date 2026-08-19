package conversation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gleneder/symphony/internal/codebase"
	"github.com/gleneder/symphony/internal/llm"
	"github.com/gleneder/symphony/internal/model"
	"github.com/gleneder/symphony/internal/store"
)

type testResearcher struct{ err error }

func (r testResearcher) Summary(string, codebase.SummaryOptions) (string, error) {
	return "context", r.err
}

type testLLM struct {
	mu        sync.Mutex
	responses []string
	calls     int
}

func (l *testLLM) Chat(context.Context, []llm.Message) (llm.Response, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.calls >= len(l.responses) {
		return llm.Response{}, fmt.Errorf("unexpected LLM call %d", l.calls)
	}
	content := l.responses[l.calls]
	l.calls++
	return llm.Response{Choices: []llm.Choice{{Message: llm.Message{Content: content}}}}, nil
}

func waitForState(t *testing.T, m *Manager, id string, want State) *Session {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := m.Session(id); got != nil && got.State == want {
			return got
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("session did not reach %s: %#v", want, m.Session(id))
	return nil
}

func waitForQuestion(t *testing.T, m *Manager, id string) *Session {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := m.Session(id); got != nil && got.State == Grilling && got.Question != nil && !got.Question.Answered {
			return got
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("session did not expose a grilling question: %#v", m.Session(id))
	return nil
}
func answerWhenReady(t *testing.T, m *Manager, id, answer string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if err := m.Answer(id, answer); err == nil {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("question did not become answerable")
}
func TestRetryAfterGrillingFailureResetsProgress(t *testing.T) {
	question := `{"done":false,"question_key":"same","question":"Choose?","options":["yes"],"total_questions":1}`
	plan := `{"title":"Plan","summary":"Summary","modules":[{"type":"criteria","heading":"Stage 1: Criteria","items":[{"text":"criterion"}]},{"type":"steps","heading":"Stage 1: Steps","items":[{"text":"step"}]},{"type":"risks","heading":"Stage 1: Risks","items":[{"text":"risk"}]},{"type":"criteria","heading":"Stage 2: Criteria","items":[{"text":"criterion"}]},{"type":"steps","heading":"Stage 2: Steps","items":[{"text":"step"}]},{"type":"risks","heading":"Stage 2: Risks","items":[{"text":"risk"}]},{"type":"decision","heading":"Decisions","items":[{"text":"decision"}]},{"type":"assumptions","heading":"Assumptions","items":[{"text":"assumption"}]},{"type":"changes","heading":"Changes","items":[{"text":"change"}]},{"type":"notes","heading":"Notes","items":[{"text":"note"}]}]}`
	llmStub := &testLLM{responses: []string{question, `not json`, question, `{"done":true}`, plan}}
	m := NewManager(store.New(t.TempDir(), nil), testResearcher{}, llmStub, ".")
	defer m.Close()
	s, err := m.Start(context.Background(), "make a plan")
	if err != nil {
		t.Fatal(err)
	}
	waitForQuestion(t, m, s.ID)
	answerWhenReady(t, m, s.ID, "yes")
	waitForState(t, m, s.ID, Failed)
	failed := m.Session(s.ID)
	if failed.QuestionCount != 1 || failed.Question == nil || len(failed.QuestionKeys) != 1 {
		t.Fatalf("failed progress = %#v", failed)
	}
	if err := m.Retry(s.ID); err != nil {
		t.Fatal(err)
	}
	waitForQuestion(t, m, s.ID)
	retried := m.Session(s.ID)
	if retried.QuestionCount != 1 || retried.Question == nil || retried.Question.Current != 1 {
		t.Fatalf("retried question = %#v", retried)
	}
	if stored := m.store.Get(s.ID); stored.Session == nil || stored.Session.QuestionCount != 1 || !stored.Session.QuestionProgressPersisted {
		t.Fatalf("persisted retry progress = %#v", stored.Session)
	}
	m.Close()
	m = NewManager(m.store, testResearcher{}, llmStub, ".")
	defer m.Close()
	retried = waitForQuestion(t, m, s.ID)
	if retried.QuestionCount != 1 || len(retried.QuestionKeys) != 1 || retried.QuestionKeys[0] != "same" {
		t.Fatalf("restarted retry progress = %#v", retried)
	}
	answerWhenReady(t, m, s.ID, "yes")
	waitForState(t, m, s.ID, Reviewing)
	stored := m.store.Get(s.ID)
	if len(stored.Messages) < 2 {
		t.Fatalf("transcript = %#v", stored.Messages)
	}
}

func TestLegacySessionReconstructsQuestionProgress(t *testing.T) {
	s := store.New(t.TempDir(), nil)
	id := "legacy-session"
	if _, err := s.CreatePlan(id, "title", "summary"); err != nil {
		t.Fatal(err)
	}
	prompt := &model.Prompt{QuestionKey: "legacy-key", Question: "Choose?", Options: []string{"yes"}, Current: 1}
	if _, err := s.AddMessage(id, "agent", prompt.Question, prompt.QuestionKey, prompt); err != nil {
		t.Fatal(err)
	}
	p := s.Get(id)
	p.Session = &model.SessionState{ID: id, State: string(Failed)}
	if err := s.Upsert(id, p); err != nil {
		t.Fatal(err)
	}
	m := NewManager(s, testResearcher{}, nil, ".")
	defer m.Close()
	got := m.Session(id)
	if got == nil || got.QuestionCount != 1 || len(got.QuestionKeys) != 1 || got.QuestionKeys[0] != prompt.QuestionKey {
		t.Fatalf("legacy progress = %#v", got)
	}
}

func TestStartUsesIndependentContextAndPersistsFailure(t *testing.T) {
	m := NewManager(store.New(t.TempDir(), nil), testResearcher{err: errors.New("research failed")}, nil, ".")
	defer m.Close()
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

func TestCloseCancelsAndJoinsExport(t *testing.T) {
	m := NewManager(store.New(t.TempDir(), nil), testResearcher{}, nil, ".")
	exportStarted := make(chan struct{})
	exportJoined := make(chan struct{})
	m.SetExporter(ExporterFunc(func(ctx context.Context, _ *model.Plan) error {
		close(exportStarted)
		<-ctx.Done()
		close(exportJoined)
		return ctx.Err()
	}))
	s := &Session{ID: "plan-export", State: Reviewing, UpdatedAt: time.Now().UTC()}
	m.mu.Lock()
	m.sessions[s.ID] = s
	m.mu.Unlock()
	if _, err := m.store.CreatePlan(s.ID, "title", "summary"); err != nil {
		t.Fatal(err)
	}
	go func() { _ = m.Approve(s.ID) }()
	select {
	case <-exportStarted:
	case <-time.After(time.Second):
		t.Fatal("export did not start")
	}
	m.Close()
	select {
	case <-exportJoined:
	default:
		t.Fatal("Close returned before export joined")
	}
	m.Close()
	if got := m.Session(s.ID); got == nil || got.ExportStatus != "failed" || got.ExportError != context.Canceled.Error() {
		t.Fatalf("export result = %#v", got)
	}
}

func TestSubscribeSendsCurrentSnapshot(t *testing.T) {
	m := NewManager(store.New(t.TempDir(), nil), testResearcher{}, nil, ".")
	defer m.Close()
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
