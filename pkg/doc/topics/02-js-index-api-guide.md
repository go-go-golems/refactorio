---
Title: JS Index API User Guide
Slug: js-index-api-user-guide
Short: A practical, start-to-finish guide for writing JavaScript scripts that query the refactor index.
Topics:
    - refactorio
    - js
    - index
    - scripting
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

## Why This Guide
The JS index API exists so you can compute reports, plan-like outputs, and audits without changing Go code. It gives you a deterministic, read-only view of the refactor index. This guide is written for someone new to the codebase and focuses on the “what, why, and how” of writing scripts.

## Quick Start
The fastest way to validate your setup is to run the example scripts that ship with refactorio.

```bash
go run ./cmd/refactorio js run \
  --script testdata/js/list_symbols.js \
  --index-db /path/to/index.sqlite \
  --run-id 1
```

Expected output is JSON containing symbol records. If you see an empty array, the index may not include the symbol you are searching for.

## Prerequisites
Before you run JS scripts, ensure you have an index database. The JS API is read-only and only sees what has been ingested.

```bash
# Create schema
go run ./cmd/refactor-index init --db /path/to/index.sqlite

# Ingest symbols (required for querySymbols/queryRefs)
go run ./cmd/refactor-index ingest symbols \
  --db /path/to/index.sqlite \
  --root /path/to/repo

# Ingest doc hits (required for queryDocHits)
go run ./cmd/refactor-index ingest doc-hits \
  --db /path/to/index.sqlite \
  --root /path/to/repo \
  --terms /path/to/terms.txt
```

## Core Concepts
The JS layer is a thin wrapper over the SQLite index. You call query functions; refactorio returns arrays of plain JS objects. Each query is deterministic and sorted in Go so that the same input yields the same output.

Key characteristics:
- **Read-only**: no apply or write operations are available.
- **Deterministic**: results are sorted and time/random can be disabled.
- **Auditable**: optional query traces (`js_trace.jsonl`).

## Where to Look in the Codebase
- JS runtime: `pkg/refactor/js/runtime.go`
- JS module registry: `pkg/refactor/js/modules/common.go`
- Index module: `pkg/refactor/js/modules/refactorindex/refactorindex.go`
- CLI runner: `cmd/refactorio/js_run.go`
- Example scripts: `testdata/js/`

These files are the best starting points if you want to extend the API.

## Writing Your First Script
A script is just a JS file executed by the runner. If it ends with an expression, that value is printed as JSON.

```javascript
const idx = require("refactor-index");

const symbols = idx.querySymbols({
  pkg: "github.com/acme/project/internal/api",
  name: "Client",
  kind: "type",
});

symbols; // returned value printed by CLI
```

Run it:

```bash
go run ./cmd/refactorio js run \
  --script /path/to/script.js \
  --index-db /path/to/index.sqlite \
  --run-id 1
```

## Common Query Patterns
### Pattern: Symbol → Refs
Many workflows start by finding a symbol, then fetching its references.

```javascript
const idx = require("refactor-index");
const symbols = idx.querySymbols({ pkg: "github.com/acme/project/internal/api", name: "Client", kind: "type" });
if (symbols.length !== 1) throw new Error(`ambiguous symbol count: ${symbols.length}`);
const refs = idx.queryRefs(symbols[0].symbol_hash);
({ symbol: symbols[0], ref_count: refs.length });
```

### Pattern: Doc Audit
```javascript
const idx = require("refactor-index");
const hits = idx.queryDocHits(["Client"], { include: ["docs/**/*.md"] });
({ hit_count: hits.length, sample: hits.slice(0, 5) });
```

### Pattern: Plan-like Output
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

## Filesets and Globs
Filesets control file filtering for `queryDocHits` and `queryFiles`. If `include` is empty, all files are included before applying `exclude`.

```javascript
const idx = require("refactor-index");
const files = idx.queryFiles({ include: ["docs/**/*.md"], exclude: ["**/vendor/**"] });
files;
```

## Query Tracing
Tracing records every query call as JSON lines. This is useful when you want to debug why a script returned a certain result count.

```bash
go run ./cmd/refactorio js run \
  --script testdata/js/trace_example.js \
  --index-db /path/to/index.sqlite \
  --run-id 1 \
  --trace /tmp/js_trace.jsonl
```

Example trace entry:

```json
{"seq":1,"action":"querySymbols","args":{"pkg":"github.com/acme/project/internal/api","name":"Client","kind":"type"},"result_count":1}
```

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Empty arrays | Index doesn’t include the target data | Re-run ingest steps and confirm `--run-id` |
| `No such built-in module` | `require()` called with wrong module | Use `require("refactor-index")` only |
| `module file does not exist` | File-based `require()` is disabled | Use native modules only |
| `queryRefs requires symbol hash` | `queryRefs` called with empty string | Use a valid `symbol_hash` from `querySymbols` |

## See Also
- `pkg/doc/topics/03-js-index-api-reference.md` — Full API reference.
- `ttmp/2026/02/04/REF-006-INDEX-LAYER-JS--index-layer-js-api/design/01-js-index-api-guide.md` — Long-form guide with deeper details.
- `testdata/js/README.md` — Script index and usage notes.
