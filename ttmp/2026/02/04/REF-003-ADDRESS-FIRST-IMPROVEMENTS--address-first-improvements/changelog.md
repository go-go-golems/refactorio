# Changelog

## 2026-02-04

- Initial workspace created


## 2026-02-04

Add task and implementation plan section for temporarily removing tree-sitter ingestion.

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/design/01-global-implementation-plan.md — Added Task 9 section
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/tasks.md — Task list updated


## 2026-02-04

Add task and implementation plan section to remove report generation temporarily.

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/design/01-global-implementation-plan.md — Added Task 10 section
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/tasks.md — Added report removal task


## 2026-02-04

Remove tree-sitter glob support task and renumber implementation plan sections.

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/design/01-global-implementation-plan.md — Renumbered tasks
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/tasks.md — Removed tree-sitter glob support task


## 2026-02-04

Update plan and task breakdown per new directives (run status/error_json + run_kv, go/packages warning metadata, skip-symbol-lookup default, root option B, oak removal).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/design/01-global-implementation-plan.md — Updated approaches
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-003-ADDRESS-FIRST-IMPROVEMENTS--address-first-improvements/tasks.md — Reorganized subtasks


## 2026-02-04

Complete Task 4: add run status/error_json + run_kv metadata table and failure handling (commit 38e81da).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/schema.go — Schema bump with status/error_json and run_kv
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/store.go — Run status/error handling + metadata helpers


## 2026-02-04

Implement best-effort symbol ingestion with warning metadata (commit b62029e).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/cmd/refactor-index/ingest_symbols.go — Expose new counts
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/ingest_symbols.go — Best-effort handling + run metadata


## 2026-02-04

Complete Task 2: add ignore-package-errors flags and code-units best-effort handling (commit 93e087b).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/cmd/refactor-index/ingest_symbols.go — Flag wiring
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/ingest_code_units.go — Best-effort handling

