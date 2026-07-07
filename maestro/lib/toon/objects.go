package toon

import (
	"fmt"
	"reflect"
	"strings"
)

// encodeObject encodes a map to TOON format.
func encodeObject(w *writer, key string, v Value, depth int, opts *EncodeOptions) error {
	// Handle OrderedMap - preserve key order
	var keys []string
	var rv reflect.Value

	if orderedMap, ok := v.(OrderedMap); ok {
		keys = orderedMap.Keys()
		rv = reflect.ValueOf(orderedMap.Values())
	} else if orderedMapPtr, ok := v.(*OrderedMap); ok {
		keys = orderedMapPtr.Keys()
		rv = reflect.ValueOf(orderedMapPtr.Values())
	} else {
		// Regular map
		rv = reflect.ValueOf(v)
		if rv.Kind() != reflect.Map {
			return &EncodeError{Message: "not a map", Value: v}
		}

		// Get and sort keys
		keys = make([]string, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			keys = append(keys, k.String())
		}
		sortStrings(keys)
	}

	// Handle empty object
	if len(keys) == 0 {
		if key != "" {
			w.push(key+colon, depth)
		}
		return nil
	}

	// If this is a keyed object, write the key first
	if key != "" {
		w.push(key+colon, depth)
		depth++
	}

	// Encode each key-value pair
	for _, k := range keys {
		var mapValue interface{}

		// Get value based on type
		if rv.Kind() == reflect.Map {
			mapKey := reflect.ValueOf(k)
			mapValue = rv.MapIndex(mapKey).Interface()
		} else {
			// For OrderedMap, rv is already the values map
			mapValue = rv.Interface().(map[string]interface{})[k]
		}

		encodedKey := encodeKey(k)

		if err := encodeValue(w, encodedKey, mapValue, depth, opts); err != nil {
			return err
		}
	}

	return nil
}

// flattenObject flattens nested maps into dotted key notation.
// Example: {"a":{"b":1}} becomes {"a.b":1}
func flattenObject(obj map[string]Value, currentPath string, depth int, opts *EncodeOptions) (map[string]Value, error) {
	result := make(map[string]Value)

	// First pass: check if any potential flattened paths would collide with literal keys
	// This implements "safe mode" collision detection
	literalKeys := make(map[string]bool)

	// Collect all literal keys at the current level
	for key := range obj {
		literalKeys[key] = true
	}

	// Check for collisions: if any potential flattened path matches a literal key, don't flatten that branch
	hasCollision := false
	for key, value := range obj {
		if isMap(value) {
			var nestedMap map[string]Value
			if orderedMap, ok := value.(OrderedMap); ok {
				nestedMap = make(map[string]Value)
				for k, v := range orderedMap.Values() {
					nestedMap[k] = v
				}
			} else if orderedMapPtr, ok := value.(*OrderedMap); ok {
				nestedMap = make(map[string]Value)
				for k, v := range orderedMapPtr.Values() {
					nestedMap[k] = v
				}
			} else if m, ok := value.(map[string]Value); ok {
				nestedMap = m
			}
			if nestedMap != nil {
				// Collect potential paths from this nested structure
				tempPaths := make(map[string]bool)
				var collectTemp func(m map[string]Value, prefix string)
				collectTemp = func(m map[string]Value, prefix string) {
					for k, v := range m {
						var path string
						if prefix == "" {
							path = k
						} else {
							path = prefix + "." + k
						}
						tempPaths[path] = true
						if isMap(v) {
							var nm map[string]Value
							if om, ok := v.(OrderedMap); ok {
								nm = make(map[string]Value)
								for k2, v2 := range om.Values() {
									nm[k2] = v2
								}
							} else if omp, ok := v.(*OrderedMap); ok {
								nm = make(map[string]Value)
								for k2, v2 := range omp.Values() {
									nm[k2] = v2
								}
							} else if m2, ok := v.(map[string]Value); ok {
								nm = m2
							}
							if nm != nil {
								collectTemp(nm, path)
							}
						}
					}
				}
				collectTemp(nestedMap, key)

				// Check if any of these paths collide with literal keys
				for path := range tempPaths {
					if literalKeys[path] {
						hasCollision = true
						break
					}
				}
			}
		}
		if hasCollision {
			break
		}
	}

	// If there's a collision, don't flatten at all
	// In strict mode, return an error; otherwise return original structure
	if hasCollision {
		if opts.Strict {
			return nil, &EncodeError{
				Message: "key collision: flattened path would conflict with existing literal key",
				Value:   obj,
			}
		}
		return obj, nil
	}

	// Helper to check if a nested structure contains any keys that need quoting
	// In safe mode, a key needs quoting if it's not a "safe key" (can't be used unquoted)
	var containsQuotedKeys func(m map[string]Value) bool
	containsQuotedKeys = func(m map[string]Value) bool {
		for k, v := range m {
			if !safeKey(k) {
				return true
			}
			if isMap(v) {
				var nm map[string]Value
				if om, ok := v.(OrderedMap); ok {
					nm = make(map[string]Value)
					for k2, v2 := range om.Values() {
						nm[k2] = v2
					}
				} else if omp, ok := v.(*OrderedMap); ok {
					nm = make(map[string]Value)
					for k2, v2 := range omp.Values() {
						nm[k2] = v2
					}
				} else if m2, ok := v.(map[string]Value); ok {
					nm = m2
				}
				if nm != nil && containsQuotedKeys(nm) {
					return true
				}
			}
		}
		return false
	}

	for key, value := range obj {
		// Skip flattening if key is not safe for use in dotted paths (would need quoting)
		isKeySafe := safeKey(key)

		// Build the full path
		var fullPath string
		if currentPath == "" {
			fullPath = key
		} else {
			fullPath = currentPath + "." + key
		}

		// Check if we should continue flattening
		// FlattenDepth limits the number of segments in the resulting dotted key
		// Count current segments in fullPath
		segmentCount := 1 // At least one segment (the current key)
		if currentPath != "" {
			segmentCount = len(strings.Split(currentPath, ".")) + 1
		}

		shouldFlatten := isMap(value) &&
			isKeySafe &&
			opts.FlattenDepth > 0 &&
			segmentCount < opts.FlattenDepth

		// Additionally, check if the nested structure contains any keys that need quoting
		// If so, don't flatten this branch at all
		if shouldFlatten && isMap(value) {
			var nestedMap map[string]Value
			if orderedMap, ok := value.(OrderedMap); ok {
				nestedMap = make(map[string]Value)
				for k, v := range orderedMap.Values() {
					nestedMap[k] = v
				}
			} else if orderedMapPtr, ok := value.(*OrderedMap); ok {
				nestedMap = make(map[string]Value)
				for k, v := range orderedMapPtr.Values() {
					nestedMap[k] = v
				}
			} else if m, ok := value.(map[string]Value); ok {
				nestedMap = m
			}
			if nestedMap != nil && containsQuotedKeys(nestedMap) {
				shouldFlatten = false
			}
		}

		if shouldFlatten {
			// Check for empty nested map
			var isEmpty bool
			var nestedMap map[string]Value

			if orderedMap, ok := value.(OrderedMap); ok {
				isEmpty = orderedMap.Len() == 0
				nestedMap = make(map[string]Value)
				for k, v := range orderedMap.Values() {
					nestedMap[k] = v
				}
			} else if orderedMapPtr, ok := value.(*OrderedMap); ok {
				isEmpty = orderedMapPtr.Len() == 0
				nestedMap = make(map[string]Value)
				for k, v := range orderedMapPtr.Values() {
					nestedMap[k] = v
				}
			} else {
				rv := reflect.ValueOf(value)
				isEmpty = rv.Len() == 0
				nestedMap = value.(map[string]Value)
			}

			if isEmpty {
				// Empty object - add at current path
				if existing, exists := result[fullPath]; exists {
					if opts.Strict {
						return nil, &EncodeError{
							Message: fmt.Sprintf("key collision: %q", fullPath),
							Value:   existing,
						}
					}
				}
				result[fullPath] = value
				continue
			}

			// Recursively flatten nested maps up to depth limit
			nested, err := flattenObject(nestedMap, fullPath, depth+1, opts)
			if err != nil {
				return nil, err
			}
			// Merge nested results
			for k, v := range nested {
				if existing, exists := result[k]; exists {
					if opts.Strict {
						return nil, &EncodeError{
							Message: fmt.Sprintf("key collision: %q", k),
							Value:   existing,
						}
					}
					// Non-strict: last value wins
				}
				result[k] = v
			}
		} else {
			// Stop flattening - add the value at current path
			// The value should remain nested (not be flattened further)
			if existing, exists := result[fullPath]; exists {
				if opts.Strict {
					return nil, &EncodeError{
						Message: fmt.Sprintf("key collision: %q", fullPath),
						Value:   existing,
					}
				}
				// Non-strict: last value wins
			}
			result[fullPath] = value
		}
	}

	return result, nil
}