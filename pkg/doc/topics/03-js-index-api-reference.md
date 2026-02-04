---
Title: JS Index API Reference
Slug: js-index-api-reference
Short: Exhaustive reference for the refactor-index JavaScript API, including data shapes and CLI flags.
Topics:
    - refactorio
    - js
    - index
    - reference
Commands:
    - refactorio
    - js
    - run
Flags:
    - --script
    - --index-db
    - --run-id
    - --trace
    - --format
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why This Reference
The user guide explains *how* to write scripts. This document is a precise reference for every API, record shape, and CLI flag so you can look up details quickly.

## Module: `refactor-index`
All JS scripts use the single native module `refactor-index`.

```javascript
const idx = require("refactor-index");
```

## Functions
### `querySymbols(filter)`
Returns symbol occurrences with definition metadata.

**Filter fields**
| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `pkg` | string | `""` | Exact package match |
| `name` | string | `""` | Exact symbol name |
| `kind` | string | `""` | `type`, `func`, `method`, `var`, `const` |
| `path` | string | `""` | Exact file path |
| `exported_only` | bool | `false` | Only exported symbols |
| `limit` | number | 0 | Max results (clamped by runtime) |
| `offset` | number | 0 | Offset for paging |

**Return shape**
```json
{
  "symbol_hash": "...",
  "pkg": "github.com/acme/project/internal/api",
  "name": "Client",
  "kind": "type",
  "recv": "",
  "signature": "",
  "def_span": "internal/api/client.go:42:6",
  "file": "internal/api/client.go",
  "line": 42,
  "col": 6,
  "is_exported": true
}
```

### `queryRefs(symbolHash)`
Returns all references for a symbol hash.

**Args**
- `symbolHash` string (required)

**Return shape**
```json
{
  "symbol_hash": "...",
  "path": "internal/api/client.go",
  "line": 55,
  "col": 10,
  "is_decl": false,
  "source": "gopls",
  "commit_hash": ""
}
```

### `queryDocHits(terms, fileset)`
Returns doc hits for the provided terms.

**Args**
- `terms`: `string[]`
- `fileset`: `{ include: string[], exclude: string[] }`

**Return shape**
```json
{
  "term": "Client",
  "path": "docs/api.md",
  "line": 18,
  "col": 2,
  "match_text": "...Client..."
}
```

### `queryFiles(fileset)`
Lists files from the index after applying fileset globs.

**Args**
- `fileset`: `{ include: string[], exclude: string[] }`

**Return shape**
```json
{
  "path": "docs/api.md",
  "ext": "md",
  "exists": true,
  "is_binary": false
}
```

## Filesets
Filesets are applied client‑side in Go after fetching rows from the database.

| Field | Type | Description |
| --- | --- | --- |
| `include` | string[] | Glob patterns to include (doublestar syntax) |
| `exclude` | string[] | Glob patterns to exclude |

If `include` is empty, all files are included before `exclude` is applied.

## Determinism Rules
- All query results are sorted in Go.
- `Date.now()` and `Math.random()` can be disabled in the runtime.
- Result limits are clamped (default 5000).

## CLI Flags (`refactorio js run`)

| Flag | Description | Required |
| --- | --- | --- |
| `--script` | Path to JS file | Yes |
| `--index-db` | Path to SQLite DB | Yes |
| `--run-id` | Restrict queries to a run (0 = all) | No |
| `--trace` | Write query trace JSONL | No |
| `--format` | `json` or `text` | No |

## Trace Output (`js_trace.jsonl`)
Each line is a JSON object with these fields:

| Field | Type | Description |
| --- | --- | --- |
| `seq` | number | Sequence number starting at 1 |
| `action` | string | Query name (`querySymbols`, `queryRefs`, `queryDocHits`, `queryFiles`) |
| `args` | object | Query arguments |
| `result_count` | number | Number of results returned |

## Failure Modes
- **`No such built-in module`**: attempted to require a non‑registered module.
- **`queryRefs requires symbol hash`**: called `queryRefs` with an empty value.
- **Empty results**: missing ingest or over‑strict filters.

## See Also
- `pkg/doc/topics/02-js-index-api-guide.md` — User guide and examples.
- `testdata/js/README.md` — Example scripts.
