package toon

import (
	"math"
	"reflect"
)

// isPrimitive checks if a value is a primitive type (nil, bool, number, or string).
func isPrimitive(v Value) bool {
	if v == nil {
		return true
	}

	switch v.(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, string:
		return true
	default:
		return false
	}
}

// isMap checks if a value is a map.
func isMap(v Value) bool {
	if v == nil {
		return false
	}
	// Check for OrderedMap
	if _, ok := v.(OrderedMap); ok {
		return true
	}
	if _, ok := v.(*OrderedMap); ok {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Map
}

// isList checks if a value is a list (slice or array).
func isList(v Value) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	k := rv.Kind()
	return k == reflect.Slice || k == reflect.Array
}

// allPrimitives checks if all elements in a slice are primitives.
func allPrimitives(v Value) bool {
	if !isList(v) {
		return false
	}

	rv := reflect.ValueOf(v)
	for i := 0; i < rv.Len(); i++ {
		if !isPrimitive(rv.Index(i).Interface()) {
			return false
		}
	}
	return true
}

// allMaps checks if all elements in a slice are maps.
func allMaps(v Value) bool {
	if !isList(v) {
		return false
	}

	rv := reflect.ValueOf(v)
	for i := 0; i < rv.Len(); i++ {
		if !isMap(rv.Index(i).Interface()) {
			return false
		}
	}
	return true
}

// sameKeys checks if all maps in a slice have the same keys.
func sameKeys(v Value) bool {
	if !isList(v) {
		return false
	}

	rv := reflect.ValueOf(v)
	if rv.Len() == 0 {
		return true
	}

	// Get keys from first map
	first := rv.Index(0).Interface()
	if !isMap(first) {
		return false
	}

	firstKeys := getMapKeys(first)
	if len(firstKeys) == 0 {
		return true
	}

	// Check all other maps have the same keys
	for i := 1; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if !isMap(item) {
			return false
		}

		keys := getMapKeys(item)
		if !sameStringSlice(firstKeys, keys) {
			return false
		}
	}

	return true
}

// getMapKeys returns sorted string keys from a map.
func getMapKeys(v Value) []string {
	// Handle OrderedMap
	if orderedMap, ok := v.(OrderedMap); ok {
		keys := make([]string, len(orderedMap.Keys()))
		copy(keys, orderedMap.Keys())
		sortStrings(keys)
		return keys
	}
	if orderedMapPtr, ok := v.(*OrderedMap); ok {
		keys := make([]string, len(orderedMapPtr.Keys()))
		copy(keys, orderedMapPtr.Keys())
		sortStrings(keys)
		return keys
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil
	}

	keys := make([]string, 0, rv.Len())
	for _, k := range rv.MapKeys() {
		keys = append(keys, k.String())
	}

	// Sort for consistent comparison
	sortStrings(keys)
	return keys
}

// sameStringSlice checks if two string slices contain the same elements (order-independent).
func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create map for O(1) lookup
	m := make(map[string]bool, len(a))
	for _, s := range a {
		m[s] = true
	}

	for _, s := range b {
		if !m[s] {
			return false
		}
	}

	return true
}

// sortStrings sorts a slice of strings in place (simple bubble sort for small slices).
func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// sortKeysWithArraysFirst sorts keys alphabetically but puts array fields first in list items
func sortKeysWithArraysFirst(keys []string, rv reflect.Value) {
	// Separate array keys from non-array keys
	arrayKeys := []string{}
	otherKeys := []string{}
	
	for _, k := range keys {
		mapKey := reflect.ValueOf(k)
		val := rv.MapIndex(mapKey).Interface()
		if isList(val) {
			arrayKeys = append(arrayKeys, k)
		} else {
			otherKeys = append(otherKeys, k)
		}
	}
	
	// Sort each group
	sortStrings(arrayKeys)
	sortStrings(otherKeys)
	
	// Rebuild keys with arrays first
	copy(keys, arrayKeys)
	copy(keys[len(arrayKeys):], otherKeys)
}

// normalize normalizes a value for encoding, converting to JSON-compatible types.
func normalize(v Value) Value {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case bool, string:
		return val

	case int, int8, int16, int32, int64:
		return reflect.ValueOf(val).Int()

	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(val).Uint())

	case float32:
		return normalizeFloat(float64(val))

	case float64:
		return normalizeFloat(val)

	case []interface{}:
		result := make([]Value, len(val))
		for i, item := range val {
			result[i] = normalize(item)
		}
		return result

	case map[string]interface{}:
		result := make(map[string]Value, len(val))
		for k, item := range val {
			result[k] = normalize(item)
		}
		return result

	case OrderedMap:
		// Preserve OrderedMap and normalize its values
		result := NewOrderedMap()
		for _, k := range val.Keys() {
			if v, ok := val.Get(k); ok {
				result.Set(k, normalize(v))
			}
		}
		return *result

	case *OrderedMap:
		// Preserve OrderedMap and normalize its values
		result := NewOrderedMap()
		for _, k := range val.Keys() {
			if v, ok := val.Get(k); ok {
				result.Set(k, normalize(v))
			}
		}
		return *result

	default:
		// Handle reflection-based conversion for other types
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			length := rv.Len()
			result := make([]Value, length)
			for i := 0; i < length; i++ {
				result[i] = normalize(rv.Index(i).Interface())
			}
			return result

		case reflect.Map:
			result := make(map[string]Value)
			for _, k := range rv.MapKeys() {
				key := k.String()
				result[key] = normalize(rv.MapIndex(k).Interface())
			}
			return result

		default:
			// Unsupported type, return nil
			return nil
		}
	}
}

// normalizeFloat handles special float values.
func normalizeFloat(f float64) Value {
	// Handle negative zero - normalize to 0
	if f == 0 && math.Signbit(f) {
		return int64(0)
	}

	// Handle NaN and Infinity
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}

	// Check if it's a whole number
	if f == math.Trunc(f) && f >= math.MinInt64 && f <= math.MaxInt64 {
		return int64(f)
	}

	return f
}

// toFloat64 converts a numeric value to float64.
func toFloat64(v Value) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	default:
		return 0, false
	}
}

// toInt64 converts a numeric value to int64.
func toInt64(v Value) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case uint:
		return int64(val), true
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		if val <= math.MaxInt64 {
			return int64(val), true
		}
		return 0, false
	case float32:
		if val == float32(int64(val)) {
			return int64(val), true
		}
		return 0, false
	case float64:
		if val == float64(int64(val)) {
			return int64(val), true
		}
		return 0, false
	default:
		return 0, false
	}
}