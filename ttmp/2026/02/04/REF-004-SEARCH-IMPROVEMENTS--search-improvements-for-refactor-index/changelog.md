# Changelog

## 2026-02-04

- Initial workspace created


## 2026-02-04

Reframe plan and tasks to remove tree-sitter scope and align work items.

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-004-SEARCH-IMPROVEMENTS--search-improvements-for-refactor-index/design-doc/01-search-improvements-implementation-plan.md — Tree-sitter removed from plan
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-004-SEARCH-IMPROVEMENTS--search-improvements-for-refactor-index/tasks.md — Rewritten task breakdown


## 2026-02-04

Complete Task 1: add multi-column FTS helper (commit 371bcd2).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/store.go — Multi-column FTS helper


## 2026-02-04

Complete Task 2: add code unit FTS table (commit d6e4ce2).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/store.go — code_unit_snapshots_fts


## 2026-02-04

Complete Task 3: add symbol defs FTS table (commit b9a30d6).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/store.go — symbol_defs_fts


## 2026-02-04

Complete Task 4: add commits FTS table (commit 3f6ff24).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/store.go — commits_fts


## 2026-02-04

Complete Task 5: add files FTS table (commit 1b0b6d8).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/store.go — files_fts


## 2026-02-04

Complete Task 6: store ISO8601 commit dates (commit 5a88229).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/ingest_commits.go — Use --date=iso-strict


## 2026-02-04

Complete Task 7: add v_last_commit_per_file view (commit 45e4934).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/schema.go — v_last_commit_per_file view


## 2026-02-04

Step 9: update search SQL examples for new FTS tables (commit 231e68b).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ttmp/2026/02/04/REF-001-TEST-INDEXING--refactorio-indexing-playbook/scripts/search-queries/glazed-03-latest-commit-for-file.sql — Use v_last_commit_per_file + files_fts


## 2026-02-04

Step 10: add FTS/view smoke checks (commit e89f7f8).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/refactorindex/ingest_commits_range_smoke_test.go — New FTS + view assertions

