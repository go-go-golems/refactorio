---
Title: Diary
Ticket: REF-004-SEARCH-IMPROVEMENTS
Status: active
Topics:
    - search
    - indexing
    - sqlite
    - refactorio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/ttmp/2026/02/04/REF-004-SEARCH-IMPROVEMENTS--search-improvements-for-refactor-index/design-doc/01-search-improvements-implementation-plan.md
      Note: Removed tree-sitter references
    - Path: refactorio/ttmp/2026/02/04/REF-004-SEARCH-IMPROVEMENTS--search-improvements-for-refactor-index/tasks.md
      Note: Reframed task breakdown
ExternalSources: []
Summary: Implementation diary for REF-004-SEARCH-IMPROVEMENTS.
LastUpdated: 2026-02-04T18:05:00-05:00
WhatFor: Track changes, decisions, and validations while implementing search improvements.
WhenToUse: Use while executing REF-004 tasks to record progress and validation steps.
---


# Diary

## Goal
Provide a step-by-step record of implementing search improvements for refactor-index, including schema changes, FTS expansions, and documentation updates.

## Step 1: Reframe plan and tasks for REF-004
I updated the task list to reflect the current scope (tree-sitter removed) and broke work into sequential tasks with subtasks. I also updated the implementation plan to remove tree-sitter references and align the scope with current refactor-index capabilities.

This sets up a clear execution path and avoids investing in removed features.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Begin the REF-004 search improvements work, starting with tasks and plan alignment.

**Inferred user intent:** Move on to the next ticket and execute its tasks sequentially with clear bookkeeping.

**Commit (code):** N/A

### What I did
- Rewrote `tasks.md` to reflect current scope and subtasks.
- Updated the implementation plan to remove tree-sitter references.

### Why
- Tree-sitter has been removed from refactor-index, so search improvements should not include its FTS.

### What worked
- The plan and tasks now match the current product scope.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/ttmp/2026/02/04/REF-004-SEARCH-IMPROVEMENTS--search-improvements-for-refactor-index/tasks.md`.
- Review `refactorio/ttmp/2026/02/04/REF-004-SEARCH-IMPROVEMENTS--search-improvements-for-refactor-index/design-doc/01-search-improvements-implementation-plan.md`.

### Technical details
- Tree-sitter references removed from plan; tasks reorganized into 10 sequential items.

## Step 2: Task 1 - Add multi-column FTS helper
I introduced a new multi-column FTS helper that preserves the existing single-column behavior while enabling multi-column triggers. This is the foundation for the upcoming FTS expansions.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Start implementing REF-004 tasks sequentially, beginning with the FTS helper.

**Inferred user intent:** Build the FTS scaffolding needed for new search capabilities.

**Commit (code):** 371bcd2 — "refactorindex: add multi-column FTS helper"

### What I did
- Added `ensureFTSColumns` and trigger helpers for multi-column FTS tables.
- Kept `ensureFTS` as a wrapper for single-column behavior.

### Why
- Multi-column FTS is required for code units, symbols, commits, and files.

### What worked
- Existing FTS usage still passes through the wrapper without changes.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Ensure the trigger SQL is correct for multi-column tables and does not regress existing FTS usage.

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/pkg/refactorindex/store.go` for FTS helper changes.
- Validate with `go test ./refactorio/pkg/refactorindex`.

### Technical details
- `ensureFTS` now calls `ensureFTSColumns` with a single column.

## Step 3: Task 2 - Add FTS for code unit snapshots
I added an FTS table for `code_unit_snapshots` covering `body_text` and `doc_text`. This extends full-text search to code unit bodies and doc comments.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Continue executing the REF-004 tasks in sequence.

**Inferred user intent:** Expand search coverage to include code unit content.

**Commit (code):** d6e4ce2 — "refactorindex: add FTS for code units"

### What I did
- Added `code_unit_snapshots_fts` via `ensureFTSColumns` during schema init.
- Bumped schema version.

### Why
- Full-text search over code unit bodies and doc text is a core search improvement.

### What worked
- FTS table is created automatically on schema init.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Verify FTS trigger behavior for nullable `doc_text` fields.

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/pkg/refactorindex/store.go` for FTS init.
- Review `refactorio/pkg/refactorindex/schema.go` for schema version bump.
- Validate with `go test ./refactorio/pkg/refactorindex`.

### Technical details
- FTS table name: `code_unit_snapshots_fts` with columns `body_text`, `doc_text`.

## Step 4: Task 3 - Add FTS for symbol defs
I added an FTS table for `symbol_defs` covering `name`, `signature`, and `pkg`. This enables faster symbol search across core fields.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Continue implementing the search improvements sequentially.

**Inferred user intent:** Expand FTS coverage to symbol definitions.

**Commit (code):** b9a30d6 — "refactorindex: add FTS for symbol defs"

### What I did
- Added `symbol_defs_fts` via `ensureFTSColumns`.
- Bumped schema version.

### Why
- Symbol search is a primary use case for refactor-index search.

### What worked
- FTS table creation is integrated into schema init.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Confirm `signature` fields are suitable for FTS tokenization.

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/pkg/refactorindex/store.go` and `schema.go`.
- Validate with `go test ./refactorio/pkg/refactorindex`.

### Technical details
- FTS table name: `symbol_defs_fts` with columns `name`, `signature`, `pkg`.

## Step 5: Task 4 - Add FTS for commits
I added an FTS table for `commits` covering `subject` and `body`. This enables fast search over commit messages.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Continue implementing remaining FTS tables.

**Inferred user intent:** Improve search coverage for commit messages.

**Commit (code):** 3f6ff24 — "refactorindex: add FTS for commits"

### What I did
- Added `commits_fts` via `ensureFTSColumns`.
- Bumped schema version.

### Why
- Commit message search is a common investigative workflow.

### What worked
- FTS table creation integrated into schema init.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Check tokenization behavior for large commit bodies.

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/pkg/refactorindex/store.go` and `schema.go`.
- Validate with `go test ./refactorio/pkg/refactorindex`.

### Technical details
- FTS table name: `commits_fts` with columns `subject`, `body`.

## Step 6: Task 5 - Add FTS for files
I added an FTS table for `files.path` to enable fast path searches.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Continue extending FTS coverage.

**Inferred user intent:** Allow fast searching of file paths.

**Commit (code):** 1b0b6d8 — "refactorindex: add FTS for files"

### What I did
- Added `files_fts` via `ensureFTSColumns`.
- Bumped schema version.

### Why
- File path search is a common search workflow.

### What worked
- The FTS table is created during schema init.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Verify file path tokenization meets search needs.

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/pkg/refactorindex/store.go` and `schema.go`.
- Validate with `go test ./refactorio/pkg/refactorindex`.

### Technical details
- FTS table name: `files_fts` with column `path`.

## Step 7: Task 6 - Store ISO8601 commit dates
I updated commit ingestion to use `--date=iso-strict`, ensuring stored dates are ISO8601 for consistent ordering and filtering.

### Prompt Context
**User prompt (verbatim):** "once you are done, work on REF-004"

**Assistant interpretation:** Normalize commit date formatting during ingestion.

**Inferred user intent:** Make commit dates sortable and consistent across runs.

**Commit (code):** 5a88229 — "refactorindex: store ISO commit dates"

### What I did
- Added `--date=iso-strict` to the git show call in `loadCommitInfo`.

### Why
- ISO8601 dates are easier to sort, filter, and compare.

### What worked
- Commit ingestion now stores normalized dates.

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Ensure downstream tooling doesn’t rely on the previous human-readable format.

### What should be done in the future
- N/A

### Code review instructions
- Review `refactorio/pkg/refactorindex/ingest_commits.go`.
- Validate with `go test ./refactorio/pkg/refactorindex`.

### Technical details
- Git date format now uses `iso-strict` for `%ad` and `%cd`.
