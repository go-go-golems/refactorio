---
Title: Backend REST API Implementation Plan
Ticket: REF-007-INDEX-BROWSE-UI
Status: active
Topics:
    - ui
    - api
    - refactorio
    - backend
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: Schema defines tables exposed by the API.
    - Path: refactorio/pkg/refactorindex/query.go
      Note: Existing query helpers to reuse or extend.
    - Path: refactorio/pkg/refactorindex/store.go
      Note: DB open/init helpers and FTS setup.
    - Path: refactorio/pkg/refactor/js/modules/refactorindex/refactorindex.go
      Note: Current JS API shapes to mirror in REST payloads.
Summary: Step-by-step plan to implement the Workbench REST API server.
LastUpdated: 2026-02-05T09:30:00-05:00
WhatFor: Guide the implementation of the backend server, from routing to query layer.
WhenToUse: Use when building the Workbench API in Go.
---

# Backend REST API Implementation Plan

## Executive Summary
This plan builds a local-first REST API for the Refactorio Workbench UI. The implementation will reuse the existing refactorindex store and schema, add a focused query layer for UI-friendly shapes, and expose endpoints that map directly to the UI spec. The goal is to ship a browsing MVP first, then layer on refactor assistance endpoints in a later phase.

## Deliverables
- A new `refactorio` subcommand that runs the API server.
- A workspace registry file to manage DB path + repo root.
- A query layer that wraps SQLite queries with typed filters.
- REST handlers for the routes defined in the analysis doc.
- Smoke tests that validate key endpoints against a sample DB.

## Proposed Package Layout
```
refactorio/
  cmd/
    refactorio/
      api.go            # new cobra command
  pkg/
    workbenchapi/
      server.go         # http server setup, router
      routes.go         # route registration
      types.go          # request/response structs
      workspace.go      # workspace registry + config
      sessions.go       # session grouping logic
      queries/
        runs.go
        search.go
        symbols.go
        code_units.go
        diffs.go
        commits.go
        docs.go
        files.go
      handlers/
        runs.go
        search.go
        symbols.go
        code_units.go
        diffs.go
        commits.go
        docs.go
        files.go
```

## Phase Plan

### Phase 1: Server Skeleton + Workspace Registry
1. Add a `refactorio api serve` Cobra command.
2. Implement workspace registry stored in a JSON or YAML file under `~/.config/refactorio/workspaces.json`.
3. Define shared response helpers and error format.
4. Add `GET /api/workspaces` and `POST /api/workspaces` routes.

### Phase 2: Runs, Sessions, and DB Info
1. Add DB open helpers that take `workspace_id` and return a `*sql.DB`.
2. Implement `GET /api/db/info` with schema and FTS detection.
3. Implement runs list, run detail, run summary, raw outputs routes.
4. Add session resolver that groups runs by `(root_path, git_from, git_to)`.

### Phase 3: Search + Core Explore Endpoints
1. Add FTS-backed search queries for symbols, code units, diffs, commits, docs, files.
2. Implement unified `POST /api/search` that calls per-type search helpers.
3. Implement symbol list and detail endpoints using `ListSymbolInventory`.
4. Implement code unit list/detail/history endpoints.
5. Implement diff list and diff file detail endpoints.
6. Implement commit list/detail/files endpoints.
7. Implement doc terms/hits endpoints.

### Phase 4: File Viewer Support
1. Add file listing endpoint with prefix tree expansion.
2. Implement file content endpoint using repo_root or `git show`.
3. Add file history endpoint using `commit_files`.

### Phase 5: Optional Refactor Assistance
1. Define plan storage location and schema.
2. Implement plan create and validate endpoints.
3. Wire gopls `prepare_rename` and ref ingestion hooks.

## Query Layer Details

### Workspace Resolution
- Prefer `workspace_id` query parameter.
- If missing, allow `db_path` for ad-hoc usage.
- If repo_root is missing, return 409 from file content endpoints.

### Typed Filters
Create filter structs mirroring the UI needs. Example:
```go
type SymbolFilter struct {
  RunID int64
  Name string
  Pkg string
  Kind string
  Path string
  ExportedOnly bool
  Limit int
  Offset int
}
```

Use the existing `refactorindex` filter types where possible, and add new ones for commits, diffs, and runs.

### FTS Queries
Implement a helper that maps a query string and FTS table to a deterministic SQL statement. Always cap `limit` and ensure the FTS table exists before querying.

## Handler Conventions
- Use `context.Context` in all query calls.
- Normalize all query parameters into filter structs.
- Return HTTP 400 for invalid filters and 409 for missing repo_root.
- Use `github.com/pkg/errors` to wrap lower-level errors.

## Testing Strategy
- Add unit tests for query builders that validate SQL and params.
- Add handler tests using an in-memory SQLite DB populated with a minimal schema.
- Add smoke tests for `GET /api/db/info`, `GET /api/runs`, and `POST /api/search`.

## Risks and Mitigations
- FTS tables may not exist in older DBs. Mitigation: check for table presence and return a warning in the response.
- Session grouping can be ambiguous. Mitigation: allow manual overrides and expose grouping metadata.
- File content retrieval depends on repo_root availability. Mitigation: return a clear error when repo_root is missing.

## Success Criteria
- UI can load a workspace, list sessions, and browse symbols, diffs, commits, and docs without N+1 calls.
- Unified search returns deterministic results across all indexed domains.
- Errors are consistent and actionable.

## Next Steps
- Implement Phase 1 and Phase 2 first, then validate with the UI spec's MVP list.
- After Phase 3, synchronize API payloads with frontend RTK Query slices.
