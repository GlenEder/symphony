package export

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gleneder/symphony/internal/model"
)

var stageHeading = regexp.MustCompile(`^Stage ([1-9][0-9]*): (.+)$`)
var planID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// Config controls work-ticket export.
type Config struct {
	OutputDir       string
	TraceabilityURL string
}

// Exporter converts approved plans into implementation tickets.
type Exporter struct{ config Config }

func New(cfg Config) *Exporter { return &Exporter{config: cfg} }

func (e *Exporter) Export(ctx context.Context, p *model.Plan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil || p.Session == nil || !planID.MatchString(p.Session.ID) {
		return errors.New("approved plan identity is required")
	}
	if p.State != "approved" {
		return errors.New("plan must be approved")
	}
	groups, globals, staged, err := validate(p)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := e.config.OutputDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, ".config", "symphony", "work_tickets")
	}
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	files := make(map[string][]byte)
	if staged {
		nums := make([]int, 0, len(groups))
		for n := range groups {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		for i, n := range nums {
			if err := ctx.Err(); err != nil {
				return err
			}
			files[filepath.Join(dir, fmt.Sprintf("%s-stage-%d.md", p.Session.ID, n))] = []byte(renderStaged(p, globals, groups[n], n, i+1, len(nums), e.config.TraceabilityURL))
		}
	} else {
		files[filepath.Join(dir, p.Session.ID+".md")] = []byte(renderSingle(p, groups[0], globals, e.config.TraceabilityURL))
	}
	return atomicWrite(ctx, dir, files)
}

func validate(p *model.Plan) (map[int][]model.Module, []model.Module, bool, error) {
	groups := map[int][]model.Module{}
	var globals []model.Module
	staged := false
	seenStages := map[int]map[string]string{}
	for _, m := range p.Modules {
		match := stageHeading.FindStringSubmatch(m.Heading)
		stageLike := strings.HasPrefix(m.Heading, "Stage ")
		if match != nil {
			staged = true
			n, _ := strconv.Atoi(match[1])
			if seenStages[n] == nil {
				seenStages[n] = map[string]string{}
			}
			if _, exists := seenStages[n][m.Type]; exists {
				return nil, nil, false, fmt.Errorf("duplicate Stage %d %s module", n, m.Type)
			}
			seenStages[n][m.Type] = m.Heading
			if m.Type == "decision" || m.Type == "assumptions" || m.Type == "notes" {
				return nil, nil, false, fmt.Errorf("%s module cannot be stage-prefixed", m.Type)
			}
			groups[n] = append(groups[n], m)
			continue
		}
		if stageLike {
			return nil, nil, false, fmt.Errorf("invalid stage heading: %q", m.Heading)
		}
		if m.Type == "criteria" || m.Type == "steps" || m.Type == "risks" {
			if staged {
				return nil, nil, false, fmt.Errorf("%s module must be stage-prefixed", m.Type)
			}
			groups[0] = append(groups[0], m)
		} else if m.Type == "decision" || m.Type == "assumptions" {
			globals = append(globals, m)
		}
	}
	if staged {
		nums := make([]int, 0, len(groups))
		for n := range groups {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		for i, n := range nums {
			if n != i+1 {
				return nil, nil, false, errors.New("stage numbers must be contiguous and unique")
			}
		}
	} else {
		if len(groups[0]) == 0 {
			groups[0] = nil
		}
	}
	return groups, globals, staged, nil
}

func renderStaged(p *model.Plan, globals, stage []model.Module, n, pos, total int, trace string) string {
	name := strings.TrimPrefix(stage[0].Heading, fmt.Sprintf("Stage %d: ", n))
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — Stage %d: %s\n\n**Status**: pending\n\n**Stage**: %d of %d of the approved %s plan\n", p.Title, n, name, pos, total, p.Session.ID)
	appendSummary(&b, p)
	appendDecisions(&b, globals)
	appendAssumptions(&b, globals)
	appendExecution(&b, stage)
	appendTrace(&b, trace, p.Session.ID)
	return b.String()
}
func renderSingle(p *model.Plan, exec, globals []model.Module, trace string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n**Status**: pending\n\n", p.Title)
	appendSummary(&b, p)
	appendCriteria(&b, exec)
	appendSteps(&b, exec)
	appendDecisions(&b, globals)
	appendRisks(&b, exec)
	appendAssumptions(&b, globals)
	appendTrace(&b, trace, p.Session.ID)
	return b.String()
}
func appendSummary(b *strings.Builder, p *model.Plan) {
	if strings.TrimSpace(p.Summary) != "" {
		b.WriteString(p.Summary + "\n\n")
	}
}
func appendExecution(b *strings.Builder, ms []model.Module) {
	appendCriteria(b, ms)
	appendSteps(b, ms)
	appendRisks(b, ms)
}
func appendCriteria(b *strings.Builder, ms []model.Module) {
	var xs []model.Item
	for _, m := range ms {
		if m.Type == "criteria" {
			xs = append(xs, m.Items...)
		}
	}
	if len(xs) == 0 {
		return
	}
	b.WriteString("\n## Acceptance Criteria\n\n")
	for _, x := range xs {
		fmt.Fprintf(b, "- [ ] %s\n", x.Text)
	}
	b.WriteString("\n")
}
func appendSteps(b *strings.Builder, ms []model.Module) {
	var xs []model.Item
	for _, m := range ms {
		if m.Type == "steps" {
			xs = append(xs, m.Items...)
		}
	}
	if len(xs) == 0 {
		return
	}
	b.WriteString("\n## Implementation Steps\n\n")
	for i, x := range xs {
		fmt.Fprintf(b, "%d. %s\n", i+1, x.Text)
	}
	b.WriteString("\n")
}
func appendDecisions(b *strings.Builder, ms []model.Module) {
	var xs []model.Item
	for _, m := range ms {
		if m.Type == "decision" {
			xs = append(xs, m.Items...)
		}
	}
	if len(xs) == 0 {
		return
	}
	b.WriteString("\n## Key Decisions\n\n")
	for _, x := range xs {
		fmt.Fprintf(b, "### %s\n- **Decision**: %s\n", x.Text, x.Answer)
		if x.Rationale != "" {
			fmt.Fprintf(b, "- **Rationale**: %s\n", x.Rationale)
		}
		if x.Options != "" {
			fmt.Fprintf(b, "- **Alternatives considered**: %s\n", x.Options)
		}
		b.WriteString("\n")
	}
}
func appendRisks(b *strings.Builder, ms []model.Module) {
	var xs []model.Item
	for _, m := range ms {
		if m.Type == "risks" {
			xs = append(xs, m.Items...)
		}
	}
	if len(xs) == 0 {
		return
	}
	b.WriteString("\n## Risks\n\n")
	for _, x := range xs {
		sev := x.Severity
		if sev == "" {
			sev = "Medium"
		}
		fmt.Fprintf(b, "- **%s**: %s", strings.Title(strings.ToLower(sev)), x.Text)
		if x.Impact != "" {
			fmt.Fprintf(b, " (impact: %s)", x.Impact)
		}
		if x.Mitigation != "" {
			fmt.Fprintf(b, " (mitigation: %s)", x.Mitigation)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}
func appendAssumptions(b *strings.Builder, ms []model.Module) {
	var xs []model.Item
	for _, m := range ms {
		if m.Type == "assumptions" {
			xs = append(xs, m.Items...)
		}
	}
	if len(xs) == 0 {
		return
	}
	b.WriteString("\n## Assumptions\n\n")
	for _, x := range xs {
		fmt.Fprintf(b, "- %s\n", x.Text)
	}
	b.WriteString("\n")
}
func appendTrace(b *strings.Builder, url, id string) {
	if url == "" {
		return
	}
	b.WriteString("\n---\n\n*Generated from [maestro plan](" + strings.TrimRight(url, "/") + "/" + id + ")*\n")
}

func atomicWrite(ctx context.Context, dir string, files map[string][]byte) error {
	// Stage the complete snapshot in a private directory, then swap every target
	// under a rollback guard. This also removes stale tickets from prior exports.
	stage, err := os.MkdirTemp(dir, ".tickets-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	keys := make([]string, 0, len(files))
	for target := range files {
		keys = append(keys, target)
	}
	sort.Strings(keys)
	for _, target := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := filepath.Base(target)
		if err := os.WriteFile(filepath.Join(stage, name), files[target], 0644); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	wanted := make(map[string]bool, len(files))
	for target := range files {
		wanted[filepath.Base(target)] = true
	}
	prefix := stalePrefix(files)
	var targets []string
	for _, entry := range entries {
		name := entry.Name()
		if wanted[name] || name == prefix+".md" || strings.HasPrefix(name, prefix+"-stage-") {
			targets = append(targets, name)
		}
	}
	sort.Strings(targets)
	backup, err := os.MkdirTemp(dir, ".tickets-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(backup)
	moved := []string{}
	installed := []string{}
	rollback := func() {
		for _, name := range installed {
			_ = os.Remove(filepath.Join(dir, name))
		}
		for _, name := range moved {
			_ = os.Rename(filepath.Join(backup, name), filepath.Join(dir, name))
		}
	}
	for _, name := range targets {
		if err := ctx.Err(); err != nil {
			rollback()
			return err
		}
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			if err := os.Rename(filepath.Join(dir, name), filepath.Join(backup, name)); err != nil {
				rollback()
				return err
			}
			moved = append(moved, name)
		}
	}
	for _, target := range keys {
		if err := ctx.Err(); err != nil {
			rollback()
			return err
		}
		name := filepath.Base(target)
		if err := os.Rename(filepath.Join(stage, name), filepath.Join(dir, name)); err != nil {
			rollback()
			return err
		}
		installed = append(installed, name)
	}
	return nil
}

func stalePrefix(files map[string][]byte) string {
	for target := range files {
		name := filepath.Base(target)
		if i := strings.Index(name, "-stage-"); i >= 0 {
			return name[:i]
		}
		return strings.TrimSuffix(name, ".md")
	}
	return ""
}
