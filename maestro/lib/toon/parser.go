package toon

import (
	"fmt"
	"strings"
)

// parser handles parsing of TOON format strings.
type parser struct {
	input  string
	pos    int
	line   int
	column int
}

// newParser creates a new parser for the given input.
func newParser(input string) *parser {
	return &parser{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

// peek returns the current character without advancing.
func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

// advance moves to the next character.
func (p *parser) advance() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	ch := p.input[p.pos]
	p.pos++
	if ch == '\n' {
		p.line++
		p.column = 1
	} else {
		p.column++
	}
	return ch
}

// skipWhitespace skips spaces and tabs (but not newlines).
func (p *parser) skipWhitespace() {
	for p.peek() == ' ' || p.peek() == '\t' {
		p.advance()
	}
}

// isEOF checks if we're at the end of input.
func (p *parser) isEOF() bool {
	return p.pos >= len(p.input)
}

func (p *parser) parseString() (string, error) {
	if p.peek() != '"' {
		return "", p.error("expected opening quote")
	}
	p.advance() // skip opening quote

	var rawContent strings.Builder
	escaped := false
	for !p.isEOF() {
		ch := p.peek()
		
		if escaped {
			rawContent.WriteByte(ch)
			p.advance()
			escaped = false
			continue
		}
		
		if ch == '\\' {
			rawContent.WriteByte(ch)
			p.advance()
			escaped = true
			continue
		}
		
		if ch == '"' {
			p.advance() // skip closing quote
			raw := rawContent.String()
			unescaped, err := validateAndUnescape(raw)
			if err != nil {
				return "", &DecodeError{
					Message: err.Error(),
					Input:   p.input,
					Line:    p.line,
					Column:  p.column,
				}
			}
			return unescaped, nil
		}
		
		rawContent.WriteByte(ch)
		p.advance()
	}
	
	return "", &DecodeError{
		Message: "unterminated string: missing closing quote",
		Input:   p.input,
		Line:    p.line,
		Column:  p.column,
	}
}

// parseKey parses a key (quoted or unquoted).
func (p *parser) parseKey() (string, error) {
	p.skipWhitespace()
	
	if p.peek() == '"' {
		// Parse quoted key - handles brackets and special chars inside quotes
		key, err := p.parseString()
		if err != nil {
			return "", err
		}
		// After parsing quoted key, we should be at the character after closing quote
		// This allows brackets [ ] to follow the quoted key
		return key, nil
	}
	
	// Parse unquoted key (letters, digits, underscore, dot)
	var result strings.Builder
	for !p.isEOF() {
		ch := p.peek()
		if ch == ':' || ch == '[' || ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		result.WriteByte(ch)
		p.advance()
	}
	
	key := result.String()
	if key == "" {
		return "", p.error("expected key")
	}
	
	return key, nil
}

// parseKeyWithQuoteInfo parses a key and returns whether it was quoted.
func (p *parser) parseKeyWithQuoteInfo() (string, bool, error) {
	p.skipWhitespace()
	
	wasQuoted := false
	if p.peek() == '"' {
		wasQuoted = true
		// Parse quoted key - handles brackets and special chars inside quotes
		key, err := p.parseString()
		return key, wasQuoted, err
	}
	
	// Parse unquoted key
	var result strings.Builder
	for !p.isEOF() {
		ch := p.peek()
		if ch == ':' || ch == '[' || ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		result.WriteByte(ch)
		p.advance()
	}
	
	key := result.String()
	if key == "" {
		return "", false, p.error("expected key")
	}
	
	return key, wasQuoted, nil
}

// expect checks for and consumes an expected character.
func (p *parser) expect(ch byte) error {
	p.skipWhitespace()
	if p.peek() != ch {
		return p.error(fmt.Sprintf("expected '%c', got '%c'", ch, p.peek()))
	}
	p.advance()
	return nil
}

// error creates a DecodeError with current position.
func (p *parser) error(msg string) error {
	// Get context (current line)
	lineStart := p.pos
	for lineStart > 0 && p.input[lineStart-1] != '\n' {
		lineStart--
	}
	lineEnd := p.pos
	for lineEnd < len(p.input) && p.input[lineEnd] != '\n' {
		lineEnd++
	}
	
	context := p.input[lineStart:lineEnd]
	
	return &DecodeError{
		Message: msg,
		Input:   p.input,
		Line:    p.line,
		Column:  p.column,
		Context: context,
	}
}