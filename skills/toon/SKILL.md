---
name: toon
description: Token-Oriented Object Notation (TOON) — compact, schema-aware JSON encoding for LLM prompts
compatibility: opencode
---

## Purpose

**TOON** (Token-Oriented Object Notation) is a compact, human-readable encoding of the JSON data model that minimizes tokens and makes structure easy for models to follow. It is designed as a drop-in, lossless representation of JSON for LLM input, typically saving 30–60% tokens compared to formatted JSON.

TOON combines YAML's indentation-based structure for nested objects with a CSV-style tabular layout for uniform arrays. Its sweet spot is uniform arrays of objects (same fields per row), achieving CSV-like compactness while adding explicit structure that helps LLMs parse and validate data reliably.

**Key facts:**
- Format version: **v2.3.0**, Spec version: **v3.3**
- Media type: `text/toon`, file extension: `.toon`
- Always UTF-8 with LF line endings
- MIT License
- Website: https://toonformat.dev

## When to Use TOON

TOON excels with **uniform arrays of objects** where every item has the same fields. For LLM prompts, the format produces deterministic, minimally quoted text with built-in validation via explicit array lengths (`[N]`) and field headers (`{fields}`).

**Best for:**
- Passing structured datasets to LLMs (employee records, product catalogs, event logs, analytics)
- Getting LLMs to generate structured output (schema-aware generation)
- Reducing token costs in prompts that include tabular/mixed data
- Detecting truncation or malformed data in LLM output

**Consider alternatives when:**
- **Deeply nested or non-uniform structures** (tabular eligibility near 0%): JSON-compact may use fewer tokens
- **Semi-uniform arrays** (~40–60% tabular eligibility): Token savings diminish; JSON may be simpler
- **Pure tabular data**: CSV is smaller, though TOON adds minimal overhead (~5–10%) for structural guardrails
- **Latency-critical applications**: Some models may process compact JSON faster despite lower token counts

## Syntax Reference

### Objects

Simple key-value pairs use `key: value` syntax, one field per line. One space follows the colon.

```
id: 123
name: Ada
active: true
```

Nested objects add one indentation level (default: 2 spaces).

```
user:
  id: 123
  name: Ada
```

A key ending with `:` and no value opens a nested object — all lines at the next indentation level belong to it.

### Primitive Arrays (Inline)

Arrays of primitives (strings, numbers, booleans, null) are rendered inline with `[N]` length:

```
tags[3]: admin,ops,dev
```

### Tabular Arrays (Uniform Objects)

When all objects in an array share the same set of primitive-valued keys, TOON uses tabular format. The header declares the length, field names, and delimiter.

```
users[2]{id,name,role}:
  1,Alice Admin,admin
  2,"Bob Smith",user
```

The header `users[2]{id,name,role}:` declares:
- **Array length**: `[2]` = 2 rows
- **Field names**: `{id,name,role}` defines column order
- **Active delimiter**: comma by default

Each row has values in the same order as the field list. Values containing the active delimiter, colons, or other special characters must be quoted.

**Tabular format requires:** identical field sets across objects, primitive values only (no nested arrays/objects), and at least one key per object.

### Mixed and Non-Uniform Arrays (List Format)

Arrays that don't meet tabular requirements use list format with hyphen markers:

```
items[3]:
  - 1
  - a: 1
  - text
```

Each element starts with `-` at one indentation level deeper than the parent array header.

### Objects as List Items

When an array element is an object, it appears as a list item:

```
items[2]:
  - id: 1
    name: First
  - id: 2
    name: Second
    extra: true
```

When a list-item object has a tabular array as its first field, the tabular header appears on the hyphen line. Rows are indented two levels deeper, other fields one level deeper:

```
items[1]:
  - users[2]{id,name}:
      1,Ada
      2,Bob
    status: active
```

**Caution — field ordering with list-format arrays:** When a list-item object has a list-format array (`items[N]:` with `-` prefixed entries) as its first field, trailing sibling fields (e.g., `type:`, `heading:`) at the same indentation level as the array items will be absorbed into the last array item by the parser. They are treated as extra properties of that item rather than as fields of the parent object.

```
# WRONG — `type:` and `heading:` get absorbed into last item
  - items[3]:
    - a: 1
    - b: 2
    - c: 3
    type: some-type
    heading: Some Heading

# RIGHT — place type/heading before items
  - type: some-type
    heading: Some Heading
    items[3]:
      - a: 1
      - b: 2
      - c: 3
```

Tabular arrays (`items[N]{fields}:` with inline data rows) do not have this issue because the tabular parser stops at lines with unquoted colons.

### Root Forms

TOON can represent:
- **Root object** (most common): Fields at depth 0 with no parent key
- **Root array**: Begins with `[N]:` or `[N]{fields}:` at depth 0
- **Root primitive**: A single primitive value (string, number, boolean, null)

```
[3]: x,y,z     # Root array of primitives
```

### Empty Containers

```
items: []       # Empty array field
{}              # Empty object → empty document (no lines)
```

## Array Headers (Detailed)

### Header Syntax

```
key[N<delimiter?>]<{fields}>:
```

- `N` = non-negative integer length
- `delimiter` (optional): absent = comma (`,`), `\t` = tab, `|` = pipe
- `{fields}` (optional for tabular arrays)

The array length `[N]` helps LLMs validate structure and detect truncation.

### Delimiter Options

Comma (default):
```
items[2]{sku,name,qty,price}:
  A1,Widget,2,9.99
  B2,Gadget,1,14.5
```

Tab:
```
items[2	]{sku	name	qty	price}:
  A1	Widget	2	9.99
  B2	Gadget	1	14.5
```

Pipe:
```
items[2|]{sku|name|qty|price}:
  A1|Widget|2|9.99
  B2|Gadget|1|14.5
```

Tab delimiters often tokenize more efficiently than commas, reduce quote-escaping, and are recommended for maximum savings.

## Quoting Rules

TOON quotes strings **only when necessary** to maximize token efficiency.

**Strings MUST be quoted if they:**
- Are empty (`""`)
- Have leading or trailing whitespace
- Equal `true`, `false`, or `null` (case-sensitive)
- Look like numbers (`"42"`, `"-3.14"`, `"1e-6"`, `"05"`)
- Contain special characters: `:`, `"`, `\`, `[`, `]`, `{`, `}`, control chars (U+0000–U+001F)
- Contain the relevant delimiter (active delimiter in array scope, or document delimiter elsewhere)
- Equal `"-"` or start with `"-"` followed by any character

**Otherwise strings can be unquoted.** Unicode, emoji, and internal spaces are safe:

```
message: Hello 世界 👋
note: This has inner spaces
```

### Escape Sequences (in quoted strings)

| Character | Escape |
|-----------|--------|
| Backslash (`\`) | `\\` |
| Double quote (`"`) | `\"` |
| Newline (U+000A) | `\n` |
| Carriage return (U+000D) | `\r` |
| Tab (U+0009) | `\t` |
| Other U+0000–U+001F | `\uXXXX` |

Other escapes (e.g., `\x`, `\0`, `\b`) are invalid. Lone-surrogate `\uXXXX` values (U+D800–U+DFFF) are rejected.

## Key Folding (Optional)

Key folding collapses chains of single-key objects into dotted paths, reducing tokens for deeply nested data.

**Without folding:**
```
data:
  metadata:
    items[2]: a,b
```

**With folding (`keyFolding: 'safe'`):**
```
data.metadata.items[2]: a,b
```

**Requirements for folding:**
- Each object in the chain has exactly one key
- All segments match `^[A-Za-z_][A-Za-z0-9_]*$` (no dots, hyphens)
- No segment would require quoting
- No collision with existing sibling keys

**Round-trip:** When decoding, enable `expandPaths: 'safe'` to split dotted keys back into nested objects.

## Using TOON with LLMs

### Sending TOON as Input

Show the format instead of describing it. Wrap encoded data in a fenced code block:

```
Data is in TOON format (2-space indent, arrays show length and fields).

```toon
users[3]{id,name,role,lastLogin}:
  1,Alice,admin,"2025-01-15T10:30:00Z"
  2,Bob,user,"2025-01-14T15:22:00Z"
  3,Charlie,user,"2025-01-13T09:45:00Z"
```

Task: Summarize the user roles and their last activity.
```

Use ````toon` or ````yaml` for code fence labels — both work fine.

### Generating TOON from LLMs

Be explicit about the expected format. Show the header and let the model fill rows:

```
Data is in TOON format (2-space indent, arrays show length and fields).

```toon
users[3]{id,name,role,lastLogin}:
  1,Alice,admin,"2025-01-15T10:30:00Z"
  2,Bob,user,"2025-01-14T15:22:00Z"
  3,Charlie,user,"2025-01-13T09:45:00Z"
```

Task: Return only users with role "user" as TOON. Use the same header format. Set [N] to match the row count.
```

The model adjusts `[N]` and generates rows — reducing generation errors by not repeating keys.

### Validation (Strict Mode)

Always validate model-generated TOON with strict mode (default):

```ts
import { decode } from '@toon-format/toon'

try {
  const data = decode(modelOutput, { strict: true })
  // Success — data is valid
} catch (error) {
  // Malformed output — count mismatch, invalid escapes, etc.
}
```

Strict mode catches: array count mismatches, field width mismatches, indentation errors, invalid escape sequences, duplicate sibling keys, and path expansion conflicts.

### Delimiter Choices for Token Efficiency

Use tab delimiters for additional savings:

```ts
const toon = encode(data, { delimiter: '\t' })
```

Tell the model "fields are tab-separated" when using tabs.

### Streaming Large Outputs

For large datasets, use `encodeLines()` to stream TOON line-by-line:

```ts
import { encodeLines } from '@toon-format/toon'

for (const line of encodeLines(largeData, { delimiter: '\t' })) {
  process.stdout.write(`${line}\n`)
}
```

For consuming streaming LLM output incrementally:

```ts
import { decodeFromLines } from '@toon-format/toon'

const lines: string[] = []
for await (const chunk of modelStream) {
  // buffer and split by newlines
  lines.push(...bufferLines(chunk))
}
const data = decodeFromLines(lines)
```

## TypeScript/JavaScript API

The `@toon-format/toon` package provides:

| Function | Description |
|----------|-------------|
| `encode(input, options?)` | Encode JSON value to TOON string |
| `decode(input, options?)` | Decode TOON string to JSON value |
| `encodeLines(input, options?)` | Generator that yields TOON lines |
| `decodeFromLines(lines, options?)` | Decode array of TOON lines |
| `decodeStream(source, options?)` | Decode from async iterable |

**Options:**
- `indent`: Indentation size (default: 2)
- `delimiter`: Array delimiter — `,`, `\t`, or `|` (default: `,`)
- `strict`: Enable strict mode validation (default: `true` for decode, `false` for encode)
- `keyFolding`: `'safe'` or `'off'` (default: `'off'`)
- `flattenDepth`: Max segments to fold (default: `Infinity`)
- `expandPaths`: `'safe'` or `'off'` (default: `'off'`)

## CLI Reference

The `@toon-format/cli` package converts JSON ↔ TOON from the command line.

### Basic Usage

```bash
# Without installation
npx @toon-format/cli input.json -o output.toon
npx @toon-format/cli data.toon -o output.json

# With global install
npm install -g @toon-format/cli
toon input.json -o output.toon
```

Auto-detects operation from file extension (`.json` → encode, `.toon` → decode). Use `--encode` or `--decode` flags to override.

### Key Options

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Output file path (stdout if omitted) |
| `-e, --encode` | Force encode mode |
| `-d, --decode` | Force decode mode |
| `--delimiter <char>` | Array delimiter: `,`, `\t` (pass as `$'\t'` in bash/zsh), or `|` |
| `--indent <number>` | Indentation size (default: 2) |
| `--stats` | Show token count estimates and savings (encode only) |
| `--no-strict` | Skip decode validation |
| `--keyFolding <mode>` | Key folding: `off`, `safe` |
| `--flattenDepth <number>` | Max segments to fold |
| `--expandPaths <mode>` | Path expansion: `off`, `safe` |
| `--verbose` | Show full stack traces for errors |

### Examples

```bash
# Token statistics
toon data.json --stats -o output.toon

# Tab delimiter for efficiency
toon data.json --delimiter $'\t' -o output.toon

# Key folding + tab delimiter + stats
toon data.json --keyFolding safe --delimiter $'\t' --stats -o output.toon

# Pipe from stdin
cat data.json | toon
echo '{"name": "Ada"}' | toon

# Decode from stdin
cat data.toon | toon --decode

# Lenient decoding
toon data.toon --no-strict -o output.json

# Round-trip with key folding
toon input.json --keyFolding safe -o folded.toon
toon folded.toon --expandPaths safe -o output.json
```

### Stdin Workflows

```bash
# Convert API response to TOON
curl https://api.example.com/data | toon --stats

# Process large dataset
cat large-dataset.json | toon --delimiter $'\t' > output.toon

# Chain with jq
jq '.results' data.json | toon > filtered.toon
```

### Streaming

Both encoding and decoding use streaming output — no full output string in memory. Peak memory scales with data depth, not total size.

## Benchmark Highlights

**Mixed-structure track (6 datasets, 4 models, 209 questions):**
- TOON: **76.4% accuracy** at **2,759 tokens** avg
- JSON: 75.0% accuracy at 4,587 tokens avg
- TOON uses **~40% fewer tokens** with comparable or better accuracy

**Per-question-type accuracy:**
| Type | TOON | JSON | CSV |
|------|------|------|-----|
| Field Retrieval | 99.6% | 99.3% | 100% |
| Aggregation | 61.9% | 61.9% | 50.9% |
| Filtering | 56.8% | 53.1% | 50.9% |
| Structure Awareness | 89.0% | 87.0% | 85.9% |
| Structural Validation | 70.0% | 60.0% | 80.0% |

**Flat-only track:** TOON is ~5.9% more tokens than CSV but includes structural guardrails (length/field declarations) that CSV lacks.

## Ecosystem

**Implementations:**
- TypeScript: `@toon-format/toon` (official reference)
- Python: `toon-format` (pypi)
- Go: `github.com/toon-format/go-toon`
- Rust: `toon-format` (crates.io)
- .NET: `ToonFormat` (NuGet)
- Java, Ruby, Swift: community ports

**Tools:**
- Playground: https://toonformat.dev/playground
- CLI: `npx @toon-format/cli`
- VS Code extension available
- GitHub: https://github.com/toon-format/spec

## Links

- Homepage: https://toonformat.dev
- Getting Started: https://toonformat.dev/guide/getting-started.html
- Format Overview: https://toonformat.dev/guide/format-overview.html
- Syntax Cheatsheet: https://toonformat.dev/reference/syntax-cheatsheet.html
- Spec: https://github.com/toon-format/spec/blob/main/SPEC.md
- API Reference: https://toonformat.dev/reference/api.html
- Playground: https://toonformat.dev/playground
- CLI: https://toonformat.dev/cli/
- Benchmarks: https://toonformat.dev/guide/benchmarks.html
- LLM-optimized docs: https://toonformat.dev/llms.txt
