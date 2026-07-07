package toon

import "strings"

// writer handles buffered output with indentation management.
type writer struct {
	buf        *strings.Builder
	indent     string
	indentSize int
}

// newWriter creates a new writer with the given indentation size.
func newWriter(indentSize int) *writer {
	return &writer{
		buf:        &strings.Builder{},
		indent:     strings.Repeat(" ", indentSize),
		indentSize: indentSize,
	}
}

// push adds a line at the specified depth level.
func (w *writer) push(line string, depth int) {
	if w.buf.Len() > 0 {
		w.buf.WriteString(newline)
	}

	// Add indentation
	for i := 0; i < depth; i++ {
		w.buf.WriteString(w.indent)
	}

	w.buf.WriteString(line)
}

// pushRaw adds content without indentation or newline.
func (w *writer) pushRaw(content string) {
	w.buf.WriteString(content)
}

// String returns the accumulated content.
func (w *writer) String() string {
	return w.buf.String()
}

// Len returns the current buffer length.
func (w *writer) Len() int {
	return w.buf.Len()
}

// Reset clears the buffer.
func (w *writer) Reset() {
	w.buf.Reset()
}