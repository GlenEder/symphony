package toon

// TOON format constants
const (
	// List markers
	listItemMarker = "-"
	listItemPrefix = "- "

	// Structure characters
	colon        = ":"
	comma        = ","
	space        = " "
	pipe         = "|"
	tab          = "\t"
	newline      = "\n"
	openBracket  = "["
	closeBracket = "]"
	openBrace    = "{"
	closeBrace   = "}"
	openParen    = "("
	closeParen   = ")"
	doubleQuote  = "\""
	backslash    = "\\"

	// Literals
	nullLiteral  = "null"
	trueLiteral  = "true"
	falseLiteral = "false"

	// Default options
	defaultIndent    = 2
	defaultDelimiter = comma
)

// Valid delimiters for array values
var validDelimiters = []string{comma, tab, pipe}

// Structure characters that require quoting in keys and values
// Note: Comma is only special when it's the active delimiter
var structureChars = []string{
	colon,
	openBracket,
	closeBracket,
	openBrace,
	closeBrace,
	openParen,
	closeParen,
	doubleQuote,
	backslash,
}

// Control characters that need escaping
var controlChars = []string{"\n", "\r", "\t", "\b", "\f"}

// Escape sequence mappings
var escapeSequences = map[string]string{
	"\\":  "\\\\",
	"\"":  "\\\"",
	"\n":  "\\n",
	"\r":  "\\r",
	"\t":  "\\t",
}

// Unescape sequence mappings
var unescapeSequences = map[string]string{
	"\\\\": "\\",
	"\\\"": "\"",
	"\\n":  "\n",
	"\\r":  "\r",
	"\\t":  "\t",
}