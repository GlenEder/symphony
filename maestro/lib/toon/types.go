package toon

// Value represents any TOON-encodable value.
// Valid types are: nil, bool, int, int64, float64, string, []Value, map[string]Value
type Value interface{}

// EncodeOptions configures encoding behavior.
type EncodeOptions struct {
	// Indent specifies the number of spaces for indentation (default: 2)
	Indent int

	// Delimiter specifies the delimiter for array values: "," | "\t" | "|" (default: ",")
	Delimiter string

	// LengthMarker specifies the prefix for array length markers (default: "")
	// Example: "#" produces "[#3]:" instead of "[3]:"
	LengthMarker string

	// FlattenPaths enables flattening of nested objects to dotted notation (default: false)
	// Example: {"a":{"b":1}} becomes "a.b: 1"
	FlattenPaths bool

	// FlattenDepth limits flattening recursion depth (default: 0 = unlimited)
	// Only applies when FlattenPaths is true
	FlattenDepth int

	// Strict enables strict collision detection when flattening paths (default: false)
	// When true, returns error on key collisions; when false, last value wins
	Strict bool
}

// DecodeOptions configures decoding behavior.
type DecodeOptions struct {
	// Keys specifies how to decode map keys (default: StringKeys)
	Keys KeyMode

	// Strict enables strict mode validation (default: true)
	// In strict mode, indentation must be consistent and arrays must have correct lengths
	Strict bool

	// IndentSize specifies the expected indentation size in spaces (default: 2)
	// Only used in strict mode for validation
	IndentSize int

	// ExpandPaths controls dotted key expansion: "off", "safe" (default: "off")
	// "safe" expands dotted keys like "a.b.c" to nested objects {"a":{"b":{"c":...}}}
	// "off" treats dotted keys as literal strings
	ExpandPaths string
}

// KeyMode specifies how to decode map keys.
type KeyMode int

const (
	// StringKeys decodes all map keys as strings
	StringKeys KeyMode = iota
)

// arrayFormat determines the array encoding format.
type arrayFormat int

const (
	arrayFormatEmpty arrayFormat = iota
	arrayFormatInline
	arrayFormatTabular
	arrayFormatList
)

// rootType indicates the type of root value in TOON input.
type rootType int

const (
	rootTypeObject rootType = iota
	rootTypeArray
	rootTypePrimitive
)