# Refactorio

Refactorio provides two complementary capabilities:

1. **Indexing** (data plane): ingest a codebase into a SQLite-backed refactor index.
2. **JavaScript query scripts** (control-plane helpers): run goja scripts against the index to compute reports or plan-like outputs.

This README focuses on the current CLI surface for indexing and the JS query layer. It does **not** yet cover apply/refactor execution, which will be added in later phases.

---

## Why This Exists
Large refactors depend on knowing where symbols, references, and docs live across a repository. The refactor-index data plane provides a stable, queryable snapshot of those facts. The JS query layer makes it easy to compute ad‑hoc analyses and prototype plan logic without changing Go code.

---

## Quick Start

### 1) Build the index
Run the refactor-index ingestion commands to populate a SQLite DB. The DB is your source of truth for all JS queries.

```bash
# Initialize and ingest symbols/doc hits as needed
# (adjust flags for your repo)

go run ./cmd/refactor-index init --db /path/to/index.sqlite

go run ./cmd/refactor-index ingest symbols \
  --db /path/to/index.sqlite \
  --root /path/to/repo

go run ./cmd/refactor-index ingest doc-hits \
  --db /path/to/index.sqlite \
  --root /path/to/repo \
  --terms /path/to/terms.txt
```

### 2) Run a JS script against the index
```bash
go run ./cmd/refactorio js run \
  --script testdata/js/list_symbols.js \
  --index-db /path/to/index.sqlite \
  --run-id 1
```

---

## Indexing CLI (refactor-index)
The indexer is a separate CLI that writes data into SQLite. Use it to ingest symbols, doc hits, diffs, and other metadata.

### Common commands

```bash
# Initialize schema
refactor-index init --db /path/to/index.sqlite

# Ingest symbols
refactor-index ingest symbols --db /path/to/index.sqlite --root /path/to/repo

# Ingest doc hits
refactor-index ingest doc-hits --db /path/to/index.sqlite --root /path/to/repo --terms /path/to/terms.txt

# List symbols
refactor-index list symbols --db /path/to/index.sqlite --name Client
```

### Why indexing matters
The JS API is read-only. It can only query what has been ingested. If your script can’t find a symbol or doc hit, confirm that the relevant ingest step has been run.

---

## JavaScript Query Layer
The JS query layer is a small goja runtime with a single native module: `refactor-index`.

### API surface

| Function | Purpose |
| --- | --- |
| `querySymbols(filter)` | Find symbol definitions and occurrences |
| `queryRefs(symbolHash)` | Find references for a symbol |
| `queryDocHits(terms, fileset)` | Find term matches in docs |
| `queryFiles(fileset)` | List indexed files |

### Fileset filters
Fileset filters accept `include` and `exclude` glob patterns. If `include` is empty, all files are included before applying excludes.

---

## Example Scripts
Scripts live in `testdata/js/` and can be used as templates.

- `list_symbols.js` — query symbols by package/name/kind
- `list_refs.js` — query refs by symbol hash
- `doc_hits.js` — query doc hits with globs
- `list_files.js` — list files using globs
- `plan_like_output.js` — build plan-like JSON from query results
- `trace_example.js` — demonstrate query tracing

---

## Query Tracing (`js_trace.jsonl`)
When you pass `--trace /path/to/js_trace.jsonl` to `refactorio js run`, each query call is recorded as a JSONL entry including the action name, arguments, and result count. This is helpful for audits and debugging.

Example:

```json
{"seq":1,"action":"querySymbols","args":{"pkg":"github.com/acme/project/internal/api","name":"Client","kind":"type"},"result_count":1}
```

### Trace Example Script
Use the provided example script to produce a trace file.

```bash
go run ./cmd/refactorio js run \
  --script testdata/js/trace_example.js \
  --index-db /path/to/index.sqlite \
  --run-id 1 \
  --trace /tmp/js_trace.jsonl
```

Sample output (`/tmp/js_trace.jsonl`):

```json
{"seq":1,"action":"querySymbols","args":{"pkg":"github.com/acme/project/internal/api","name":"Client","kind":"type"},"result_count":1}
{"seq":2,"action":"queryDocHits","args":{"terms":["Client"],"fileset":{"include":["docs/**/*.md"],"exclude":[]}},"result_count":4}
```

---

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `No such built-in module` | `require()` tried to load a non-allowed module | Use only `require("refactor-index")` |
| Empty results | Index not ingested for that data | Run the relevant ingest command |
| `module file does not exist` | File-based module loading is disabled | Use native modules only |
| `queryRefs requires symbol hash` | Script called `queryRefs` with empty hash | Pass a valid `symbol_hash` |

---

## See Also
- `ttmp/2026/02/04/REF-006-INDEX-LAYER-JS--index-layer-js-api/design/01-js-index-api-guide.md` — Full JS API usage guide.
- `ttmp/2026/02/04/REF-006-INDEX-LAYER-JS--index-layer-js-api/plan/01-js-index-layer-implementation-plan.md` — Implementation plan and task breakdown.
- `ttmp/2026/02/04/REF-005-REFACTORING-TOOLS--refactorio-refactoring-tools/design/03-goja-query-api-architecture-proposal-a.md` — Proposal A architecture spec.
