---
Title: Global Implementation Plan
Ticket: REF-003-ADDRESS-FIRST-IMPROVEMENTS
Status: active
Topics:
    - refactorio
    - indexing
    - refactor-index
    - improvements
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Implementation plan for the first improvement set identified in the sqleton refactor-index investigation."
LastUpdated: 2026-02-04T11:46:06-05:00
WhatFor: "Plan and scope refactor-index improvements with code-level touch points."
WhenToUse: "Use before implementing the first batch of refactor-index fixes and enhancements."
---

# Global Implementation Plan

## Overview
This plan breaks the improvement list into discrete tasks. Each task includes a brief codebase analysis, a recommended implementation approach, and likely touch points. Tasks are listed in the same order as the ticket task list so the plan can be executed incrementally.

## Task 1: Best-effort symbol ingestion (partial package load)

### Current state
`refactorio/pkg/refactorindex/ingest_symbols.go` uses `packages.Load` with `packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax` and calls `packages.PrintErrors(pkgs)`. Any non-zero error count returns `package load errors`, aborting the run. This blocks symbol ingestion for repos that do not compile cleanly.

### Implementation approach
- Extend `IngestSymbolsConfig` with `IgnorePackageErrors bool`.
- If `packages.PrintErrors(pkgs) > 0` and the ignore flag is set, continue processing only packages with `Types`, `TypesInfo`, and `Fset` populated.
- Track counts for `packages_with_errors` and `packages_skipped` in the result and surface in the CLI output.
- Capture the `go/packages` error strings and store them as run metadata with `severity=warning` (via the run metadata table introduced in Task 4).

### Files to touch
- `refactorio/pkg/refactorindex/ingest_symbols.go`
- `refactorio/cmd/refactor-index/ingest_symbols.go` (surface new counts)
- `refactorio/pkg/refactorindex/ingest_range.go` (pass flag through when symbol ingestion is used)

### Validation
- Run `go test ./refactorio/pkg/refactorindex -run IngestSymbols` or existing smoke tests.
- Add a new unit test that simulates a broken package and asserts partial ingestion when the flag is set.

## Task 2: Expose --ignore-package-errors for symbol/code-unit ingestion

### Current state
`IngestSymbols` and `IngestCodeUnits` behave identically regarding package load errors: any errors abort the run. CLI commands `ingest symbols` and `ingest code-units` do not expose a flag to override this.

### Implementation approach
- Add a `--ignore-package-errors` (or `--best-effort`) flag to both CLI commands.
- Thread the flag into `IngestSymbolsConfig` and `IngestCodeUnitsConfig` and use the same behavior described in Task 1 for both pipelines.
- For `ingest range`, add a matching flag (e.g., `--symbols-ignore-package-errors`, `--code-units-ignore-package-errors`, or a shared `--ignore-package-errors` used for both when enabled).

### Files to touch
- `refactorio/cmd/refactor-index/ingest_symbols.go`
- `refactorio/cmd/refactor-index/ingest_code_units.go`
- `refactorio/cmd/refactor-index/ingest_range.go`
- `refactorio/pkg/refactorindex/ingest_symbols.go`
- `refactorio/pkg/refactorindex/ingest_code_units.go`

### Validation
- CLI smoke test: run `ingest symbols` / `ingest code-units` on a repo with intentional compile errors and confirm non-zero counts plus warning output.

## Task 3: Add --limit/--offset flags for list commands

### Current state
`list diff-files` lacks `--limit`/`--offset`; it always loads all rows. `list symbols` has `--limit` but no `--offset`. Store query helpers (`ListDiffFiles`, `ListSymbolInventory`) do not accept offsets.

### Implementation approach
- Extend `ListDiffFilesSettings` with `limit` and `offset` and thread to `Store.ListDiffFiles`.
- Update `Store.ListDiffFiles` to accept `limit` and `offset` parameters and append `LIMIT ? OFFSET ?` when values are non-zero.
- Extend `SymbolInventoryFilter` with `Offset` and update `ListSymbolInventory` to include offset and optional limit.
- Update CLI help text to describe limit/offset usage.

### Files to touch
- `refactorio/cmd/refactor-index/list_diff_files.go`
- `refactorio/cmd/refactor-index/list_symbols.go`
- `refactorio/pkg/refactorindex/query.go`

### Validation
- Add/adjust tests in `ingest_diff_smoke_test.go` or a new test to assert limit/offset behavior.

## Task 4: Record run failures with status + error_json and run metadata table

### Current state
Runs are created in `meta_runs` and marked complete via `finished_at`. Errors during ingestion return early without recording error details. Failed runs remain with `finished_at = NULL` but no reason.

### Implementation approach
- Add `meta_runs.status` and `meta_runs.error_json` columns.
- Add a `run_kv` table for arbitrary key/value metadata (at minimum: `run_id`, `key`, `value`, `created_at`).
- Add store helpers to set status/error_json and to insert run metadata rows.
- Update ingestion flows to set `status=success` on finish, and `status=failed` + `error_json` on failure once a run has been created.
- Use `run_kv` to store warnings (e.g., go/packages errors with `severity=warning`) instead of failing runs.

### Files to touch
- `refactorio/pkg/refactorindex/schema.go`
- `refactorio/pkg/refactorindex/store.go`
- All ingestion files that create runs: `ingest_diff.go`, `ingest_commits.go`, `ingest_symbols.go`, `ingest_code_units.go`, `ingest_doc_hits.go`, `ingest_gopls_refs.go`

### Validation
- Add a test that triggers a failure after run creation and asserts a `run_errors` row is written.

## Task 5: Gopls ingestion tolerant of missing symbol_defs

### Current state
`IngestGoplsReferences` requires `SymbolHash` and calls `Store.GetSymbolDefIDByHash`. If symbol ingestion failed, gopls ingestion cannot proceed.

### Implementation approach
- Default `--skip-symbol-lookup` to true in `ingest gopls-refs` (allow opt-out if needed).
- When skip is enabled (or hash missing), allow targets to proceed using only `file_path`, `line`, `col`.
- Store results in a new table that does not require `symbol_def_id`, e.g. `symbol_refs_unresolved` with columns `run_id`, `commit_id`, `file_id`, `line`, `col`, `is_decl`, `source`, plus optional `symbol_hash` string for later reconciliation.
- Expose a query helper to list unresolved refs if we need visibility without symbol_defs.

### Files to touch
- `refactorio/pkg/refactorindex/schema.go`
- `refactorio/pkg/refactorindex/ingest_gopls_refs.go`
- `refactorio/cmd/refactor-index/ingest_gopls_refs.go`

### Validation
- Add a test that passes a target with no symbol hash and confirms unresolved refs are inserted.

## Task 6: Document required build state and surface clear CLI message

### Current state
When `packages.Load` emits errors, the CLI returns `package load errors` with no guidance. The tutorial does not clearly call out that symbol/code-unit ingestion requires a buildable repo.

### Implementation approach
- Improve the error message when package loading fails: mention that repo must compile and suggest `--ignore-package-errors` if available.
- Update the tutorial at `refactorio/pkg/doc/tutorials/01-refactor-index-how-to-use.md` to add a short note in Step 4 about compilation requirements and the best-effort flag.
- If we add run error metadata, ensure the error text includes the package errors summary.

### Files to touch
- `refactorio/pkg/refactorindex/ingest_symbols.go`
- `refactorio/pkg/refactorindex/ingest_code_units.go`
- `refactorio/pkg/doc/tutorials/01-refactor-index-how-to-use.md`

### Validation
- Spot-check the CLI error output when packages fail to load.

## Task 7: Handle root commit explicitly

### Current state
Commit ingestion uses `git rev-list from..to`, which excludes the `from` commit itself. If `from` is the root commit, it is never ingested. Diff ingestion uses `git diff from to` which will fail if `from` is a root commit and we attempt `from^` in range ingestion.

### Implementation approach
- Detect when `from` is the root commit (e.g., `git rev-list --max-parents=0 <from>`). If so, include it explicitly in the commit list (Option B default).
- For diff ingestion in range mode, handle root commit by diffing against the empty tree. Use `git diff --root <commit>` or `git show --root` to generate the patch for the initial commit.
- Update range ingestion to avoid `hash^` for root commits and to use the root diff path instead.

### Files to touch
- `refactorio/pkg/refactorindex/ingest_commits.go`
- `refactorio/pkg/refactorindex/ingest_diff.go`
- `refactorio/pkg/refactorindex/ingest_range.go`
- `refactorio/cmd/refactor-index/ingest_commits.go`

### Validation
- Add a test that ingests from the root commit on a small repo and verifies the root commit appears and diff ingestion succeeds.

## Task 8: Rip out tree-sitter functionality (temporary removal)

### Current state
Tree-sitter ingestion is wired via `refactorio/pkg/refactorindex/ingest_tree_sitter.go` and exposed in the CLI (ingest range and `ingest tree-sitter`). The schema includes `ts_captures`, and help/tutorial docs describe the flow. This feature is currently a footgun due to glob limitations and external dependencies.

### Implementation approach
- Remove the `ingest tree-sitter` command from CLI wiring (`cmd/refactor-index/root.go` + command file).
- Remove tree-sitter from range ingestion flags/config and skip invoking it in `IngestCommitRange`.
- Leave schema tables in place for now to avoid breaking existing DBs, but mark them as unused in docs.
- Update tutorial/help docs to remove tree-sitter steps (or mark as temporarily disabled).
- Remove the Oak dependency from refactorio (go.mod/go.sum/go.work as applicable) once tree-sitter code is removed.

### Files to touch
- `refactorio/cmd/refactor-index/root.go`
- `refactorio/cmd/refactor-index/ingest_tree_sitter.go`
- `refactorio/cmd/refactor-index/ingest_range.go`
- `refactorio/pkg/refactorindex/ingest_range.go`
- `refactorio/pkg/refactorindex/ingest_tree_sitter.go`
- `refactorio/pkg/doc/tutorials/01-refactor-index-how-to-use.md`

### Validation
- Build `refactor-index` and ensure CLI help no longer lists tree-sitter.
- Run existing ingestion tests to confirm no references to tree-sitter remain.

## Task 9: Temporarily remove report generation

### Current state
The `report` command generates markdown summaries such as `diff-files.md` via `refactorio/pkg/refactorindex/report.go` and the CLI wiring in `cmd/refactor-index`. This is secondary to core indexing and adds maintenance surface while other ingestion gaps are being addressed.

### Implementation approach
- Remove the `report` command from CLI wiring (`cmd/refactor-index/root.go` plus the command source file).
- Remove or stub the `refactorio/pkg/refactorindex/report.go` entrypoint (leave helper types if they’re reused elsewhere; otherwise delete).
- Update tutorials/docs to remove report steps and references.
- Keep any schema tables untouched (reporting is a read-only layer).

### Files to touch
- `refactorio/cmd/refactor-index/root.go`
- `refactorio/cmd/refactor-index/report.go`
- `refactorio/pkg/refactorindex/report.go`
- `refactorio/pkg/doc/tutorials/01-refactor-index-how-to-use.md`

### Validation
- Build `refactor-index` and confirm `report` no longer appears in help output.
