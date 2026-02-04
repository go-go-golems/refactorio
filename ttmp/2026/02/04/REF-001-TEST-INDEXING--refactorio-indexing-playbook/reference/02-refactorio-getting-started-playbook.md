---
Title: Refactorio Getting Started Playbook
Ticket: REF-001-TEST-INDEXING
Status: active
Topics:
    - refactorio
    - indexing
    - playbook
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../refactorio/AGENT.md
      Note: Build/test conventions
    - Path: ../../../../../../../refactorio/Makefile
      Note: Lint/test/build targets
    - Path: ../../../../../../../refactorio/cmd/refactor-index/root.go
      Note: CLI wiring and subcommands
    - Path: ../../../../../../../refactorio/cmd/refactor-index/ingest_range.go
      Note: Range ingest command wiring
    - Path: ../../../../../../../refactorio/pkg/refactorindex/schema.go
      Note: SQLite schema and version
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_range.go
      Note: Range ingestion orchestration
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_diff.go
      Note: Diff ingestion implementation
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_symbols.go
      Note: AST symbol inventory implementation
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_gopls_refs.go
      Note: gopls reference ingestion
    - Path: ../../../../../../../refactorio/pkg/refactorindex/report.go
      Note: Report rendering
    - Path: ../../../../../../../refactorio/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/reference/02-validation-playbook.md
      Note: Full validation steps and sample commands
    - Path: ../../../../../../../refactorio/ttmp/2026/02/03/GL-008-CREATE-REFACTORING-TOOL--create-refactoring-tool/design-doc/01-refactoring-tool-design.md
      Note: Roadmap for refactorio runner/plan/apply
ExternalSources: []
Summary: "Start-to-finish playbook for running the refactor-index tool, understanding the data pipeline, and locating the core code/docs."
LastUpdated: 2026-02-04T10:57:30-05:00
WhatFor: "Provide a single, copy/paste-friendly on-ramp to refactorio indexing workflows and the refactor tool suite roadmap."
WhenToUse: "Use when you need to run refactor-index on a repo, validate ingestion, or orient in the codebase and design docs."
---

# Refactorio Getting Started Playbook

## Goal
This tutorial walks you from zero to a working refactor-index run, with practical explanations of each ingestion pass and pointers to the implementation so you can dig deeper.

## Context
Refactorio is a Go module in this workspace (`refactorio/`). The active CLI today is `refactor-index`, a SQLite-backed indexer that ingests repository metadata (diffs, symbols, doc hits, etc.) and makes it queryable for future refactoring workflows. The higher-level `refactorio` runner/plan/apply tool suite is designed and documented, but not wired into a working CLI yet.

One important workspace detail: `refactorio/go.mod` contains a replace for `github.com/go-go-golems/oak` pointing to `../oak`. If `../oak` is missing, Go commands will fail. The fix is either to ensure that module is present locally or update/remove the replace.

## Quick Reference

### Where Everything Lives
- Repo root: `refactorio/`
- CLI entry: `refactorio/cmd/refactor-index`
- Core package: `refactorio/pkg/refactorindex`
- Design and validation docs: `refactorio/ttmp/2026/02/03/...`

### CLI Surface (refactor-index)
Top-level:
- `refactor-index init`
- `refactor-index ingest ...`
- `refactor-index list ...`
- `refactor-index report`

Ingest subcommands:
- `ingest diff`
- `ingest commits`
- `ingest symbols`
- `ingest code-units`
- `ingest doc-hits`
- `ingest tree-sitter`
- `ingest gopls-refs`
- `ingest range`

List subcommands:
- `list diff-files`
- `list symbols`

### Data Pipeline (Mental Model)
1. Initialize SQLite schema.
2. Ingest commit lineage and diffs.
3. Ingest AST-derived symbols and code units.
4. Ingest doc hits via ripgrep.
5. Ingest tree-sitter captures for rich structural queries.
6. Ingest gopls semantic references.
7. Query or report the resulting data.

## Step-by-Step Tutorial

### Step 1: Sanity Check Your Workspace
Start by confirming you can run Go commands and that optional dependencies are available.

```bash
cd /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio

# Optional tools that improve ingestion coverage
which gopls || true
which rg || true
```

If you see errors about `../oak`, fix that first. The rest of the tutorial assumes Go commands run successfully.

Implementation pointer: the replace causing this is in `refactorio/go.mod`.

### Step 2: Create a Workspace for Outputs
You want a temp directory for the SQLite DB and intermediate files. This keeps your repo clean and makes cleanup easy.

```bash
ROOT=/home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio
TMP=$(mktemp -d)
DB=$TMP/index.sqlite
SOURCES=$TMP/sources
TERMS=$TMP/terms.txt
QUERY=$TMP/queries.yaml

cd "$ROOT"
```

### Step 3: Prepare Inputs for Doc Hits and Tree-Sitter
Doc hits and tree-sitter require small input files. We’ll build a simple terms list and a tree-sitter query file.

```bash
echo "refactor" > "$TERMS"

cat > "$QUERY" <<'QEOF'
language: go
queries:
  funcs: |
    (function_declaration name: (identifier) @name)
QEOF
```

This query finds all Go function declarations and captures the name identifier. You can expand this later for more complex queries.

Implementation pointer: tree-sitter ingestion is in `refactorio/pkg/refactorindex/ingest_tree_sitter.go`.

### Step 4: Initialize the SQLite Schema
The index tool stores everything in SQLite. Initialize the schema first.

```bash
go run ./cmd/refactor-index init --db "$DB"
```

You should see a row output describing the schema version. If it fails, check the Go dependency setup.

Implementation pointer: schema initialization is in `refactorio/pkg/refactorindex/schema.go` and `refactorio/pkg/refactorindex/store.go`.

### Step 5: Ingest Commit Lineage and Diffs
This pass tells the indexer what changed between two refs. For a quick test, use the last commit.

```bash
go run ./cmd/refactor-index ingest commits --db "$DB" --repo . --from HEAD~1 --to HEAD
go run ./cmd/refactor-index ingest diff --db "$DB" --repo . --from HEAD~1 --to HEAD
```

This produces rows for commits, files, blobs, and diff hunks. It’s the baseline for everything that follows.

Implementation pointer: diff ingestion lives in `refactorio/pkg/refactorindex/ingest_diff.go`.

### Step 6: Ingest Symbols and Code Units
This pass builds an AST inventory of symbols and structural code units.

```bash
go run ./cmd/refactor-index ingest symbols --db "$DB" --root .
go run ./cmd/refactor-index ingest code-units --db "$DB" --root .
```

You should see counts for symbols, occurrences, code units, and snapshots. These tables are the backbone for refactoring analysis.

Implementation pointer: `refactorio/pkg/refactorindex/ingest_symbols.go` and `refactorio/pkg/refactorindex/ingest_code_units.go`.

### Step 7: Ingest Doc Hits (Optional, Requires rg)
Doc hits scan the repo for terms of interest. This is helpful for refactors that include docs and configs.

```bash
go run ./cmd/refactor-index ingest doc-hits --db "$DB" --root . --terms "$TERMS" --sources-dir "$SOURCES"
```

If `rg` isn’t installed, skip this step.

Implementation pointer: `refactorio/pkg/refactorindex/ingest_doc_hits.go`.

### Step 8: Ingest Tree-Sitter Captures
Tree-sitter ingestion stores query matches for richer structural analysis.

```bash
go run ./cmd/refactor-index ingest tree-sitter --db "$DB" --root . --language go --queries "$QUERY" --file-glob "$ROOT/cmd/refactor-index/*.go"
```

You should see a non-zero number of captures if the query matches.

Implementation pointer: `refactorio/pkg/refactorindex/ingest_tree_sitter.go`.

### Step 9: Ingest gopls References (Optional, Requires gopls)
This step needs a target spec that points to a specific symbol occurrence. The validation playbook shows how to extract one from the DB.

Use the playbook here:
- `refactorio/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/reference/02-validation-playbook.md`

Implementation pointer: `refactorio/pkg/refactorindex/ingest_gopls_refs.go`.

### Step 10: Query the Index
Once the DB is populated, list outputs to verify everything is in place.

```bash
go run ./cmd/refactor-index list diff-files --db "$DB"
go run ./cmd/refactor-index list symbols --db "$DB" --limit 5
```

This gives you immediate confidence that the index is usable.

Implementation pointer: list command wiring is in `refactorio/cmd/refactor-index/root.go` and query logic in `refactorio/pkg/refactorindex/query.go`.

### Step 11: Generate a Report
Reports are SQL-backed and render to output directories. Use them for high-level summaries.

```bash
# Example: list available run IDs, then generate a report
go run ./cmd/refactor-index report --db "$DB" --run-id 1 --out "$TMP/reports"
```

Implementation pointer: `refactorio/pkg/refactorindex/report.go`.

## Implementation Map (Link Back to Code)
Use this when you want to trace behavior from CLI to implementation.

1. CLI wiring and command tree: `refactorio/cmd/refactor-index/root.go`
2. Schema + storage: `refactorio/pkg/refactorindex/schema.go`, `refactorio/pkg/refactorindex/store.go`
3. Range orchestration: `refactorio/pkg/refactorindex/ingest_range.go`
4. Diff ingestion: `refactorio/pkg/refactorindex/ingest_diff.go`
5. Symbol ingestion: `refactorio/pkg/refactorindex/ingest_symbols.go`
6. gopls refs: `refactorio/pkg/refactorindex/ingest_gopls_refs.go`
7. Reports: `refactorio/pkg/refactorindex/report.go`

## Roadmap: The Refactor Tool Suite
When you’re ready to move beyond indexing into automated refactors, read the design doc:
- `refactorio/ttmp/2026/02/03/GL-008-CREATE-REFACTORING-TOOL--create-refactoring-tool/design-doc/01-refactoring-tool-design.md`

It describes the future `refactorio run/plan/apply/audit/report` workflow and how it builds on the index you just created.

## Usage Examples

### Example: One-Shot Ingest for a Git Range
Use `ingest range` to run multiple passes at once.

```bash
go run ./cmd/refactor-index ingest range --db "$DB" --repo . --from HEAD~5 --to HEAD
```

Implementation pointer: orchestration logic is in `refactorio/pkg/refactorindex/ingest_range.go` and CLI wiring in `refactorio/cmd/refactor-index/ingest_range.go`.

### Example: Minimal Symbols-Only Run
If you only need symbols for a quick inventory:

```bash
go run ./cmd/refactor-index ingest symbols --db "$DB" --root .
go run ./cmd/refactor-index list symbols --db "$DB" --limit 10
```

## Related
- `refactorio/AGENT.md`
- `refactorio/Makefile`
- `refactorio/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/reference/02-validation-playbook.md`
- `refactorio/ttmp/2026/02/03/GL-008-CREATE-REFACTORING-TOOL--create-refactoring-tool/design-doc/01-refactoring-tool-design.md`
