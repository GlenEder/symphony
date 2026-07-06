package main

import "encoding/json"

type Plan struct {
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Response string   `json:"response,omitempty"`
	Approved bool     `json:"approved"`
	Modules  []Module `json:"modules"`
}

type Module struct {
	Type    string `json:"type"`
	Heading string `json:"heading"`
	Items   []Item `json:"items"`
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
}

// FlatPlan is a JSON-friendly version with less nesting for the API.
type FlatPlan struct {
	Title    string       `json:"title"`
	Summary  string       `json:"summary"`
	Response string       `json:"response,omitempty"`
	Approved bool         `json:"approved"`
	Modules  []FlatModule `json:"modules"`
}

type FlatModule struct {
	Type    string     `json:"type"`
	Heading string     `json:"heading"`
	Items   []FlatItem `json:"items"`
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
	fp := FlatPlan{Title: p.Title, Summary: p.Summary, Response: p.Response, Approved: p.Approved}
	for _, m := range p.Modules {
		fm := FlatModule{Type: m.Type, Heading: m.Heading}
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
