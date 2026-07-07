package toon

// decode decodes a TOON format string to a value.
func decode(input string, opts *DecodeOptions) (Value, error) {
	if input == "" {
		return map[string]Value{}, nil
	}
	
	// Apply defaults to options
	opts = getDecodeOptions(opts)
	
	// Create structural parser
	sp := newStructuralParser(input, opts)
	
	// Parse the input
	result, err := sp.parse()
	if err != nil {
		return nil, err
	}
	
	return result, nil
}