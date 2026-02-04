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
    - Path: refactorio/pkg/refactorindex/ingest_tree_sitter.go
      Note: Tree-sitter captures define structural search domain.
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: Defines the index tables and existing FTS coverage used in search.
    - Path: refactorio/pkg/refactorindex/store.go
      Note: Creates schema and FTS triggers; informs how new search indexes should be added.
ExternalSources: []
Summary: Design for a unified search system over refactorindex data (diffs, symbols, code units, commits, doc hits, tree-sitter, gopls refs).
LastUpdated: 2026-02-04T11:15:59.380767787-05:00
WhatFor: Define queries, CLI verbs, REST API, and implementation plan for search.
WhenToUse: Use when designing or implementing search over the refactor index.
---


# Search System Design

## Goals
- Provide unified search across code units, symbols, references, diffs, commits, tree-sitter captures, and doc hits.
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
- `ts_captures` stores tree-sitter query captures and snippets.
- `doc_hits` stores ripgrep hits for explicit terms, with `doc_hits_fts` already present.

Implications for search:
- Search must be explicit about which `run_id` it targets because runs represent distinct snapshots or ingestion jobs.
- `commit_id` is scoped to a specific `run_id` in the current schema, so commit hash filters must join through `commits`.
- Two FTS tables already exist (`doc_hits_fts`, `diff_lines_fts`), but most other text columns are not indexed for full-text search.

## Search Domains and Core Queries
Each domain has its own query shape, filters, and ranking signals. The system should support them independently and via a unified entry point.

### Symbols
Use case: “Find symbols named `Foo` in package `bar` that are exported and defined in a specific file.”
- Primary table: `symbol_defs`, `symbol_occurrences`, `files`.
- Filters: `kind`, `name`, `pkg`, `file path`, `exported`, `run_id`.
- Suggested index additions: FTS on `symbol_defs.name`, `symbol_defs.signature`, and optionally `symbol_defs.pkg` for fuzzy lookup.

### Symbol References
Use case: “Find all references to symbol hash `X` in commit `Y`.”
- Primary table: `symbol_refs`, `files`, `commits`.
- Filters: `symbol_hash`, `run_id`, `commit_hash`, `file path`, `is_decl`.
- Notes: Requires join from `symbol_defs` to `symbol_refs` by symbol hash.

### Code Units
Use case: “Search for functions containing `ctx` and `errgroup` in their body or doc comment.”
- Primary table: `code_units`, `code_unit_snapshots`, `files`.
- Filters: `kind`, `name`, `pkg`, `recv`, `file path`, `run_id`, `commit_hash`.
- Suggested index additions: FTS on `code_unit_snapshots.body_text` and `code_unit_snapshots.doc_text`.
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
- Suggested index additions: FTS on `subject` and `body`.

### Tree-Sitter Captures
Use case: “Find all `call_expression` captures matching a query named `http-calls`.”
- Primary table: `ts_captures`.
- Filters: `query_name`, `capture_name`, `node_type`, `file path`, `run_id`, `commit_hash`.
- Suggested index additions: FTS on `snippet` and optional text columns for `query_name` and `capture_name`.

### Files
Use case: “Find files by path or extension.”
- Primary table: `files`.
- Filters: `path`, `ext`, `is_binary`, `file_exists`.
- Suggested index additions: FTS on `path` for fast partial matching.

## Unified Search Query Model
A single query entry point should accept both free-text and structured filters. Suggested fields:
- `query`: text to search in FTS-backed domains.
- `types`: list of domains, e.g., `code_units`, `diffs`, `commits`, `symbols`, `docs`, `tree_sitter`.
- `run_id`: required for most searches unless `commit_hash` is provided.
- `commit_hash`: join to commit_id for commit-scoped searches.
- `filters`: structured filters such as `kind`, `pkg`, `file`, `capture_name`, `author`.
- `limit` and `offset`.

The query layer should be able to dispatch to per-domain queries, normalize outputs, and return a unified result set.

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

## Implementation Plan
1. Add a `search` package in `pkg/refactorindex` with typed filters and result structs.
2. Implement per-domain query functions in the store, each returning a normalized result struct.
3. Add FTS tables for domains that need full-text search.
4. Add a unified search dispatcher that merges per-domain results and normalizes ranking.
5. Wire new CLI verbs using glazed command descriptions and common flag sections.
6. Add optional REST API service for search and indexing metadata.
7. Add tests and smoke tests for each query type.

### Schema and Index Changes
Suggested additions:
- `code_unit_snapshots_fts` on `body_text` and `doc_text`.
- `symbol_defs_fts` on `name`, `signature`, `pkg`.
- `commits_fts` on `subject` and `body`.
- `ts_captures_fts` on `snippet`.
- `files_fts` on `path`.

Use the same FTS5 pattern already present in `store.ensureFTS` so it integrates with triggers and rebuilds.

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
- Consider adding an index or view that maps `commit_hash` to a stable commit identity across runs if cross-run search is needed later.

### Data Quality and Limitations
- `doc_hits` only includes matches for terms in the terms file; it is not a general text index.
- For general text search, add a `file_contents` ingest step or add a code-unit-level text search using `body_text` and `doc_text`.
- `gopls` references depend on `gopls` availability and symbol target specs; failures should be captured in raw outputs.

### Testing Strategy
- Add per-domain search smoke tests similar to existing ingest smoke tests.
- Include at least one FTS query test for each new FTS table.
- Verify `run_id` scoping and `commit_hash` filtering.

## Open Questions
- Should search default to the latest run, or require explicit `--run-id`?
- Do we want a persistent unified `search_documents` table for global search, or is a per-domain dispatcher sufficient?
- Should we support cross-run queries (e.g., “search across all runs for symbol X”) and how should those be ranked?
