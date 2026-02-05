---
Title: Workbench REST API Reference
Slug: workbench-api-reference
Short: End-to-end reference for the refactorio workbench browsing API.
Topics:
    - refactorio
    - api
    - rest
    - reference
Commands:
    - refactorio
    - api
    - serve
Flags:
    - --addr
    - --base-path
    - --workspace-config
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why This Reference
Use this page when you need the precise REST surface for the workbench API: routes, filters, and response shapes. It is intended to be a lookup document for UI integration and tooling.

## Base URL and Auth
The server is launched with `refactorio api serve`. By default, the API is mounted at `/api`, but you can change it with `--base-path`.

There is no built-in authentication; assume the service is local or protected by your own gateway.

## Workspace Selection (Required)
All endpoints that read indexed data require one of these query parameters:

- `workspace_id`: Resolve DB and repo paths from the workspace config file.
- `db_path`: Absolute or relative path to the SQLite DB (resolved to absolute on the server).

If neither is provided, the API returns `400 invalid_argument`.

The workspace config path is set via `--workspace-config`. If omitted, it defaults to the OS user config directory (`$XDG_CONFIG_HOME/refactorio/workspaces.json` on Linux).

## Error Format
Errors are returned as:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "...",
    "details": {"field": "..."}
  }
}
```

## Pagination
Most list endpoints accept `limit` and `offset`. The server clamps `limit` to an endpoint-specific max. When omitted, defaults are applied (listed per endpoint below).

## Health and Database

### `GET /health`
Returns a simple OK payload.

**Response**
```json
{"status":"ok"}
```

### `GET /db/info`
Returns schema details and feature availability for the selected DB.

**Query**
- `workspace_id` or `db_path` (required)

**Response**
```json
{
  "workspace_id": "smoke",
  "db_path": "/abs/path/index.sqlite",
  "repo_root": "/abs/path/repo",
  "schema_version": 17,
  "tables": {"meta_runs": true},
  "fts_tables": {"symbol_defs_fts": true},
  "features": {"fts": true, "gopls_refs": false, "doc_hits": true, "tree_sitter": false},
  "views": {"v_last_commit_per_file": true}
}
```

## Workspaces
Workspaces are stored in the workspace config file and used by `workspace_id` resolution.

### `GET /workspaces`
Lists stored workspaces.

**Response**
```json
{"items": [{"id":"main","name":"main","db_path":"/abs/path/index.sqlite","repo_root":"/abs/path/repo"}]}
```

### `POST /workspaces`
Creates a new workspace entry.

**Body**
```json
{
  "id": "main",
  "name": "main",
  "db_path": "/abs/path/index.sqlite",
  "repo_root": "/abs/path/repo"
}
```

**Response**: workspace JSON (with `created_at`, `updated_at`).

### `GET /workspaces/{id}`
Fetches a workspace entry.

### `PATCH /workspaces/{id}`
Updates `name`, `db_path`, or `repo_root`.

**Body**
```json
{"name":"New Name","db_path":"/abs/path/index.sqlite","repo_root":"/abs/path/repo"}
```

### `DELETE /workspaces/{id}`
Deletes the workspace entry.

**Response**
```json
{"deleted":"main"}
```

## Runs and Raw Outputs

### `GET /runs`
Lists runs in `meta_runs`.

**Query**
- `status`, `root_path`, `git_from`, `git_to`
- `started_after`, `started_before` (RFC3339 timestamps)
- `limit` (default 100, max 1000), `offset`

**Response**
```json
{"items":[{"id":1,"started_at":"...","status":"success","tool_version":"...","git_from":"...","git_to":"...","root_path":"...","args_json":"{}","sources_dir":"..."}],"limit":100,"offset":0}
```

### `GET /runs/{id}`
Returns a single run by ID.

### `GET /runs/{id}/summary`
Returns counts for known tables by run.

**Response**
```json
{"run_id":1,"counts":{"diff_files":12,"code_unit_snapshots":200}}
```

### `GET /runs/{id}/raw-outputs`
Lists raw outputs for a run (if `raw_outputs` table exists).

### `GET /raw-outputs`
Lists raw outputs across runs.

**Query**
- `source`, `run_id`
- `limit` (default 100, max 1000), `offset`

## Sessions
Sessions group runs by `{root_path, git_from, git_to}` and surface the latest run ID for each domain.

### `GET /sessions`
Lists computed sessions (and applies workspace overrides if present).

**Response**
```json
{"items":[{"id":"db:HEAD..HEAD:abcd1234","workspace_id":"main","root_path":"/repo","git_from":"HEAD~1","git_to":"HEAD","runs":{"diff":12,"symbols":11},"availability":{"diff":true,"symbols":true},"last_updated":"2026-02-05T00:00:00Z"}]}
```

### `GET /sessions/{id}`
Returns a single session by ID.

### `POST /sessions`
Creates or replaces a session override (requires `workspace_id`).

**Body**
```json
{
  "id": "custom-session",
  "root_path": "/repo",
  "git_from": "HEAD~1",
  "git_to": "HEAD",
  "runs": {"diff": 12, "symbols": 11, "code_units": 10}
}
```

## Search
All `/search/*` endpoints require FTS tables. If missing, the API responds with `400 search_error` and a message like `fts table not available: symbol_defs_fts`.

### `POST /search`
Unified search across multiple types.

**Body**
```json
{
  "query": "Client",
  "types": ["symbols", "code_units", "diffs", "commits", "docs", "files"],
  "filters": {"path": "internal/api/client.go", "pkg": "github.com/acme/project/internal/api", "symbol_kind": "type", "term": "Client"},
  "limit": 50,
  "offset": 0,
  "run_ids": {"symbols": 12, "code_units": 12, "diffs": 12, "commits": 12, "docs": 12}
}
```

**Response**
```json
{"items":[{"type":"symbol","primary":"Client","secondary":"github.com/acme/project/internal/api","path":"internal/api/client.go","line":10,"col":5,"run_id":12,"payload":{...}}]}
```

### `GET /search/symbols`
**Query**
- `q` (required)
- `path`, `pkg`, `kind`, `run_id`
- `limit` (default 100, max 1000), `offset`

**Response item**: `SymbolSearchRecord`

### `GET /search/code-units`
**Query**
- `q` (required)
- `path`, `pkg`, `kind`, `run_id`
- `limit` (default 100, max 1000), `offset`

**Response item**: `CodeUnitSearchRecord`

### `GET /search/diff`
**Query**
- `q` (required)
- `path`, `run_id`
- `limit` (default 100, max 1000), `offset`

**Response item**: `DiffSearchRecord`

### `GET /search/commits`
**Query**
- `q` (required)
- `run_id`
- `limit` (default 100, max 1000), `offset`

**Response item**: `CommitSearchRecord`

### `GET /search/docs`
**Query**
- `q` (required)
- `term`, `path`, `run_id`
- `limit` (default 100, max 1000), `offset`

**Response item**: `DocSearchRecord`

### `GET /search/files`
**Query**
- `q` (required)
- `limit` (default 100, max 1000), `offset`

**Response item**: `FileSearchRecord`

## Symbols

### `GET /symbols`
Lists symbol occurrences.

**Query**
- `run_id`
- `exported_only` (bool)
- `kind`, `name`, `pkg`, `path`
- `limit` (default 100, max 1000), `offset`

**Response item**
```json
{"symbol_hash":"...","name":"Client","kind":"type","pkg":"github.com/acme/project/internal/api","file":"internal/api/client.go","line":10,"col":5,"is_exported":true,"run_id":12}
```

### `GET /symbols/{hash}`
Returns a representative occurrence for a symbol hash.

**Query**
- `run_id` (optional)

### `GET /symbols/{hash}/refs`
Returns gopls references if `symbol_refs` table exists.

**Query**
- `run_id`
- `limit` (default 200, max 2000), `offset`

**Response**
```json
{"items":[{"run_id":12,"symbol_hash":"...","path":"internal/api/client.go","line":55,"col":10,"is_decl":false,"source":"gopls"}],"refs_available":true,"limit":200,"offset":0}
```

## Code Units

### `GET /code-units`
Lists code unit snapshots.

**Query**
- `run_id`
- `path`, `pkg`, `kind`, `name`
- `body_q` (FTS search over body text)
- `limit` (default 100, max 1000), `offset`

### `GET /code-units/{hash}`
Returns the most recent snapshot for a code unit.

**Query**
- `run_id` (optional)

### `GET /code-units/{hash}/history`
Returns snapshot history.

**Query**
- `limit` (default 50, max 500), `offset`

### `POST /code-units/{hash}/diff`
Computes a simple line diff between two runs.

**Body**
```json
{"left_run_id": 10, "right_run_id": 12}
```

**Response**
```json
{"diff":["-old line","+new line"]}
```

## Diffs

### `GET /diff-runs`
Lists runs that have diff files.

**Query**
- `session_id` (optional; returns the session's diff run if present)
- `limit` (default 100, max 1000), `offset`

### `GET /diff/{run_id}/files`
Lists diff files for a run.

**Query**
- `limit` (default 100, max 1000), `offset`

### `GET /diff/{run_id}/file`
Returns hunks and lines for a single diff file.

**Query**
- `path` (required)

## Commits

### `GET /commits`
Lists commits, optionally filtered by FTS or metadata.

**Query**
- `q` (full-text search; requires `commits_fts`)
- `run_id`, `path`, `author`, `after`, `before`
- `limit` (default 100, max 1000), `offset`

### `GET /commits/{hash}`
Returns a single commit (most recent by hash).

### `GET /commits/{hash}/files`
Lists files touched by a commit.

### `GET /commits/{hash}/diff`
Returns diff data by locating the diff run whose `git_to` matches the commit hash.

**Query**
- `path` (optional, to return hunks for a single file)

## Docs

### `GET /docs/terms`
Lists top doc terms by count.

**Query**
- `run_id`
- `path_prefix`
- `limit` (default 100, max 1000), `offset`

### `GET /docs/hits`
Lists doc hits.

**Query**
- `run_id`, `term`, `path`
- `limit` (default 200, max 2000), `offset`

## Files

### `GET /files`
Lists files or immediate subdirectories under an optional prefix.

**Query**
- `prefix` (directory prefix)
- `ext`
- `exists` (bool)
- `is_binary` (bool)
- `limit` (default 1000, max 5000), `offset`

**Response item**
```json
{"path":"internal/api","kind":"dir"}
```

### `GET /file`
Returns file contents from `repo_root` (workspace config required).

**Query**
- `path` (required)
- `ref` (git ref; when omitted reads from disk)
- `with_lines` (bool; include `lines` array)

**Response**
```json
{"path":"internal/api/client.go","ref":"HEAD","content":"...","lines":["..."],"line_count":120}
```

### `GET /files/history`
Lists commits touching a file.

**Query**
- `path` (required)
- `run_id`
- `limit` (default 100, max 1000), `offset`

## Tree-sitter

### `GET /tree-sitter/captures`
Lists tree-sitter captures (if `ts_captures` exists).

**Query**
- `run_id`
- `query_name`, `capture_name`, `node_type`, `path`
- `limit` (default 200, max 2000), `offset`

## See Also
- `pkg/workbenchapi/` for the Go handlers.
- `pkg/doc/topics/03-js-index-api-reference.md` for the JS index API.
