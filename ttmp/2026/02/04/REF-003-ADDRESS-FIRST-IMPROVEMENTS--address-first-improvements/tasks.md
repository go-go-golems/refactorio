# Tasks

## TODO

## Task 1: Best-effort symbol ingestion (partial package load)
- [x] Allow partial package processing when package load errors are present and ignore flag is set.
- [x] Record go/packages errors as run metadata with severity=warning.
- [x] Extend symbol ingestion result/CLI output to include skipped/error package counts.
- [x] Add/update tests for partial ingestion.

## Task 2: Expose --ignore-package-errors for symbol/code-unit ingestion
- [x] Add CLI flags for symbols and code-units.
- [x] Thread flags through ingest range.
- [x] Apply best-effort behavior for code-units.

## Task 3: Add --limit/--offset flags for list commands
- [x] Add limit/offset to list diff-files.
- [x] Add offset to list symbols.
- [x] Update query helpers and tests.

## Task 4: Record run failures with status + error_json and run metadata table
- [x] Add meta_runs.status and meta_runs.error_json columns.
- [x] Add run_kv table for arbitrary metadata (run_id, key, value).
- [x] Add store helpers for setting status/error_json and writing run metadata.
- [x] Record failures consistently across ingestion functions.

## Task 5: Gopls ingestion tolerant of missing symbol_defs (skip-symbol-lookup default true)
- [x] Default skip-symbol-lookup to true in CLI.
- [x] Allow targets without symbol_hash; store unresolved refs in a new table.
- [x] Add list/query support for unresolved refs (if needed).

## Task 6: Document required build state and surface clear CLI message
- [x] Improve package load error messages (include best-effort hint).
- [x] Update tutorial documentation to describe compile requirements and ignore flag.

## Task 7: Handle root commit explicitly (Option B default)
- [x] Detect root commit and include it automatically.
- [x] Diff root commit against empty tree in range ingestion.
- [x] Add tests for root commit ingestion.

## Task 8: Remove tree-sitter functionality and oak dependency
- [x] Remove tree-sitter CLI commands and range ingestion wiring.
- [x] Remove tree-sitter ingestion implementation.
- [x] Remove oak dependency from refactorio (go.mod/go.sum/go.work if applicable).
- [x] Update docs to remove tree-sitter steps.

## Task 9: Remove report generation functionality
- [x] Remove report CLI command and wiring.
- [x] Remove report implementation.
- [x] Update docs to remove report steps.
