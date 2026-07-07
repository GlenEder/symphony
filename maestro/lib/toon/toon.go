// Package toon implements encoding and decoding of TOON (Token-Oriented Object Notation) format.
//
// TOON is a compact data format optimized for LLM token efficiency, achieving
// 30-60% token reduction compared to JSON while maintaining readability.
//
// Features:
//   - Token Efficient: 30-60% fewer tokens than JSON
//   - Human Readable: Indentation-based structure like YAML
//   - Three Array Formats: Inline, tabular, and list formats
//   - Type Safe: Strong typing with clear interfaces
//
// Author: Stefano Straus (https://github.com/sstraus)
// Copyright (c) 2025 Stefano Straus
//
// Basic usage:
//
//	// Encoding
//	data := map[string]interface{}{
//		"name": "Alice",
//		"age":  30,
//	}
//	encoded, err := toon.Marshal(data, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(encoded))
//	// Output: age: 30
//	//         name: Alice
//
//	// Decoding
//	input := `name: Alice
//	age: 30`
//	var result map[string]interface{}
//	err = toon.Unmarshal([]byte(input), &result, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("%+v\n", result)
//	// Output: map[age:30 name:Alice]
package toon

// Version is the current version of the TOON library.
const Version = "1.0.0"

// Marshal encodes a Go value to TOON format.
//
// The value v can be any JSON-compatible type: nil, bool, int, int64, float64,
// string, []interface{}, or map[string]interface{}.
//
// Options can be nil to use defaults, or specify custom encoding behavior.
//
// Example:
//
//	data := map[string]interface{}{
//		"tags": []interface{}{"go", "toon"},
//	}
//	encoded, err := toon.Marshal(data, nil)
//	// Output: tags[2]: go,toon
func Marshal(v interface{}, opts *EncodeOptions) ([]byte, error) {
	// Validate options
	opts = getEncodeOptions(opts)
	if err := validateEncodeOptions(opts); err != nil {
		return nil, err
	}

	// Normalize the value
	normalized := normalize(v)

	// Encode
	result, err := encode(normalized, opts)
	if err != nil {
		return nil, err
	}

	return []byte(result), nil
}

// Unmarshal decodes TOON format data into a Go value.
//
// The value pointed to by v will be populated with the decoded data.
// v must be a pointer to a value of the appropriate type.
//
// Options can be nil to use defaults, or specify custom decoding behavior.
//
// Example:
//
//	input := []byte("name: Alice\nage: 30")
//	var result map[string]interface{}
//	err := toon.Unmarshal(input, &result, nil)
//	// result: map[string]interface{}{"name": "Alice", "age": 30}
func Unmarshal(data []byte, v interface{}, opts *DecodeOptions) error {
	// Validate options
	opts = getDecodeOptions(opts)
	if err := validateDecodeOptions(opts); err != nil {
		return err
	}

	// Decode
	result, err := decode(string(data), opts)
	if err != nil {
		return err
	}

	// Assign result to v
	return assignResult(result, v)
}

// assignResult assigns the decoded result to the target variable.
func assignResult(result Value, v interface{}) error {
	switch target := v.(type) {
	case *map[string]interface{}:
		if m, ok := result.(map[string]Value); ok {
			// Convert map[string]Value to map[string]interface{}
			converted := make(map[string]interface{}, len(m))
			for k, val := range m {
				converted[k] = val
			}
			*target = converted
		} else {
			return &DecodeError{Message: "cannot assign non-map to map target"}
		}
	case *[]interface{}:
		if arr, ok := result.([]Value); ok {
			// Convert []Value to []interface{}
			converted := make([]interface{}, len(arr))
			for i, val := range arr {
				converted[i] = val
			}
			*target = converted
		} else {
			return &DecodeError{Message: "cannot assign non-array to array target"}
		}
	case *interface{}:
		*target = result
	default:
		return &DecodeError{Message: "unsupported target type"}
	}
	return nil
}