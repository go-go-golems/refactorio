---
Title: Frontend/Backend API Alignment Plan (Clean Cutover)
Ticket: REF-008-API-CONTRACT-ALIGNMENT
Status: active
Topics:
    - ui
    - api
    - refactorio
    - frontend
    - backend
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/pkg/workbenchapi/files.go
      Note: Backend files contract
    - Path: refactorio/pkg/workbenchapi/routes.go
      Note: Backend route wiring
    - Path: refactorio/pkg/workbenchapi/search.go
      Note: Backend search contract
    - Path: refactorio/sources/ui-design.md
      Note: UI design spec
    - Path: refactorio/ui/src/api/baseApi.ts
      Note: Shared API base for all slices
    - Path: refactorio/ui/src/api/workspaces.ts
      Note: Workspace API usage
    - Path: refactorio/ui/src/types/api.ts
      Note: UI type assumptions enumerated
ExternalSources: []
Summary: Exhaustive alignment analysis for all RTK Query slices and backend endpoints, including clean cutover contract decisions and missing-information inventory.
LastUpdated: 2026-02-05T16:50:00-05:00
WhatFor: Provide a complete, endpoint-by-endpoint plan to align the Refactorio Workbench frontend and backend API with a clean cutover.
WhenToUse: Use when implementing the contract alignment work or reviewing the required changes across UI and backend.
---


# Frontend/Backend API Alignment Plan (Clean Cutover)

## Goal
Align the Refactorio Workbench frontend and backend API contracts with a clean cutover (no backward compatibility), covering every RTK Query slice and every backend endpoint. This document includes an exhaustive mismatch inventory and a definitive target contract for each endpoint.

## Decision: Source of Truth for the Contract
Use the backend as the wire-contract source of truth, with the following modifications:
- Keep backend response shapes as the canonical baseline where possible.
- Add missing fields that the current UI or UI design spec requires.
- Update the UI to match the canonical backend contract.
- Do not support dual or legacy shapes; clean cutover only.

## Contract Normalization Rules (Clean Cutover)
Applies to all endpoints in scope:
- All list endpoints return `{ items, limit, offset, total? }`.
- IDs are explicit and consistent:
  - Numeric IDs remain `id` (not `run_id` in list/detail responses).
  - Hash IDs remain `hash` (or `symbol_hash`, `unit_hash`) depending on domain.
- File paths use `path` consistently for backend responses.
- Package paths use `pkg` consistently in backend responses.
- Line/column fields use `line`, `col` for a single position, and `start_line`, `start_col`, `end_line`, `end_col` for ranges.
- Times are RFC3339Nano strings.
- Errors use the existing backend error schema (already consistent across endpoints).

## Inventory: RTK Query Slices
Slices and current endpoints as implemented:
- `workspaces.ts`: `GET /workspaces`, `GET /db/info`, `GET /workspaces/:id`, `POST /workspaces`, `PATCH /workspaces/:id`, `DELETE /workspaces/:id`
- `runs.ts`: `GET /runs`, `GET /runs/:id`, `GET /runs/:id/summary`
- `sessions.ts`: `GET /sessions`, `GET /sessions/:id`
- `symbols.ts`: `GET /symbols`, `GET /symbols/:hash`, `GET /symbols/:hash/refs`
- `codeUnits.ts`: `GET /code-units`, `GET /code-units/:hash`, `GET /code-units/:hash/history`
- `commits.ts`: `GET /commits`, `GET /commits/:hash`, `GET /commits/:hash/files`
- `diffs.ts`: `GET /diff-runs`, `GET /diff/:run_id/files`, `GET /diff/:run_id/file`
- `docs.ts`: `GET /docs/terms`, `GET /docs/hits`
- `files.ts`: `GET /files`, `GET /file`, `GET /files/history`
- `search.ts`: `POST /search`, `GET /search/symbols`, `GET /search/code-units`, `GET /search/commits`

## Inventory: Backend Endpoints
All current backend endpoints in `refactorio/pkg/workbenchapi`:
- `GET /health`
- `GET /workspaces`, `POST /workspaces`
- `GET /workspaces/:id`, `PATCH /workspaces/:id`, `DELETE /workspaces/:id`
- `GET /db/info`
- `GET /sessions`, `POST /sessions`, `GET /sessions/:id`
- `GET /runs`, `GET /runs/:id`, `GET /runs/:id/summary`
- `GET /raw-outputs`, `GET /runs/:id/raw-outputs`
- `GET /symbols`, `GET /symbols/:hash`, `GET /symbols/:hash/refs`
- `GET /code-units`, `GET /code-units/:hash`, `GET /code-units/:hash/history`, `POST /code-units/:hash/diff`
- `GET /commits`, `GET /commits/:hash`, `GET /commits/:hash/files`, `GET /commits/:hash/diff`
- `GET /diff-runs`, `GET /diff/:run_id/files`, `GET /diff/:run_id/file`
- `GET /docs/terms`, `GET /docs/hits`
- `GET /files`, `GET /file`, `GET /files/history`
- `POST /search`
- `GET /search/symbols`, `GET /search/code-units`, `GET /search/diff`, `GET /search/commits`, `GET /search/docs`, `GET /search/files`
- `GET /tree-sitter/captures`

## Slice-by-Slice Alignment (Exhaustive)

### 1) Workspaces Slice
Current UI expectations:
- `GET /workspaces` returns `{ workspaces: Workspace[] }`.
- `Workspace` includes `id`, `name`, `db_path`, `repo_root`, `created_at`, `updated_at`.
- `POST /workspaces` accepts `name`, `db_path`, `repo_root`.
- `GET /db/info` returns `tables: string[]`, `fts_tables: string[]`, `row_counts`.

Backend contract today:
- `GET /workspaces` returns `{ items: Workspace[] }`.
- `POST /workspaces` requires `id` and normalizes paths.
- `GET /db/info` returns `tables: map`, `fts_tables: map`, no row counts, and `features` map.

Clean cutover target:
- Keep backend as canonical and update UI.
- `GET /workspaces` uses `{ items }`.
- `POST /workspaces` requires `id` explicitly. UI must provide `id` or auto-generate.
- `GET /db/info` remains as backend contract; UI must render maps and optionally derived counts.

Required frontend changes:
- Update `workspaces.ts` to transform `{ items }`.
- Add `id` field to `WorkspaceForm` or auto-generate from name.
- Update `DBInfo` type and dashboard rendering to use `tables` and `fts_tables` maps.

Required backend changes:
- Optional: add `row_counts` if the UI needs DB summary cards without extra queries.

### 2) Sessions Slice
Current UI expectations:
- `GET /sessions` returns `{ sessions: Session[] }`.
- `Session` uses fixed `SessionAvailability` fields and `runs` as fixed fields.

Backend contract today:
- `GET /sessions` returns `{ items: Session[] }`.
- `availability` is a map keyed by domain.

Clean cutover target:
- Use backend `items` response and map `availability` into a fixed typed structure in UI (or keep it as map).

Required frontend changes:
- Update `sessions.ts` transform to use `{ items }`.
- Update UI types to handle `availability` map.

Required backend changes:
- None.

### 3) Runs Slice
Current UI expectations:
- `Run` uses `run_id` and expects `started_at`, `root_path`, `git_from`, `git_to`.
- `GET /runs/:id/summary` returns fields like `symbols_count`, `diff_files_count` directly.

Backend contract today:
- `RunRecord` uses `id`.
- `GET /runs/:id/summary` returns `{ run_id, counts: {...} }`.

Clean cutover target:
- Use backend `id` fields.
- Use backend summary `counts` map.

Required frontend changes:
- Rename `run_id` to `id` in UI types and code.
- Update summary rendering to use `counts` map.

Required backend changes:
- Optional: add `total` to run lists if pagination needs it.

### 4) Symbols Slice
Current UI expectations:
- `Symbol` fields: `symbol_hash`, `package_path`, `file_path`, `start_line`, `start_col`, `end_line`, `end_col`, `exported`.
- `SymbolRef` fields: `file_path`, `start_line`, `start_col`, `end_line`, `end_col`, `is_declaration`.

Backend contract today:
- `SymbolRecord` uses `pkg`, `file`, `line`, `col`, `is_exported`.
- Symbol refs use `path`, `line`, `col`, `is_decl`.

Clean cutover target:
- Keep backend fields as canonical, update UI to use `pkg`, `file`, `line`, `col`, `is_exported`.
- If end range is required, add `end_line`, `end_col` to backend by joining span data if available. Otherwise remove range display from UI.

Required frontend changes:
- Update types and renderers to `pkg`, `file`, `line`, `col`.
- Adjust SymbolDetail to handle missing `end_line`.

Required backend changes:
- Optional: add end range fields if needed by UI design.

### 5) Code Units Slice
Current UI expectations:
- List uses `code_unit_hash`, `package_path`, `file_path`.
- Detail uses `body` and `doc_comment`.

Backend contract today:
- Uses `unit_hash`, `pkg`, `file`, `body_text`, `doc_text`.

Clean cutover target:
- Use backend names (`unit_hash`, `pkg`, `file`, `body_text`, `doc_text`).

Required frontend changes:
- Update types and renderers to match backend.
- Update `codeUnits.ts` to use `body_q` query param when searching body text.

Required backend changes:
- None.

### 6) Commits Slice
Current UI expectations:
- `Commit` uses `commit_hash` and `commit_date`.
- Commit files include `additions` and `deletions`.

Backend contract today:
- Commits use `hash`, `committer_date`.
- Commit files include `path`, `status`, `old_path`, `new_path`, `blob_old`, `blob_new`.

Clean cutover target:
- Use backend names (`hash`, `committer_date`).
- If additions/deletions are required, add them to backend by joining diff stats.

Required frontend changes:
- Update commit types and renderers.
- Update commit files renderer to handle missing additions/deletions or remove those columns.

Required backend changes:
- Optional: add additions/deletions to commit files.

### 7) Diffs Slice
Current UI expectations:
- `DiffRun` includes `files_count`.
- `DiffFile` includes `file_path`, `hunks_count`, `additions`, `deletions`.
- `DiffHunk` uses `old_count`, `new_count` and `DiffLine` uses `old_line`, `new_line`, `content`.

Backend contract today:
- Diff runs reuse `RunRecord` (no `files_count`).
- Diff files use `path`, `status`, `old_path`, `new_path`.
- Diff hunks use `old_lines`, `new_lines` and line fields `line_no_old`, `line_no_new`, `text`.

Clean cutover target:
- Use backend names and add missing counts if needed.
- Diff runs should include `files_count` if UI depends on it.
- Diff files should include additions/deletions/hunks_count if UI keeps those columns.

Required frontend changes:
- Update diff types and renderers to backend field names.

Required backend changes:
- Add counts to diff run/file responses if needed for UI tables.

### 8) Docs Slice
Current UI expectations:
- Doc hits use `file_path`.

Backend contract today:
- Doc hits use `path`.

Clean cutover target:
- Use backend `path`.

Required frontend changes:
- Update `DocHit` and Docs page to render `path`.

Required backend changes:
- None.

### 9) Files Slice
Current UI expectations:
- File entries use `is_dir` and `children_count`.
- Tree assumes nested structure or expandable paths.

Backend contract today:
- File entries use `kind` (`file` or `dir`) with flat segments.
- The endpoint supports `prefix` filtering but does not return nested children in a single call.

Clean cutover target:
- Use backend `kind` and `prefix`-based expansion.
- UI must implement lazy loading and maintain a `childrenMap`.

Required frontend changes:
- Update `FileEntry` to `kind`.
- Implement `onExpand` to request `/files?prefix=...` and populate `childrenMap`.

Required backend changes:
- Optional: add `children_count` to improve UI display.

### 10) Search Slice
Current UI expectations:
- `SearchResult` includes `id`, `label`, `location`.

Backend contract today:
- Search results include `primary`, `secondary`, `path`, `line`, `col`, `snippet`.
- Typed search endpoints return domain-specific records, not unified results.

Clean cutover target:
- Use backend fields and map them to UI-friendly labels in a thin adapter layer.
- Define a stable `id` for UI selection, derived from backend fields.

Required frontend changes:
- Update `SearchResult` type to match backend or add a mapping layer in the UI.
- Add support for `/search/diff`, `/search/docs`, `/search/files` if the UI needs them.

Required backend changes:
- None, unless a unified UI-friendly search schema is preferred.

## Missing Information Inventory

### Missing From Backend for Current UI
- `DBInfo.row_counts` used by Dashboard cards.
- `DiffRun.files_count` used in Diffs list.
- `DiffFile.additions`, `DiffFile.deletions`, `DiffFile.hunks_count` used in Diffs list.
- `CommitFile.additions`, `CommitFile.deletions` used in Commits inspector.
- Symbol range end positions (`end_line`, `end_col`) expected in UI.
- Search result `id`, `label`, `location` expected in UI.
- File entry `is_dir` and `children_count` expected in UI.
- Run summary counts in flat fields (`symbols_count`, `diff_files_count`) expected in UI.

### Missing From Frontend for Current Backend
- Workspace creation requires `id` but UI does not provide it.
- Session scoping is not applied in UI queries (`session_id` or `run_id`).
- File tree does not use `prefix` expansion.
- Code unit body search uses `q`, backend expects `body_q`.
- Symbols list uses `q` but backend list endpoint expects `name`; symbol search should use `/search/symbols`.
- Search UI does not use `/search/diff`, `/search/docs`, `/search/files`.
- UI does not use `raw-outputs`, `commit diff`, or `code-unit diff` endpoints.
- UI does not surface `tree-sitter` captures even though endpoint exists.

### Missing From Both vs UI Design Spec
Items mentioned in `refactorio/sources/ui-design.md` but not implemented in UI or backend:
- Refactor plan builder, plan storage, and plan execution endpoints.
- Audit workflows and reports.
- Command palette and quick actions.
- Workspace schema upgrade action.
- Global session/workspace validation flows in the UI.
- Admin screens for raw outputs, schema info, and settings.
- Full “Refactor” section: Plans, Runs, Audits, Reports.

## Clean Cutover Execution Plan

Phase 1: Contract Definition
- Lock the canonical contract for each endpoint in this document.
- Decide which missing fields must be added to backend vs removed from UI.

Phase 2: Backend Updates
- Add fields needed by UI (only those explicitly required after Phase 1).
- Add any missing endpoints required by the UI design spec.

Phase 3: Frontend Updates
- Update RTK Query slices to the canonical contract.
- Update UI types and renderers to backend field names.
- Implement session scoping across all pages.
- Implement file tree lazy loading using `prefix`.

Phase 4: Validation
- Run the UI against the live backend using the playbook.
- Verify all pages render without undefined fields.
- Confirm that session scoping produces consistent results.

## Related Files (Suggested)
- `refactorio/ui/src/api/*.ts`
- `refactorio/ui/src/types/api.ts`
- `refactorio/ui/src/pages/*.tsx`
- `refactorio/pkg/workbenchapi/*.go`
- `refactorio/sources/ui-design.md`
