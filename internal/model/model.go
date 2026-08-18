package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Plan struct {
	Title     string        `json:"title"`
	Summary   string        `json:"summary"`
	State     string        `json:"state"`
	UpdatedAt string        `json:"updated_at,omitempty"`
	Messages  []Message     `json:"messages,omitempty"`
	Modules   []Module      `json:"modules"`
	Session   *SessionState `json:"session,omitempty"`
}

type SessionState struct {
	ID             string    `json:"id"`
	Prompt         string    `json:"prompt"`
	State          string    `json:"state"`
	Context        string    `json:"context,omitempty"`
	DecisionAreas  string    `json:"decision_areas,omitempty"`
	Error          string    `json:"error,omitempty"`
	Question       *Prompt   `json:"question,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
	QuestionCount  int       `json:"question_count"`
	QuestionKeys   []string  `json:"question_keys,omitempty"`
	TotalQuestions int       `json:"total_questions,omitempty"`
}
type Prompt struct {
	QuestionKey    string   `json:"question_key"`
	Question       string   `json:"question,omitempty"`
	Options        []string `json:"options"`
	AllowCustom    bool     `json:"allow_custom"`
	TotalQuestions int      `json:"total_questions,omitempty"`
	Answered       bool     `json:"answered"`
	Answer         string   `json:"answer,omitempty"`
	Recommended    int      `json:"recommended,omitempty"`
	Current        int      `json:"current,omitempty"`
}
type Message struct {
	ID        string  `json:"id"`
	Role      string  `json:"role"`
	Text      string  `json:"text"`
	ItemRef   string  `json:"item_ref,omitempty"`
	Prompt    *Prompt `json:"prompt,omitempty"`
	CreatedAt string  `json:"created_at"`
}
type Column struct {
	Heading string `json:"heading"`
	Key     string `json:"key"`
}
type Module struct {
	Type       string   `json:"type"`
	Heading    string   `json:"heading"`
	Items      []Item   `json:"items"`
	Columns    []Column `json:"columns,omitempty"`
	HideRowNum bool     `json:"hideRowNum,omitempty"`
}
type Item struct {
	Text       string `json:"text"`
	Severity   string `json:"severity,omitempty"`
	Impact     string `json:"impact,omitempty"`
	Mitigation string `json:"mitigation,omitempty"`
	Status     string `json:"status,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Answered   bool   `json:"answered,omitempty"`
	Answer     string `json:"answer,omitempty"`
	ChangeType string `json:"type,omitempty"`
	Options    string `json:"options,omitempty"`
	Rationale  string `json:"rationale,omitempty"`
}
type FlatPlan struct {
	Title       string       `json:"title"`
	Summary     string       `json:"summary"`
	State       string       `json:"state"`
	AgentStatus string       `json:"agent_status"`
	UpdatedAt   string       `json:"updated_at,omitempty"`
	Messages    []Message    `json:"messages,omitempty"`
	Modules     []FlatModule `json:"modules"`
}
type FlatModule struct {
	Type       string     `json:"type"`
	Heading    string     `json:"heading"`
	Items      []FlatItem `json:"items"`
	Columns    []Column   `json:"columns,omitempty"`
	HideRowNum bool       `json:"hideRowNum,omitempty"`
}
type FlatItem struct {
	Text       string `json:"text"`
	Severity   string `json:"severity,omitempty"`
	Impact     string `json:"impact,omitempty"`
	Mitigation string `json:"mitigation,omitempty"`
	Status     string `json:"status,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Answered   bool   `json:"answered,omitempty"`
	Answer     string `json:"answer,omitempty"`
	ChangeType string `json:"changeType,omitempty"`
	Options    string `json:"options,omitempty"`
	Rationale  string `json:"rationale,omitempty"`
}

func ToFlatPlan(p *Plan, status string) FlatPlan {
	f := FlatPlan{Title: p.Title, Summary: p.Summary, State: p.State, AgentStatus: status, UpdatedAt: p.UpdatedAt, Messages: p.Messages}
	for _, m := range p.Modules {
		fm := FlatModule{Type: m.Type, Heading: m.Heading, Columns: m.Columns, HideRowNum: m.HideRowNum}
		for _, i := range m.Items {
			fm.Items = append(fm.Items, FlatItem{Text: i.Text, Severity: i.Severity, Impact: i.Impact, Mitigation: i.Mitigation, Status: i.Status, Owner: i.Owner, Answered: i.Answered, Answer: i.Answer, ChangeType: i.ChangeType, Options: i.Options, Rationale: i.Rationale})
		}
		f.Modules = append(f.Modules, fm)
	}
	return f
}
func (f FlatPlan) JSON() []byte { b, _ := json.Marshal(f); return b }
func NewMessageID() string      { return fmt.Sprintf("msg_%x", time.Now().UnixNano()) }
