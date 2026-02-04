---
Title: Search improvements implementation plan
Ticket: REF-004-SEARCH-IMPROVEMENTS
Status: active
Topics:
    - search
    - indexing
    - sqlite
    - refactorio
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/pkg/refactorindex/ingest_code_units.go
      Note: Source of code_unit_snapshots data
    - Path: refactorio/pkg/refactorindex/ingest_commits.go
      Note: |-
        Commit date format (ISO8601)
        Commit date format
    - Path: refactorio/pkg/refactorindex/ingest_diff.go
      Note: Source of diff_lines FTS data
    - Path: refactorio/pkg/refactorindex/ingest_symbols.go
      Note: Source of symbol_defs data
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: |-
        Schema additions (FTS tables, view)
        Schema + view definitions
    - Path: refactorio/pkg/refactorindex/store.go
      Note: |-
        ensureFTS helpers and schema init
        FTS helpers and schema init
ExternalSources: []
Summary: Implementation plan to add FTS coverage, store ISO commit dates, and add last-commit-per-file view for refactor-index search.
LastUpdated: 2026-02-04T12:15:00-05:00
WhatFor: Provide a concrete plan for improving search index coverage and query ergonomics.
WhenToUse: Use when implementing search improvements in refactor-index.
---


# Search improvements implementation plan

## Executive Summary
We will expand FTS coverage to enable fast search over code units, symbols, commits, and file paths. We will normalize commit dates to ISO8601 in ingestion, and add a `v_last_commit_per_file` view to simplify common history queries. The work is scoped to schema updates, ingestion tweaks, and updated helper functions in the refactorindex store.

## Problem Statement
The current index only exposes FTS for diff lines and doc hits, which limits search. Commit dates are stored in human-readable form, making ordering and filtering unreliable. Finally, users must hand-write multi-join queries to find the “last commit that touched this file,” which should be a first-class query surface.

## Proposed Solution
1. **FTS expansion**
   - Add FTS tables and triggers for:
     - `code_unit_snapshots.body_text` and `code_unit_snapshots.doc_text`
     - `symbol_defs.name`, `symbol_defs.signature`, `symbol_defs.pkg`
     - `commits.subject`, `commits.body`
     - `files.path`
    - (tree-sitter removed; no FTS for `ts_captures`)
   - Either:
     - Extend `ensureFTS` to handle multiple columns, or
     - Create separate FTS tables per column (simpler but more tables).

2. **ISO8601 commit dates**
   - Update `ingest_commits.go` to call git with `--date=iso-strict`.
   - Store ISO8601 values in the existing `author_date`/`committer_date` columns.

3. **Last commit per file view**
   - Add a SQLite view `v_last_commit_per_file` that exposes the latest commit for each file (per run), including file path, commit hash, committer_date, and status.

## Design Decisions
1. **FTS multi-column vs per-column tables**
   - Preferred: multi-column FTS (fewer tables, simpler query syntax).
   - This requires updating `ensureFTS` and trigger creation to handle multiple columns.

2. **Commit date storage**
   - We will store ISO8601 directly in existing columns. This is a breaking semantic change for new runs but avoids schema expansion.
   - Old runs retain the previous date format; documenting this is required.

3. **View definition**
   - Use commit row id (`commits.id`) as the ordering signal within a run, since it follows `rev-list --reverse` ingestion order.
   - Include `run_id` in the view to avoid cross-run ambiguity.

## Alternatives Considered
- **Separate ISO date columns** (e.g., `author_date_iso`) to avoid format mixing.
  - Rejected for now to keep schema minimal, but may be revisited if backward compatibility becomes a concern.
- **Materialized table for last commit per file**
  - Rejected due to complexity and need for refresh; view is sufficient for now.

## Implementation Plan
1. **FTS helper update**
   - Update `ensureFTS`/`ensureFTSTriggers` to accept a list of columns and create multi-column FTS5 tables.
   - Add new helpers if multi-column support is too invasive.

2. **Schema updates**
   - Add new FTS tables and triggers for:
     - `code_unit_snapshots` (body_text, doc_text)
     - `symbol_defs` (name, signature, pkg)
     - `commits` (subject, body)
     - `files` (path)
    - (tree-sitter removed; no `ts_captures` FTS)
   - Add `CREATE VIEW IF NOT EXISTS v_last_commit_per_file` using window functions:

```sql
CREATE VIEW IF NOT EXISTS v_last_commit_per_file AS
SELECT * FROM (
  SELECT
    c.run_id,
    f.path AS file_path,
    cf.file_id,
    c.id AS commit_id,
    c.hash,
    c.committer_date,
    cf.status,
    cf.old_path,
    cf.new_path,
    ROW_NUMBER() OVER (PARTITION BY c.run_id, cf.file_id ORDER BY c.id DESC) AS rn
  FROM commit_files cf
  JOIN commits c ON c.id = cf.commit_id
  JOIN files f ON f.id = cf.file_id
) WHERE rn = 1;
```

3. **Commit date normalization**
   - Update `loadCommitInfo` in `ingest_commits.go` to call git with `--date=iso-strict`.
   - Document that older runs may use human-readable dates.

4. **Tests and validation**
   - Extend smoke tests to assert FTS table existence and non-zero row counts for new FTS tables.
   - Add a simple test query against `v_last_commit_per_file` to ensure it returns one row per file.

5. **Docs and examples**
   - Update search design docs and help/tutorial pages to use the new FTS tables.
   - Add example queries that use `v_last_commit_per_file`.

## Open Questions
- Do we want to backfill ISO dates for existing runs, or treat re-ingestion as the migration path?
- Should `v_last_commit_per_file` resolve renames more aggressively using `old_path`/`new_path`?

## Risks
- Multi-column FTS changes may require careful trigger logic to avoid partial updates.
- Large FTS rebuilds can be expensive for large repos; need to monitor runtime.
- Mixed date formats across runs could confuse downstream consumers if not documented clearly.

## Implementation Notes
- Update `store.go` to call new FTS helpers from `InitSchema`.
- Ensure `schema_versions` remains forward-compatible with new view/table definitions.

## See Also
- `../../../../../../../refactorio/ttmp/2026/02/04/REF-002-SEARCH-FUNCTIONALITY--search-functionality/design/01-search-system-design.md` - source of search use cases
