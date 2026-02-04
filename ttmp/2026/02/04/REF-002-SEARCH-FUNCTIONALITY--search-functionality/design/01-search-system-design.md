---
Title: Search System Design
Ticket: REF-002-SEARCH-FUNCTIONALITY
Status: active
Topics: []
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/pkg/refactorindex/ingest_code_units.go
      Note: Code unit snapshots provide body/doc text for full-text search.
    - Path: refactorio/pkg/refactorindex/ingest_doc_hits.go
      Note: Doc hits ingestion and existing FTS table for text search.
    - Path: refactorio/pkg/refactorindex/ingest_gopls_refs.go
      Note: Gopls references define symbol reference search domain.
    - Path: refactorio/pkg/refactorindex/ingest_symbols.go
      Note: Symbol ingestion shapes symbol search capabilities.
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: Defines the index tables and existing FTS coverage used in search.
    - Path: refactorio/pkg/refactorindex/store.go
      Note: Creates schema and FTS triggers; informs how new search indexes should be added.
ExternalSources: []
Summary: Design for a unified search system over refactorindex data (diffs, symbols, code units, commits, doc hits, files, gopls refs; tree-sitter optional).
LastUpdated: 2026-02-04T14:20:00-05:00
WhatFor: Define queries, CLI verbs, REST API, and implementation plan for search.
WhenToUse: Use when designing or implementing search over the refactor index.
---


# Search System Design

## Goals
- Provide unified search across code units, symbols, references, diffs, commits, doc hits, and file paths (tree-sitter captures optional).
- Support structured filters by run, commit, file path, package, symbol kind, capture name, and change type.
- Keep SQLite-first, relying on FTS5 where possible.
- Offer both CLI and REST interfaces with consistent query semantics.
- Make search results traceable back to the underlying index rows.

## Non-Goals
- Real-time indexing of live editor buffers.
- Cross-repo federated search.
- Search-specific storage beyond SQLite unless scalability requires it later.

## Current Index Structure (Summary)
The refactor index already captures multiple data domains keyed by `run_id` and optional `commit_id`:
- `meta_runs` and `raw_outputs` track ingestion metadata and raw tool outputs.
- `files` records normalized file paths.
- `diff_files`, `diff_hunks`, `diff_lines` capture git diff data. `diff_lines_fts` already exists for full-text search of diff line text.
- `symbol_defs` and `symbol_occurrences` store symbol definitions and their file/position occurrences.
- `code_units` and `code_unit_snapshots` store function/type snapshots, including `body_text` and optional `doc_text`.
- `commits`, `commit_files`, and `file_blobs` store commit history and per-file changes.
- `symbol_refs` stores gopls reference locations.
- `ts_captures` stores tree-sitter query captures and snippets (ingestion currently disabled).
- `doc_hits` stores ripgrep hits for explicit terms, with `doc_hits_fts` already present.

Implications for search:
- Search must be explicit about which `run_id` it targets because runs represent distinct snapshots or ingestion jobs.
- `commit_id` is scoped to a specific `run_id` in the current schema, so commit hash filters must join through `commits`.
- FTS tables now exist for `doc_hits`, `diff_lines`, `code_unit_snapshots`, `symbol_defs`, `commits`, and `files`.

## Search Domains and Core Queries
Each domain has its own query shape, filters, and ranking signals. The system should support them independently and via a unified entry point.

### Symbols
Use case: “Find symbols named `Foo` in package `bar` that are exported and defined in a specific file.”
- Primary table: `symbol_defs`, `symbol_occurrences`, `files`.
- Filters: `kind`, `name`, `pkg`, `file path`, `exported`, `run_id`.
- FTS: `symbol_defs_fts` on `name`, `signature`, `pkg`.

### Symbol References
Use case: “Find all references to symbol hash `X` in commit `Y`.”
- Primary table: `symbol_refs`, `files`, `commits`.
- Filters: `symbol_hash`, `run_id`, `commit_hash`, `file path`, `is_decl`.
- Notes: Requires join from `symbol_defs` to `symbol_refs` by symbol hash.

### Code Units
Use case: “Search for functions containing `ctx` and `errgroup` in their body or doc comment.”
- Primary table: `code_units`, `code_unit_snapshots`, `files`.
- Filters: `kind`, `name`, `pkg`, `recv`, `file path`, `run_id`, `commit_hash`.
- FTS: `code_unit_snapshots_fts` on `body_text` and `doc_text`.
- Ranking: `bm25` with higher weight on `doc_text` and on exact name matches.

### Doc Hits
Use case: “Find documents containing terms from a predefined list.”
- Primary table: `doc_hits`, `files`.
- Filters: `term`, `run_id`, `commit_hash`, `file path`.
- Current limitation: only terms in the terms file are indexed; this is not a general full-text index.
- Suggested enhancement: optional `file_contents` index for general repo text search, or a scheduled re-index for dynamic term sets.

### Diffs
Use case: “Search for added lines containing `context.Background` in a diff run.”
- Primary table: `diff_lines` joined through `diff_hunks` and `diff_files` to `files`.
- Filters: `kind` (added/removed/context), `run_id`, `file path`, `old_path`, `new_path`.
- Full-text: `diff_lines_fts` already exists.

### Commits
Use case: “Search commit messages mentioning `refactor` between two refs.”
- Primary table: `commits`.
- Filters: `author_name`, `author_email`, `date range`, `run_id`.
- FTS: `commits_fts` on `subject` and `body`.

### Tree-Sitter Captures
Use case: “Find all `call_expression` captures matching a query named `http-calls`.”
- Primary table: `ts_captures`.
- Filters: `query_name`, `capture_name`, `node_type`, `file path`, `run_id`, `commit_hash`.
- If tree-sitter is reintroduced, add FTS on `snippet` and optional text columns for `query_name` and `capture_name`.

### Files
Use case: “Find files by path or extension.”
- Primary table: `files`.
- Filters: `path`, `ext`, `is_binary`, `file_exists`.
- FTS: `files_fts` on `path` for fast partial matching.

## Unified Search Query Model
A single query entry point should accept both free-text and structured filters. Suggested fields:
- `query`: text to search in FTS-backed domains.
- `types`: list of domains, e.g., `code_units`, `diffs`, `commits`, `symbols`, `docs`, `tree_sitter`.
- `run_id`: required for most searches unless `commit_hash` is provided.
- `commit_hash`: join to commit_id for commit-scoped searches.
- `filters`: structured filters such as `kind`, `pkg`, `file`, `capture_name`, `author`.
- `limit` and `offset`.

The query layer should be able to dispatch to per-domain queries, normalize outputs, and return a unified result set.

## Real-World Use Cases (Cross-Table, Multi-Commit)
These scenarios describe user-level workflows that require joins across multiple tables and often span multiple commits. They should drive the search API/CLI shape and influence ranking and result normalization.

### 1) “Who introduced the new error handling path and where did it spread?”
User story: After noticing a new error wrapping pattern, a developer wants to identify the commit where it was introduced and follow how it propagated across the codebase.
Tables involved: `commits`, `commit_files`, `diff_lines` (+ `diff_lines_fts`), `files`, `code_unit_snapshots`.
Cross-commit aspect: starts from a commit message or diff snippet, then scans subsequent commits for the same signature pattern.
Query steps:
- Search commits by message/body for “error wrapping”, “errors.Wrap”, or “fmt.Errorf”.
- For each candidate commit, search diff lines (added lines) containing the pattern.
- Link to code unit snapshots in later commits that now contain the same pattern.
Desired output:
- A timeline of commits with matching diffs, and per-commit file locations with snippets.

### 2) “Which code units changed after a symbol rename?”
User story: A symbol rename should update references across multiple files and commits; the user wants to confirm where it changed and where it did not.
Tables involved: `symbol_defs`, `symbol_refs`, `commits`, `commit_files`, `files`, `diff_lines`.
Cross-commit aspect: compare ref counts and file lists before and after the rename across a commit range.
Query steps:
- Resolve the old/new symbol hashes (or name + pkg + kind).
- Retrieve references for each commit in a range and diff their file sets.
- Combine with diff lines for the commits that did the rename.
Desired output:
- A report of files where the symbol changed, grouped by commit, with missing references flagged.

### 3) “Find all new external API calls added in the last 20 commits”
User story: Security review wants to see new outbound calls to `http.NewRequest` or `grpc.Dial`.
Tables involved: `commits`, `diff_lines` (+ FTS), `files`, `ts_captures` (if tree-sitter queries exist for call expressions).
Cross-commit aspect: scan a commit range and aggregate hits by symbol or capture name.
Query steps:
- Use tree-sitter captures for `call_expression` or `selector_expression` matching known APIs.
- Cross-check diff lines for added calls to those APIs.
- Tie each match to the commit and file path.
Desired output:
- A grouped list of new API calls with commit hash, file, and line snippet.

### 4) “Which functions were touched by a specific author and how did they evolve?”
User story: A reviewer wants to see how a particular author changed a function over time.
Tables involved: `commits`, `commit_files`, `code_unit_snapshots`, `files`.
Cross-commit aspect: list code unit snapshots across multiple commits ordered by commit date.
Query steps:
- Filter commits by `author_email` and a date range.
- Map commit files to code unit snapshots within those files and commits.
- Optionally diff snapshot bodies across consecutive commits for the same code unit hash.
Desired output:
- A per-function history (commit timeline with snippet diffs).

### 5) “Track introduction of a TODO term into docs and code”
User story: A team wants to eliminate specific TODO phrases across repo history.
Tables involved: `doc_hits` (+ FTS), `diff_lines` (+ FTS), `commits`, `files`.
Cross-commit aspect: show first appearance and follow-up changes across commits.
Query steps:
- Search doc hits and diff lines for the TODO term across runs/commits.
- Identify the earliest commit where the term appears.
- List subsequent commits that add/remove occurrences.
Desired output:
- First introduction commit, followed by a chronological list of add/remove events with file locations.

### 6) “Which refactors introduced new tree-sitter patterns?”
User story: Query tree-sitter captures for a pattern (e.g., new `interface` usage) and map it to refactor commits.
Tables involved: `ts_captures`, `commits`, `commit_files`, `files`.
Cross-commit aspect: find commits where capture count changes significantly.
Query steps:
- Count captures per commit for a given query name/capture name.
- Join with commit metadata and diff files to see impacted areas.
Desired output:
- A list of commits with capture count deltas and file-level breakdowns.

### 7) “Search for code-unit doc comments that mention a deprecated concept, then link to change history”
User story: Documentation hygiene requires replacing deprecated terminology; identify doc comments that still mention it and see when they last changed.
Tables involved: `code_unit_snapshots` (+ FTS), `commits`, `files`, `commit_files`.
Cross-commit aspect: for each matching code unit, find the most recent commit that touched its file.
Query steps:
- Full-text search in `code_unit_snapshots.doc_text`.
- For each match, find the latest commit (in range) that modified the file.
Desired output:
- List of code units with the outdated term and their most recent change metadata.

## CLI Design
Add a `search` command group to `refactor-index`.

Suggested verbs and flags:
- `refactor-index search` (global search)
- `refactor-index search symbols`
- `refactor-index search refs`
- `refactor-index search code-units`
- `refactor-index search docs`
- `refactor-index search diff`
- `refactor-index search commits`
- `refactor-index search tree-sitter`
- `refactor-index search files`

Common flags:
- `--db PATH` (required)
- `--run-id ID` (optional but strongly recommended)
- `--commit-hash HASH` (optional)
- `--query TEXT` (optional for structured-only searches)
- `--limit N` and `--offset N`
- `--format` for glazed output styles

Example CLI flows:
- `refactor-index search --db index.db --run-id 42 --query "errgroup" --types code-units,docs,diff`
- `refactor-index search symbols --db index.db --run-id 42 --kind func --name NewStore`
- `refactor-index search diff --db index.db --run-id 42 --query "context.Background" --kind +`
- `refactor-index search commits --db index.db --query "refactor" --author "manuel"`

## REST API Design
Expose a minimal HTTP API that mirrors the CLI semantics.

### POST /api/search
Request body:
- `query` string
- `types` array
- `run_id` integer
- `commit_hash` string
- `filters` map
- `limit` integer
- `offset` integer

Response body:
- `results` array of search hits
- `total` integer
- `next_offset` integer

Each result should include:
- `type` (domain)
- `score` (fts score if applicable)
- `run_id`
- `commit_hash` (if available)
- `file_path` and `span` (line/col range)
- `snippet` (highlighted content)
- `metadata` (domain-specific fields)

### Domain-specific endpoints
- `GET /api/search/symbols`
- `GET /api/search/refs`
- `GET /api/search/code-units`
- `GET /api/search/docs`
- `GET /api/search/diff`
- `GET /api/search/commits`
- `GET /api/search/tree-sitter`
- `GET /api/search/files`

### Supporting endpoints
- `GET /api/runs` and `GET /api/runs/{id}`
- `GET /api/commits/{hash}`
- `GET /api/symbols/{hash}`

## Alternative Entry Points (Reduce /search Overload)
Instead of a single, highly-general `/search`, provide multiple entry points tailored to user intent. These can coexist with `/api/search` as a power-user endpoint.

### Option A: Domain-Specific Search Endpoints (Strongly Typed)
Purpose: Keep query parameters focused and stable per domain.
- `GET /api/search/symbols`
- `GET /api/search/refs`
- `GET /api/search/code-units`
- `GET /api/search/diff`
- `GET /api/search/commits`
- `GET /api/search/tree-sitter`
- `GET /api/search/docs`
- `GET /api/search/files`
Benefits: simpler parameter sets, easier caching, clearer validation rules.

### Option B: Intent-Based Endpoints (User Workflow Focused)
Purpose: Align with real tasks instead of storage shape.
- `POST /api/trace/introductions` (find first introduction of a term/pattern)
- `POST /api/trace/propagation` (track how a pattern spread across commits)
- `POST /api/impact/symbol-rename` (detect reference changes after a rename)
- `POST /api/audit/external-calls` (scan commit ranges for outbound API usage)
- `POST /api/hygiene/deprecated-terms` (find deprecated terms in docs/code units)
Benefits: matches user stories, reduces need to build complex client-side orchestration.

### Option C: Exploration Endpoints (Cross-Domain Discovery)
Purpose: Provide safe defaults for broad queries without full generality.
- `GET /api/explore/changes` (diff-driven text search; default to added lines)
- `GET /api/explore/definitions` (symbols + code units with name filters)
- `GET /api/explore/usage` (symbol refs + doc hits around a term)
Benefits: curated search scope with fewer parameters.

### Option D: Aggregation Endpoints (Summary/Stats)
Purpose: Return aggregates instead of raw results for dashboards.
- `GET /api/stats/term-frequency`
- `GET /api/stats/capture-deltas`
- `GET /api/stats/symbol-refs`
Benefits: supports metrics and trend views without large result sets.

### Option E: Saved Search / Named Query Endpoints
Purpose: Persist complex queries and reuse them.
- `POST /api/search/saved`
- `GET /api/search/saved/{id}`
- `POST /api/search/saved/{id}/run`
Benefits: reduces client complexity; improves reproducibility.

### CLI Parallels (Optional)
Match entry points with CLI verbs to keep UX consistent:
- `refactor-index search symbols ...`
- `refactor-index trace introduction ...`
- `refactor-index audit external-calls ...`
- `refactor-index stats term-frequency ...`

### Recommendation
Keep `/api/search` for advanced users, but emphasize domain-specific or intent-based endpoints for most workflows. This reduces parameter overload and makes query validation and documentation clearer.

## Implementation Plan
1. Add a `search` package in `pkg/refactorindex` with typed filters and result structs.
2. Implement per-domain query functions in the store, each returning a normalized result struct.
3. Add FTS tables for domains that need full-text search.
4. Add a unified search dispatcher that merges per-domain results and normalizes ranking.
5. Wire new CLI verbs using glazed command descriptions and common flag sections.
6. Add optional REST API service for search and indexing metadata.
7. Add tests and smoke tests for each query type.

### Schema and Index Changes
Current additions:
- `code_unit_snapshots_fts` on `body_text` and `doc_text`.
- `symbol_defs_fts` on `name`, `signature`, `pkg`.
- `commits_fts` on `subject` and `body`.
- `files_fts` on `path`.
- `v_last_commit_per_file` view for latest commit per file per run.
- `diff_lines_fts` and `doc_hits_fts` already exist for diff and doc hit search.
- Tree-sitter FTS is not present while ingestion is disabled.

Use the same FTS5 patterns in `store.ensureFTS` / `ensureFTSColumns` so they integrate with triggers and rebuilds.

### Query Result Normalization
Every domain-specific query should map to a shared result struct:
- `type`
- `primary_key` or `row_id`
- `score`
- `run_id` and `commit_hash`
- `file_path`
- `start_line`, `start_col`, `end_line`, `end_col`
- `snippet`
- `metadata`

### Ranking and Snippets
- Use FTS5 `bm25` for per-domain ranking where available.
- Prefer exact symbol or file path matches above FTS results for deterministic output.
- Use `snippet()` or `highlight()` to produce preview text.

### Run and Commit Semantics
- Require `run_id` for most queries, since the schema is run-scoped.
- When `commit_hash` is provided, resolve it to `commit_id` through `commits` for the same run.
- Use `v_last_commit_per_file` for “latest commit touching file” lookups instead of manual joins.
- Consider adding an index or view that maps `commit_hash` to a stable commit identity across runs if cross-run search is needed later.

### Data Quality and Limitations
- `doc_hits` only includes matches for terms in the terms file; it is not a general text index.
- For general text search, add a `file_contents` ingest step or add a code-unit-level text search using `body_text` and `doc_text`.
- `gopls` references depend on `gopls` availability and symbol target specs; failures should be captured in raw outputs.
- Tree-sitter ingestion is currently disabled, so `ts_captures` may be empty unless re-enabled.

### Testing Strategy
- Add per-domain search smoke tests similar to existing ingest smoke tests.
- Include at least one FTS query test for each new FTS table.
- Verify `run_id` scoping and `commit_hash` filtering.

## Open Questions
- Should search default to the latest run, or require explicit `--run-id`?
- Do we want a persistent unified `search_documents` table for global search, or is a per-domain dispatcher sufficient?
- Should we support cross-run queries (e.g., “search across all runs for symbol X”) and how should those be ranked?
