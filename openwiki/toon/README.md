# TOON Format — Token-Oriented Object Notation

TOON is a compact, human-readable encoding of JSON data designed to minimize token usage for LLM interactions. It typically saves 30–60% tokens compared to formatted JSON.

The full format specification and TypeScript reference implementation are available at [toonformat.dev](https://toonformat.dev). The Maestro planning server previously used TOON for plan storage but has since migrated to JSON.

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

## Usage in Symphony

The TOON format is available as a skill (`toon`) for use by LLM agents when generating compact structured output.
Maestro now stores plans as JSON on disk; it previously used TOON as its on-disk format.
