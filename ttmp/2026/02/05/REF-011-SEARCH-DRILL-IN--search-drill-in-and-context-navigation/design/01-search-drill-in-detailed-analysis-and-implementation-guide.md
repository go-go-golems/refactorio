---
Title: Search Drill-In Detailed Analysis and Implementation Guide
Ticket: REF-011-SEARCH-DRILL-IN
Status: active
Topics:
    - ui
    - refactorio
    - frontend
    - search
    - deep-linking
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/ui/src/pages/SearchPage.tsx
      Note: Current unified search page and query sync baseline.
    - Path: refactorio/ui/src/components/data-display/SearchResults.tsx
      Note: Current search results rendering and selection behavior.
    - Path: refactorio/ui/src/types/api.ts
      Note: Shared SearchResult and SearchRequest contracts.
    - Path: refactorio/ui/src/App.tsx
      Note: Route map and global navigation entry points.
    - Path: refactorio/ui/src/pages/SymbolsPage.tsx
      Note: Symbol list and inspector target for symbol drill-in.
    - Path: refactorio/ui/src/pages/CodeUnitsPage.tsx
      Note: Code unit list and inspector target for code unit drill-in.
    - Path: refactorio/ui/src/pages/CommitsPage.tsx
      Note: Commit list and inspector target for commit drill-in.
    - Path: refactorio/ui/src/pages/DiffsPage.tsx
      Note: Three-pane diff exploration target for diff drill-in.
    - Path: refactorio/ui/src/pages/DocsPage.tsx
      Note: Doc terms and hits target for doc drill-in.
    - Path: refactorio/ui/src/pages/FilesPage.tsx
      Note: File tree and file content target for file drill-in.
    - Path: refactorio/ui/src/hooks/useSessionContext.ts
      Note: Session-scoped run mapping used by search requests.
    - Path: refactorio/pkg/workbenchapi/search.go
      Note: Unified search payload shape and per-type payload records.
    - Path: refactorio/pkg/workbenchapi/symbols.go
      Note: Symbol lookup and references endpoints used in deep links.
    - Path: refactorio/pkg/workbenchapi/code_units.go
      Note: Code unit lookup endpoint used in deep links.
    - Path: refactorio/pkg/workbenchapi/commits.go
      Note: Commit lookup and files endpoints used in deep links.
    - Path: refactorio/pkg/workbenchapi/diffs.go
      Note: Diff run and diff file endpoints used in deep links.
    - Path: refactorio/pkg/workbenchapi/docs.go
      Note: Doc hits endpoint used in deep links.
    - Path: refactorio/pkg/workbenchapi/files.go
      Note: File list/content endpoints used in deep links.
    - Path: refactorio/sources/ui-design.md
      Note: Source design spec that requires stable contextual drill-in.
ExternalSources: []
Summary: Detailed analysis and phased implementation plan for search result drill-in, deep links, and cross-view context navigation.
LastUpdated: 2026-02-06T00:00:00-05:00
WhatFor: Define exactly how search results navigate to contextual detail views, what the UX should look like, and which code paths must change.
WhenToUse: Use when implementing or reviewing search drill-in and deep-link behavior across Workbench pages.
---

# Search Drill-In Detailed Analysis and Implementation Guide

## 1. Executive Summary

`REF-011-SEARCH-DRILL-IN` introduces the missing link between global search and contextual investigation workflows. Today, users can search across symbols, code units, commits, diffs, docs, and files, but cannot consistently open a result into the correct destination view with preserved context.

This guide defines:

1. What drill-in means in Workbench and what user experience it should provide.
2. A canonical deep-link contract for each result type.
3. A concrete file-by-file implementation plan across UI and API layers.
4. A phased delivery sequence aligned with nearby tickets (`REF-009`, `REF-010`, `REF-012`, `REF-013`).
5. A complete validation strategy (unit, component, integration, and manual QA).

Primary outcome:

- Every search result row becomes actionable and resolves to a stable contextual location in the app.
- Users can share links to exact entities and reopen them reliably after reload.
- Search transitions from "result list only" to "result list plus contextual navigation pipeline".

## 2. Problem Statement

## 2.1 Current Behavior

Current Search view (`ui/src/pages/SearchPage.tsx`) executes `POST /api/search` and renders `SearchResults`, grouped by type. Result rows call `onSelect` if provided, but SearchPage does not pass selection handlers and does not navigate on click.

Current limitations:

1. Clicking a search result does not open any detail view.
2. Result selection state is ephemeral and not encoded in URL.
3. Destination pages maintain selection in local state only.
4. There is no shared routing contract for "open this symbol/code unit/commit/diff/doc/file in context".

## 2.2 Product and UX Misalignment

The UI design spec in `sources/ui-design.md` explicitly expects:

1. Drill-in from search to stable contextual detail.
2. Read-in-context navigation between definitions, references, diffs, commits, docs, and files.
3. Deep links that can be copied and shared.

Current implementation satisfies result rendering but not contextual drill-in behavior.

## 2.3 Why This Matters

Without drill-in:

1. Search output cannot drive investigation tasks.
2. Users must manually re-find items in destination pages.
3. Session-scoped context is lost between search and exploration.
4. "Find -> inspect -> act" loops are slower and error-prone.

## 3. Scope Definition

## 3.1 In Scope for REF-011

1. Result-click behavior on Search page.
2. Canonical route/query contract per result type.
3. URL-driven selection hydration on destination pages.
4. Optional in-page preview panel hooks for immediate context.
5. "Copy link" generation from search results.
6. Keyboard drill-in (`Enter`) from selected result row.

## 3.2 Out of Scope for REF-011

1. Full global URL-state normalization for all filters (owned by `REF-013`).
2. Topbar session selector redesign (owned by `REF-009`).
3. File tree lazy-loading architecture work (owned by `REF-010`).
4. Full error-surface standardization (owned by `REF-012`).
5. Plan/audit/report workflows (future tickets).

## 3.3 Integration Boundaries with Neighbor Tickets

1. `REF-009`: Drill-in links should continue to work regardless of whether session is selected via dashboard or topbar.
2. `REF-010`: File drill-in should not block on lazy loading; it may force-load needed path prefixes.
3. `REF-012`: Drill-in must surface not-found/missing-data errors clearly, but shared error components can be added later.
4. `REF-013`: REF-011 introduces minimal link contract; REF-013 can later generalize URL-state management patterns.

## 4. Target User Experience

## 4.1 UX Goals

1. A result click always does something useful.
2. The destination opens with the target entity pre-selected.
3. If detail data exists, inspector/detail panes open automatically.
4. URL reflects target context so refresh and sharing preserve intent.
5. Back navigation returns users to search query context.

## 4.2 Search Page Visual Behavior

When a query is present:

1. Results remain grouped by type.
2. Each row offers primary click behavior: "Open detail".
3. Secondary controls appear on hover/focus:
4. `Open` (same tab)
5. `Open in new tab`
6. `Copy link`

Keyboard behavior:

1. Up/down changes focused result.
2. `Enter` opens focused result.
3. `Cmd/Ctrl+Enter` opens focused result in new tab (optional phase 2).

## 4.3 Destination Behavior by Type

1. `symbol`: open Symbols page, prefilter and preselect symbol, open Symbol inspector.
2. `code_unit`: open Code Units page, prefilter and preselect code unit, open Code Unit inspector.
3. `commit`: open Commits page, prefilter and preselect commit, open Commit inspector.
4. `diff`: open Diffs page with run preselected, file preselected, optional line-focused anchor.
5. `doc`: open Docs page with term selected and target hit highlighted.
6. `file`: open Files page with path selected and file content loaded.

## 4.4 Missing-Data UX Rules

If target entity is missing in active session/run:

1. Keep user on destination page.
2. Show an inline callout near list/inspector:
3. "Target from search link not found in current scope."
4. Include debugging metadata (type, id/hash/path, run_id).
5. Preserve link query params to support retries after session switch.

## 5. Canonical Deep-Link Contract

This section defines the minimum stable contract for drill-in links generated from search results.

## 5.1 Shared Parameters

Common query params:

1. `from=search`
2. `q=<original search query>`
3. `session_id=<active session id>` (when available)
4. `run_id=<result run id>` (when meaningful)

Rule: destination pages must ignore unknown params and only consume known keys.

## 5.2 Route Mapping Table

| Search result type | Destination route | Required params | Optional params |
| --- | --- | --- | --- |
| `symbol` | `/symbols` | `symbol_hash` | `run_id`, `line`, `path`, `q`, `session_id` |
| `code_unit` | `/code-units` | `unit_hash` | `run_id`, `line`, `path`, `q`, `session_id` |
| `commit` | `/commits` | `commit_hash` | `run_id`, `q`, `session_id` |
| `diff` | `/diffs` | `run_id`, `path` | `line_new`, `line_old`, `hunk_id`, `q`, `session_id` |
| `doc` | `/docs` | `term` | `path`, `line`, `col`, `run_id`, `q`, `session_id` |
| `file` | `/files` | `path` | `line`, `q`, `session_id` |

## 5.3 Result Identity Keys

Use stable keys from current backend payload:

1. `symbol`: `payload.symbol_hash` preferred; fallback tuple `(primary,path,line,col,run_id)`.
2. `code_unit`: `payload.unit_hash` preferred.
3. `commit`: `commit_hash` or `payload.hash`.
4. `diff`: tuple `(run_id,path,line_new,line_old,payload.hunk_id)`.
5. `doc`: tuple `(run_id,term,path,line,col)`.
6. `file`: `path`.

## 5.4 URL Examples

1. Symbol:
`/symbols?symbol_hash=abc123&run_id=22&from=search&q=CommandProcessor`

2. Code unit:
`/code-units?unit_hash=def456&run_id=23&from=search&q=execute`

3. Commit:
`/commits?commit_hash=789abcd&run_id=14&from=search&q=context`

4. Diff:
`/diffs?run_id=14&path=pkg/workbenchapi/search.go&line_new=260&from=search&q=query`

5. Doc:
`/docs?term=CommandProcessor&path=docs/api.md&line=23&run_id=18&from=search&q=CommandProcessor`

6. File:
`/files?path=ui/src/pages/SearchPage.tsx&line=59&from=search&q=SearchPage`

## 6. Backend Data Contract Analysis

## 6.1 Current Unified Search Payload

`pkg/workbenchapi/search.go` returns `SearchResult` with:

1. `type`
2. `primary`
3. `secondary`
4. `path`
5. `line`
6. `col`
7. `snippet`
8. `run_id`
9. `commit_hash`
10. `payload` (type-specific record)

This is sufficient for baseline drill-in without backend changes.

## 6.2 Type-Specific Payload Availability

1. Symbol payload includes `symbol_hash`, `kind`, `pkg`, file location.
2. Code unit payload includes `unit_hash`, `kind`, `pkg`, file range.
3. Diff payload includes `run_id`, `path`, `line_no_old`, `line_no_new`, `hunk_id`, `diff_file_id`.
4. Commit payload includes `hash`, subject/body, author metadata.
5. Doc payload includes `term`, `path`, `line`, `col`, `match_text`.
6. File payload includes `path`, extension, binary/existence flags.

## 6.3 Backend Gaps (Non-Blocking)

1. Unified search does not return a first-class stable `id` field.
2. Diff drill-in has no dedicated endpoint for "open by hunk_id".
3. Session override semantics are implicit in `run_ids` map, not explicit in result metadata.

These are optional improvements for phase 2 or separate tickets.

## 6.4 Backend Changes Recommended (Optional)

If implemented, these improve robustness but are not required for phase 1:

1. Add `id` to unified `SearchResult` as type-specific stable key.
2. Add `location` object with normalized shape `{path,line,col}`.
3. Add optional `session_id` echo in response for diagnostics.
4. Add `/diff/:run_id/hunk/:id` endpoint for exact hunk opening.

## 7. Frontend Architecture Plan

## 7.1 High-Level Approach

Implement drill-in as a routing contract, not as a new global store.

Principles:

1. Search page builds links from result records.
2. Destination pages hydrate selection from URL params.
3. Existing local state remains, but URL seed occurs on initial load.
4. Unknown params are ignored to preserve forward compatibility.

## 7.2 New Utilities

Add a shared helper module:

`ui/src/features/search-drill-in/`

Suggested files:

1. `linkBuilder.ts`
2. `parsers.ts`
3. `types.ts`

Responsibilities:

1. Convert `SearchResult + query + session` into destination URL.
2. Parse destination URL params into typed drill-in intents.
3. Validate required fields and emit typed errors for missing data.

## 7.3 Search Results Component Changes

`ui/src/components/data-display/SearchResults.tsx`

Current behavior:

- `onSelect` callback only.

Planned changes:

1. Add `onOpen?: (result: SearchResult) => void` prop.
2. Default row click uses `onOpen` when present; fallback to `onSelect`.
3. Add optional row-level action controls:
4. Open
5. Copy link
6. Open new tab (if `openInNewTab` callback provided)
7. Improve accessibility labels for clickable rows.

## 7.4 Search Page Changes

`ui/src/pages/SearchPage.tsx`

Planned behavior:

1. Use `useNavigate()` for same-tab navigation.
2. Build destination URL via `linkBuilder`.
3. Pass `onOpen` to `SearchResults`.
4. Add copy-link action using `navigator.clipboard` with fallback.
5. Preserve search query in outgoing links via `q` param.

Pseudo-code:

```ts
const handleOpenResult = (result: SearchResult) => {
  const href = buildSearchDrillInHref({
    result,
    query,
    sessionId,
    source: 'search',
  })
  navigate(href)
}
```

## 7.5 Destination Page Hydration Changes

Each page adds a small URL-hydration effect.

Pattern:

1. Parse `useSearchParams()`.
2. If expected key exists (e.g., `symbol_hash`), attempt to resolve target from loaded list.
3. If not in current page list, use detail endpoint to fetch exact entity when possible.
4. Set local `selected` state and optional filters to bring target into view.

## 8. Per-Page Detailed Implementation

## 8.1 Symbols Page (`ui/src/pages/SymbolsPage.tsx`)

### Incoming Params

1. `symbol_hash`
2. `run_id` (optional override)
3. `path`, `line` (optional context)

### Behavior

1. If `symbol_hash` exists:
2. Try to find in current `symbols` list.
3. If absent, call `useGetSymbolQuery({hash, run_id})`.
4. Set `selectedSymbol` from found/fetched record.
5. Keep existing inspector behavior unchanged.

### Required Additions

1. URL parsing effect.
2. Optional query override for `run_id` when provided.
3. Not-found callout state.

## 8.2 Code Units Page (`ui/src/pages/CodeUnitsPage.tsx`)

### Incoming Params

1. `unit_hash`
2. `run_id`

### Behavior

1. Resolve code unit by hash from list or detail endpoint.
2. Preselect and open inspector.
3. Optionally set `searchQuery` to code unit name if direct resolution fails (fallback).

### Required Additions

1. URL parsing effect.
2. Fallback detail fetch by hash.
3. Not-found callout.

## 8.3 Commits Page (`ui/src/pages/CommitsPage.tsx`)

### Incoming Params

1. `commit_hash`
2. `run_id` (optional)

### Behavior

1. Resolve commit from list or `useGetCommitQuery`.
2. Set `selected` commit and open inspector.
3. Preload files with existing `useGetCommitFilesQuery` flow.

### Required Additions

1. URL parsing effect.
2. Optional run scoping if `run_id` present.
3. Not-found callout.

## 8.4 Diffs Page (`ui/src/pages/DiffsPage.tsx`)

### Incoming Params

1. `run_id`
2. `path`
3. `line_new` / `line_old` (optional)
4. `hunk_id` (optional)

### Behavior

1. Preselect run by `run_id`.
2. Load file list for run.
3. Preselect file by exact `path`.
4. Load hunks and optionally highlight target line/hunk.

### Required Additions

1. URL parsing effect for run/file selection.
2. Optional line highlight state for DiffViewer.
3. Not-found callout for missing run/file.

## 8.5 Docs Page (`ui/src/pages/DocsPage.tsx`)

### Incoming Params

1. `term`
2. `path` (optional)
3. `line`, `col` (optional)
4. `run_id` (optional)

### Behavior

1. Select term row matching `term`.
2. Load hits via existing endpoint.
3. If `path+line` present, auto-scroll and highlight matching hit row.

### Required Additions

1. URL parsing effect.
2. Hit highlight state.
3. Optional run override support.

## 8.6 Files Page (`ui/src/pages/FilesPage.tsx`)

### Incoming Params

1. `path`
2. `line` (optional)

### Behavior

1. Ensure needed directory prefixes are expanded to reveal file.
2. Set selected file path.
3. Load file content and optionally scroll to line (future if CodeViewer supports line anchors).

### Required Additions

1. Prefix expansion helper from path segments.
2. Selection hydration from URL.
3. Optional line anchor state plumbing into CodeViewer.

## 9. UX and Interaction Details

## 9.1 Row Interaction States

Search result rows should support:

1. Default
2. Hover
3. Focus-visible
4. Active/pressed
5. Selected (keyboard navigation)

## 9.2 Action Affordances

Recommended row actions (right-aligned on hover/focus):

1. `Open`
2. `Copy link`
3. `Open in new tab`

For mobile or narrow width:

1. Keep row tap as primary open action.
2. Move secondary actions to kebab menu.

## 9.3 Empty and Error Messaging

Search page should differentiate:

1. No search term entered.
2. Query loading.
3. Query completed with no results.
4. Query failed (API error).
5. Drill-in target not found after navigation.

## 9.4 Accessibility Requirements

1. Result rows must be keyboard reachable.
2. Enter key must invoke open action.
3. Action buttons require `aria-label`.
4. Copy-link success/error should be announced via `aria-live` region (or toast).

## 10. Code Touch Matrix

This matrix lists expected modifications and how each file changes.

## 10.1 Frontend Routing and Search

1. `ui/src/pages/SearchPage.tsx`
- Add drill-in navigation handler.
- Pass open/copy actions to SearchResults.
- Preserve query/session context in outgoing links.

2. `ui/src/components/data-display/SearchResults.tsx`
- Add open action API and row action controls.
- Add keyboard open behavior.

3. `ui/src/types/api.ts`
- Optional tighten `payload` typing by discriminated union (recommended).

4. `ui/src/App.tsx`
- No route additions required for phase 1.
- Optional query-preserving back-link helpers.

## 10.2 Destination Pages

1. `ui/src/pages/SymbolsPage.tsx`
2. `ui/src/pages/CodeUnitsPage.tsx`
3. `ui/src/pages/CommitsPage.tsx`
4. `ui/src/pages/DiffsPage.tsx`
5. `ui/src/pages/DocsPage.tsx`
6. `ui/src/pages/FilesPage.tsx`

Each page adds:

1. URL param parser.
2. Initial hydration effect.
3. Not-found callout state.

## 10.3 Shared Utility Additions

New files (recommended):

1. `ui/src/features/search-drill-in/linkBuilder.ts`
2. `ui/src/features/search-drill-in/parseDrillInParams.ts`
3. `ui/src/features/search-drill-in/searchDrillInTypes.ts`
4. `ui/src/features/search-drill-in/index.ts`

## 10.4 API Layer

Likely no changes required for baseline implementation.

Optional future enhancements:

1. `pkg/workbenchapi/search.go` add stable `id` and normalized `location` object.
2. `ui/src/api/search.ts` adapt if backend contract expands.

## 11. Data and Type Design

## 11.1 Strongly Typed Result Payloads (Recommended)

Current `payload?: unknown` limits compile-time safety.

Recommended frontend type shape:

```ts
type SearchPayloadByType = {
  symbol: { symbol_hash: string; kind: string; pkg: string; file: string; line: number; col: number; run_id: number }
  code_unit: { unit_hash: string; kind: string; pkg: string; file: string; start_line: number; start_col: number; run_id: number }
  commit: { hash: string; subject?: string; author_name?: string; run_id: number }
  diff: { run_id: number; path: string; line_no_old?: number; line_no_new?: number; hunk_id: number; diff_file_id: number }
  doc: { run_id: number; term: string; path: string; line: number; col: number }
  file: { path: string; ext?: string; exists?: boolean; is_binary?: boolean }
}
```

Then define discriminated union using `type`.

## 11.2 Drill-In Intent Type

```ts
type DrillInIntent =
  | { type: 'symbol'; symbolHash: string; runId?: number }
  | { type: 'code_unit'; unitHash: string; runId?: number }
  | { type: 'commit'; commitHash: string; runId?: number }
  | { type: 'diff'; runId: number; path: string; lineNew?: number; lineOld?: number; hunkId?: number }
  | { type: 'doc'; term: string; path?: string; line?: number; col?: number; runId?: number }
  | { type: 'file'; path: string; line?: number }
```

This intent can be generated from search results and parsed from URL params.

## 12. Implementation Phases

## Phase 1: Core Drill-In and URL Hydration (MVP)

Deliverables:

1. Clickable search results that navigate correctly by type.
2. Destination pages hydrate and preselect target entity.
3. Copy-link support from search result rows.

Tasks:

1. Add link builder + param parser utility.
2. Wire SearchPage -> SearchResults open handler.
3. Add URL hydration to all six destination pages.
4. Add basic not-found callouts.

Exit criteria:

1. Manual QA passes all type mappings.
2. Reload preserves selected target via URL.

## Phase 2: UX Polish and Robustness

Deliverables:

1. Keyboard drill-in polish.
2. Open-in-new-tab controls.
3. Diff line/hunk highlight behavior.
4. Better copy-link feedback/toast.

Tasks:

1. Add row action controls.
2. Add focus and keyboard handling.
3. Add line/hunk highlight support in DiffsPage/DiffViewer.

Exit criteria:

1. Accessibility checks pass keyboard-only flow.
2. Links open equivalent context in new tab.

## Phase 3: Contract Hardening (Optional)

Deliverables:

1. Strongly typed payload union.
2. Optional backend `id/location` enrichment.
3. Runtime guardrails for malformed result payloads.

Tasks:

1. Type tightening in `ui/src/types/api.ts`.
2. Optional backend response extension.
3. Parser tests for invalid payload scenarios.

Exit criteria:

1. No `any`/`unknown` casts in drill-in path.
2. Contract tests validate backward compatibility.

## 13. Testing Strategy

## 13.1 Unit Tests

Target modules:

1. `linkBuilder.ts`
2. `parseDrillInParams.ts`

Coverage:

1. Per-type link generation.
2. Missing field handling.
3. Query/session propagation.
4. URL encoding for complex paths.

## 13.2 Component Tests

1. `SearchResults` row click triggers `onOpen` with expected result.
2. Row keyboard `Enter` opens selected result.
3. Copy-link action uses generated href.

## 13.3 Page Integration Tests

For each destination page:

1. Given URL params, page selects target entity.
2. Inspector opens when target resolved.
3. Not-found callout appears when resolution fails.

## 13.4 API Contract Tests

Current `pkg/workbenchapi/api_test.go` can be extended with:

1. Search result payload invariants per type.
2. Presence of required key fields (`symbol_hash`, `unit_hash`, etc.) in payload.

## 13.5 Manual QA Checklist

1. Search symbol -> open -> symbol inspector selected.
2. Search code unit -> open -> code unit inspector selected.
3. Search commit -> open -> commit inspector selected.
4. Search diff -> open -> run + file loaded.
5. Search doc -> open -> term selected + hit visible.
6. Search file -> open -> file content loaded.
7. Browser reload preserves target context for all above.
8. Copy link from result and open in fresh tab reproduces context.

## 14. Risk Analysis and Mitigation

## 14.1 URL Contract Drift Across Tickets

Risk:

- `REF-013` introduces wider URL-state changes that may conflict with REF-011 params.

Mitigation:

1. Keep REF-011 params minimal and namespaced by destination semantics.
2. Parse permissively and ignore unknown params.
3. Document param ownership in this guide.

## 14.2 Session/Run Inconsistency

Risk:

- Deep link references run not present in active session.

Mitigation:

1. Prefer explicit `run_id` from result when available.
2. Show clear not-found + suggested action (switch session).
3. Preserve params so retry works after session switch.

## 14.3 Large Data Resolution Delays

Risk:

- Destination list query may not include target due to pagination.

Mitigation:

1. Add direct detail fetch by unique key (`hash`/`unit_hash`/`commit_hash`).
2. Avoid dependency on list pagination for resolution.

## 14.4 Diff Drill-In Precision

Risk:

- Path-only diff opening may not highlight exact line/hunk.

Mitigation:

1. Carry `line_new`/`line_old`/`hunk_id` in URL.
2. Implement best-effort highlight in viewer.
3. Consider backend hunk endpoint in future phase.

## 15. Rollout and Backward Compatibility

## 15.1 Rollout Plan

1. Merge phase 1 behind default behavior (no feature flag required).
2. Keep existing non-drill-in functionality intact.
3. Add telemetry/logging only if already available in project conventions.

## 15.2 Backward Compatibility

1. Existing `/search?q=...` links remain valid.
2. Destination pages still work without drill-in params.
3. Unknown params are ignored.

## 16. Acceptance Criteria

A feature-complete REF-011 implementation must satisfy all:

1. Each search result type has a deterministic destination route.
2. Search result click navigates to destination with contextual params.
3. Destination resolves and selects target entity when data exists.
4. URL round-trips (copy/paste/reload) preserve context.
5. Not-found states are explicit and actionable.
6. Unit and integration tests cover link generation and hydration logic.

## 17. Implementation Checklist (Actionable)

## 17.1 Foundation

- [ ] Create `ui/src/features/search-drill-in/` with builder/parser/types.
- [ ] Add test coverage for builder/parser.

## 17.2 Search Surface

- [ ] Add `onOpen` behavior to `SearchResults`.
- [ ] Wire SearchPage row click to navigation.
- [ ] Add copy-link action.

## 17.3 Destination Hydration

- [ ] Symbols page hydration from `symbol_hash`.
- [ ] Code Units page hydration from `unit_hash`.
- [ ] Commits page hydration from `commit_hash`.
- [ ] Diffs page hydration from `run_id + path`.
- [ ] Docs page hydration from `term(+path/line)`.
- [ ] Files page hydration from `path(+line)`.

## 17.4 QA

- [ ] Manual QA pass on all six result types.
- [ ] Add/extend integration tests.
- [ ] Verify browser back/forward behavior.

## 18. Concrete "Code Touch" Summary

If executed exactly as planned, primary touched files will be:

1. `ui/src/pages/SearchPage.tsx`
2. `ui/src/components/data-display/SearchResults.tsx`
3. `ui/src/pages/SymbolsPage.tsx`
4. `ui/src/pages/CodeUnitsPage.tsx`
5. `ui/src/pages/CommitsPage.tsx`
6. `ui/src/pages/DiffsPage.tsx`
7. `ui/src/pages/DocsPage.tsx`
8. `ui/src/pages/FilesPage.tsx`
9. `ui/src/types/api.ts` (optional but recommended)
10. `ui/src/features/search-drill-in/linkBuilder.ts` (new)
11. `ui/src/features/search-drill-in/parseDrillInParams.ts` (new)
12. `ui/src/features/search-drill-in/searchDrillInTypes.ts` (new)
13. `ui/src/features/search-drill-in/index.ts` (new)
14. `pkg/workbenchapi/search.go` (optional enhancements only)

## 19. Suggested Review Order

For reviewers and implementers, use this order:

1. `design/01-search-drill-in-detailed-analysis-and-implementation-guide.md` (this document)
2. `ui/src/pages/SearchPage.tsx`
3. `ui/src/components/data-display/SearchResults.tsx`
4. Destination pages in this order: Symbols -> Code Units -> Commits -> Diffs -> Docs -> Files
5. Optional backend changes in `pkg/workbenchapi/search.go`

## 20. Final Notes

This ticket is the bridge between "global discovery" and "contextual investigation". The proposed plan intentionally avoids broad architecture churn while still establishing a stable deep-link contract that future tickets can expand.

After phase 1 lands, Workbench gains a meaningful workflow upgrade:

1. find the entity,
2. open it with context,
3. share that context reliably.

That is the minimum viable behavior required by the design spec and by day-to-day refactor investigation workflows.
