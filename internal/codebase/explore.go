// Package codebase explores source trees without external Go dependencies.
package codebase

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	MaxDepth, MaxEntries int
	MaxFileSize          int64
	IgnoreDirs           []string
}
type Explorer struct{ Options }
type SearchMatch struct {
	File    string
	Line    int
	Text    string
	Context []string
}

type TreeResult struct {
	Tree      string
	Truncated bool
}
type ProjectType struct {
	Language  string
	Framework string
	Files     []string
}

var defaultIgnored = map[string]bool{".git": true, "node_modules": true, "vendor": true, "dist": true, "build": true, "target": true, ".idea": true, ".vscode": true}

func New(opts Options) *Explorer {
	if opts.MaxFileSize <= 0 {
		opts.MaxFileSize = 256 * 1024
	}
	if opts.MaxEntries <= 0 {
		opts.MaxEntries = 10000
	}
	return &Explorer{opts}
}
func (e *Explorer) ignored(name string) bool {
	if defaultIgnored[name] {
		return true
	}
	for _, n := range e.IgnoreDirs {
		if name == n {
			return true
		}
	}
	return false
}
func (e *Explorer) Tree(root string) (string, error) {
	result, err := e.TreeResult(root)
	return result.Tree, err
}

func (e *Explorer) TreeResult(root string) (TreeResult, error) {
	info, err := os.Stat(root)
	if err != nil {
		return TreeResult{}, err
	}
	if !info.IsDir() {
		return TreeResult{}, fmt.Errorf("codebase: %s is not a directory", root)
	}
	var b strings.Builder
	count := 0
	truncated := false
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		depth := len(strings.Split(rel, string(filepath.Separator)))
		if e.MaxDepth > 0 && depth > e.MaxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() && e.ignored(d.Name()) {
			return fs.SkipDir
		}
		if count >= e.MaxEntries {
			truncated = true
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		count++
		b.WriteString(strings.Repeat("  ", depth-1))
		b.WriteString("- " + d.Name())
		if d.IsDir() {
			b.WriteByte('/')
		}
		b.WriteByte('\n')
		return nil
	})
	return TreeResult{Tree: b.String(), Truncated: truncated}, err
}
func (e *Explorer) ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > e.MaxFileSize {
		return nil, fmt.Errorf("codebase: file exceeds size limit")
	}
	return ioReadAll(f, e.MaxFileSize)
}
func ioReadAll(r *os.File, max int64) ([]byte, error) {
	var b bytes.Buffer
	if _, err := b.ReadFrom(io.LimitReader(r, max+1)); err != nil {
		return nil, err
	}
	if int64(b.Len()) > max {
		return nil, fmt.Errorf("codebase: file exceeds size limit")
	}
	return b.Bytes(), nil
}

func (e *Explorer) Search(root, pattern string, contextLines int) ([]SearchMatch, error) {
	if _, err := exec.LookPath("rg"); err == nil {
		return e.searchRG(root, pattern, contextLines)
	}
	return e.searchGo(root, pattern, contextLines)
}
func (e *Explorer) searchRG(root, pattern string, contextLines int) ([]SearchMatch, error) {
	if contextLines < 0 {
		contextLines = 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := []string{"--line-number", "--with-filename", "--no-heading", "--color", "never", "-C", strconv.Itoa(contextLines), pattern, root}
	cmd := exec.CommandContext(ctx, "rg", args...)
	var output limitedBuffer
	cmd.Stdout = &output
	outErr := cmd.Run()
	out := output.Bytes()
	err := outErr
	if output.truncated {
		return nil, fmt.Errorf("codebase: search output exceeds limit")
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil && len(out) == 0 {
		if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
			return nil, err
		}
	}
	matches := parseRG(string(out))
	if len(matches) > e.MaxEntries {
		matches = matches[:e.MaxEntries]
	}
	return matches, nil
}

const maxSearchOutput = 4 * 1024 * 1024

type limitedBuffer struct {
	bytes.Buffer
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := maxSearchOutput - b.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.Buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}
	return b.Buffer.Write(p)
}

func parseRG(s string) []SearchMatch {
	var out []SearchMatch
	re := regexp.MustCompile(`^(.+?)(:|-)([0-9]+)(:|-)(.*)$`)
	var current *SearchMatch
	for _, line := range strings.Split(s, "\n") {
		m := re.FindStringSubmatch(line)
		if len(m) == 6 && m[2] == ":" && m[4] == ":" {
			n, _ := strconv.Atoi(m[3])
			out = append(out, SearchMatch{File: m[1], Line: n, Text: m[5], Context: []string{m[5]}})
			current = &out[len(out)-1]
		} else if current != nil && (line == "--" || strings.Contains(line, "-")) {
			current.Context = append(current.Context, line)
		} else if line == "" {
			current = nil
		}
	}
	return out
}
func (e *Explorer) searchGo(root, pattern string, contextLines int) ([]SearchMatch, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var out []SearchMatch
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != root && e.ignored(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > e.MaxFileSize {
			return nil
		}
		f, x := os.Open(path)
		if x != nil {
			return nil
		}
		defer f.Close()
		var lines []string
		s := bufio.NewScanner(f)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		for i, l := range lines {
			if len(out) >= e.MaxEntries {
				return nil
			}
			if re.MatchString(l) {
				start := i - contextLines
				if start < 0 {
					start = 0
				}
				end := i + contextLines + 1
				if end > len(lines) {
					end = len(lines)
				}
				out = append(out, SearchMatch{File: path, Line: i + 1, Text: l, Context: append([]string(nil), lines[start:end]...)})
				if len(out) >= e.MaxEntries {
					return nil
				}
			}
		}
		return nil
	})
	return out, err
}

func (e *Explorer) Detect(root string) ([]ProjectType, error) {
	names := map[string]ProjectType{"go.mod": {"Go", "", nil}, "package.json": {"JavaScript/TypeScript", "", nil}, "Cargo.toml": {"Rust", "", nil}, "requirements.txt": {"Python", "", nil}, "pyproject.toml": {"Python", "", nil}, "Gemfile": {"Ruby", "", nil}, "pom.xml": {"Java", "", nil}, "build.gradle": {"Java", "", nil}, "composer.json": {"PHP", "", nil}}
	var out []ProjectType
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	frameworks := map[string][]struct{ name, framework string }{
		"package.json":     {{"react", "React"}, {"next", "Next.js"}},
		"requirements.txt": {{"django", "Django"}},
		"pyproject.toml":   {{"django", "Django"}},
		"Gemfile":          {{"rails", "Rails"}},
		"go.mod":           {{"gin-gonic/gin", "Gin"}},
		"pom.xml":          {{"spring-boot", "Spring"}, {"springframework", "Spring"}},
		"build.gradle":     {{"spring-boot", "Spring"}, {"springframework", "Spring"}},
	}
	for _, x := range entries {
		if p, ok := names[x.Name()]; ok {
			p.Files = []string{x.Name()}
			if candidates := frameworks[x.Name()]; len(candidates) > 0 {
				data, readErr := os.ReadFile(filepath.Join(root, x.Name()))
				if readErr != nil {
					return nil, readErr
				}
				content := strings.ToLower(string(data))
				for _, candidate := range candidates {
					if strings.Contains(content, strings.ToLower(candidate.name)) {
						p.Framework = candidate.framework
						break
					}
				}
			}
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Language < out[j].Language })
	return out, nil
}
