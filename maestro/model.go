package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Plan represents a structured planning document with a conversation thread.
type Plan struct {
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	State    string    `json:"state"`              // "draft" or "approved"
	UpdatedAt string   `json:"updated_at,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Modules  []Module  `json:"modules"`
}

// Message is a single entry in the plan's conversation thread.
type Message struct {
	ID        string `json:"id"`
	Role      string `json:"role"`                // "agent" or "human"
	Text      string `json:"text"`
	ItemRef   string `json:"item_ref,omitempty"`  // "moduleIndex:itemIndex" positional ref
	CreatedAt string `json:"created_at"`
}

// Column defines a single column in a table module.
type Column struct {
	Heading string `json:"heading"`
	Key     string `json:"key"`
}

// Module is a typed section of a plan.
type Module struct {
	Type       string   `json:"type"`
	Heading    string   `json:"heading"`
	Items      []Item   `json:"items"`
	Columns    []Column `json:"columns,omitempty"`
	HideRowNum bool     `json:"hideRowNum,omitempty"`
}

// Item is a single entry within a module.
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
}

// FlatPlan is the JSON-friendly version with less nesting for the API and WebSocket.
type FlatPlan struct {
	Title    string       `json:"title"`
	Summary  string       `json:"summary"`
	State    string       `json:"state"`
	UpdatedAt string      `json:"updated_at,omitempty"`
	Messages []Message    `json:"messages,omitempty"`
	Modules  []FlatModule `json:"modules"`
}

type FlatModule struct {
	Type       string       `json:"type"`
	Heading    string       `json:"heading"`
	Items      []FlatItem   `json:"items"`
	Columns    []Column     `json:"columns,omitempty"`
	HideRowNum bool         `json:"hideRowNum,omitempty"`
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
}

func toFlatPlan(p *Plan) FlatPlan {
	fp := FlatPlan{
		Title:     p.Title,
		Summary:   p.Summary,
		State:     p.State,
		UpdatedAt: p.UpdatedAt,
		Messages:  p.Messages,
	}
	for _, m := range p.Modules {
		fm := FlatModule{Type: m.Type, Heading: m.Heading, Columns: m.Columns, HideRowNum: m.HideRowNum}
		for _, it := range m.Items {
			fm.Items = append(fm.Items, FlatItem{
				Text:       it.Text,
				Severity:   it.Severity,
				Impact:     it.Impact,
				Mitigation: it.Mitigation,
				Status:     it.Status,
				Owner:      it.Owner,
				Answered:   it.Answered,
				Answer:     it.Answer,
				ChangeType: it.ChangeType,
			})
		}
		fp.Modules = append(fp.Modules, fm)
	}
	return fp
}

func (fp FlatPlan) JSON() []byte {
	b, _ := json.Marshal(fp)
	return b
}

// newMsgID generates a short unique message ID based on nanosecond timestamp.
func newMsgID() string {
	return fmt.Sprintf("msg_%x", time.Now().UnixNano())
}
