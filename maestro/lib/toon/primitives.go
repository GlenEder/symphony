package toon

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// encodePrimitive encodes a primitive value to TOON format.
func encodePrimitive(v Value, delimiter string) (string, error) {
	if v == nil {
		return nullLiteral, nil
	}

	switch val := v.(type) {
	case bool:
		if val {
			return trueLiteral, nil
		}
		return falseLiteral, nil

	case string:
		return encodeString(val, delimiter), nil

	case int, int8, int16, int32, int64:
		if i, ok := toInt64(val); ok {
			return strconv.FormatInt(i, 10), nil
		}
		return "", &EncodeError{Message: "invalid integer", Value: val}

	case uint, uint8, uint16, uint32, uint64:
		if i, ok := toInt64(val); ok {
			return strconv.FormatInt(i, 10), nil
		}
		// Fallback for large uint64
		if u, ok := val.(uint64); ok {
			return strconv.FormatUint(u, 10), nil
		}
		return "", &EncodeError{Message: "invalid unsigned integer", Value: val}

	case float32, float64:
		return encodeFloat(val)

	default:
		return "", &EncodeError{
			Message: "unsupported primitive type",
			Value:   val,
		}
	}
}

// encodeFloat encodes a float value with proper formatting.
func encodeFloat(v Value) (string, error) {
	f, ok := toFloat64(v)
	if !ok {
		return "", &EncodeError{Message: "invalid float", Value: v}
	}

	// Handle NaN and Infinity
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nullLiteral, nil
	}

	// Handle negative zero
	if f == 0 && math.Signbit(f) {
		return "0", nil
	}

	// Check if it's a whole number
	if f == math.Trunc(f) && f >= math.MinInt64 && f <= math.MaxInt64 {
		return strconv.FormatInt(int64(f), 10), nil
	}

	// Format float without scientific notation
	str := strconv.FormatFloat(f, 'f', -1, 64)
	
	// Remove trailing zeros after decimal point
	if strings.Contains(str, ".") {
		str = strings.TrimRight(str, "0")
		str = strings.TrimRight(str, ".")
	}

	return str, nil
}

// encodeString encodes a string value, adding quotes if necessary.
func encodeString(s string, delimiter string) string {
	if needsQuoting(s, delimiter) {
		return doubleQuote + escapeString(s) + doubleQuote
	}
	return s
}

// needsQuoting determines if a string needs to be quoted.
func needsQuoting(s string, delimiter string) bool {
	if s == "" {
		return true
	}

	// Check for leading or trailing spaces
	if strings.HasPrefix(s, space) || strings.HasSuffix(s, space) {
		return true
	}

	// Check if it's a literal
	if s == trueLiteral || s == falseLiteral || s == nullLiteral {
		return true
	}

	// Check if it looks like a number
	if looksLikeNumber(s) {
		return true
	}

	// Check for structure characters
	for _, char := range structureChars {
		if strings.Contains(s, char) {
			return true
		}
	}

	// Check for active delimiter
	if strings.Contains(s, delimiter) {
		return true
	}

	// Check for control characters
	for _, char := range controlChars {
		if strings.Contains(s, char) {
			return true
		}
	}

	// Check if starts with hyphen
	if strings.HasPrefix(s, "-") {
		return true
	}

	return false
}

// looksLikeNumber checks if a string looks like a number.
func looksLikeNumber(s string) bool {
	// Leading zeros indicate it's a string (except "0" itself)
	if len(s) > 1 && s[0] == '0' && s[1] >= '0' && s[1] <= '9' {
		return true // Quote it as it looks like a number but shouldn't be parsed as one
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// escapeString escapes special characters in a string.
// Order matters: escape backslashes FIRST, then control chars.
func escapeString(s string) string {
	result := s
	// Escape backslashes first to avoid double-escaping
	result = strings.ReplaceAll(result, "\\", "\\\\")
	// Then escape quotes and control characters
	result = strings.ReplaceAll(result, "\"", "\\\"")
	result = strings.ReplaceAll(result, "\t", "\\t")
	result = strings.ReplaceAll(result, "\n", "\\n")
	result = strings.ReplaceAll(result, "\r", "\\r")
	return result
}

// encodeKey encodes a map key, applying stricter rules than values.
func encodeKey(k string) string {
	if safeKey(k) {
		return k
	}
	return doubleQuote + escapeString(k) + doubleQuote
}

// safeKey checks if a key can be used unquoted.
// Keys must match ^[A-Z_][\w.]*$ to be unquoted.
func safeKey(k string) bool {
	if k == "" {
		return false
	}

	// First character must be letter or underscore
	first := rune(k[0])
	if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z') || first == '_') {
		return false
	}

	// Remaining characters must be alphanumeric, underscore, or dot
	for _, ch := range k[1:] {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '.') {
			return false
		}
	}

	return true
}

// unescapeString unescapes special characters in a string.
// Order matters: unescape control chars FIRST, then backslashes.
func unescapeString(s string) string {
	result := s
	// Unescape control characters and quotes first
	result = strings.ReplaceAll(result, "\\\"", "\"")
	result = strings.ReplaceAll(result, "\\t", "\t")
	result = strings.ReplaceAll(result, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\r", "\r")
	// Unescape backslashes last
	result = strings.ReplaceAll(result, "\\\\", "\\")
	return result
}

// validateAndUnescape validates escape sequences and unescapes a string.
func validateAndUnescape(s string) (string, error) {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' {
			if i+1 >= len(s) {
				return "", &DecodeError{Message: "unterminated string: unexpected end in escape sequence"}
			}
			next := s[i+1]
			switch next {
			case '\\', '"', 'n', 'r', 't':
				// Valid escape sequence
				if next == '\\' {
					result.WriteByte('\\')
				} else if next == '"' {
					result.WriteByte('"')
				} else if next == 'n' {
					result.WriteByte('\n')
				} else if next == 'r' {
					result.WriteByte('\r')
				} else if next == 't' {
					result.WriteByte('\t')
				}
				i += 2
			default:
				return "", &DecodeError{Message: fmt.Sprintf("invalid escape sequence: \\%c", next)}
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	
	// Check for odd number of trailing backslashes (unterminated)
	// This shouldn't happen if we've processed correctly, but validate
	trailingBackslashes := 0
	for j := len(s) - 1; j >= 0 && s[j] == '\\'; j-- {
		trailingBackslashes++
	}
	if trailingBackslashes%2 == 1 {
		return "", &DecodeError{Message: "unterminated string: odd number of trailing backslashes"}
	}
	
	return result.String(), nil
}

// parseValue parses a primitive value from a string.
func parseValue(s string) (Value, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return "", nil
	}

	// Check for null
	if s == nullLiteral {
		return nil, nil
	}

	// Check for boolean
	if s == trueLiteral {
		return true, nil
	}
	if s == falseLiteral {
		return false, nil
	}

	// Check for quoted string
	if strings.HasPrefix(s, doubleQuote) {
		// Check for unterminated string
		if !strings.HasSuffix(s, doubleQuote) || len(s) < 2 {
			return "", &DecodeError{Message: "unterminated string: missing closing quote"}
		}
		content := s[1 : len(s)-1]
		// Validate escape sequences
		unescaped, err := validateAndUnescape(content)
		if err != nil {
			return "", err
		}
		return unescaped, nil
	}

	// Try to parse as number
	if num, err := parseNumber(s); err == nil {
		return num, nil
	}

	// Otherwise, it's an unquoted string
	return s, nil
}

// parseNumber attempts to parse a string as a number.
func parseNumber(s string) (Value, error) {
	// Leading zeros indicate it should remain a string (except "0" and negative numbers)
	if len(s) > 1 && s[0] == '0' && s[1] >= '0' && s[1] <= '9' {
		return nil, fmt.Errorf("not a number (leading zero)")
	}
	
	// Try integer first
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	return nil, fmt.Errorf("not a number")
}