package toon

import "fmt"

// validateEncodeOptions validates and normalizes encoding options.
func validateEncodeOptions(opts *EncodeOptions) error {
	if opts == nil {
		return nil
	}

	// Validate indent
	if opts.Indent < 0 {
		return &EncodeError{
			Message: "indent must be non-negative",
			Value:   opts.Indent,
		}
	}

	// Validate delimiter
	if opts.Delimiter != "" && !isValidDelimiter(opts.Delimiter) {
		return &EncodeError{
			Message: fmt.Sprintf("invalid delimiter %q, must be one of: %q, %q, %q",
				opts.Delimiter, comma, tab, pipe),
			Value: opts.Delimiter,
		}
	}

	return nil
}

// validateDecodeOptions validates and normalizes decoding options.
func validateDecodeOptions(opts *DecodeOptions) error {
	if opts == nil {
		return nil
	}

	// Validate indent size
	if opts.IndentSize < 1 {
		return &DecodeError{
			Message: "indent_size must be positive",
		}
	}

	return nil
}

// isValidDelimiter checks if a delimiter is valid.
func isValidDelimiter(delimiter string) bool {
	for _, valid := range validDelimiters {
		if delimiter == valid {
			return true
		}
	}
	return false
}

// getEncodeOptions returns options with defaults applied.
func getEncodeOptions(opts *EncodeOptions) *EncodeOptions {
	if opts == nil {
		opts = &EncodeOptions{}
	}

	// Apply defaults
	result := &EncodeOptions{
		Indent:       opts.Indent,
		Delimiter:    opts.Delimiter,
		LengthMarker: opts.LengthMarker,
		FlattenPaths: opts.FlattenPaths,
		FlattenDepth: opts.FlattenDepth,
		Strict:       opts.Strict,
	}
	
	// Handle FlattenDepth defaults for infinite folding
	// -1 means "not set" (use infinite)
	// 0 means "explicitly disabled" (no folding)
	// >0 means fold up to that depth
	if result.FlattenPaths && result.FlattenDepth == -1 {
		// Not set - default to infinite folding
		result.FlattenDepth = 9999
	}

	if result.Indent == 0 {
		result.Indent = defaultIndent
	}

	if result.Delimiter == "" {
		result.Delimiter = defaultDelimiter
	}

	return result
}

// getDecodeOptions returns options with defaults applied.
func getDecodeOptions(opts *DecodeOptions) *DecodeOptions {
	if opts == nil {
		// Default to strict mode when no options provided
		return &DecodeOptions{
			Keys:        StringKeys,
			Strict:      true,
			IndentSize:  defaultIndent,
			ExpandPaths: "off",
		}
	}

	// Copy provided options
	result := &DecodeOptions{
		Keys:        opts.Keys,
		Strict:      opts.Strict,
		IndentSize:  opts.IndentSize,
		ExpandPaths: opts.ExpandPaths,
	}

	if result.IndentSize == 0 {
		result.IndentSize = defaultIndent
	}

	if result.ExpandPaths == "" {
		result.ExpandPaths = "off"
	}

	return result
}