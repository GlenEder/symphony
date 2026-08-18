package codebase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SummaryOptions struct {
	MaxBytes    int
	MaxFiles    int
	MaxFileSize int64
}

func (e *Explorer) Summary(root string, opts SummaryOptions) (string, error) {
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 12000
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 50
	}
	effectiveFileSize := e.MaxFileSize
	if opts.MaxFileSize > 0 {
		effectiveFileSize = opts.MaxFileSize
	}
	tree, err := e.Tree(root)
	if err != nil {
		return "", err
	}
	types, _ := e.Detect(root)
	var b strings.Builder
	b.WriteString("Project context\n\nDirectory tree:\n")
	b.WriteString(tree)
	b.WriteString("\nDetected types:\n")
	for _, p := range types {
		fmt.Fprintf(&b, "- %s", p.Language)
		if p.Framework != "" {
			fmt.Fprintf(&b, " (%s)", p.Framework)
		}
		b.WriteByte('\n')
	}
	files := []string{"go.mod", "package.json", "Cargo.toml", "requirements.txt", "pyproject.toml", "Gemfile"}
	n := 0
	for _, name := range files {
		if n >= opts.MaxFiles {
			break
		}
		data, x := readFileLimit(filepath.Join(root, name), effectiveFileSize)
		if x != nil {
			continue
		}
		section := fmt.Sprintf("\n%s:\n%s", name, data)
		if b.Len()+len(section) > opts.MaxBytes {
			break
		}
		b.WriteString(section)
		n++
	}
	result := b.String()
	if len(result) > opts.MaxBytes {
		result = result[:opts.MaxBytes]
	}
	return result, nil
}

func readFileLimit(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioReadAll(f, max)
}
