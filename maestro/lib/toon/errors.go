package toon

import "fmt"

// EncodeError represents an error that occurred during encoding.
type EncodeError struct {
	Message string
	Value   Value
	Cause   error
}

// Error implements the error interface.
func (e *EncodeError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Value)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *EncodeError) Unwrap() error {
	return e.Cause
}

// DecodeError represents an error that occurred during decoding.
type DecodeError struct {
	Message string
	Input   string
	Line    int
	Column  int
	Token   string
	Context string
	Cause   error
}

// Error implements the error interface.
func (e *DecodeError) Error() string {
	msg := e.Message

	if e.Line > 0 || e.Column > 0 {
		if e.Line > 0 && e.Column > 0 {
			msg = fmt.Sprintf("%s at line %d, column %d", msg, e.Line, e.Column)
		} else if e.Line > 0 {
			msg = fmt.Sprintf("%s at line %d", msg, e.Line)
		} else {
			msg = fmt.Sprintf("%s at column %d", msg, e.Column)
		}
	}

	if e.Token != "" {
		msg = fmt.Sprintf("%s (token: '%s')", msg, e.Token)
	}

	if e.Context != "" {
		msg = fmt.Sprintf("%s\n\nContext:\n%s", msg, e.Context)
	}

	return msg
}

// Unwrap returns the underlying error.
func (e *DecodeError) Unwrap() error {
	return e.Cause
}