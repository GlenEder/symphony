# TOON Format — Token-Oriented Object Notation

TOON is a compact, human-readable encoding of JSON data designed to minimize token usage for LLM interactions. It typically saves 30–60% tokens compared to formatted JSON.

**This project includes a Go implementation of TOON** at `maestro/lib/toon/` used by the Maestro planning server to store and load plan files. The full format specification and TypeScript reference implementation are available at [toonformat.dev](https://toonformat.dev).

## Overview

| Property | Value |
|----------|-------|
| Format version | v2.3.0 |
| Spec version | v3.3 |
| Media type | `text/toon` |
| File extension | `.toon` |
| Encoding | UTF-8 with LF line endings |
| License | MIT |

## Key Features

- **Compact tabular arrays** — uniform arrays of objects use CSV-style rows with explicit headers
- **Token efficiency** — minimal quoting, deterministic structure, 30-60% savings vs JSON
- **Built-in validation** — array headers declare length `[N]` and fields `{fields}` for truncation detection
- **LLM-friendly** — indentation-based nesting without braces, easy for models to parse and generate
- **Lossless JSON round-trip** — any valid JSON can be encoded to TOON and decoded back

## Syntax Highlights

### Objects

```
name: Ada
active: true
nested_key:
  inner: value
```

### Primitive Arrays

```
tags[3]: admin,ops,dev
```

### Tabular Arrays (the main feature)

```
users[2]{id,name,role}:
  1,Alice Admin,admin
  2,"Bob Smith",user
```

The header `users[2]{id,name,role}:` declares:
- `[2]` = row count
- `{id,name,role}` = column order
- Commas as default delimiter (tabs and pipes also supported)

Tabular format requires: identical fields across objects, primitive values only, at least one key.

### Mixed Arrays (list format)

```
items[2]:
  - id: 1
    name: First
  - name: Second
    extra: true
```

## Quoting Rules

TOON quotes strings only when necessary:

**Must quote if value:**
- Is empty (`""`)
- Has leading/trailing whitespace
- Equals `true`, `false`, `null`
- Looks like a number (`"42"`)
- Contains `:`, `"`, `\`, `[`, `]`, `{`, `}`
- Contains the active delimiter
- Starts with `-`

**Safe without quotes:** Unicode, emoji, internal spaces, most text.

## Key Folding

Collapses chains of single-key objects into dotted paths:

```toon
# Without folding
data:
  metadata:
    items[2]: a,b

# With folding (safe mode)
data.metadata.items[2]: a,b
```

## Go Library

The project at `maestro/lib/toon/` is a Go port of the TypeScript `@toon-format/toon` library.

### Source Files

| File | Purpose |
|------|---------|
| `toon.go` | Top-level encode/decode API |
| `encode.go` | JSON-to-TOON encoding |
| `decode.go` | TOON-to-JSON decoding |
| `parser.go` | Core parser structure |
| `structural_parser.go` | Low-level structural parser (48KB, the largest file) |
| `primitives.go` | Primitive value parsing |
| `arrays.go` | Array (tabular + list) handling |
| `objects.go` | Object parsing and rendering |
| `orderedmap.go` | Ordered map for stable key iteration |
| `types.go` | Type definitions |
| `constants.go` | Configuration constants |
| `options.go` | Encode/Decode options |
| `errors.go` | Error types |
| `writer.go` | Output writer |
| `utils.go` | Utility functions |

### Usage in Maestro

Maestro uses TOON for plan file storage. The round-trip:

```
Plan struct → json.Marshal → map[string]any → toon.Marshal → .toon file
.toon file → toon.Unmarshal → map[string]any → json.Marshal → json.Unmarshal → Plan struct
```

This is done in `maestro/store.go`:

```go
// Decoding (loadFile → decodePlan)
func decodePlan(data []byte) (*Plan, error) {
    var raw map[string]any
    toon.Unmarshal(data, &raw, &toon.DecodeOptions{Strict: false})
    js, _ := json.Marshal(raw)
    var plan Plan
    json.Unmarshal(js, &plan)
    return &plan, nil
}

// Encoding (persistPlan)
toonBytes, _ := toon.Marshal(raw, &toon.EncodeOptions{Indent: 2})
os.WriteFile(path, toonBytes, 0644)
```

### Important Notes for Developers

- The `replace` directive in `maestro/go.mod` maps the import `github.com/sstraus/toon_go/toon` to the local `./lib/toon` directory.
- The Go library is **not tested** — there are no `*_test.go` files in `maestro/lib/toon/`.
- The parser has a known behavior: when a list-item object has a list-format array as its first field, trailing sibling fields at the same indentation level can be absorbed into the last array item. See the [TOON skill](../skills/README.md) for the workaround (place `type`/`heading` before items).
- Strict mode is disabled (`Strict: false`) in Maestro's plan decoding to tolerate hand-edited files.
