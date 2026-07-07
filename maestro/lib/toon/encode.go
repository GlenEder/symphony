package toon

import "fmt"

// encode encodes a normalized value to TOON format string.
func encode(v Value, opts *EncodeOptions) (string, error) {
	w := newWriter(opts.Indent)
	
	if err := encodeValue(w, "", v, 0, opts); err != nil {
		return "", err
	}
	
	return w.String(), nil
}

// encodeValue encodes a value with an optional key.
func encodeValue(w *writer, key string, v Value, depth int, opts *EncodeOptions) error {
	if v == nil {
		if key != "" {
			w.push(key+colon+space+nullLiteral, depth)
		} else {
			w.push(nullLiteral, depth)
		}
		return nil
	}

	// Handle primitives
	if isPrimitive(v) {
		encoded, err := encodePrimitive(v, opts.Delimiter)
		if err != nil {
			return err
		}
		if key != "" {
			w.push(key+colon+space+encoded, depth)
		} else {
			if depth == 0 && w.Len() == 0 {
				// Root primitive without indentation
				w.pushRaw(encoded)
			} else {
				w.push(encoded, depth)
			}
		}
		return nil
	}

	// Handle maps
	if isMap(v) {
		// Convert OrderedMap to regular map for flattening if needed
		var mapValue map[string]Value
		if orderedMap, ok := v.(OrderedMap); ok {
			// Convert map[string]interface{} to map[string]Value
			mapValue = make(map[string]Value)
			for k, val := range orderedMap.Values() {
				mapValue[k] = val
			}
		} else if orderedMapPtr, ok := v.(*OrderedMap); ok {
			// Convert map[string]interface{} to map[string]Value
			mapValue = make(map[string]Value)
			for k, val := range orderedMapPtr.Values() {
				mapValue[k] = val
			}
		} else if m, ok := v.(map[string]Value); ok {
			mapValue = m
		} else {
			return &EncodeError{
				Message: "unsupported map type",
				Value:   v,
			}
		}

		// Apply flattening if enabled (only at root level, depth==0)
		// After flattening, we don't want to flatten nested maps again
		if opts.FlattenPaths && depth == 0 {
			flattened, err := flattenObject(mapValue, "", 0, opts)
			if err != nil {
				return fmt.Errorf("flatten failed: %w", err)
			}
			// Create new options with flattening disabled for nested encoding
			nestedOpts := *opts
			nestedOpts.FlattenPaths = false
			return encodeObject(w, key, flattened, depth, &nestedOpts)
		}
		return encodeObject(w, key, v, depth, opts)
	}

	// Handle arrays
	if isList(v) {
		return encodeArray(w, key, v, depth, opts)
	}

	return &EncodeError{
		Message: "unsupported type",
		Value:   v,
	}
}