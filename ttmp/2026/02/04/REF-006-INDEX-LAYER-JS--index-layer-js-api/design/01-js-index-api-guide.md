---
Title: JS Index API Guide
Ticket: REF-006-INDEX-LAYER-JS
Status: active
Topics:
    - refactorio
    - js
    - index
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/cmd/refactorio/js_run.go
      Note: JS runner CLI
    - Path: refactorio/pkg/refactor/js/modules/refactorindex/refactorindex.go
      Note: JS query module
    - Path: refactorio/testdata/js/README.md
      Note: Example scripts overview
    - Path: refactorio/testdata/js/doc_hits.js
      Note: Example script
    - Path: refactorio/testdata/js/list_files.js
      Note: Example script
    - Path: refactorio/testdata/js/list_refs.js
      Note: Example script
    - Path: refactorio/testdata/js/list_symbols.js
      Note: Example script
    - Path: refactorio/testdata/js/plan_like_output.js
      Note: Example script
    - Path: refactorio/testdata/js/trace_example.js
      Note: Trace example script
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-04T16:12:30-05:00
WhatFor: ""
WhenToUse: ""
---



# JS Index API Guide

## Why This Exists
The JS index API lets you write small JavaScript scripts that query the refactor-index data plane and return structured results. This enables fast, scriptable exploration of symbols, references, doc hits, and files without requiring Go changes for every query idea.

The API is intentionally read-only. It does not apply refactors. It gives you a deterministic, auditable query layer so you can compute plans or reports safely.

## Quick Start
This is the smallest runnable example. It queries symbols and prints the result as JSON.

```bash
go run ./cmd/refactorio js run \
  --script testdata/js/list_symbols.js \
  --index-db /path/to/index.sqlite \
  --run-id 1
```

```javascript
// testdata/js/list_symbols.js
const idx = require("refactor-index");
const symbols = idx.querySymbols({
  pkg: "github.com/acme/project/internal/api",
  name: "Client",
  kind: "type",
});
console.log(JSON.stringify(symbols, null, 2));
symbols;
```

## Core Concepts
The JS API is a thin wrapper over the refactor-index store. It exposes a single module named `refactor-index` with query functions. The Go runtime ensures results are stable and deterministic, then returns arrays of plain objects that are easy to inspect or transform.

When you return a value from a script, the CLI prints it. This allows scripts to be used as building blocks for reports, plan generation, or audits.

## Runtime Model
The JS runtime is based on goja with Node-style `require()`. Only allow-listed native modules are registered. File-based module loading is disabled by default, and time/random access can be turned off for determinism.

This means:
- Your script can import `refactor-index` but not `fs` or `exec`.
- No filesystem writes occur from JS.
- Results are stable because queries are sorted in Go.

## API Reference
The `refactor-index` module exports four functions. Each returns an array of plain objects.

| Function | Purpose | Input | Output |
| --- | --- | --- | --- |
| `querySymbols(filter)` | Find symbol definitions and occurrences | `filter` object | `SymbolRecord[]` |
| `queryRefs(symbolHash)` | List references for a symbol | symbol hash string | `RefRecord[]` |
| `queryDocHits(terms, fileset)` | Find doc hit matches | `terms[]`, `fileset` | `DocHitRecord[]` |
| `queryFiles(fileset)` | List files from the index | `fileset` | `FileRecord[]` |

### SymbolRecord
A symbol record is a single occurrence with definition metadata.

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

### RefRecord
A reference record points at a location that mentions a symbol.

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

### DocHitRecord
A doc hit record indicates a term match in text files.

```json
{
  "term": "Client",
  "path": "docs/api.md",
  "line": 18,
  "col": 2,
  "match_text": "...Client..."
}
```

### FileRecord
A file record is a row from the indexed file list.

```json
{
  "path": "docs/api.md",
  "ext": "md",
  "exists": true,
  "is_binary": false
}
```

## Filters and Filesets
Each query uses a small filter or fileset. Filters control server-side query predicates. Filesets control client-side glob filtering using doublestar.

### Symbol Filter Fields
- `pkg` string
- `name` string
- `kind` string
- `path` string
- `exported_only` bool
- `limit` int
- `offset` int

### Fileset
- `include` array of glob patterns
- `exclude` array of glob patterns

If `include` is empty, all files are included before applying excludes.

## Determinism Guarantees
The runtime enforces deterministic behavior by disabling time and randomness (when configured) and by sorting all query results in Go before returning them to JS. This ensures that repeated runs over the same index produce identical outputs.

If you need stable plan generation, always avoid JavaScript code that depends on map iteration order or object property ordering.

## Query Tracing (`js_trace.jsonl`)
The refactor-index module can emit a JSONL trace of every query call. Each entry includes the action name, arguments, and result count. This trace is intended for audits, debugging, and reproducibility.

Example trace entry:

```json
{"seq":1,"action":"querySymbols","args":{"pkg":"github.com/acme/project/internal/api","name":"Client","kind":"type"},"result_count":1}
```

## Trace Output Reference
Each line in `js_trace.jsonl` is a JSON object. These fields are stable and meant to be parsed by tooling.

| Field | Type | Meaning | Notes |
| --- | --- | --- | --- |
| `seq` | number | Monotonic sequence number | Starts at 1 for each script run |
| `action` | string | The query name | One of `querySymbols`, `queryRefs`, `queryDocHits`, `queryFiles` |
| `args` | object | The exact arguments passed to the query | Mirrors the JS inputs |
| `result_count` | number | Number of records returned | After fileset filtering and sorting |

**Why this matters:** The trace lets you answer “what did the script actually ask the index?” without reading JS code. It also makes debugging deterministic issues (unexpected counts, empty results) much faster.

## Writing Scripts
Scripts should return a value explicitly or implicitly by ending with an expression. This is what the CLI prints.

### Example: Compute a Simple Summary
```javascript
const idx = require("refactor-index");

const symbols = idx.querySymbols({ pkg: "github.com/acme/project/internal/api", name: "Client", kind: "type" });
const refs = symbols.length === 1 ? idx.queryRefs(symbols[0].symbol_hash) : [];

const summary = {
  symbol_count: symbols.length,
  ref_count: refs.length,
};

summary;
```

### Example: Plan-like Output
```javascript
const idx = require("refactor-index");
const symbols = idx.querySymbols({ pkg: "github.com/acme/project/internal/api", name: "Client", kind: "type" });
if (symbols.length !== 1) throw new Error(`ambiguous symbol count: ${symbols.length}`);
const sym = symbols[0];

const plan = {
  plan_version: 1,
  ops: [
    {
      type: "go.gopls.rename",
      selector: { symbol_hash: sym.symbol_hash },
      resolved: { def_span: sym.def_span, old_name: sym.name, new_name: "APIClient" },
    },
  ],
};

plan;
```

## CLI Usage
The JS runner accepts a script path and an index DB path.

```bash
go run ./cmd/refactorio js run \
  --script testdata/js/doc_hits.js \
  --index-db /path/to/index.sqlite \
  --run-id 1 \
  --trace /tmp/js_trace.jsonl
```

When `--trace` is provided, the JS module writes query traces to the given path.

## Example Scripts
These scripts live under `testdata/js/` and are intended to be runnable.

- `testdata/js/list_symbols.js` — query symbols by package/name/kind
- `testdata/js/list_refs.js` — query refs by symbol hash
- `testdata/js/doc_hits.js` — query doc hits with fileset globs
- `testdata/js/list_files.js` — list files by fileset
- `testdata/js/plan_like_output.js` — construct plan-like output
- `testdata/js/trace_example.js` — demonstrate query tracing

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `No such built-in module` | Module name is not registered | Use `require("refactor-index")` only |
| `queryRefs requires symbol hash` | `queryRefs` called with empty value | Pass a valid `symbol_hash` string |
| Empty result arrays | Filters too strict or run ID mismatch | Remove filters or set `--run-id 0` |
| `module file does not exist` | File-based `require()` is disabled | Use only native modules |

## Troubleshooting by Error Message
These are the most common errors you’ll see while writing scripts, along with direct fixes.

| Error message | Likely cause | Fix |
| --- | --- | --- |
| `No such built-in module` | A script called `require()` for a module that isn’t registered. | Only `require("refactor-index")` is allowed. Remove `require("fs")`, `require("exec")`, etc. |
| `module file does not exist` | File-based `require()` is disabled. | Inline your script or use only native modules. |
| `queryRefs requires symbol hash` | `queryRefs` called with `null`, `undefined`, or empty string. | Pass a concrete `symbol_hash` from `querySymbols`. |
| `ambiguous symbol count` (custom) | Your selector matches multiple symbols. | Add filters (`pkg`, `kind`, `path`) or refine with `exported_only`. |
| Empty array results | Filters are too strict or wrong run ID. | Remove filters or set `--run-id 0` to query all runs. |
| JSON output is `null` | The script didn’t return a value. | End your script with an expression (e.g., `result;`). |

## See Also
- `ttmp/2026/02/04/REF-006-INDEX-LAYER-JS--index-layer-js-api/plan/01-js-index-layer-implementation-plan.md` — Implementation plan for the JS index layer.
- `ttmp/2026/02/04/REF-005-REFACTORING-TOOLS--refactorio-refactoring-tools/design/03-goja-query-api-architecture-proposal-a.md` — Proposal A architecture spec.
- `testdata/js/README.md` — Quick usage notes for JS examples.
