package toon

import (
	"reflect"
	"strconv"
	"strings"
)

// encodeArray encodes an array to TOON format.
func encodeArray(w *writer, key string, v Value, depth int, opts *EncodeOptions) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return &EncodeError{Message: "not an array", Value: v}
	}

	// Detect format
	format := detectArrayFormat(v)

	switch format {
	case arrayFormatEmpty:
		return encodeEmptyArray(w, key, depth, opts)
	case arrayFormatInline:
		return encodeInlineArray(w, key, v, depth, opts)
	case arrayFormatTabular:
		return encodeTabularArray(w, key, v, depth, opts)
	case arrayFormatList:
		return encodeListArray(w, key, v, depth, opts)
	default:
		return &EncodeError{Message: "unknown array format", Value: v}
	}
}

// detectArrayFormat determines the appropriate array encoding format.
func detectArrayFormat(v Value) arrayFormat {
	rv := reflect.ValueOf(v)
	length := rv.Len()

	if length == 0 {
		return arrayFormatEmpty
	}

	// Check if all elements are primitives
	if allPrimitives(v) {
		return arrayFormatInline
	}

	// Check if all elements are maps with same keys and all primitive values
	if allMaps(v) && sameKeys(v) {
		// Check if all values in all maps are primitives
		allPrim := true
		for i := 0; i < length; i++ {
			item := rv.Index(i).Interface()

			// Handle OrderedMap vs regular map
			if orderedMap, ok := item.(OrderedMap); ok {
				for _, k := range orderedMap.Keys() {
					val, _ := orderedMap.Get(k)
					if !isPrimitive(val) {
						allPrim = false
						break
					}
				}
			} else if orderedMapPtr, ok := item.(*OrderedMap); ok {
				for _, k := range orderedMapPtr.Keys() {
					val, _ := orderedMapPtr.Get(k)
					if !isPrimitive(val) {
						allPrim = false
						break
					}
				}
			} else {
				itemRv := reflect.ValueOf(item)
				if itemRv.Kind() == reflect.Map {
					for _, k := range itemRv.MapKeys() {
						val := itemRv.MapIndex(k).Interface()
						if !isPrimitive(val) {
							allPrim = false
							break
						}
					}
				}
			}
			if !allPrim {
				break
			}
		}
		if allPrim {
			return arrayFormatTabular
		}
	}

	// Otherwise use list format
	return arrayFormatList
}

// encodeEmptyArray encodes an empty array.
func encodeEmptyArray(w *writer, key string, depth int, opts *EncodeOptions) error {
	lengthMarker := formatLengthMarker(0, opts.LengthMarker)
	
	if key != "" {
		// Empty array syntax: key[0]: (no space after colon)
		w.push(key+openBracket+lengthMarker+closeBracket+colon, depth)
	} else {
		w.push(openBracket+lengthMarker+closeBracket+colon, depth)
	}
	
	return nil
}

// encodeInlineArray encodes an array of primitives in inline format.
func encodeInlineArray(w *writer, key string, v Value, depth int, opts *EncodeOptions) error {
	rv := reflect.ValueOf(v)
	length := rv.Len()
	
	lengthMarker := formatLengthMarker(length, opts.LengthMarker)
	delimiterMarker := ""
	if opts.Delimiter != comma {
		delimiterMarker = opts.Delimiter
	}

	// Encode values
	values := make([]string, length)
	for i := 0; i < length; i++ {
		item := rv.Index(i).Interface()
		encoded, err := encodePrimitive(item, opts.Delimiter)
		if err != nil {
			return err
		}
		values[i] = encoded
	}

	joined := strings.Join(values, opts.Delimiter)

	if key != "" {
		line := key + openBracket + lengthMarker + delimiterMarker + closeBracket + colon + space + joined
		w.push(line, depth)
	} else {
		line := openBracket + lengthMarker + delimiterMarker + closeBracket + colon + space + joined
		w.push(line, depth)
	}

	return nil
}

// encodeTabularArray encodes an array of uniform objects in tabular format.
func encodeTabularArray(w *writer, key string, v Value, depth int, opts *EncodeOptions) error {
	rv := reflect.ValueOf(v)
	length := rv.Len()
	
	if length == 0 {
		return encodeEmptyArray(w, key, depth, opts)
	}

	// Get keys from first object
	first := rv.Index(0).Interface()
	var keys []string

	// Handle OrderedMap vs regular map
	if orderedMap, ok := first.(OrderedMap); ok {
		keys = orderedMap.Keys()
	} else if orderedMapPtr, ok := first.(*OrderedMap); ok {
		keys = orderedMapPtr.Keys()
	} else {
		firstRv := reflect.ValueOf(first)
		keys = make([]string, 0, firstRv.Len())
		for _, k := range firstRv.MapKeys() {
			keys = append(keys, k.String())
		}
		sortStrings(keys)
	}

	// Format header
	lengthMarker := formatLengthMarker(length, opts.LengthMarker)
	delimiterMarker := ""
	if opts.Delimiter != comma {
		delimiterMarker = opts.Delimiter
	}

	encodedKeys := make([]string, len(keys))
	for i, k := range keys {
		encodedKeys[i] = encodeKey(k)
	}
	fields := strings.Join(encodedKeys, opts.Delimiter)

	var header string
	if key != "" {
		header = key + openBracket + lengthMarker + delimiterMarker + closeBracket + 
			openBrace + fields + closeBrace + colon
	} else {
		header = openBracket + lengthMarker + delimiterMarker + closeBracket + 
			openBrace + fields + closeBrace + colon
	}
	
	w.push(header, depth)

	// Format data rows
	for i := 0; i < length; i++ {
		item := rv.Index(i).Interface()

		values := make([]string, len(keys))
		for j, k := range keys {
			var val interface{}

			// Handle OrderedMap vs regular map
			if orderedMap, ok := item.(OrderedMap); ok {
				val, _ = orderedMap.Get(k)
			} else if orderedMapPtr, ok := item.(*OrderedMap); ok {
				val, _ = orderedMapPtr.Get(k)
			} else {
				itemRv := reflect.ValueOf(item)
				mapKey := reflect.ValueOf(k)
				val = itemRv.MapIndex(mapKey).Interface()
			}

			encoded, err := encodePrimitive(val, opts.Delimiter)
			if err != nil {
				return err
			}
			values[j] = encoded
		}

		row := strings.Join(values, opts.Delimiter)
		w.push(row, depth+1)
	}

	return nil
}

// encodeListArray encodes an array in list format (for mixed or non-uniform arrays).
func encodeListArray(w *writer, key string, v Value, depth int, opts *EncodeOptions) error {
	rv := reflect.ValueOf(v)
	length := rv.Len()
	
	lengthMarker := formatLengthMarker(length, opts.LengthMarker)
	delimiterMarker := ""
	if opts.Delimiter != comma {
		delimiterMarker = opts.Delimiter
	}

	// Write header
	var header string
	if key != "" {
		header = key + openBracket + lengthMarker + delimiterMarker + closeBracket + colon
	} else {
		header = openBracket + lengthMarker + delimiterMarker + closeBracket + colon
	}
	w.push(header, depth)

	// Encode each item
	for i := 0; i < length; i++ {
		item := rv.Index(i).Interface()
		if err := encodeListItem(w, item, depth+1, opts, true); err != nil {
			return err
		}
	}

	return nil
}

// encodeListItem encodes a single item in a list array.
func encodeListItem(w *writer, item Value, depth int, opts *EncodeOptions, isFirst bool) error {
	// Handle primitives
	if isPrimitive(item) {
		encoded, err := encodePrimitive(item, opts.Delimiter)
		if err != nil {
			return err
		}
		w.push(listItemPrefix+encoded, depth)
		return nil
	}

	// Handle arrays
	if isList(item) {
		rv := reflect.ValueOf(item)
		length := rv.Len()
		
		if length == 0 {
			lengthMarker := formatLengthMarker(0, opts.LengthMarker)
			delimiterMarker := ""
			if opts.Delimiter != comma {
				delimiterMarker = opts.Delimiter
			}
			// Empty array syntax without space
			w.push(listItemPrefix+openBracket+lengthMarker+delimiterMarker+closeBracket+colon, depth)
			return nil
		}
		
		// Check if all primitives for inline format
		if allPrimitives(item) {
			lengthMarker := formatLengthMarker(length, opts.LengthMarker)
			delimiterMarker := ""
			if opts.Delimiter != comma {
				delimiterMarker = opts.Delimiter
			}
			
			values := make([]string, length)
			for i := 0; i < length; i++ {
				val := rv.Index(i).Interface()
				encoded, err := encodePrimitive(val, opts.Delimiter)
				if err != nil {
					return err
				}
				values[i] = encoded
			}
			joined := strings.Join(values, opts.Delimiter)
			
			line := listItemPrefix + openBracket + lengthMarker + delimiterMarker + closeBracket + colon + space + joined
			w.push(line, depth)
			return nil
		}
		
		// Complex nested array - recursively encode with full multi-level support
		lengthMarker := formatLengthMarker(length, opts.LengthMarker)
		delimiterMarker := ""
		if opts.Delimiter != comma {
			delimiterMarker = opts.Delimiter
		}
		
		header := listItemPrefix + openBracket + lengthMarker + delimiterMarker + closeBracket + colon
		w.push(header, depth)
		
		// Recursively encode nested items supporting multi-level arrays and objects
		for i := 0; i < length; i++ {
			nested := rv.Index(i).Interface()
			// Full recursive encoding for any depth of nesting
			if err := encodeListItem(w, nested, depth+1, opts, false); err != nil {
				return err
			}
		}
		
		return nil
	}

	// Handle maps
	if isMap(item) {
		// Check if it's an OrderedMap - if so, preserve its key order
		var keys []string
		var itemRv reflect.Value

		if orderedMap, ok := item.(OrderedMap); ok {
			// Use the OrderedMap's preserved key order
			keys = orderedMap.Keys()
			itemRv = reflect.ValueOf(orderedMap.Values())
		} else if orderedMapPtr, ok := item.(*OrderedMap); ok {
			// Handle pointer to OrderedMap
			keys = orderedMapPtr.Keys()
			itemRv = reflect.ValueOf(orderedMapPtr.Values())
		} else {
			// Regular map - extract keys and sort them
			itemRv = reflect.ValueOf(item)
			keys = make([]string, 0, itemRv.Len())
			for _, k := range itemRv.MapKeys() {
				keys = append(keys, k.String())
			}
			// Sort keys, but put arrays first in list items
			sortKeysWithArraysFirst(keys, itemRv)
		}

		// Calculate alignment offset for subsequent keys (list marker "- " is 2 chars)
		alignmentOffset := 2
		
		for idx, k := range keys {
			mapKey := reflect.ValueOf(k)
			val := itemRv.MapIndex(mapKey).Interface()
			encodedKey := encodeKey(k)
			
			// First entry gets the list marker
			if idx == 0 {
				if isPrimitive(val) {
					encoded, err := encodePrimitive(val, opts.Delimiter)
					if err != nil {
						return err
					}
					w.push(listItemPrefix+encodedKey+colon+space+encoded, depth)
				} else if isList(val) {
					// Array on hyphen line - encode inline without line break
					if err := encodeArray(w, listItemPrefix+encodedKey, val, depth, opts); err != nil {
						return err
					}
				} else {
					// Complex value (object) - proper nesting support
					w.push(listItemPrefix+encodedKey+colon, depth)
					if err := encodeValue(w, "", val, depth+1, opts); err != nil {
						return err
					}
				}
			} else {
				// Subsequent entries are indented with alignment offset to match first key
				baseIndent := depth * opts.Indent
				alignedLine := strings.Repeat(" ", baseIndent+alignmentOffset) + encodedKey

				// Calculate effective depth for aligned content
				// alignedLine has (depth * Indent + alignmentOffset) spaces
				// So effective depth = depth + (alignmentOffset / Indent)
				effectiveDepth := depth + alignmentOffset/opts.Indent

				if isPrimitive(val) {
					encoded, err := encodePrimitive(val, opts.Delimiter)
					if err != nil {
						return err
					}
					w.pushRaw(newline + alignedLine + colon + space + encoded)
				} else if isList(val) {
					// Array field - encode inline on same line
					// Don't use pushRaw for line break, let encodeArray handle it
					if err := encodeArray(w, alignedLine, val, depth, opts); err != nil {
						return err
					}
				} else {
					// Complex value (object) - proper nesting with correct depth
					w.pushRaw(newline + alignedLine + colon)
					if err := encodeValue(w, "", val, effectiveDepth+1, opts); err != nil {
						return err
					}
				}
			}
		}
		
		return nil
	}

	return &EncodeError{Message: "unsupported list item type", Value: item}
}

// formatLengthMarker formats the length marker with optional prefix.
func formatLengthMarker(length int, marker string) string {
	if marker == "" {
		return strconv.Itoa(length)
	}
	return marker + strconv.Itoa(length)
}