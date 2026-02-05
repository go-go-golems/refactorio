---
Title: Backend REST API Architecture & Route Analysis
Ticket: REF-007-INDEX-BROWSE-UI
Status: active
Topics:
    - ui
    - api
    - refactorio
    - backend
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/pkg/refactorindex/ingest_code_units.go
      Note: Code unit snapshots (body/doc text, spans).
    - Path: refactorio/pkg/refactorindex/ingest_commits.go
      Note: Commit ingestion semantics, commit/file blobs, ISO dates.
    - Path: refactorio/pkg/refactorindex/ingest_diff.go
      Note: Diff ingestion semantics (diff_files/hunks/lines, raw outputs).
    - Path: refactorio/pkg/refactorindex/ingest_symbols.go
      Note: Symbol definitions ingestion (symbol_defs + symbol_occurrences).
    - Path: refactorio/pkg/refactorindex/query.go
      Note: Existing query helpers for symbols/refs/doc hits/files used as baseline for API.
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: Core DB schema (runs, diffs, symbols, code units, commits, doc hits, FTS tables).
    - Path: refactorio/pkg/refactorindex/store.go
      Note: DB initialization, FTS setup, run metadata, raw outputs.
ExternalSources: []
Summary: Architecture + route-by-route REST API spec for the Refactorio Workbench UI.
LastUpdated: 2026-02-05T09:30:00-05:00
WhatFor: Provide a concrete backend design that maps UI needs to SQLite queries and API contracts.
WhenToUse: Use when implementing the Workbench REST API server.
---


# Backend REST API Architecture & Route Analysis

## Executive Summary
This document defines a REST API for the Refactorio Workbench UI that sits on top of the refactor-index SQLite database and an optional repo checkout. The goal is to turn the existing schema and ingestion outputs into UI-ready endpoints without forcing the frontend into N+1 query patterns. The API design follows a thin server pattern: minimal state, deterministic queries, and consistent pagination. It also anticipates optional refactor-assist features (plans, runs) without blocking the initial browsing-focused MVP.

## Primary Source
The UI requirements are derived from the imported spec at `refactorio/ttmp/2026/02/04/REF-007-INDEX-BROWSE-UI--index-browse-ui-backend-api/sources/local/ui-design.md.md` (source: `refactorio/sources/ui-design.md`).

## Scope and Non-Goals
Scope includes workspaces, sessions (grouped runs), search, exploration endpoints for files, symbols, code units, commits, diffs, docs, and optional tree-sitter captures. The API also serves raw outputs and run metadata for transparency. Non-goals include implementing the refactor plan execution engine, background job orchestration, or a full multi-user auth layer. The server remains local-first and assumes trusted access.

## Architecture Overview
The backend is a read-heavy HTTP service layered over SQLite, with a small amount of config state for workspaces and manual session mapping.

**Core layers**
- API server: `net/http` with structured routing and JSON responses.
- Workspace registry: local config file storing `{id, name, db_path, repo_root}`.
- Session resolver: runtime grouping of runs by `(root_path, git_from, git_to)` with manual overrides.
- Query layer: SQL helpers that map UI filters to efficient queries and FTS searches.
- File content provider: reads file content from repo root or `git show <ref>:<path>`.
- Response shaping: consistent paging, sorting, and error structure.

**Data inputs**
- SQLite tables from `refactorindex` schema.
- Raw outputs on disk (gopls, rg, git diffs).
- Optional repo checkout for file contents and `git show` queries.

## API Conventions
**Base URL**: `/api`

**Workspace selection**
Every request that touches a DB must resolve a workspace. Use a `workspace_id` query param, with a fallback `db_path` query param for ad-hoc usage. If both are present, `workspace_id` wins.

**Pagination**
All list endpoints accept `limit` and `offset`. Defaults: `limit=100`, `offset=0`. Endpoints return `{items, limit, offset, total}` when `total` is cheap, otherwise omit `total`.

**Errors**
Return JSON error payloads with a stable shape.

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "workspace_id is required",
    "details": {"field": "workspace_id"}
  }
}
```

**Sorting**
Each list endpoint documents its default ordering. Custom ordering is supported via `sort` and `direction` (asc/desc) when trivial to implement.

## Core Entity Shapes
These are canonical JSON shapes referenced throughout the routes.

**Workspace**
```json
{
  "id": "glazed",
  "name": "glazed",
  "db_path": "/abs/path/to/index.sqlite",
  "repo_root": "/abs/path/to/repo",
  "created_at": "2026-02-05T09:00:00Z",
  "updated_at": "2026-02-05T09:00:00Z"
}
```

**Run**
```json
{
  "id": 42,
  "status": "success",
  "started_at": "2026-02-04T10:12:00Z",
  "finished_at": "2026-02-04T10:14:00Z",
  "tool_version": "dev",
  "git_from": "HEAD~20",
  "git_to": "HEAD",
  "root_path": "/abs/path/to/repo",
  "args_json": "{...}",
  "error_json": null,
  "sources_dir": "/abs/path/to/sources"
}
```

**Session**
```json
{
  "id": "glazed:HEAD~20..HEAD",
  "workspace_id": "glazed",
  "root_path": "/abs/path/to/repo",
  "git_from": "HEAD~20",
  "git_to": "HEAD",
  "runs": {
    "commits": 12,
    "diff": 15,
    "symbols": 18,
    "code_units": 19,
    "doc_hits": 20,
    "gopls_refs": 21,
    "tree_sitter": null
  },
  "availability": {
    "commits": true,
    "diff": true,
    "symbols": true,
    "code_units": true,
    "doc_hits": true,
    "gopls_refs": false,
    "tree_sitter": false
  },
  "last_updated": "2026-02-04T10:14:00Z"
}
```

**SearchResult**
```json
{
  "type": "symbol",
  "primary": "Client",
  "secondary": "github.com/acme/project/internal/api",
  "path": "internal/api/client.go",
  "line": 42,
  "col": 6,
  "snippet": "type Client struct { ... }",
  "run_id": 18,
  "commit_hash": "",
  "score": 0.87,
  "payload": {"symbol_hash": "..."}
}
```

## Route Catalog and Detailed Specifications

### Workspace Routes

#### GET /api/workspaces
Purpose: list configured workspaces.

Request: no body. Optional query: `include_stats` (bool).

Response schema:
```json
{
  "items": [Workspace],
  "limit": 0,
  "offset": 0
}
```

Pseudocode:
```go
func listWorkspaces(w http.ResponseWriter, r *http.Request) {
  cfg := loadWorkspaceConfig()
  writeJSON(w, map[string]any{"items": cfg.Workspaces})
}
```

#### POST /api/workspaces
Purpose: register a new workspace (db path + optional repo root).

Request schema:
```json
{
  "id": "glazed",
  "name": "glazed",
  "db_path": "/abs/path/index.sqlite",
  "repo_root": "/abs/path/repo"
}
```

Response schema: Workspace.

Pseudocode:
```go
func createWorkspace(w http.ResponseWriter, r *http.Request) {
  body := decodeJSON(r)
  validatePaths(body)
  cfg := loadWorkspaceConfig()
  cfg.Add(body)
  cfg.Save()
  writeJSON(w, body)
}
```

#### GET /api/workspaces/:id
Purpose: fetch a workspace definition.

Request: path param `id`.

Response: Workspace.

Pseudocode:
```go
func getWorkspace(w http.ResponseWriter, r *http.Request) {
  ws := resolveWorkspaceByID(r)
  writeJSON(w, ws)
}
```

#### PATCH /api/workspaces/:id
Purpose: update workspace fields (db_path, repo_root, name).

Request schema (partial):
```json
{
  "name": "glazed",
  "db_path": "/abs/path/index.sqlite",
  "repo_root": "/abs/path/repo"
}
```

Response: Workspace.

Pseudocode:
```go
func updateWorkspace(w http.ResponseWriter, r *http.Request) {
  ws := resolveWorkspaceByID(r)
  patch := decodeJSON(r)
  ws.Apply(patch)
  saveConfig()
  writeJSON(w, ws)
}
```

#### DELETE /api/workspaces/:id
Purpose: remove a workspace from config.

Pseudocode:
```go
func deleteWorkspace(w http.ResponseWriter, r *http.Request) {
  cfg := loadWorkspaceConfig()
  cfg.Remove(id)
  cfg.Save()
  writeJSON(w, map[string]any{"deleted": id})
}
```

### DB Info and Health

#### GET /api/db/info
Purpose: read schema version, table/FTS availability, and feature flags.

Request: query `workspace_id` or `db_path`.

Response schema:
```json
{
  "workspace_id": "glazed",
  "db_path": "/abs/path/index.sqlite",
  "schema_version": 17,
  "tables": {"meta_runs": true, "diff_files": true},
  "fts_tables": {"diff_lines_fts": true, "doc_hits_fts": true},
  "features": {"fts": true, "tree_sitter": false, "gopls_refs": true}
}
```

Pseudocode:
```go
func dbInfo(w http.ResponseWriter, r *http.Request) {
  ws := resolveWorkspace(r)
  db := openDB(ws.DBPath)
  defer db.Close()
  schemaVersion := queryInt(db, "SELECT max(version) FROM schema_versions")
  tables := scanSQLiteMaster(db)
  fts := filterFTSTables(tables)
  features := inferFeatures(tables, fts)
  writeJSON(w, info)
}
```

### Runs and Raw Outputs

#### GET /api/runs
Purpose: list runs with basic metadata.

Request query: `workspace_id`, optional `status`, `root_path`, `git_from`, `git_to`, `started_after`, `started_before`, `limit`, `offset`.

Response schema:
```json
{
  "items": [Run],
  "limit": 100,
  "offset": 0
}
```

Pseudocode:
```go
func listRuns(w http.ResponseWriter, r *http.Request) {
  ws := resolveWorkspace(r)
  db := openDB(ws.DBPath)
  filters := parseRunFilters(r)
  rows := queryRuns(db, filters)
  writeJSON(w, rows)
}
```

SQL sketch:
```sql
SELECT id, started_at, finished_at, status, tool_version, git_from, git_to, root_path, args_json, error_json, sources_dir
FROM meta_runs
WHERE (? = '' OR status = ?)
ORDER BY started_at DESC
LIMIT ? OFFSET ?;
```

#### GET /api/runs/:id
Purpose: return a single run.

Response: Run.

Pseudocode:
```go
func getRun(w http.ResponseWriter, r *http.Request) {
  run := queryRunByID(db, id)
  writeJSON(w, run)
}
```

#### GET /api/runs/:id/summary
Purpose: return row counts per domain for a run.

Response schema:
```json
{
  "run_id": 42,
  "counts": {
    "diff_files": 120,
    "diff_hunks": 350,
    "diff_lines": 4200,
    "symbols": 800,
    "code_units": 500,
    "doc_hits": 240,
    "commits": 20,
    "commit_files": 340,
    "symbol_refs": 1200,
    "ts_captures": 0
  }
}
```

Pseudocode:
```go
func runSummary(w http.ResponseWriter, r *http.Request) {
  counts := map[string]int{}
  for _, table := range tablesByRunID {
    counts[table] = queryInt(db, "SELECT count(*) FROM "+table+" WHERE run_id = ?", runID)
  }
  writeJSON(w, summary)
}
```

#### GET /api/runs/:id/raw-outputs
Purpose: list raw outputs for a run.

Response schema:
```json
{
  "items": [
    {"id": 1, "run_id": 42, "source": "gopls-references", "path": "...", "created_at": "..."}
  ]
}
```

Pseudocode:
```go
func runRawOutputs(w http.ResponseWriter, r *http.Request) {
  rows := queryRawOutputs(db, runID)
  writeJSON(w, map[string]any{"items": rows})
}
```

SQL sketch:
```sql
SELECT id, run_id, source, path, created_at
FROM raw_outputs
WHERE run_id = ?
ORDER BY created_at DESC;
```

#### GET /api/raw-outputs
Purpose: list raw outputs across runs (for data admin view).

Request query: `workspace_id`, optional `source`, `run_id`, `limit`, `offset`.

Pseudocode: same as run raw outputs with extra filters.

#### GET /api/raw-outputs/:id/download
Purpose: stream the raw output file from disk.

Pseudocode:
```go
func downloadRawOutput(w http.ResponseWriter, r *http.Request) {
  row := queryRawOutputByID(db, id)
  http.ServeFile(w, r, row.Path)
}
```

### Sessions (UI Concept)

#### GET /api/sessions
Purpose: list computed sessions for a workspace.

Request query: `workspace_id`.

Response schema:
```json
{
  "items": [Session]
}
```

Pseudocode:
```go
func listSessions(w http.ResponseWriter, r *http.Request) {
  ws := resolveWorkspace(r)
  runs := queryRuns(db, RunFilter{RootPath: ws.RepoRoot})
  groups := groupRunsByRootAndRange(runs)
  sessions := mergeManualSessionOverrides(ws, groups)
  writeJSON(w, map[string]any{"items": sessions})
}
```

Grouping logic:
- Prefer `(root_path, git_from, git_to)` as the session key.
- If git_from/git_to is missing, fall back to the closest run window by `started_at` within the same root_path.
- Allow manual overrides stored in the workspace config file.

#### GET /api/sessions/:id
Purpose: return one session, with run ids per domain.

Pseudocode:
```go
func getSession(w http.ResponseWriter, r *http.Request) {
  session := resolveSessionByID(ws, id)
  writeJSON(w, session)
}
```

#### POST /api/sessions
Purpose: allow manual session definitions or corrections.

Request schema:
```json
{
  "id": "custom-session",
  "root_path": "/abs/path/repo",
  "git_from": "HEAD~20",
  "git_to": "HEAD",
  "runs": {"commits": 12, "diff": 15}
}
```

Pseudocode:
```go
func createSession(w http.ResponseWriter, r *http.Request) {
  session := decodeJSON(r)
  ws := resolveWorkspace(r)
  ws.Sessions.Upsert(session)
  saveConfig()
  writeJSON(w, session)
}
```

### Unified Search

#### POST /api/search
Purpose: perform a multi-domain search with a single request.

Request schema:
```json
{
  "workspace_id": "glazed",
  "session_id": "glazed:HEAD~20..HEAD",
  "query": "Client",
  "types": ["symbols", "code_units", "diffs", "commits", "docs", "files"],
  "filters": {
    "path_glob": "internal/**/*.go",
    "pkg": "github.com/acme/project/internal/api",
    "symbol_kind": "type",
    "commit_hash": "",
    "run_ids": {"symbols": 18, "diffs": 15}
  },
  "limit": 50,
  "offset": 0
}
```

Response schema:
```json
{
  "items": [SearchResult],
  "by_type": {
    "symbols": {"count": 12, "items": [SearchResult]},
    "diffs": {"count": 4, "items": [SearchResult]}
  }
}
```

Pseudocode:
```go
func unifiedSearch(w http.ResponseWriter, r *http.Request) {
  req := decodeJSON(r)
  ws := resolveWorkspace(req.WorkspaceID)
  db := openDB(ws.DBPath)
  session := resolveSession(req)
  results := []SearchResult{}
  if includes(req.Types, "symbols") { results = append(results, searchSymbols(db, req, session)...) }
  if includes(req.Types, "code_units") { results = append(results, searchCodeUnits(db, req, session)...) }
  if includes(req.Types, "diffs") { results = append(results, searchDiffLines(db, req, session)...) }
  if includes(req.Types, "commits") { results = append(results, searchCommits(db, req, session)...) }
  if includes(req.Types, "docs") { results = append(results, searchDocHits(db, req, session)...) }
  if includes(req.Types, "files") { results = append(results, searchFiles(db, req, session)...) }
  writeJSON(w, map[string]any{"items": results})
}
```

#### GET /api/search/symbols
Purpose: search symbols only (FTS + filters).

Request query: `workspace_id`, `q`, `run_id`, `pkg`, `name`, `kind`, `path`, `limit`, `offset`.

SQL sketch:
```sql
SELECT d.symbol_hash, d.name, d.kind, d.pkg, d.recv, d.signature,
       f.path, o.line, o.col, o.is_exported
FROM symbol_defs_fts fts
JOIN symbol_defs d ON d.id = fts.rowid
JOIN symbol_occurrences o ON o.symbol_def_id = d.id
JOIN files f ON f.id = o.file_id
WHERE symbol_defs_fts MATCH ?
  AND (? = 0 OR o.run_id = ?)
ORDER BY d.pkg, d.name
LIMIT ? OFFSET ?;
```

Pseudocode: call `searchSymbols` with FTS query and filters.

#### GET /api/search/code-units
Purpose: search code unit snapshots by body/doc FTS.

Request query: `workspace_id`, `q`, `run_id`, `pkg`, `name`, `kind`, `path`, `limit`, `offset`.

SQL sketch:
```sql
SELECT cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.end_line, s.body_text
FROM code_unit_snapshots_fts fts
JOIN code_unit_snapshots s ON s.id = fts.rowid
JOIN code_units cu ON cu.id = s.code_unit_id
JOIN files f ON f.id = s.file_id
WHERE code_unit_snapshots_fts MATCH ?
  AND (? = 0 OR s.run_id = ?)
ORDER BY cu.pkg, cu.name
LIMIT ? OFFSET ?;
```

#### GET /api/search/diff
Purpose: search diff lines by FTS match.

Request query: `workspace_id`, `q`, `run_id`, `path`, `limit`, `offset`.

SQL sketch:
```sql
SELECT df.run_id, f.path, dl.text, dl.line_no_old, dl.line_no_new, dl.kind
FROM diff_lines_fts fts
JOIN diff_lines dl ON dl.id = fts.rowid
JOIN diff_hunks dh ON dh.id = dl.hunk_id
JOIN diff_files df ON df.id = dh.diff_file_id
JOIN files f ON f.id = df.file_id
WHERE diff_lines_fts MATCH ?
  AND (? = 0 OR df.run_id = ?)
ORDER BY f.path, dl.id
LIMIT ? OFFSET ?;
```

#### GET /api/search/commits
Purpose: search commits by subject/body FTS.

Request query: `workspace_id`, `q`, `run_id`, `author`, `limit`, `offset`.

#### GET /api/search/docs
Purpose: search doc hits by match_text FTS, optionally filter by term.

Request query: `workspace_id`, `q`, `term`, `run_id`, `path`, `limit`, `offset`.

#### GET /api/search/files
Purpose: search file paths by FTS.

Request query: `workspace_id`, `q`, `limit`, `offset`.

### Files and File Viewer

#### GET /api/files
Purpose: list files or directory prefixes for the tree.

Request query: `workspace_id`, `prefix`, `ext`, `exists`, `is_binary`, `limit`, `offset`.

Response schema:
```json
{
  "items": [
    {"path": "internal/api", "kind": "dir"},
    {"path": "internal/api/client.go", "kind": "file", "ext": "go"}
  ]
}
```

Pseudocode:
```go
func listFiles(w http.ResponseWriter, r *http.Request) {
  filters := parseFileFilters(r)
  files := queryFiles(db, filters)
  items := expandToTreeLevel(files, filters.Prefix)
  writeJSON(w, map[string]any{"items": items})
}
```

#### GET /api/file
Purpose: return file content and metadata, optionally at a git ref.

Request query: `workspace_id`, `path` (required), `ref` (optional), `with_lines` (bool).

Response schema:
```json
{
  "path": "internal/api/client.go",
  "ref": "HEAD",
  "content": "...",
  "lines": ["line1", "line2"],
  "line_count": 120
}
```

Pseudocode:
```go
func getFile(w http.ResponseWriter, r *http.Request) {
  ws := resolveWorkspace(r)
  path := r.URL.Query().Get("path")
  ref := r.URL.Query().Get("ref")
  content := readFileFromRepo(ws.RepoRoot, ref, path)
  writeJSON(w, map[string]any{"path": path, "ref": refOrHEAD, "content": content, "lines": splitLines(content)})
}
```

Note: if repo_root is missing, return 409 with a message indicating file content is unavailable.

#### GET /api/files/history
Purpose: list commits that touched a file (History tab).

Request query: `workspace_id`, `path`, `run_id`, `limit`, `offset`.

SQL sketch:
```sql
SELECT c.hash, c.subject, c.committer_date, cf.status, cf.old_path, cf.new_path
FROM commit_files cf
JOIN commits c ON c.id = cf.commit_id
JOIN files f ON f.id = cf.file_id
WHERE f.path = ? AND (? = 0 OR c.run_id = ?)
ORDER BY c.id DESC
LIMIT ? OFFSET ?;
```

### Symbols

#### GET /api/symbols
Purpose: list symbol definitions with filters.

Request query: `workspace_id`, `run_id`, `exported_only`, `kind`, `name`, `pkg`, `path`, `limit`, `offset`.

Response schema:
```json
{
  "items": [
    {
      "symbol_hash": "...",
      "name": "Client",
      "kind": "type",
      "pkg": "github.com/acme/project/internal/api",
      "recv": "",
      "signature": "",
      "file": "internal/api/client.go",
      "line": 42,
      "col": 6,
      "is_exported": true
    }
  ]
}
```

Pseudocode: call existing `ListSymbolInventory` with filters.

#### GET /api/symbols/:hash
Purpose: return a symbol definition and its primary location.

SQL sketch:
```sql
SELECT d.symbol_hash, d.name, d.kind, d.pkg, d.recv, d.signature,
       f.path, o.line, o.col, o.is_exported
FROM symbol_defs d
JOIN symbol_occurrences o ON o.symbol_def_id = d.id
JOIN files f ON f.id = o.file_id
WHERE d.symbol_hash = ?
ORDER BY o.line ASC
LIMIT 1;
```

Pseudocode:
```go
func getSymbol(w http.ResponseWriter, r *http.Request) {
  hash := param(r, "hash")
  record := querySymbolByHash(db, hash)
  writeJSON(w, record)
}
```

#### GET /api/symbols/:hash/refs
Purpose: return gopls references for a symbol.

Request query: `workspace_id`, `run_id`, `limit`, `offset`.

SQL sketch:
```sql
SELECT r.run_id, c.hash, d.symbol_hash, f.path, r.line, r.col, r.is_decl, r.source
FROM symbol_refs r
JOIN symbol_defs d ON d.id = r.symbol_def_id
JOIN files f ON f.id = r.file_id
LEFT JOIN commits c ON c.id = r.commit_id
WHERE d.symbol_hash = ?
  AND (? = 0 OR r.run_id = ?)
ORDER BY f.path, r.line, r.col
LIMIT ? OFFSET ?;
```

If refs are missing, return an empty list plus a hint: `{ "refs_available": false }`.

### Code Units

#### GET /api/code-units
Purpose: list code units with filters.

Request query: `workspace_id`, `run_id`, `name`, `pkg`, `kind`, `path`, `body_q`, `limit`, `offset`.

SQL sketch (non-FTS path):
```sql
SELECT cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.end_line, s.body_hash
FROM code_unit_snapshots s
JOIN code_units cu ON cu.id = s.code_unit_id
JOIN files f ON f.id = s.file_id
WHERE (? = 0 OR s.run_id = ?)
ORDER BY cu.pkg, cu.name
LIMIT ? OFFSET ?;
```

If `body_q` is provided, use `code_unit_snapshots_fts` instead.

#### GET /api/code-units/:hash
Purpose: return a code unit snapshot for the selected run (or latest).

Request query: `workspace_id`, `run_id` optional.

SQL sketch:
```sql
SELECT cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.end_line, s.body_text, s.doc_text, s.body_hash
FROM code_units cu
JOIN code_unit_snapshots s ON s.code_unit_id = cu.id
JOIN files f ON f.id = s.file_id
WHERE cu.unit_hash = ?
  AND (? = 0 OR s.run_id = ?)
ORDER BY s.run_id DESC
LIMIT 1;
```

#### GET /api/code-units/:hash/history
Purpose: list snapshots for a code unit.

Request query: `workspace_id`, `limit`, `offset`.

Pseudocode: same as detail, without limit 1, ordered by `s.run_id DESC`.

#### POST /api/code-units/:hash/diff
Purpose: diff two snapshots by `snapshot_id` or `run_id`.

Request schema:
```json
{"left_run_id": 18, "right_run_id": 19}
```

Pseudocode: load both bodies and return a unified diff (simple line diff).

### Commits

#### GET /api/commits
Purpose: list commits with filters.

Request query: `workspace_id`, `run_id`, `author`, `after`, `before`, `path`, `q`, `limit`, `offset`.

If `q` is provided, use `commits_fts` and join `commits`.

#### GET /api/commits/:hash
Purpose: return commit metadata.

SQL sketch:
```sql
SELECT id, run_id, hash, author_name, author_email, author_date, committer_date, subject, body
FROM commits
WHERE hash = ?
ORDER BY id DESC
LIMIT 1;
```

#### GET /api/commits/:hash/files
Purpose: list files changed in a commit.

SQL sketch:
```sql
SELECT f.path, cf.status, cf.old_path, cf.new_path, cf.blob_old, cf.blob_new
FROM commit_files cf
JOIN commits c ON c.id = cf.commit_id
JOIN files f ON f.id = cf.file_id
WHERE c.hash = ?
ORDER BY f.path;
```

#### GET /api/commits/:hash/diff
Purpose: render a diff for a commit.

Implementation note: if a per-commit diff run exists, use `diff_lines`. Otherwise, run `git show` against repo_root.

### Diffs

#### GET /api/diff-runs
Purpose: list diff runs for a session or workspace.

Request query: `workspace_id`, `session_id`.

Pseudocode: resolve session, return its diff run id if available.

#### GET /api/diff/:run_id/files
Purpose: list diff files within a diff run.

Request query: `workspace_id`, `limit`, `offset`.

Pseudocode: call existing `ListDiffFiles` with run id.

#### GET /api/diff/:run_id/file
Purpose: return hunks and lines for one diff file.

Request query: `workspace_id`, `path`.

SQL sketch:
```sql
SELECT dh.id, dh.old_start, dh.old_lines, dh.new_start, dh.new_lines
FROM diff_hunks dh
JOIN diff_files df ON df.id = dh.diff_file_id
JOIN files f ON f.id = df.file_id
WHERE df.run_id = ? AND f.path = ?
ORDER BY dh.id;
```

Then load diff lines by `hunk_id`:
```sql
SELECT kind, line_no_old, line_no_new, text
FROM diff_lines
WHERE hunk_id = ?
ORDER BY id;
```

### Docs / Terms

#### GET /api/docs/terms
Purpose: list doc-hit terms with counts.

Request query: `workspace_id`, `run_id`, `path_prefix`, `limit`, `offset`.

SQL sketch:
```sql
SELECT term, count(*) AS hits
FROM doc_hits h
JOIN files f ON f.id = h.file_id
WHERE (? = 0 OR h.run_id = ?)
  AND (? = '' OR f.path LIKE ?)
GROUP BY term
ORDER BY hits DESC
LIMIT ? OFFSET ?;
```

#### GET /api/docs/hits
Purpose: list doc hits for a term or file.

Request query: `workspace_id`, `run_id`, `term`, `path`, `limit`, `offset`.

Pseudocode: call existing `ListDocHits` with term filter and optional path filter.

### Tree-sitter (Optional)

#### GET /api/tree-sitter/captures
Purpose: list tree-sitter captures.

Request query: `workspace_id`, `run_id`, `query_name`, `capture_name`, `node_type`, `path`, `limit`, `offset`.

SQL sketch:
```sql
SELECT t.query_name, t.capture_name, t.node_type, t.start_line, t.end_line, t.snippet, f.path
FROM ts_captures t
JOIN files f ON f.id = t.file_id
WHERE (? = 0 OR t.run_id = ?)
ORDER BY f.path, t.start_line
LIMIT ? OFFSET ?;
```

### Refactor Assistance (Optional Phase)

These routes require new storage (JSON files or new tables). They are included for completeness but should be staged after the browsing MVP.

#### POST /api/refactor/plans
Request schema:
```json
{
  "name": "Rename Client",
  "session_id": "glazed:HEAD~20..HEAD",
  "targets": [{"symbol_hash": "...", "old_name": "Client", "new_name": "APIClient"}],
  "doc_terms": [{"from": "Client", "to": "APIClient", "include": ["docs/**/*.md"]}]
}
```

Pseudocode: validate targets exist, persist plan JSON to a local directory, return plan id.

#### POST /api/refactor/plans/:id/validate
Purpose: run gopls `prepare_rename` and optionally ingest refs.

Pseudocode: use existing gopls ingestion utilities and store results as plan metadata.

## Data Source Mapping (UI View → Tables)

| UI View | Tables / sources |
| --- | --- |
| Workspace info | `schema_versions`, `sqlite_master` |
| Runs list | `meta_runs`, `run_kv`, `raw_outputs` |
| Sessions | computed from `meta_runs` |
| Symbols | `symbol_defs`, `symbol_occurrences`, `files` |
| Symbol refs | `symbol_refs`, `symbol_refs_unresolved`, `files` |
| Code units | `code_units`, `code_unit_snapshots`, `files` |
| Commits | `commits`, `commit_files`, `files`, `file_blobs` |
| Diffs | `diff_files`, `diff_hunks`, `diff_lines`, `files` |
| Docs | `doc_hits`, `files` |
| Tree-sitter | `ts_captures`, `files` |
| Search | `*_fts` tables + joins |

## Key Gaps and Mitigations

**File content**: The DB stores file paths and code unit bodies, not full file contents. The file viewer must read from repo_root or use `git show` for a ref. If repo_root is missing, return a clear error and disable file previews.

**Session grouping**: Runs are stored separately per domain. The UI depends on a stable session, so the backend must group runs or accept manual mappings. The grouping logic should be deterministic and visible in session responses.

**Missing gopls refs**: If `symbol_refs` is empty, the API should return `refs_available=false` and allow the UI to surface a “compute refs” action.

**FTS availability**: The API should surface which FTS tables exist. If FTS tables are missing (older schema), the search endpoints should fall back to non-FTS queries with explicit warnings about performance and completeness.

## Implementation Notes

- Prefer read-only queries in the first milestone. Any write endpoints should write to a separate plan storage folder rather than mutating the index DB.
- Use prepared statements or a minimal query builder to avoid ad-hoc string concatenation for FTS queries.
- Return stable field names matching the JS index API where possible to reduce frontend adaptation.
- Keep search results deterministic by sorting in SQL or in Go after fetching.

## Appendix: Endpoint Summary Table

| Route | Purpose |
| --- | --- |
| GET /api/workspaces | List workspaces |
| POST /api/workspaces | Create workspace |
| GET /api/workspaces/:id | Workspace detail |
| PATCH /api/workspaces/:id | Update workspace |
| DELETE /api/workspaces/:id | Delete workspace |
| GET /api/db/info | DB info + feature flags |
| GET /api/runs | Runs list |
| GET /api/runs/:id | Run detail |
| GET /api/runs/:id/summary | Run row counts |
| GET /api/runs/:id/raw-outputs | Raw outputs for run |
| GET /api/raw-outputs | Raw outputs list |
| GET /api/raw-outputs/:id/download | Download raw output |
| GET /api/sessions | Session list |
| GET /api/sessions/:id | Session detail |
| POST /api/sessions | Create or override session |
| POST /api/search | Unified search |
| GET /api/search/symbols | Symbol search |
| GET /api/search/code-units | Code unit search |
| GET /api/search/diff | Diff line search |
| GET /api/search/commits | Commit search |
| GET /api/search/docs | Doc hit search |
| GET /api/search/files | File path search |
| GET /api/files | File list/tree |
| GET /api/file | File content |
| GET /api/files/history | File commit history |
| GET /api/symbols | Symbol list |
| GET /api/symbols/:hash | Symbol detail |
| GET /api/symbols/:hash/refs | Symbol refs |
| GET /api/code-units | Code unit list |
| GET /api/code-units/:hash | Code unit detail |
| GET /api/code-units/:hash/history | Code unit history |
| POST /api/code-units/:hash/diff | Code unit diff |
| GET /api/commits | Commit list |
| GET /api/commits/:hash | Commit detail |
| GET /api/commits/:hash/files | Commit files |
| GET /api/commits/:hash/diff | Commit diff |
| GET /api/diff-runs | Diff run list |
| GET /api/diff/:run_id/files | Diff files list |
| GET /api/diff/:run_id/file | Diff file detail |
| GET /api/docs/terms | Doc terms list |
| GET /api/docs/hits | Doc hits list |
| GET /api/tree-sitter/captures | Tree-sitter captures |
| POST /api/refactor/plans | Create refactor plan |
| POST /api/refactor/plans/:id/validate | Validate plan |
