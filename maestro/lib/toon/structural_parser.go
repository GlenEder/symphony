package toon

import (
	"fmt"
	"strconv"
	"strings"
)

// lineInfo represents a preprocessed line with metadata.
type lineInfo struct {
	content    string
	indent     int
	lineNumber int
	original   string
	isBlank    bool
}

// structuralParser handles indentation-based parsing of TOON format.
type structuralParser struct {
	lines []lineInfo
	pos   int
	opts  *DecodeOptions
}

// newStructuralParser creates a new structural parser.
func newStructuralParser(input string, opts *DecodeOptions) *structuralParser {
	lines := preprocessLines(input)
	return &structuralParser{
		lines: lines,
		pos:   0,
		opts:  opts,
	}
}

// preprocessLines splits input into line information structures.
func preprocessLines(input string) []lineInfo {
	rawLines := strings.Split(input, "\n")
	lines := make([]lineInfo, 0, len(rawLines))
	
	for i, line := range rawLines {
		indent := calculateIndent(line)
		content := strings.TrimLeft(line, " \t")
		isBlank := strings.TrimSpace(line) == ""
		
		lines = append(lines, lineInfo{
			content:    content,
			indent:     indent,
			lineNumber: i + 1,
			original:   line,
			isBlank:    isBlank,
		})
	}
	
	// Remove trailing blank lines
	for len(lines) > 0 && lines[len(lines)-1].isBlank {
		lines = lines[:len(lines)-1]
	}
	
	return lines
}

// calculateIndent returns the number of leading spaces.
func calculateIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			// Tabs not allowed in indentation, but count for error detection
			count += 4 // Treat tab as 4 spaces for counting
		} else {
			break
		}
	}
	return count
}

// parse parses the entire input and returns the decoded value.
func (sp *structuralParser) parse() (Value, error) {
	if len(sp.lines) == 0 {
		return map[string]Value{}, nil
	}
	
	// Validate indentation in strict mode
	if sp.opts.Strict {
		if err := sp.validateIndentation(); err != nil {
			return nil, err
		}
	}
	
	// Detect root type
	rootType, err := sp.detectRootType()
	if err != nil {
		return nil, err
	}
	
	switch rootType {
	case rootTypeArray:
		return sp.parseRootArray()
	case rootTypePrimitive:
		return sp.parseRootPrimitive()
	case rootTypeObject:
		return sp.parseObject(0, 0)
	default:
		return nil, &DecodeError{Message: "unknown root type"}
	}
}

// validateIndentation checks indentation rules in strict mode.
func (sp *structuralParser) validateIndentation() error {
	for _, line := range sp.lines {
		if line.isBlank {
			continue
		}
		
		// Check for tabs in indentation
		leadingSpace := ""
		for i := 0; i < len(line.original) && (line.original[i] == ' ' || line.original[i] == '\t'); i++ {
			leadingSpace += string(line.original[i])
		}
		
		if strings.Contains(leadingSpace, "\t") {
			return &DecodeError{
				Message: "tab characters not allowed in indentation (strict mode)",
				Line:    line.lineNumber,
				Context: line.original,
			}
		}
		
		// Check if indent is multiple of indent_size
		if line.indent > 0 && line.indent%sp.opts.IndentSize != 0 {
			return &DecodeError{
				Message: "indentation must be multiple of indent size (strict mode)",
				Line:    line.lineNumber,
				Context: line.original,
			}
		}
	}
	return nil
}

// detectRootType determines the type of the root value.
func (sp *structuralParser) detectRootType() (rootType, error) {
	if len(sp.lines) == 0 {
		return rootTypeObject, nil
	}
	
	firstLine := sp.lines[0]
	
	// Check for root array patterns
	if strings.HasPrefix(firstLine.content, "[") {
		return rootTypeArray, nil
	}
	
	// Check if single line (primitive)
	if len(sp.lines) == 1 {
		// If starts with quote, check if it's a quoted key (has : after closing quote)
		if strings.HasPrefix(firstLine.content, "\"") {
			// Look for closing quote
			inQuote := false
			escaped := false
			for i := 0; i < len(firstLine.content); i++ {
				ch := firstLine.content[i]
				if i == 0 {
					inQuote = true
					continue
				}
				if escaped {
					escaped = false
					continue
				}
				if ch == '\\' {
					escaped = true
					continue
				}
				if ch == '"' && inQuote {
					// Found closing quote, check what comes after
					remaining := strings.TrimSpace(firstLine.content[i+1:])
					if strings.HasPrefix(remaining, ":") || strings.HasPrefix(remaining, "[") {
						// It's a quoted key
						return rootTypeObject, nil
					}
					// It's a quoted primitive value
					return rootTypePrimitive, nil
				}
			}
			// Unterminated quote - treat as primitive (will error during parse)
			return rootTypePrimitive, nil
		}
		
		// Check if it has a colon (key-value pair = object)
		if strings.Contains(firstLine.content, ":") {
			// If line ends with colon (empty value), it's an object
			if strings.HasSuffix(strings.TrimSpace(firstLine.content), ":") {
				return rootTypeObject, nil
			}
			// Could be object or primitive with colon in value
			// Need to parse to determine
			parts := strings.SplitN(firstLine.content, ":", 2)
			if len(parts) == 2 {
				// If there's content after colon or key has brackets, it's an object
				after := strings.TrimSpace(parts[1])
				if after != "" || strings.Contains(parts[0], "[") {
					return rootTypeObject, nil
				}
			}
		}
		// No colon or empty after colon (no key) = primitive
		return rootTypePrimitive, nil
	}
	
	// Multiple lines = object
	return rootTypeObject, nil
}

// parseRootPrimitive parses a single primitive value.
func (sp *structuralParser) parseRootPrimitive() (Value, error) {
	if len(sp.lines) == 0 {
		return nil, &DecodeError{Message: "empty input"}
	}
	
	line := sp.lines[0]
	return parseValue(line.content)
}

// parseRootArray parses a root-level array.
func (sp *structuralParser) parseRootArray() (Value, error) {
	return sp.parseArrayFromLine(sp.lines[0], 0)
}

// parseObject parses an object starting from the current position.
func (sp *structuralParser) parseObject(baseIndent int, startPos int) (Value, error) {
	result := make(map[string]Value)
	sp.pos = startPos
	
	for sp.pos < len(sp.lines) {
		line := sp.lines[sp.pos]
		
		// Skip blank lines in objects - allowed between fields
		// Blank lines within arrays are handled by array parsing functions
		if line.isBlank {
			sp.pos++
			continue
		}
		
		// Check if we've moved to a different indent level
		if line.indent < baseIndent {
			break
		}
		
		if line.indent > baseIndent {
			// Skip lines that are more indented (they'll be parsed as part of nested values)
			sp.pos++
			continue
		}
		
		// Parse key-value pair
		key, wasQuoted, value, err := sp.parseKeyValueLineWithQuoteInfo(line, baseIndent)
		if err != nil {
			return nil, err
		}
		
		// Apply path expansion if enabled, key contains dots, was not quoted, and all segments are valid identifiers
		if sp.opts.ExpandPaths == "safe" && !wasQuoted && strings.Contains(key, ".") && isExpandablePath(key) {
			if err := sp.expandDottedKey(key, value, result); err != nil {
				return nil, err
			}
		} else {
			// Check for conflict with already-expanded paths in strict mode
			if sp.opts.Strict && sp.opts.ExpandPaths == "safe" {
				if existing, exists := result[key]; exists {
					existingType := getValueType(existing)
					newType := getValueType(value)
					if existingType == "object" && newType != "object" && newType != "null" {
						return nil, &DecodeError{
							Message: fmt.Sprintf("path expansion conflict: key %q conflicts with expanded path (type %s cannot overwrite %s)", key, newType, existingType),
						}
					}
				}
			}
			result[key] = value
		}
		sp.pos++
	}
	
	return result, nil
}

// expandDottedKey expands a dotted key path into nested maps with conflict resolution.
// Example: "a.b.c" with value 1 creates {"a":{"b":{"c":1}}}
func (sp *structuralParser) expandDottedKey(path string, value Value, target map[string]Value) error {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return &DecodeError{Message: "empty path in expandDottedKey"}
	}
	
	// Handle path ending with '.' (e.g., "data.") to create empty array
	if parts[len(parts)-1] == "" {
		// Remove trailing empty part
		parts = parts[:len(parts)-1]
		if len(parts) == 0 {
			return &DecodeError{Message: "invalid path with only dot"}
		}
		// Set empty array as value
		value = []interface{}{}
	}
	
	// Single part - direct assignment
	if len(parts) == 1 {
		key := parts[0]
		// Handle empty nested object case (e.g., "a.b:")
		if key == "" && value == nil {
			return nil
		}
		// Check for conflict in strict mode
		if sp.opts.Strict {
			if existing, exists := target[key]; exists {
				// Check type compatibility - primitive vs object/array conflict
				existingType := getValueType(existing)
				newType := getValueType(value)
				// Conflict if trying to assign different non-null types
				if existingType != newType && existingType != "null" && newType != "null" {
					return &DecodeError{
						Message: fmt.Sprintf("path expansion conflict: key %q already exists with type %s, cannot assign type %s", key, existingType, newType),
					}
				}
			}
		}
		target[key] = value
		return nil
	}
	
	// Multi-part path - recursive expansion
	firstKey := parts[0]
	remainingPath := strings.Join(parts[1:], ".")
	
	// Handle empty key segment
	if firstKey == "" {
		return &DecodeError{Message: "empty key segment in path"}
	}
	
	// Get or create intermediate map
	existing, exists := target[firstKey]
	if !exists {
		// Create new nested map
		nested := make(map[string]Value)
		target[firstKey] = nested
		return sp.expandDottedKey(remainingPath, value, nested)
	}
	
	// Existing value - check compatibility
	nestedMap, isMap := existing.(map[string]Value)
	if !isMap {
		// Conflict: existing value is not a map
		if sp.opts.Strict {
			existingType := getValueType(existing)
			return &DecodeError{
				Message: fmt.Sprintf("path expansion conflict: key %q has type %s, cannot expand as object", firstKey, existingType),
			}
		}
		// Non-strict: overwrite with new nested structure (LWW)
		nested := make(map[string]Value)
		target[firstKey] = nested
		return sp.expandDottedKey(remainingPath, value, nested)
	}
	
	// Recurse into existing map
	return sp.expandDottedKey(remainingPath, value, nestedMap)
}

// getValueType returns a string describing the type of a value for error messages
func getValueType(v Value) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case map[string]Value:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	default:
		return "unknown"
	}
}

// areValuesCompatible checks if two values have compatible types for merging.
func areValuesCompatible(v1, v2 Value) bool {
	// Both nil
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil || v2 == nil {
		return false
	}
	
	// Check if both are maps
	_, isMap1 := v1.(map[string]Value)
	_, isMap2 := v2.(map[string]Value)
	if isMap1 && isMap2 {
		return true
	}
	
	// Check if both are arrays
	_, isArray1 := v1.([]Value)
	_, isArray2 := v2.([]Value)
	if isArray1 && isArray2 {
		return true
	}
	
	// Different types - incompatible
	if (isMap1 || isArray1) != (isMap2 || isArray2) {
		return false
	}
	
	// Both primitives - compatible
	return !isMap1 && !isArray1 && !isMap2 && !isArray2
}

// isExpandablePath checks if a dotted path can be safely expanded.
// Returns false if any segment would need quoting (contains special chars, hyphens, etc.)
func isExpandablePath(path string) bool {
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if !isValidIdentifier(part) {
			return false
		}
	}
	return true
}

// isValidIdentifier checks if a string is a valid unquoted identifier.
// Valid identifiers: start with letter/underscore, contain only letters/digits/underscores
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	
	// First character: letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	
	// Remaining characters: letter, digit, or underscore
	for _, ch := range s[1:] {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	
	return true
}

// parseKeyValueLine parses a single key-value line.
func (sp *structuralParser) parseKeyValueLine(line lineInfo, baseIndent int) (string, Value, error) {
	p := newParser(line.content)
	
	// Parse key
	key, err := p.parseKey()
	if err != nil {
		return "", nil, err
	}
	
	// Check for array marker
	if p.peek() == '[' {
		// This is an array
		value, err := sp.parseArrayFromLine(line, baseIndent)
		return key, value, err
	}
	
	// Expect colon
	if err := p.expect(':'); err != nil {
		return "", nil, err
	}
	
	p.skipWhitespace()
	
	// Check if value is on same line or next lines
	if p.isEOF() || p.peek() == '\n' {
		// Value is on next lines (nested object or array)
		sp.pos++
		if sp.pos >= len(sp.lines) {
			// Empty value
			return key, nil, nil
		}
		
		nextLine := sp.lines[sp.pos]
		if nextLine.indent <= baseIndent {
			// Empty value
			sp.pos--
			return key, nil, nil
		}
		
		// Parse nested value
		value, err := sp.parseObject(nextLine.indent, sp.pos)
		if err != nil {
			return "", nil, err
		}
		sp.pos-- // Will be incremented in main loop
		return key, value, nil
	}
	
	// Parse inline value
	remaining := p.input[p.pos:]
	value, err := parseValue(remaining)
	if err != nil {
		return "", nil, err
	}
	
	return key, value, nil
}

// parseKeyValueLineWithQuoteInfo parses a single key-value line and returns quote info.
func (sp *structuralParser) parseKeyValueLineWithQuoteInfo(line lineInfo, baseIndent int) (string, bool, Value, error) {
	p := newParser(line.content)
	
	// Parse key with quote information
	key, wasQuoted, err := p.parseKeyWithQuoteInfo()
	if err != nil {
		return "", false, nil, err
	}
	
	// Check for array marker
	if p.peek() == '[' {
		// This is an array
		value, err := sp.parseArrayFromLine(line, baseIndent)
		return key, wasQuoted, value, err
	}
	
	// Expect colon
	if err := p.expect(':'); err != nil {
		return "", false, nil, err
	}
	
	p.skipWhitespace()
	
	// Check if value is on same line or next lines
	if p.isEOF() || p.peek() == '\n' {
		// Value is on next lines (nested object or array)
		sp.pos++
		if sp.pos >= len(sp.lines) {
			// Empty value - return empty map for object key
			return key, wasQuoted, map[string]Value{}, nil
		}
		
		nextLine := sp.lines[sp.pos]
		if nextLine.indent <= baseIndent {
			// Empty value - return empty map for object key
			sp.pos--
			if strings.HasSuffix(key, "[0]") {
				return key, wasQuoted, []Value{}, nil
			}
			return key, wasQuoted, map[string]Value{}, nil
		}
		
		// Parse nested value
		value, err := sp.parseObject(nextLine.indent, sp.pos)
		if err != nil {
			return "", false, nil, err
		}
		sp.pos-- // Will be incremented in main loop
		return key, wasQuoted, value, nil
	}
	
	// Parse inline value
	remaining := p.input[p.pos:]
	if strings.HasPrefix(remaining, "[") {
		value, err := sp.parseArray(p, baseIndent)
		if err != nil {
			return "", false, nil, err
		}
		return key, wasQuoted, value, nil
	}
	value, err := parseValue(remaining)
	if err != nil {
		return "", false, nil, err
	}
	
	return key, wasQuoted, value, nil
}

// parseArrayFromLine parses an array starting from a line.
func (sp *structuralParser) parseArrayFromLine(line lineInfo, baseIndent int) (Value, error) {
	p := newParser(line.content)

	// Skip to opening bracket (handles both quoted and unquoted keys)
	// Need to skip past quoted keys to avoid finding '[' inside quotes like "key[test]"[3]
	if p.peek() == '"' {
		// Skip past quoted key, handling escaped quotes and inner brackets
		p.advance() // skip opening quote
		for !p.isEOF() {
			ch := p.peek()
			if ch == '\\' {
				// Skip escape sequence (both backslash and next char)
				p.advance()
				if !p.isEOF() {
					p.advance()
				}
				continue
			}
			if ch == '"' {
				p.advance() // skip closing quote
				break
			}
			// Continue advancing even if we see brackets inside quotes
			p.advance()
		}
	}

	// Now skip to the array bracket
	for p.peek() != '[' && !p.isEOF() {
		p.advance()
	}

	if err := p.expect('['); err != nil {
		return nil, err
	}
	
	// Parse length and delimiter marker together
	lengthStr := ""
	delimiter := ""
	for p.peek() != ']' && !p.isEOF() {
		ch := p.peek()
		if ch >= '0' && ch <= '9' {
			lengthStr += string(ch)
			p.advance()
		} else if ch == '#' {
			// Length marker prefix
			p.advance()
		} else if ch == '\t' {
			delimiter = "\t"
			lengthStr += string(ch)
			p.advance()
		} else if ch == '|' {
			delimiter = "|"
			lengthStr += string(ch)
			p.advance()
		} else {
			break
		}
	}
	
	if err := p.expect(']'); err != nil {
		return nil, err
	}
	
	// Check for tabular format {keys}
	if p.peek() == '{' {
		return sp.parseTabularArray(line, baseIndent, lengthStr, delimiter)
	}
	
	if err := p.expect(':'); err != nil {
		return nil, err
	}
	
	p.skipWhitespace()
	
	// Check if values are inline or on next lines
	if p.isEOF() || p.peek() == '\n' {
		// Values on next lines (list format)
		return sp.parseListArray(baseIndent, lengthStr, delimiter)
	}
	
	// Inline array
	return sp.parseInlineArray(p, lengthStr, delimiter)
}

// parseInlineArray parses an inline array of primitives.
func (sp *structuralParser) parseInlineArray(p *parser, lengthStr string, delimiter string) (Value, error) {
	// Use provided delimiter or default to comma
	if delimiter == "" {
		delimiter = ","
	}
	
	remaining := p.input[p.pos:]
	
	// Don't validate delimiter consistency here - commas in data are allowed with tab/pipe delimiters
	// Only the splitRowByDelimiter function handles delimiter parsing correctly
	
	// Use splitRowByDelimiter to handle delimiters properly (respects quotes)
	parts := sp.splitRowByDelimiter(remaining, delimiter)
	
	result := make([]Value, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		value, err := parseValue(trimmed)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	
	// Validate array length if specified
	if lengthStr != "" && sp.opts.Strict {
		// Extract numeric length from lengthStr
		numStr := ""
		for _, ch := range lengthStr {
			if ch >= '0' && ch <= '9' {
				numStr += string(ch)
			}
		}
		if numStr != "" {
			expectedLen, _ := strconv.Atoi(numStr)
			if len(result) != expectedLen {
				return nil, &DecodeError{
					Message: fmt.Sprintf("array length mismatch: expected %d, got %d", expectedLen, len(result)),
				}
			}
		}
	}
	
	return result, nil
}

// parseTabularArray parses a tabular array format.
func (sp *structuralParser) parseTabularArray(line lineInfo, baseIndent int, lengthStr string, delimiter string) (Value, error) {
	p := newParser(line.content)
	
	// Skip to opening brace
	for p.peek() != '{' && !p.isEOF() {
		p.advance()
	}
	p.advance() // skip {
	
	// Use provided delimiter (tab/pipe) or default to comma
	headerDelimiter := ","
	if delimiter == "\t" || delimiter == "|" {
		headerDelimiter = delimiter
	}
	
	// Parse keys respecting quotes and delimiter
	keys := []string{}
	current := ""
	inQuotes := false
	escaped := false
	
	for p.peek() != '}' && !p.isEOF() {
		ch := p.peek()
		p.advance()
		
		if escaped {
			current += string([]byte{ch})
			escaped = false
			continue
		}
		
		if ch == '\\' && inQuotes {
			escaped = true
			continue
		}
		
		if ch == '"' {
			if !inQuotes {
				inQuotes = true
			} else {
				inQuotes = false
				// Add key
				keys = append(keys, current)
				current = ""
			}
			continue
		}
		
		if !inQuotes {
			if ch == ' ' {
				// Skip spaces outside quotes
				continue
			}
			if string(ch) == headerDelimiter {
				// Delimiter - save current key if any
				if current != "" {
					keys = append(keys, strings.TrimSpace(current))
					current = ""
				}
				continue
			}
		}
		
		current += string([]byte{ch})
	}
	
	// Add last key if any
	if current != "" {
		keys = append(keys, strings.TrimSpace(current))
	}
	
	// Parse data rows with delimiter-aware splitting
	result := make([]Value, 0)
	sp.pos++
	rowCount := 0

	// Debug: Check parser state
	if sp.pos >= len(sp.lines) {
		// No lines available after header - this can happen if temp parser not set up correctly
		if lengthStr != "" && sp.opts.Strict {
			numStr := ""
			for _, ch := range lengthStr {
				if ch >= '0' && ch <= '9' {
					numStr += string(ch)
				}
			}
			if numStr != "" {
				expectedLen, _ := strconv.Atoi(numStr)
				if expectedLen > 0 {
					return nil, &DecodeError{
						Message: fmt.Sprintf("tabular array length mismatch: expected %d rows, got 0 (pos=%d, total lines=%d, baseIndent=%d)", expectedLen, sp.pos, len(sp.lines), baseIndent),
					}
				}
			}
		}
		return result, nil
	}

	for sp.pos < len(sp.lines) {
		line := sp.lines[sp.pos]
		
		if line.isBlank {
			if sp.opts.Strict {
				return nil, &DecodeError{
					Message: "blank line not allowed within tabular array in strict mode",
					Line:    line.lineNumber,
					Context: line.original,
				}
			}
			sp.pos++
			continue
		}

		// For tabular arrays on hyphen lines, data rows can be at same indent as header
		// Use < instead of <= to allow same-indent data rows
		if line.indent < baseIndent {
			sp.pos--
			break
		}

		// Check if this line is an object field (has unquoted colon) rather than a data row
		// Data rows may have quoted colons, but object fields have unquoted colons
		if sp.hasUnquotedColon(line.content) {
			// This looks like an object field, not a tabular row - stop parsing
			sp.pos--
			break
		}

		// Parse row respecting quoted values
		parts := sp.splitRowByDelimiter(line.content, headerDelimiter)
		
		if sp.opts.Strict && len(parts) != len(keys) {
			return nil, &DecodeError{
				Message: fmt.Sprintf("tabular array row has wrong number of values: expected %d, got %d", len(keys), len(parts)),
				Line:    line.lineNumber,
				Context: line.original,
			}
		}
		
		row := make(map[string]Value)
		for i, key := range keys {
			if i < len(parts) {
				value, err := parseValue(strings.TrimSpace(parts[i]))
				if err != nil {
					return nil, err
				}
				row[key] = value
			}
		}
		
		result = append(result, row)
		rowCount++
		sp.pos++
	}
	
	// Validate row count if length specified
	if lengthStr != "" && sp.opts.Strict {
		numStr := ""
		for _, ch := range lengthStr {
			if ch >= '0' && ch <= '9' {
				numStr += string(ch)
			}
		}
		if numStr != "" {
			expectedLen, _ := strconv.Atoi(numStr)
			if rowCount != expectedLen {
				return nil, &DecodeError{
					Message: fmt.Sprintf("tabular array length mismatch: expected %d rows, got %d", expectedLen, rowCount),
				}
			}
		}
	}
	
	return result, nil
}

// splitRowByDelimiter splits a row by delimiter respecting quoted strings
func (sp *structuralParser) splitRowByDelimiter(content string, delimiter string) []string {
	parts := make([]string, 0)
	current := ""
	inQuotes := false
	escaped := false
	
	for i := 0; i < len(content); i++ {
		ch := content[i]
		
		if escaped {
			current += string([]byte{ch})
			escaped = false
			continue
		}
		
		if ch == '\\' && inQuotes {
			current += string([]byte{ch})
			escaped = true
			continue
		}
		
		if ch == '"' {
			inQuotes = !inQuotes
			current += string([]byte{ch})
			continue
		}
		
		if !inQuotes && string(ch) == delimiter {
			parts = append(parts, current)
			current = ""
			continue
		}
		
		current += string([]byte{ch})
	}
	
	parts = append(parts, current)
	return parts
}

// hasUnquotedColon checks if a line contains an unquoted colon,
// which typically indicates an object field (key: value) rather than data.
func (sp *structuralParser) hasUnquotedColon(content string) bool {
	inQuotes := false
	escaped := false

	for i := 0; i < len(content); i++ {
		ch := content[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && inQuotes {
			escaped = true
			continue
		}

		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}

		// Check for unquoted colon followed by space (typical object field pattern)
		if !inQuotes && ch == ':' {
			// Check if followed by space or end of string (typical for "key: value" or "key:")
			if i+1 >= len(content) || content[i+1] == ' ' {
				return true
			}
		}
	}

	return false
}

// parseListArray parses a list-style array.
func (sp *structuralParser) parseListArray(baseIndent int, lengthStr string, delimiter string) (Value, error) {
	result := make([]Value, 0)
	sp.pos++
	itemCount := 0
	
	for sp.pos < len(sp.lines) {
		line := sp.lines[sp.pos]
		
		if line.isBlank {
			// Check if array has ended (next non-blank line is at baseIndent or less)
			// Look ahead to see if this blank line is between array items or after array
			nextNonBlank := sp.pos + 1
			for nextNonBlank < len(sp.lines) && sp.lines[nextNonBlank].isBlank {
				nextNonBlank++
			}
			if nextNonBlank < len(sp.lines) && sp.lines[nextNonBlank].indent <= baseIndent {
				// Blank line is after array ends - this is allowed
				sp.pos--
				break
			}
			// In strict mode, blank lines within arrays are not allowed
			if sp.opts.Strict {
				return nil, &DecodeError{
					Message: "blank lines not allowed within arrays in strict mode",
					Line:    line.lineNumber,
					Context: line.original,
				}
			}
			sp.pos++
			continue
		}
		
		if line.indent <= baseIndent {
			sp.pos--
			break
		}
		
		// Check for list marker
		if strings.HasPrefix(line.content, "-") {
			// Parse list item
			content := strings.TrimPrefix(line.content, "-")
			content = strings.TrimSpace(content)
			
			// Check if empty (just a hyphen) - should be empty object
			if content == "" {
				result = append(result, map[string]Value{})
				itemCount++
				sp.pos++
			} else if strings.HasPrefix(content, "[") {
				// Nested array on hyphen line
				value, err := sp.parseArrayFromLine(line, line.indent)
				if err != nil {
					return nil, err
				}
				result = append(result, value)
				itemCount++
				sp.pos++
			} else if !strings.Contains(content, ":") {
				// Simple value
				value, err := parseValue(content)
				if err != nil {
					return nil, err
				}
				result = append(result, value)
				itemCount++
				sp.pos++
			} else {
				// Object - collect all lines for this item
				itemLines := []lineInfo{line}
				itemIndent := line.indent
				sp.pos++
				
				for sp.pos < len(sp.lines) {
					nextLine := sp.lines[sp.pos]
					if nextLine.isBlank {
						if sp.opts.Strict {
							return nil, &DecodeError{
								Message: "blank lines not allowed within arrays in strict mode",
								Line:    nextLine.lineNumber,
								Context: nextLine.original,
							}
						}
						sp.pos++
						continue
					}

					// Check if this is a new parent-level item
					if nextLine.indent < itemIndent {
						break // Less indented = definitely parent level
					}

					// Same indent with hyphen = new sibling item, always break
					if nextLine.indent == itemIndent && strings.HasPrefix(nextLine.content, "-") {
						break
					}

					// Greater indent with hyphen: check if current item is array definition
					// If first line is array (e.g., "matrix[2]:"), hyphens at greater indent are nested array items
					if nextLine.indent > itemIndent && strings.HasPrefix(nextLine.content, "-") {
						// Get first line of current item (strip hyphen)
						firstContent := strings.TrimPrefix(itemLines[0].content, "-")
						firstContent = strings.TrimSpace(firstContent)

						// If first line is array header (ends with : and has [N]),
						// then this hyphen is a nested array item
						hasArrayNotation := strings.Contains(firstContent, "[") && strings.Contains(firstContent, "]")
						endsWithColon := strings.HasSuffix(firstContent, ":")

						if hasArrayNotation && endsWithColon {
							// Continue - this is nested array item
							itemLines = append(itemLines, nextLine)
							sp.pos++
							continue
						}

						// Otherwise it's content with a hyphen, treat as regular content
					}

					// More indented = part of current item (or we determined hyphen is nested array item above)
					itemLines = append(itemLines, nextLine)
					sp.pos++
				}

				// Parse item as object
				item, err := sp.parseListItem(itemLines, itemIndent, delimiter)
				if err != nil {
					return nil, err
				}
				result = append(result, item)
				itemCount++
			}
		} else {
			// Line doesn't start with hyphen - stop array parsing
			// This line is likely a sibling field to the array
			sp.pos--
			break
		}
	}
	
	// Validate array length if specified
	if lengthStr != "" && sp.opts.Strict {
		numStr := ""
		for _, ch := range lengthStr {
			if ch >= '0' && ch <= '9' {
				numStr += string(ch)
			}
		}
		if numStr != "" {
			expectedLen, _ := strconv.Atoi(numStr)
			if itemCount != expectedLen {
				return nil, &DecodeError{
					Message: fmt.Sprintf("list array length mismatch: expected %d, got %d", expectedLen, itemCount),
				}
			}
		}
	}
	
	return result, nil
}

// parseListItem parses a single list item (which may be an object).
func (sp *structuralParser) parseListItem(lines []lineInfo, baseIndent int, parentDelimiter string) (Value, error) {
	if len(lines) == 0 {
		return nil, &DecodeError{Message: "empty list item"}
	}
	
	firstLine := lines[0]
	content := strings.TrimPrefix(firstLine.content, "-")
	content = strings.TrimSpace(content)
	
	// Check for array on hyphen line - handle key[...]syntax
	if strings.HasPrefix(content, "[") {
		return sp.parseArrayFromLine(firstLine, baseIndent)
	}
	
	// Check if first line contains key with array notation
	if strings.Contains(content, "[") && strings.Contains(content, "]") {
		// Parse key with array marker
		p := newParser(content)
		key, err := p.parseKey()
		if err == nil && p.peek() == '[' {
			// This is an array definition on the hyphen line
			// Create temp parser with just the lines for this item
			adjustedLines := make([]lineInfo, len(lines))
			copy(adjustedLines, lines)
			adjustedLines[0].content = content // Remove "- " prefix

			tempSP := &structuralParser{
				lines: adjustedLines,
				pos:   0,
				opts:  sp.opts,
			}

			// Parse array using temp parser with proper base indent
			value, err := tempSP.parseArrayFromLine(adjustedLines[0], baseIndent)
			if err != nil {
				return nil, err
			}

			// Start building result with the array
			result := map[string]Value{key: value}

			// Check if there are remaining lines at the same or greater indent (sibling fields)
			// tempSP.pos points to the last line consumed by array parsing
			// Lines after tempSP.pos are sibling fields at same indent as hyphen
			if tempSP.pos < len(adjustedLines)-1 {
				// Parse remaining lines as object fields
				remainingStartPos := tempSP.pos + 1
				for i := remainingStartPos; i < len(adjustedLines); i++ {
					line := adjustedLines[i]

					// Skip blank lines
					if line.isBlank {
						continue
					}

					// Parse as key-value line
					if strings.Contains(line.content, ":") {
						k, v, err := sp.parseKeyValueLine(line, baseIndent)
						if err != nil {
							return nil, err
						}
						result[k] = v
					}
				}
			}

			return result, nil
		}
	}
	
	// If only one line, parse as key-value or simple value
	if len(lines) == 1 {
		if !strings.Contains(content, ":") {
			// Try to parse as value (may be array or primitive)
			return parseValue(content)
		}
	}
	
	// Parse as object with proper nesting support
	result := make(map[string]Value)
	
	// Check if first line is tabular array header (e.g., users[2]{id,name}:)
	if strings.Contains(content, "[") && strings.Contains(content, "{") && strings.Contains(content, "}") {
		// This might be a tabular array on the hyphen line
		p := newParser(content)
		key, err := p.parseKey()
		if err == nil && p.peek() == '[' {
			// This is a tabular array on hyphen line - need to parse with full context
			// Adjust lines to remove the "- " prefix from first line for proper parsing
			adjustedLines := make([]lineInfo, len(lines))
			copy(adjustedLines, lines)
			adjustedLines[0].content = content // Already has "- " removed

			// For tabular arrays on hyphen lines, the data rows should be at indent > first line indent
			// Create a temp parser with ONLY the lines for this array
			tempSP := &structuralParser{
				lines: adjustedLines,
				pos:   0,
				opts:  sp.opts,
			}

			// Parse the array starting from first line with proper indent
			value, err := tempSP.parseArrayFromLine(adjustedLines[0], adjustedLines[0].indent)
			if err != nil {
				return nil, err
			}
			
			// Build result object with the array and any remaining fields
			result := make(map[string]Value)
			result[key] = value
			
			// Check if there are more fields after the array (tempSP.pos points to next unparsed line)
			for i := tempSP.pos + 1; i < len(adjustedLines); i++ {
				line := adjustedLines[i]
				if line.isBlank {
					continue
				}
				if line.indent <= baseIndent {
					break
				}
				
				// Parse additional fields
				lp := newParser(line.content)
				fkey, ferr := lp.parseKey()
				if ferr != nil {
					continue
				}
				if ferr := lp.expect(':'); ferr != nil {
					continue
				}
				lp.skipWhitespace()
				remaining := lp.input[lp.pos:]
				if fval, ferr := parseValue(remaining); ferr == nil {
					result[fkey] = fval
				}
			}
			
			return result, nil
		}
	}
	
	// Parse first line
	var firstKey string
	if strings.Contains(content, ":") {
		parts := strings.SplitN(content, ":", 2)
		key := strings.TrimSpace(parts[0])
		valueStr := ""
		if len(parts) > 1 {
			valueStr = strings.TrimSpace(parts[1])
		}
		
		// Handle keys with array notation
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			baseKey := key[:strings.Index(key, "[")]
			if valueStr == "" {
				// Check if this is truly empty array (key[0]:) or has nested content
				if strings.HasSuffix(key, "[0]") {
					// Explicitly marked as empty array
					result[baseKey] = []interface{}{}
				}
				// else: Array with size but empty value - will be handled by loop at i=0
			} else {
				// Parse the value
				value, err := parseValue(valueStr)
				if err != nil {
					return nil, err
				}
				result[baseKey] = value
			}
		} else if valueStr != "" {
			// Parse value - may be array, object, or primitive
			value, err := parseValue(valueStr)
			if err != nil {
				return nil, err
			}
			result[key] = value
		} else if !strings.Contains(key, "[") && len(lines) > 1 {
			// Key with empty value and NO array notation (e.g., "properties:") - might be nested object wrapper
			// Only treat as wrapper if there are nested lines with greater indent
			actualKey := key
			// Find all subsequent lines with greater indent
			nestedLines := []lineInfo{}
			baseIndent := lines[0].indent
			for j := 1; j < len(lines); j++ {
				if lines[j].isBlank {
					continue
				}
				if lines[j].indent > baseIndent {
					nestedLines = append(nestedLines, lines[j])
				} else {
					break
				}
			}
			if len(nestedLines) > 0 {
				// Parse nested lines as object with adjusted indent
				tempSP := newStructuralParser("", sp.opts)
				tempSP.lines = nestedLines
				tempSP.pos = 0
				nestedObj, err := tempSP.parseObject(nestedLines[0].indent, 0)
				if err != nil {
					return nil, err
				}
				result[actualKey] = nestedObj
				return result, nil
			}
			// No nested lines found, treat as regular empty value
			firstKey = actualKey
		} else {
			// Empty value - mark this key for special handling
			// Don't add to result yet, let the loop handle nested content
			firstKey = key
		}
	}

	// Parse remaining lines with multi-level recursion support and depth tracking
	// Start at i=0 if first line has array notation with empty value (needs to be processed by loop)
	i := 1
	if strings.Contains(content, "[") && strings.Contains(content, "]") && strings.Contains(content, ":") {
		parts := strings.SplitN(content, ":", 2)
		if len(parts) > 1 && strings.TrimSpace(parts[1]) == "" {
			key := strings.TrimSpace(parts[0])
			if !strings.HasSuffix(key, "[0]") {
				// Non-empty array with no inline value - start loop at i=0 to process it
				i = 0
			}
		}
	}
	
	// Special handling: if firstKey was set but not added to result (empty value on first line),
	// parse all remaining lines as nested content under firstKey - but ONLY for non-array keys
	if firstKey != "" && !strings.Contains(firstKey, "[") {
		_, exists := result[firstKey]
		if !exists && len(lines) > 1 {
			nestedLines := lines[1:]
			if len(nestedLines) > 0 {
				// Parse as nested object
				tempSP := newStructuralParser("", sp.opts)
				tempSP.lines = nestedLines
				tempSP.pos = 0
				
				nestedObj, err := tempSP.parseObject(nestedLines[0].indent, 0)
				if err != nil {
					return nil, err
				}
				result[firstKey] = nestedObj
				return result, nil
			}
		}
	}
	
	for i < len(lines) {
		line := lines[i]
		if line.isBlank {
			i++
			continue
		}
		
		p := newParser(line.content)
		key, err := p.parseKey()
		if err != nil {
			i++
			continue
		}
		
		// Check for array marker
		if p.peek() == '[' {
			// This is an array - parse it with proper depth
			tempSP := newStructuralParser(line.content, sp.opts)
			tempSP.lines = []lineInfo{line}
			tempSP.pos = 0
			
			// Collect nested lines for this array (multi-level recursion)
			currentIndent := line.indent
			j := i + 1
			for j < len(lines) {
				nextLine := lines[j]
				if nextLine.isBlank {
					j++
					continue
				}
				if nextLine.indent <= currentIndent {
					break
				}
				tempSP.lines = append(tempSP.lines, nextLine)
				j++
			}
			
			value, err := tempSP.parseArrayFromLine(line, currentIndent)
			if err != nil {
				i++
				continue
			}
			result[key] = value
			i = j
			continue
		}
		
		if err := p.expect(':'); err != nil {
			i++
			continue
		}
		
		p.skipWhitespace()
		remaining := p.input[p.pos:]
		
		// Check if value is on same line or nested
		if remaining == "" || strings.TrimSpace(remaining) == "" {
			// Empty value - check for nested content with multi-level support
			nestedLines := []lineInfo{}
			currentIndent := line.indent
			j := i + 1
			for j < len(lines) {
				nextLine := lines[j]
				if nextLine.isBlank {
					j++
					continue
				}
				if nextLine.indent <= currentIndent {
					break
				}
				nestedLines = append(nestedLines, nextLine)
				j++
			}
			
			if len(nestedLines) > 0 {
				// Recursively parse nested content with depth limit (prevent stack overflow)
				if len(nestedLines) > 100 {
					return nil, &DecodeError{Message: "nesting depth exceeded limit"}
				}
				
				// Check if this is the first key with empty value - if so, wrap nested content
				if key == firstKey && firstKey != "" {
					// This is the first key (like "properties:") - wrap all nested content under it
					// Create a temporary structural parser to parse the nested content as an object
					tempSP := newStructuralParser("", sp.opts)
					tempSP.lines = nestedLines
					tempSP.pos = 0
					
					// Parse as a nested object starting from the first nested line's indent
					nestedObj, err := tempSP.parseObject(nestedLines[0].indent, 0)
					if err != nil {
						return nil, err
					}
					result[key] = nestedObj
					i = j
					continue
				}
				
				// Use parseListItem recursively for multi-level nested structures
				nestedResult := make(map[string]Value)
				k := 0
				for k < len(nestedLines) {
					nestedLine := nestedLines[k]
					np := newParser(nestedLine.content)
					nkey, nerr := np.parseKey()
					if nerr != nil {
						k++
						continue
					}
					
					// Check for nested arrays
					if np.peek() == '[' {
						tempSP := newStructuralParser(nestedLine.content, sp.opts)
						tempSP.lines = []lineInfo{nestedLine}
						tempSP.pos = 0
						
						// Collect deeply nested lines
						nestedIndent := nestedLine.indent
						m := k + 1
						for m < len(nestedLines) {
							deepLine := nestedLines[m]
							if deepLine.isBlank {
								m++
								continue
							}
							if deepLine.indent <= nestedIndent {
								break
							}
							tempSP.lines = append(tempSP.lines, deepLine)
							m++
						}
						
						nvalue, nerr := tempSP.parseArrayFromLine(nestedLine, nestedIndent)
						if nerr != nil {
							k++
							continue
						}
						nestedResult[nkey] = nvalue
						k = m
						continue
					}
					
					if nerr := np.expect(':'); nerr != nil {
						k++
						continue
					}
					np.skipWhitespace()
					nremaining := np.input[np.pos:]
					
					// Check for deeply nested objects
					if nremaining == "" || strings.TrimSpace(nremaining) == "" {
						deepNestedLines := []lineInfo{}
						nestedIndent := nestedLine.indent
						m := k + 1
						for m < len(nestedLines) {
							deepLine := nestedLines[m]
							if deepLine.isBlank {
								m++
								continue
							}
							if deepLine.indent <= nestedIndent {
								break
							}
							deepNestedLines = append(deepNestedLines, deepLine)
							m++
						}
						
						if len(deepNestedLines) > 0 {
							// Recursively parse deeper nesting
							deepNested := make(map[string]Value)
							for _, deepLine := range deepNestedLines {
								dnp := newParser(deepLine.content)
								dnkey, dnerr := dnp.parseKey()
								if dnerr != nil {
									continue
								}
								if dnp.peek() == '[' {
									tempSP := newStructuralParser(deepLine.content, sp.opts)
									tempSP.lines = []lineInfo{deepLine}
									dnvalue, dnerr := tempSP.parseArrayFromLine(deepLine, deepLine.indent)
									if dnerr == nil {
										deepNested[dnkey] = dnvalue
									}
									continue
								}
								if dnerr := dnp.expect(':'); dnerr == nil {
									dnp.skipWhitespace()
									dnremaining := dnp.input[dnp.pos:]
									dnvalue, dnerr := parseValue(dnremaining)
									if dnerr == nil {
										deepNested[dnkey] = dnvalue
									}
								}
							}
							nestedResult[nkey] = deepNested
							k = m
							continue
						} else {
							nestedResult[nkey] = map[string]Value{}
						}
					} else {
						nvalue, nerr := parseValue(nremaining)
						if nerr != nil {
							k++
							continue
						}
						nestedResult[nkey] = nvalue
					}
					k++
				}
				result[key] = nestedResult
				i = j
				continue
			} else {
				result[key] = map[string]Value{}
			}
		} else {
			// Parse inline value - may be array or primitive
			value, err := parseValue(remaining)
			if err != nil {
				i++
				continue
			}
			result[key] = value
		}
		i++
	}
	
	if len(result) == 0 {
		return parseValue(content)
	}
	
	return result, nil
}
func (sp *structuralParser) parseArray(p *parser, baseIndent int) (Value, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}

	// Parse length and delimiter marker together
	lengthStr := ""
	delimiter := ""
	for p.peek() != ']' && !p.isEOF() {
		ch := p.peek()
		if ch >= '0' && ch <= '9' {
			lengthStr += string(ch)
			p.advance()
		} else if ch == '#' {
			// Length marker prefix
			p.advance()
		} else if ch == '\t' {
			delimiter = "\t"
			lengthStr += string(ch)
			p.advance()
		} else if ch == '|' {
			delimiter = "|"
			lengthStr += string(ch)
			p.advance()
		} else {
			break
		}
	}

	if err := p.expect(']'); err != nil {
		return nil, err
	}

	// Check for tabular format {keys}
	if p.peek() == '{' {
		p.advance() // skip {
		headerStart := p.pos
		for p.peek() != '}' && !p.isEOF() {
			p.advance()
		}
		header := p.input[headerStart:p.pos]
		p.advance() // skip }
		keys := sp.parseHeader(header, delimiter)
		if err := p.expect(':'); err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.isEOF() || p.peek() == '\n' {
			sp.pos++
			return sp.parseTabularRows(baseIndent, lengthStr, delimiter, keys)
		} else {
			return nil, &DecodeError{Message: "tabular array must have rows on separate lines"}
		}
	}

	if err := p.expect(':'); err != nil {
		return nil, err
	}

	p.skipWhitespace()

	if p.isEOF() || p.peek() == '\n' {
		sp.pos++
		return sp.parseListArray(baseIndent, lengthStr, delimiter)
	}

	return sp.parseInlineArray(p, lengthStr, delimiter)
}

func parseLengthAndDelimiter(p *parser) (string, string) {
	lengthStr := ""
	delimiter := ""
	for p.peek() != ']' && !p.isEOF() {
		ch := p.peek()
		if ch >= '0' && ch <= '9' {
			lengthStr += string(ch)
			p.advance()
		} else if ch == '#' {
			p.advance()
		} else if ch == '\t' {
			delimiter = "\t"
			lengthStr += "\t"
			p.advance()
		} else if ch == '|' {
			delimiter = "|"
			lengthStr += "|"
			p.advance()
		} else {
			break
		}
	}
	return lengthStr, delimiter
}

func (sp *structuralParser) parseHeader(header string, delimiter string) []string {
	if delimiter == "" {
		delimiter = ","
	}
	parts := sp.splitRowByDelimiter(header, delimiter)
	keys := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
			unescaped, err := validateAndUnescape(part[1:len(part)-1])
			if err != nil {
				keys = append(keys, part)
			} else {
				keys = append(keys, unescaped)
			}
		} else {
			keys = append(keys, part)
		}
	}
	return keys
}

func (sp *structuralParser) parseTabularRows(baseIndent int, lengthStr string, delimiter string, keys []string) ([]Value, error) {
	result := make([]Value, 0)
	rowCount := 0
	for sp.pos < len(sp.lines) {
		line := sp.lines[sp.pos]
		if line.isBlank {
			if sp.opts.Strict {
				return nil, &DecodeError{
					Message: "blank line not allowed within tabular array in strict mode",
					Line:    line.lineNumber,
					Context: line.original,
				}
			}
			sp.pos++
			continue
		}
		if line.indent <= baseIndent {
			break
		}
		parts := sp.splitRowByDelimiter(line.content, delimiter)
		if sp.opts.Strict && len(parts) != len(keys) {
			return nil, &DecodeError{
				Message: fmt.Sprintf("tabular array row has wrong number of values: expected %d, got %d", len(keys), len(parts)),
				Line:    line.lineNumber,
				Context: line.original,
			}
		}
		row := make(map[string]Value)
		for i, k := range keys {
			if i < len(parts) {
				v, err := parseValue(strings.TrimSpace(parts[i]))
				if err != nil {
					return nil, err
				}
				row[k] = v
			}
		}
		result = append(result, row)
		rowCount++
		sp.pos++
	}
	// validate length
	if sp.opts.Strict && lengthStr != "" {
		numStr := ""
		for _, ch := range lengthStr {
			if ch >= '0' && ch <= '9' {
				numStr += string(ch)
			}
		}
		if numStr != "" {
			expected, _ := strconv.Atoi(numStr)
			if rowCount != expected {
				return nil, &DecodeError{
					Message: fmt.Sprintf("tabular array length mismatch: expected %d rows, got %d", expected, rowCount),
				}
			}
		}
	}
	return result, nil
}

func parseArrayHeader(key string) (lengthStr string, delimiter string, isTabular bool, header string, err error) {
	lastOpen := strings.LastIndex(key, "[")
	if lastOpen == -1 {
		return "", "", false, "", nil
	}
	closePos := strings.Index(key[lastOpen:], "]")
	if closePos == -1 {
		return "", "", false, "", fmt.Errorf("unmatched [ in key %q", key)
	}
	closePos += lastOpen
	lengthStr = key[lastOpen+1 : closePos]
	if strings.Contains(lengthStr, "\t") {
		delimiter = "\t"
	} else if strings.Contains(lengthStr, "|") {
		delimiter = "|"
	} else {
		delimiter = ""
	}
	pos := closePos + 1
	if pos < len(key) && key[pos] == '{' {
		hStart := pos + 1
		hClose := strings.Index(key[hStart:], "}")
		if hClose == -1 {
			return "", "", false, "", fmt.Errorf("unmatched { in key %q", key)
		}
		header = key[hStart : hStart+hClose]
		isTabular = true
		pos += hClose + 1
	}
	if pos < len(key) && key[pos] != ':' {
		return "", "", false, "", fmt.Errorf("expected : after array header in key %q", key)
	}
	return lengthStr, delimiter, isTabular, header, nil
}