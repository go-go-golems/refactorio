---
Title: How to Use refactor-index
Slug: refactor-index-how-to-use
Short: Step-by-step tutorial for building a refactor index, understanding each ingestion pass, and mapping results back to the code.
Topics:
    - refactorio
    - indexing
    - sqlite
    - gopls
    - documentation
Commands:
    - refactor-index
    - init
    - ingest
    - list
    - report
Flags:
    - --db
    - --root
    - --repo
    - --from
    - --to
    - --terms
    - --sources-dir
    - --target
    - --targets-file
    - --targets-json
    - --limit
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# How to Use refactor-index

## What You'll Build
You will create a local SQLite index for a Git repository, run the main ingestion passes, and confirm the results with list/report commands. By the end, you'll be able to connect the CLI output to the implementation files so you can extend or debug the pipeline with confidence.

## Prerequisites
This tutorial assumes you are in the refactorio workspace and can run Go commands. Some ingestion passes require optional tools; the tutorial calls those out when needed.

- Go toolchain available
- A Git repository to index (the refactorio repo works fine)
- Optional: `rg` for doc hits
- Optional: `gopls` for semantic references

## Step 1 - Create a Clean Workspace
A clean workspace keeps your repo tidy and makes it easy to throw away indexes as you iterate. You'll create a temp directory for the SQLite database and supporting files.

```bash
ROOT=/home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio
TMP=$(mktemp -d)
DB=$TMP/index.sqlite
SOURCES=$TMP/sources
TERMS=$TMP/terms.txt

cd "$ROOT"
```

## Step 2 - Initialize the Database
The `init` command creates the SQLite schema and records the schema version. This is the foundation for every other pass.

```bash
GOWORK=off go run ./cmd/refactor-index init --db "$DB"
```

Expected output is a single row with the schema version. If the command fails, fix dependency issues before continuing.

Implementation links:
- Schema definition: `refactorio/pkg/refactorindex/schema.go`
- Schema init and inserts: `refactorio/pkg/refactorindex/store.go`

## Step 3 - Ingest Commits and Diffs
Commit and diff ingestion tell the indexer what changed over a range, which is the core timeline for the rest of the pipeline.

```bash
GOWORK=off go run ./cmd/refactor-index ingest commits --db "$DB" --repo . --from HEAD~1 --to HEAD
GOWORK=off go run ./cmd/refactor-index ingest diff --db "$DB" --repo . --from HEAD~1 --to HEAD
```

You should see non-zero counts for commits and diff files. If the range is too small, use a larger window like `HEAD~5` to `HEAD`.

Implementation links:
- Diff orchestration: `refactorio/pkg/refactorindex/ingest_diff.go`
- Diff parsing helpers: `refactorio/pkg/refactorindex/diff_parse.go`

## Step 4 - Ingest Symbols and Code Units
Symbol and code unit ingestion builds an AST inventory of the repo. This enables precise refactoring queries later.

```bash
GOWORK=off go run ./cmd/refactor-index ingest symbols --db "$DB" --root .
GOWORK=off go run ./cmd/refactor-index ingest code-units --db "$DB" --root .
```

You should see counts for symbols, occurrences, code units, and snapshots. These rows are the basis for list queries and gopls targeting.

Note: symbol/code-unit ingestion requires a buildable repo. If `go/packages` reports errors, fix the build or pass `--ignore-package-errors` to proceed with partial results.

Implementation links:
- Symbol ingestion: `refactorio/pkg/refactorindex/ingest_symbols.go`
- Code unit ingestion: `refactorio/pkg/refactorindex/ingest_code_units.go`

## Step 5 - Ingest Doc Hits (Optional)
Doc hits scan text files for terms you care about. This is especially useful for refactors that touch markdown or configuration files.

```bash
echo "refactor" > "$TERMS"
GOWORK=off go run ./cmd/refactor-index ingest doc-hits --db "$DB" --root . --terms "$TERMS" --sources-dir "$SOURCES"
```

If `rg` is not installed, this command will fail. Skip it and continue with other passes.

Implementation link:
- Doc hits ingestion: `refactorio/pkg/refactorindex/ingest_doc_hits.go`

## Step 6 - Ingest gopls References (Optional)
Gopls references provide semantic rename safety, but they require a target spec. The list command can generate one for you.

```bash
GOWORK=off go run ./cmd/refactor-index list symbols --db "$DB" --limit 1
```

Look for the `target_spec` field in the output. It has the format:

```
SYMBOL_HASH|path/to/file.go|line|col
```

Use it directly with `ingest gopls-refs`:

```bash
GOWORK=off go run ./cmd/refactor-index ingest gopls-refs --db "$DB" --repo . --sources-dir "$SOURCES" \
  --target "<target_spec>"
```

Implementation link:
- gopls reference ingestion: `refactorio/pkg/refactorindex/ingest_gopls_refs.go`

## Step 7 - Query the Index
Listing commands are the fastest way to verify that your index is populated.

```bash
GOWORK=off go run ./cmd/refactor-index list diff-files --db "$DB"
GOWORK=off go run ./cmd/refactor-index list symbols --db "$DB" --limit 5
```

If these are empty, revisit earlier steps or use a larger commit range.

Implementation links:
- List symbols wiring: `refactorio/cmd/refactor-index/list_symbols.go`
- SQL query helpers: `refactorio/pkg/refactorindex/query.go`

## Step 8 - Generate a Report
Reports turn the indexed data into structured summaries that can be shared or archived.

```bash
GOWORK=off go run ./cmd/refactor-index report --db "$DB" --run-id 1 --out "$TMP/reports"
```

Implementation link:
- Report rendering: `refactorio/pkg/refactorindex/report.go`

## Discovering Help Topics
The help system exposes documentation sections via the CLI. Use it to explore topics, examples, and tutorials without leaving the terminal.

```bash
GOWORK=off go run ./cmd/refactor-index help
GOWORK=off go run ./cmd/refactor-index help refactor-index-how-to-use
GOWORK=off go run ./cmd/refactor-index help --query "type:tutorial AND topic:indexing"
```

## Complete Example
This combines the core passes into a single run that you can paste and execute.

```bash
ROOT=/home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio
TMP=$(mktemp -d)
DB=$TMP/index.sqlite
SOURCES=$TMP/sources
TERMS=$TMP/terms.txt

cd "$ROOT"

echo "refactor" > "$TERMS"

GOWORK=off go run ./cmd/refactor-index init --db "$DB"
GOWORK=off go run ./cmd/refactor-index ingest commits --db "$DB" --repo . --from HEAD~1 --to HEAD
GOWORK=off go run ./cmd/refactor-index ingest diff --db "$DB" --repo . --from HEAD~1 --to HEAD
GOWORK=off go run ./cmd/refactor-index ingest symbols --db "$DB" --root .
GOWORK=off go run ./cmd/refactor-index ingest code-units --db "$DB" --root .
GOWORK=off go run ./cmd/refactor-index ingest doc-hits --db "$DB" --root . --terms "$TERMS" --sources-dir "$SOURCES"
GOWORK=off go run ./cmd/refactor-index list diff-files --db "$DB"
GOWORK=off go run ./cmd/refactor-index list symbols --db "$DB" --limit 5
```

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `gopls` errors or empty refs | `gopls` not installed or target spec invalid. | Install `gopls` and ensure `target_spec` uses `symbol_hash|path|line|col`. |
| `rg` not found | Ripgrep missing. | Install `rg`, or skip `ingest doc-hits`. |
| Empty list results | Indexing passes didn't run or range too small. | Re-run ingest commands or use a larger `--from`/`--to` window. |

## See Also
- `../../../ttmp/2026/02/04/REF-001-TEST-INDEXING--refactorio-indexing-playbook/reference/02-refactorio-getting-started-playbook.md` - Extended onboarding playbook and context.
- `../../../ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/reference/02-validation-playbook.md` - Full validation playbook with two tracks.
- `../../../ttmp/2026/02/03/GL-008-CREATE-REFACTORING-TOOL--create-refactoring-tool/design-doc/01-refactoring-tool-design.md` - Roadmap for the higher-level refactor suite.
