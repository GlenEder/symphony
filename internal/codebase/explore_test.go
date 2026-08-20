package codebase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"src", ".git", "node_modules", "custom"} {
		if err := os.Mkdir(filepath.Join(root, d), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\nneedle here\nthird\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example\n"), 0644)
	_ = os.WriteFile(filepath.Join(root, ".git", "secret"), []byte("needle"), 0644)
	return root
}
func TestTreeReadAndDetect(t *testing.T) {
	root := fixture(t)
	e := New(Options{MaxFileSize: 100})
	tree, err := e.Tree(root)
	if err != nil || strings.Contains(tree, ".git") || !strings.Contains(tree, "main.go") {
		t.Fatalf("tree: %s %v", tree, err)
	}
	data, err := e.ReadFile(filepath.Join(root, "src", "main.go"))
	if err != nil || len(data) == 0 {
		t.Fatal(err)
	}
	types, _ := e.Detect(root)
	if len(types) != 1 || types[0].Language != "Go" {
		t.Fatalf("types: %+v", types)
	}
}
func TestSearchFallbackAndSummaryBound(t *testing.T) {
	root := fixture(t)
	e := New(Options{MaxFileSize: 100})
	matches, err := e.searchGo(root, "needle", 1)
	if err != nil || len(matches) != 1 || len(matches[0].Context) != 3 {
		t.Fatalf("matches: %+v %v", matches, err)
	}
	summary, err := e.Summary(root, SummaryOptions{MaxBytes: 40})
	if err != nil || len(summary) > 40 {
		t.Fatalf("summary: %d %v", len(summary), err)
	}
}
func TestReadLimit(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "x")
	if err := os.WriteFile(p, []byte("12345"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := New(Options{MaxFileSize: 2}).ReadFile(p); err == nil {
		t.Fatal("expected size error")
	}
}
