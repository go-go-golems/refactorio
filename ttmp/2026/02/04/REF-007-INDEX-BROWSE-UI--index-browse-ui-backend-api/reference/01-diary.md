---
Title: Diary
Ticket: REF-007-INDEX-BROWSE-UI
Status: active
Topics:
    - ui
    - api
    - refactorio
    - backend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/cmd/refactorio/api.go
      Note: New API serve command
    - Path: refactorio/cmd/refactorio/root.go
      Note: Wired API command into CLI
    - Path: refactorio/pkg/refactorindex/ingest_code_units.go
      Note: Code unit snapshot details reviewed
    - Path: refactorio/pkg/refactorindex/ingest_commits.go
      Note: Commit ingestion details reviewed
    - Path: refactorio/pkg/refactorindex/ingest_diff.go
      Note: Diff ingestion details reviewed
    - Path: refactorio/pkg/refactorindex/ingest_symbols.go
      Note: Symbol ingestion details reviewed
    - Path: refactorio/pkg/refactorindex/query.go
      Note: Query helpers referenced for API mapping
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: Schema reviewed while mapping UI requirements
    - Path: refactorio/pkg/refactorindex/store.go
      Note: DB initialization and FTS setup reviewed
    - Path: refactorio/pkg/workbenchapi/db.go
      Note: Workspace-aware DB open helper
    - Path: refactorio/pkg/workbenchapi/db_info.go
      Note: DB info endpoint and schema/FTS detection
    - Path: refactorio/pkg/workbenchapi/decode.go
      Note: Strict JSON decoding helper
    - Path: refactorio/pkg/workbenchapi/json.go
      Note: Shared JSON and error response helpers
    - Path: refactorio/pkg/workbenchapi/routes.go
      Note: Base route registration
    - Path: refactorio/pkg/workbenchapi/runs.go
      Note: Run list/detail/summary and raw outputs endpoints
    - Path: refactorio/pkg/workbenchapi/search.go
      Note: FTS-backed search endpoints and unified search
    - Path: refactorio/pkg/workbenchapi/server.go
      Note: Server config
    - Path: refactorio/pkg/workbenchapi/workspace.go
      Note: Workspace config model and CRUD handlers
ExternalSources: []
Summary: Diary for backend API documentation and analysis.
LastUpdated: 2026-02-05T09:30:00-05:00
WhatFor: Track analysis steps and documentation work for the Workbench REST API.
WhenToUse: Use when reviewing how the API design docs were produced.
---







# Diary

## Goal
Capture the analysis and documentation steps used to define the Workbench REST API for the UI spec.

## Step 1: Analyze UI Spec and Map to Backend API
I imported the UI spec into the new ticket workspace, then reviewed the refactorindex schema, ingestion behavior, and existing JS query API to understand which tables and data shapes the UI can rely on. This established the baseline for API contracts and highlighted where new query helpers are required.

I then drafted two documents: a detailed route-by-route API analysis with schemas and pseudocode, and an implementation plan describing the server layout, phases, and risk mitigation. This keeps the UX-driven requirements anchored to concrete SQL and handler logic.

### Prompt Context
**User prompt (verbatim):** "Create a new docmgr ticket REF-007-INDEX-BROWSE-UI and use docmgr import --file ... --ticket ... to import refactorio/sources/ui-design.md into the ticket, and write a detailed implementation document on building the backend API REST API for the UI presented there.

Keep a detailed diary as you analyze the document and look around in the source to find relevant functions and APIs to use to implement the REST API, and create a detailed analysis document which goes over the overall architecture and design, and then digs into every route and its schema and pseudocode implementation for its backend, and a prose paragraphs explanation and all that.

Once the document is written, upload it to remarkable."

**Assistant interpretation:** Create a new ticket, import the UI spec, analyze the refactorindex code and existing APIs, produce a deep analysis doc and a practical implementation plan for the REST API, keep a diary of the work, and upload the final docs to reMarkable.

**Inferred user intent:** Get a UI-driven backend API spec that is grounded in the current refactorindex data model and ready for implementation.

**Commit (code):** N/A (docs-only)

### What I did
- Created the docmgr ticket `REF-007-INDEX-BROWSE-UI`.
- Imported `refactorio/sources/ui-design.md` into the ticket sources.
- Reviewed refactorindex schema and ingestion logic to map tables to UI needs.
- Reviewed existing query helpers and JS index API for response shapes.
- Wrote the route-by-route analysis document with schemas and pseudocode.
- Wrote the implementation plan for server structure and phases.

### Why
- The UI spec requires stable, scoped sessions and cross-domain search that the current DB supports but does not expose as REST endpoints.
- A clear mapping from UI views to SQL and API contracts reduces implementation risk and avoids N+1 query patterns.

### What worked
- The refactorindex schema already covers most entities required for the UI.
- The JS index module provides a consistent response shape for symbols, refs, doc hits, and files that can be mirrored in REST.

### What didn't work
- N/A

### What I learned
- File content is not stored in the DB, so the API must read from repo_root or `git show` for historical refs.
- Session grouping must be derived from run metadata and needs manual override support to avoid ambiguity.

### What was tricky to build
The hardest part to reason about was session grouping and file content retrieval. Runs are independent per domain, so the session resolver needs deterministic grouping by `(root_path, git_from, git_to)` plus heuristics for missing git ranges. File content is not stored, so the API must either read from a repo checkout or fall back to `git show`, which implies the repo_root is mandatory for file previews.

### What warrants a second pair of eyes
- Session grouping heuristics and how they handle missing `git_from`/`git_to` in symbol or doc runs.
- Search query construction for FTS tables to ensure correct joins and performance.

### What should be done in the future
- Define a persistent storage format for refactor plans and run artifacts if the refactor-assist phase is implemented.

### Code review instructions
- Start with `refactorio/ttmp/2026/02/04/REF-007-INDEX-BROWSE-UI--index-browse-ui-backend-api/analysis/01-backend-rest-api-architecture-route-analysis.md` for the full route catalog.
- Then review `refactorio/ttmp/2026/02/04/REF-007-INDEX-BROWSE-UI--index-browse-ui-backend-api/design/01-backend-rest-api-implementation-plan.md` for the phased implementation plan.

### Technical details
Commands run:
```bash
docmgr ticket create-ticket --ticket REF-007-INDEX-BROWSE-UI --title "Index Browse UI Backend API" --topics ui,api,refactorio,backend
docmgr import file --ticket REF-007-INDEX-BROWSE-UI --file /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/sources/ui-design.md --name "ui-design.md"
docmgr doc add --ticket REF-007-INDEX-BROWSE-UI --doc-type analysis --title "Backend REST API Architecture & Route Analysis"
docmgr doc add --ticket REF-007-INDEX-BROWSE-UI --doc-type design --title "Backend REST API Implementation Plan"
docmgr doc add --ticket REF-007-INDEX-BROWSE-UI --doc-type reference --title "Diary"
```
Key sources reviewed:
- `refactorio/sources/ui-design.md`
- `refactorio/pkg/refactorindex/schema.go`
- `refactorio/pkg/refactorindex/query.go`
- `refactorio/pkg/refactorindex/ingest_commits.go`
- `refactorio/pkg/refactorindex/ingest_diff.go`
- `refactorio/pkg/refactor/js/modules/refactorindex/refactorindex.go`

## Step 2: Scaffold Workbench API Server
I added the initial Workbench API server package and a Cobra command to launch it. This puts a stable execution entry point in place so upcoming tasks can focus on endpoint behavior rather than wiring.

I kept the scaffolding intentionally small: a server config, base path normalization, a health route, and shared JSON/error helpers. This ensures later handlers share a consistent response pattern.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Create the task list, implement the first task (server scaffold), commit the code, and update the diary to reflect progress.

**Inferred user intent:** Build the backend incrementally with visible progress, tasks, and commit history.

**Commit (code):** b525dcb — "Add workbench API server scaffold"

### What I did
- Added `refactorio/pkg/workbenchapi` with `Server`, base routes, and JSON/error helpers.
- Added `refactorio/cmd/refactorio/api.go` to expose `refactorio api serve`.
- Wired the new command into `refactorio/cmd/refactorio/root.go`.
- Ran `gofmt` on new and modified files.

### Why
- A stable server entry point is required before implementing endpoint behavior.
- Shared JSON/error helpers prevent inconsistent response shapes later.

### What worked
- The server now exposes `GET /api/health` and normalizes base paths cleanly.
- The CLI command structure aligns with the rest of the repo.

### What didn't work
- N/A

### What I learned
- Keeping base path normalization isolated simplifies route registration later.

### What was tricky to build
Normalizing the base path while allowing a root-level mount is subtle; it needs to avoid double slashes and handle `/` as a special case so the mux works as expected.

### What warrants a second pair of eyes
- Confirm the base path + `StripPrefix` logic behaves correctly for `/api` and `/`.

### What should be done in the future
- Implement workspace registry + CRUD routes as the next task.

### Code review instructions
- Start with `refactorio/pkg/workbenchapi/server.go` and `refactorio/pkg/workbenchapi/routes.go`.
- Verify CLI wiring in `refactorio/cmd/refactorio/api.go`.
- Optional manual check: `go run ./cmd/refactorio api serve --addr :8090` and call `/api/health`.

### Technical details
Commands run:
```bash
gofmt -w refactorio/cmd/refactorio/api.go refactorio/cmd/refactorio/root.go refactorio/pkg/workbenchapi/server.go refactorio/pkg/workbenchapi/routes.go refactorio/pkg/workbenchapi/json.go
git -C refactorio add cmd/refactorio/api.go cmd/refactorio/root.go pkg/workbenchapi/server.go pkg/workbenchapi/routes.go pkg/workbenchapi/json.go
git -C refactorio commit -m "Add workbench API server scaffold"
```

## Step 3: Implement Workspace Registry + CRUD Endpoints
I implemented the workspace registry backed by a JSON config file and exposed CRUD routes under `/api/workspaces`. This lets the UI persist workspace metadata (db path + repo root) and aligns with the UI spec’s “Workspace selection” requirements.

I also added decoding helpers and a configurable config path, so the API can run with a default location but still be overridden for tests or alternative setups.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete the next task by building workspace persistence and CRUD endpoints, then commit and update the diary.

**Inferred user intent:** Incrementally deliver core API functionality with traceable commits and documentation.

**Commit (code):** e6b2cd6 — "Add workspace registry endpoints"

### What I did
- Added workspace config load/save utilities and JSON schema in `pkg/workbenchapi/workspace.go`.
- Added `/api/workspaces` and `/api/workspaces/:id` handlers with GET/POST/PATCH/DELETE.
- Introduced a JSON decode helper with strict decoding.
- Added `--workspace-config` flag and config path plumbing in the server.

### Why
- The Workbench UI needs to store and retrieve workspace definitions to connect to index DBs reliably.
- A strict JSON decoder prevents silent input errors and enforces a clean API contract.

### What worked
- CRUD operations persist cleanly to a stable config file under the user config dir.
- The base server can now be configured for different environments and tests.

### What didn't work
- N/A

### What I learned
- Workspace IDs are the natural stable key; normalizing paths early reduces later ambiguity.

### What was tricky to build
Updating the `ServeMux` routing to support both collection and item routes required a clean path parsing strategy, since the standard mux does not provide path parameters.

### What warrants a second pair of eyes
- Confirm the workspace path parsing and `StripPrefix` behavior in combination with `/api`.
- Confirm update semantics (especially clearing `repo_root`) are acceptable.

### What should be done in the future
- Add validation endpoints to check that `db_path` points to a valid index DB.

### Code review instructions
- Start with `refactorio/pkg/workbenchapi/workspace.go` and `refactorio/pkg/workbenchapi/routes.go`.
- Verify CLI changes in `refactorio/cmd/refactorio/api.go`.

### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/server.go refactorio/pkg/workbenchapi/routes.go refactorio/pkg/workbenchapi/workspace.go refactorio/pkg/workbenchapi/decode.go refactorio/cmd/refactorio/api.go
git -C refactorio add cmd/refactorio/api.go pkg/workbenchapi/server.go pkg/workbenchapi/routes.go pkg/workbenchapi/workspace.go pkg/workbenchapi/decode.go
git -C refactorio commit -m "Add workspace registry endpoints"
```

## Step 4: Add DB Open Helper + /api/db/info
I added a workspace-aware DB resolver and a `/api/db/info` endpoint that reports schema version, table availability, FTS presence, and feature flags. This matches the UI’s need to validate a workspace and understand what data is available.

The handler uses SQLite metadata and a simple schema version query, and it is careful to handle missing tables gracefully (older DBs). This should make the UI’s “DB info” panel robust across versions.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement the DB info endpoint with schema/FTS detection and commit the changes with diary updates.

**Inferred user intent:** Build out core data discovery endpoints before larger browse/search work.

**Commit (code):** f1fe1bd — "Add db info endpoint"

### What I did
- Added a workspace-aware DB open helper in `pkg/workbenchapi/db.go`.
- Implemented `/api/db/info` in `pkg/workbenchapi/db_info.go`.
- Registered the DB route with the server’s route setup.

### Why
- The UI needs to validate schema versions and FTS support to enable/disable features.
- Standardized DB access is required before larger query work.

### What worked
- The endpoint cleanly reports table presence, FTS tables, and feature flags.
- Workspace resolution supports either `workspace_id` or `db_path`.

### What didn't work
- N/A

### What I learned
- SQLite metadata queries can be used safely even if schema tables are missing, as long as we guard lookups.

### What was tricky to build
The main nuance is handling older DBs that lack `schema_versions` while still returning a useful response; this required table existence checks before querying.

### What warrants a second pair of eyes
- Ensure the feature flags (gopls refs, tree-sitter, doc hits) map cleanly to the UI expectations.

### What should be done in the future
- Add an explicit validation endpoint that checks the expected schema version and required tables.

### Code review instructions
- Start with `refactorio/pkg/workbenchapi/db_info.go` and `refactorio/pkg/workbenchapi/db.go`.
- Verify route registration in `refactorio/pkg/workbenchapi/routes.go`.

### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/db.go refactorio/pkg/workbenchapi/db_info.go refactorio/pkg/workbenchapi/routes.go
git -C refactorio add pkg/workbenchapi/db.go pkg/workbenchapi/db_info.go pkg/workbenchapi/routes.go
git -C refactorio commit -m "Add db info endpoint"
```

## Step 5: Implement Runs + Raw Outputs Endpoints
I implemented the runs listing/detail endpoints, run summaries, and raw output listings. This powers the UI’s “Runs” and “Raw Outputs” pages and provides the run-level metadata needed for session grouping later.

The summary endpoint computes per-run counts using targeted SQL queries and guards against missing tables, which keeps it safe on partially populated DBs.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Build the runs endpoints and commit the implementation with diary updates.

**Inferred user intent:** Surface run metadata early so the UI can navigate data availability and diagnostics.

**Commit (code):** 8619d3c — "Add run and raw output endpoints"

### What I did
- Added `/api/runs`, `/api/runs/:id`, `/api/runs/:id/summary`, and `/api/runs/:id/raw-outputs`.
- Added `/api/raw-outputs` list endpoint with source/run filters.
- Implemented limit/offset parsing and count helpers for run summaries.

### Why
- Run visibility is a prerequisite for the session dashboard and data debugging views.
- Raw outputs provide traceability for ingestion artifacts and errors.

### What worked
- Run list supports basic filtering and pagination.
- Summaries return counts even when some tables are missing.

### What didn't work
- N/A

### What I learned
- Joining diff tables for run-level counts requires multi-step joins through diff_files.

### What was tricky to build
The summary query set had to handle tables that may not exist in older schemas. Guarding with `sqlite_master` checks prevents hard errors while still returning meaningful counts where possible.

### What warrants a second pair of eyes
- Confirm summary counts match UI expectations (especially diff line counts and commit-related joins).

### What should be done in the future
- Add optional totals for `/api/runs` once a paging strategy for large datasets is finalized.

### Code review instructions
- Start with `refactorio/pkg/workbenchapi/runs.go` for query logic and handlers.
- Verify registration in `refactorio/pkg/workbenchapi/routes.go`.

### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/runs.go refactorio/pkg/workbenchapi/routes.go
git -C refactorio add pkg/workbenchapi/runs.go pkg/workbenchapi/routes.go
git -C refactorio commit -m "Add run and raw output endpoints"
```

## Step 6: Implement Session Resolver + /api/sessions Endpoints
I added the session resolver that groups runs by `(root_path, git_from, git_to)` and exposes `/api/sessions` and `/api/sessions/:id`. The resolver infers which domain each run covers by checking for rows in relevant tables, and produces availability flags so the UI can highlight missing data.

I also added optional session overrides stored on the workspace config, allowing manual adjustments via `POST /api/sessions`. Overrides are merged into the computed sessions list and can replace computed entries with custom run mappings.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement the session resolver and session endpoints, commit, and log the progress.

**Inferred user intent:** Make session selection possible so the UI can group runs into coherent slices for browsing.

**Commit (code):** 3ab4469 — "Add sessions endpoints"

### What I did
- Added `Session`, `SessionRuns`, and `SessionOverride` types.
- Implemented `/api/sessions` and `/api/sessions/:id` endpoints.
- Added session grouping and availability computation.
- Added a session override handler (`POST /api/sessions`) that persists overrides in workspace config.
- Registered session routes in the server.

### Why
- Sessions provide the UI with stable, human-meaningful groupings of runs.
- Overrides handle incomplete metadata cases where git ranges are missing.

### What worked
- Sessions are grouped deterministically and include availability flags.
- Overrides persist alongside workspace configuration.

### What didn't work
- N/A

### What I learned
- Run data presence checks need to be table-aware to avoid errors on older schemas.

### What was tricky to build
Balancing readable session IDs with uniqueness required a short hash in the ID. This avoids collisions when two root paths share the same git range while keeping IDs relatively stable.

### What warrants a second pair of eyes
- The decision to fall back to per-run sessions when `git_from`/`git_to` are missing.
- Whether the short hash in session IDs is sufficient for UI deep-linking expectations.

### What should be done in the future
- Add a DELETE endpoint for session overrides if we want users to remove manual mappings.

### Code review instructions
- Start with `refactorio/pkg/workbenchapi/sessions.go` and `refactorio/pkg/workbenchapi/session_types.go`.
- Verify workspace config changes in `refactorio/pkg/workbenchapi/workspace.go`.
- Confirm route registration in `refactorio/pkg/workbenchapi/routes.go`.

### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/session_types.go refactorio/pkg/workbenchapi/sessions.go refactorio/pkg/workbenchapi/workspace.go refactorio/pkg/workbenchapi/routes.go
go test ./pkg/workbenchapi
git -C refactorio add pkg/workbenchapi/session_types.go pkg/workbenchapi/sessions.go pkg/workbenchapi/workspace.go pkg/workbenchapi/routes.go
git -C refactorio commit -m "Add sessions endpoints"
```

## Step 7: Stabilize Session IDs
I fixed a subtle issue where session IDs could change when multiple runs mapped to the same session key. The ID is now assigned only once at session creation time, avoiding accidental suffixing as additional runs are processed.

This keeps session deep links stable and prevents confusing UI behavior when multiple runs share the same git range.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Correct the session ID generation behavior and commit the fix.

**Inferred user intent:** Maintain stable session identifiers for consistent UI navigation.

**Commit (code):** 31f44fe — "Fix session id stability"

### What I did
- Updated session ID assignment to occur only when the session builder is created.
- Ensured subsequent runs for the same session do not override the ID.

### Why
- Session IDs should be stable per session key, not dependent on run iteration order.

### What worked
- IDs now remain consistent even when multiple runs share a session.

### What didn't work
- N/A

### What I learned
- Assigning identifiers inside the per-run loop can create accidental churn.

### What was tricky to build
It was easy to miss that ID assignment occurred on every run iteration because of the map lookup flow; isolating it to the session-creation branch fixes the instability.

### What warrants a second pair of eyes
- Confirm the updated ID logic still avoids collisions across distinct session keys.

### What should be done in the future
- Add a small unit test for session ID stability once we add tests for session logic.

### Code review instructions
- Review `refactorio/pkg/workbenchapi/sessions.go` around session creation and ID assignment.

### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/sessions.go
go test ./pkg/workbenchapi
git -C refactorio add pkg/workbenchapi/sessions.go
git -C refactorio commit -m "Fix session id stability"
```

## Step 8: Sanitize Session IDs for URL Safety
I updated session ID generation to sanitize `git_from` and `git_to` so IDs are safe for URL path segments. This prevents accidental `/` characters or other unsafe symbols from breaking `/api/sessions/:id` routes.

The sanitized label keeps the IDs readable while still anchoring uniqueness on a short hash of `(root_path, git_from, git_to)`.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Apply the review finding about session ID safety and commit the fix.

**Inferred user intent:** Ensure session endpoints are robust for a variety of git ref strings.

**Commit (code):** ba92eb9 — "Sanitize session ids"

### What I did
- Added `sanitizeSessionLabel` to replace non-url-safe characters with `_`.
- Applied sanitization in `buildSessionID`.
- Ran `go test ./pkg/workbenchapi`.

### Why
- Git refs can include slashes or characters that break path-based IDs.

### What worked
- Session IDs now remain stable and URL-safe without changing uniqueness behavior.

### What didn't work
- N/A

### What I learned
- Relying on raw git refs in path segments is brittle; sanitization is safer than URL escaping because Go unescapes `%2F`.

### What was tricky to build
Ensuring IDs stay readable without leaking unsafe path characters required a conservative character whitelist.

### What warrants a second pair of eyes
- Confirm the sanitize rules are acceptable for UI display and debugging.

### What should be done in the future
- Consider exposing a separate `display_label` field if we want the exact git ref shown to users.

### Code review instructions
- Review `buildSessionID` and `sanitizeSessionLabel` in `refactorio/pkg/workbenchapi/sessions.go`.

### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/sessions.go
go test ./pkg/workbenchapi
git -C refactorio add pkg/workbenchapi/sessions.go
git -C refactorio commit -m "Sanitize session ids"
```

## Step 9: Implement Search Endpoints
I added the typed search endpoints (`/api/search/*`) and the unified `POST /api/search` dispatcher. Each search uses the corresponding FTS table and joins back to the core tables to return UI-friendly records, with optional `run_id` and basic filters.

The unified search aggregates per-type results into a single list of `SearchResult` objects, preserving type-specific payloads for richer UI previews.

### Prompt Context
**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement task 7 (search endpoints) after reviewing sessions, then commit and update the diary.

**Inferred user intent:** Provide fast, FTS-backed search across all indexed domains to power the UI’s unified search view.

**Commit (code):** 91f02bb — "Add search endpoints"

### What I did
- Added `/api/search/symbols`, `/api/search/code-units`, `/api/search/diff`, `/api/search/commits`, `/api/search/docs`, `/api/search/files`.\n- Added `POST /api/search` to aggregate cross-domain results.\n- Implemented FTS-backed query helpers for each domain.\n- Registered search routes in the server.\n
### Why
- The UI requires fast cross-domain search, and FTS tables are the intended backend for that.\n
### What worked
- All search endpoints use parameterized SQL and handle missing FTS tables with clear errors.\n- Unified search returns typed results with payloads for preview panels.\n
### What didn't work
- N/A\n
### What I learned
- Keeping the typed endpoints alongside unified search reduces UI complexity and keeps debugging simple.\n
### What was tricky to build
Balancing a shared `SearchResult` shape with domain-specific fields required a `payload` field to avoid losing type-specific context.\n
### What warrants a second pair of eyes
- Verify FTS queries and joins for correctness, especially `diff_lines` and `code_unit_snapshots` joins.\n- Confirm search result field names match frontend expectations.\n
### What should be done in the future
- Add fallback behavior when FTS tables are missing (e.g., LIKE-based search with warnings).\n
### Code review instructions
- Start with `refactorio/pkg/workbenchapi/search.go` and the route registration in `refactorio/pkg/workbenchapi/routes.go`.\n- Verify per-type SQL queries match the schema.\n
### Technical details
Commands run:
```bash
gofmt -w refactorio/pkg/workbenchapi/search.go refactorio/pkg/workbenchapi/routes.go
go test ./pkg/workbenchapi
git -C refactorio add pkg/workbenchapi/search.go pkg/workbenchapi/routes.go
git -C refactorio commit -m "Add search endpoints"
```
