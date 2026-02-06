---
Title: UI Cleanup and API Alignment Engineering Postmortem
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
    - Path: pkg/workbenchapi/diffs.go
      Note: |-
        Final mismatch fix for session-scoped diff run lookup.
        Final runtime mismatch fix for override session diff-runs lookup
    - Path: pkg/workbenchapi/sessions.go
      Note: Canonical backend session computation and availability model.
    - Path: ttmp/2026/02/04/REF-007-INDEX-BROWSE-UI--index-browse-ui-backend-api/reference/01-diary.md
      Note: |-
        Foundational implementation narrative for backend API + UI scaffold.
        Primary source for foundational implementation chronology
    - Path: ttmp/2026/02/05/REF-008-API-CONTRACT-ALIGNMENT--ui-api-contract-alignment-clean-cutover/reference/01-diary.md
      Note: |-
        Primary alignment execution diary with step-by-step corrective changes.
        Primary source for alignment and bug-fix chronology
    - Path: ttmp/2026/02/05/REF-015-REMOVE-TREESITTER--remove-tree-sitter-surface-area-temporary/reference/01-diary.md
      Note: |-
        Tree-sitter removal rationale and execution details.
        Source for tree-sitter surface removal decisions
    - Path: ui/src/App.tsx
      Note: Session/workspace orchestration in the React shell.
    - Path: ui/src/hooks/useSessionContext.ts
      Note: Session-scoping abstraction for domain run IDs.
    - Path: ui/src/pages/SymbolsPage.tsx
      Note: Final duplicate-key fix in live validation pass
ExternalSources: []
Summary: Detailed technical postmortem of the UI cleanup and API-alignment program across REF-007, REF-008, and REF-015.
LastUpdated: 2026-02-05T19:14:22-05:00
WhatFor: Deep engineering retrospective and handoff document for the full UI/API stabilization effort.
WhenToUse: Use when onboarding contributors, planning follow-up tickets, or auditing architecture and correctness decisions made during cleanup.
---


# UI Cleanup and API Alignment Engineering Postmortem

## 1. Purpose and Reader Contract

This document is an engineering-grade postmortem for the UI cleanup and bug-fixing campaign executed across the most recent Refactorio UI tickets, centered on `REF-007-INDEX-BROWSE-UI`, `REF-008-API-CONTRACT-ALIGNMENT`, and `REF-015-REMOVE-TREESITTER`.

The target reader is a contributor who needs to understand not only *what changed* but *why it changed*, what failed before changes landed, how correctness was validated, and where to continue if future work is delegated. This is intentionally long-form and explicit. It is designed to substitute for tribal memory and fragmented chat history.

This analysis is written from the perspective of system behavior under real runtime conditions, not just compile-time correctness. It emphasizes architecture-level decisions, API contract decisions, state management consistency, and validation strategy under live data.

## 2. Scope

### Included workstreams

1. Backend API surface maturing for UI consumption (`REF-007`).
2. Frontend and backend contract alignment with clean cutover semantics (`REF-008`).
3. Storybook and MSW stabilization after session-scoped queries were introduced (`REF-008`).
4. Live backend bring-up and data-path stabilization with real SQLite workspace data (`REF-008`).
5. Session selection and session-scoped rendering correctness fixes (`REF-008`).
6. Tree-sitter contract/surface removal for temporary de-scope (`REF-015`).
7. TypeScript build-blocker cleanup and final runtime mismatch closure (`REF-008`).

### Excluded workstreams

1. Product feature implementation outside index-browse and contract alignment scope.
2. Visual redesign pass (explicitly out of scope per request; Storybook visuals were considered acceptable).
3. Deep historical cleanup of legacy ticket artifacts beyond affected active tickets.
4. Full schema removal of `ts_captures` from SQLite migration path (we removed runtime/API exposure, not historical schema support).

## 3. Source Material Used

Primary narrative and technical sources:

1. `ttmp/2026/02/04/REF-007-INDEX-BROWSE-UI--index-browse-ui-backend-api/reference/01-diary.md`
2. `ttmp/2026/02/05/REF-008-API-CONTRACT-ALIGNMENT--ui-api-contract-alignment-clean-cutover/reference/01-diary.md`
3. `ttmp/2026/02/05/REF-015-REMOVE-TREESITTER--remove-tree-sitter-surface-area-temporary/reference/01-diary.md`
4. `ttmp/2026/02/05/REF-008-API-CONTRACT-ALIGNMENT--ui-api-contract-alignment-clean-cutover/tasks.md`
5. `ttmp/2026/02/05/REF-008-API-CONTRACT-ALIGNMENT--ui-api-contract-alignment-clean-cutover/changelog.md`
6. `sources/ui-design.md`
7. Commits on branch `task/implement-refactorio-refactoring` through `35152de`

In practical terms, this postmortem synthesizes ticket logs, final code state, and live runtime validation outcomes (HTTP probes + Playwright traversal + compiler/test signals).

## 4. System Baseline Before Cleanup

### 4.1 Architectural shape (intended)

The intended architecture is straightforward but strict in contract boundaries:

1. `refactorio` Go backend provides REST endpoints under `/api` for workspace-centric and session-scoped data exploration.
2. React app provides a navigation shell and per-domain pages (runs, symbols, code units, commits, diffs, docs, files, search).
3. RTK Query handles API interactions and transforms responses into client-consumable data models.
4. Session object is the central context primitive for coherent cross-domain exploration.

### 4.2 Reality before stabilization

Before alignment and cleanup, the system was in a classic transition state:

1. UI partially implemented against provisional data shapes.
2. Backend evolved independently with stronger, more concrete envelopes and field naming.
3. Storybook mocks drifted from actual endpoint contracts.
4. Session selection UX was visibly present but behaviorally incomplete.
5. Runtime failures appeared under real data conditions despite compile-green subsets.

The work therefore had to solve *semantic drift* and *execution drift* simultaneously.

## 5. Key Design Decision: Backend as Source of Truth

A core decision was made and repeatedly reinforced: for API contract disagreements, backend endpoint behavior is canonical; frontend adapts without compatibility shims unless explicitly needed.

### Why this was defensible

1. Backend is the single producer for live data.
2. Session semantics are assembled server-side from run metadata and table presence.
3. The UI objective is exploratory fidelity, not a generic abstraction layer disconnected from backend behavior.
4. Backward-compatibility shims in UI would have hidden unresolved contract debt and delayed full convergence.

### Tradeoff accepted

This introduced a concentrated burst of UI changes and Storybook breakage. That was intentional and considered lower risk than prolonged dual-contract maintenance.

## 6. Timeline and Program Phases

## 6.1 Phase A: Foundational build (`REF-007`)

`REF-007` established backend browse APIs and initial UI scaffold/storybook component ecosystem. It delivered most route surfaces and page skeletons, but by nature of rapid expansion left residual integration gaps.

Representative commits from this era include scaffold and component accretion (`5d6521e`, `c23d60d`, `40f114a`, `cfba20b`, `fc50865`, `9ad80c9`, `bb8fe1e`, `516b03a`, `815906b`).

### Phase A outcome

High feature coverage, medium contract stability.

## 6.2 Phase B: Contract cutover and session scoping (`REF-008`)

`REF-008` performed the decisive cutover:

1. Endpoint envelopes and field names aligned to backend.
2. Page queries wired to session-derived domain run IDs.
3. Selection resets and empty-state semantics added for missing domain runs.
4. Storybook/MSW contract corrected.

Core commit: `ee1e0bc`.

### Phase B outcome

Contract coherence improved substantially, but runtime and toolchain gaps remained.

## 6.3 Phase C: Live runtime recovery and observability (`REF-008`)

Work then shifted from type-level correctness to live system operability:

1. Added zerolog request/error traces.
2. Migrated/refilled workspace DB to current schema.
3. Fixed `/api/files` and file search null-scan faults.

Core commit: `b44d71e` + follow-up live stabilization commit `be28060`.

### Phase C outcome

Live API became debuggable and usable with real data.

## 6.4 Phase D: Session UX correctness (`REF-008`)

After basic operability, session behavior correctness became priority:

1. Topbar selectors moved from placeholder to Redux-wired controls.
2. Domain tables stopped leaking stale rows across session switches.
3. Auto-selection favored highest-coverage session.

Core commits: `0525237`, `2476c0c`.

### Phase D outcome

Session behavior became user-trustworthy.

## 6.5 Phase E: Temporary tree-sitter de-scope (`REF-015`)

Because backend tree-sitter implementation was intentionally removed, runtime/API/UI contract still exposing tree-sitter was technical debt and source of confusion. `REF-015` removed this from active runtime surfaces.

Core commit: `5031d68` + ticket close `ebaa7c0`.

### Phase E outcome

Runtime contract simplified and de-risked.

## 6.6 Phase F: Final compile/runtime mismatch closure (`REF-008`)

Final pass eliminated residual TypeScript blockers and two runtime mismatches found only under live traversal:

1. TS build blocker cleanup (`f8bd48d`).
2. Session-scoped diff-runs override bug and symbol key collision fix (`3388bbb`).
3. Validation and diary closure step (`35152de`).

### Phase F outcome

Full `REF-008` task set reached completion with live confirmation.

## 7. Technical Deep Dive: Backend API Alignment

## 7.1 Session model as the architectural center

The backend session model in `pkg/workbenchapi/sessions.go` computes grouped run-context from `meta_runs` + domain table presence and emits:

1. `runs` map: domain -> run_id pointer.
2. `availability` map: domain -> bool.
3. metadata (`id`, `git_from`, `git_to`, `root_path`, `last_updated`).

This became the central orientation mechanism for the UI.

### Session computation pattern (conceptual)

```go
for run in meta_runs:
  key := grouping(run.root_path, run.git_from, run.git_to, run.id)
  session := getOrCreate(key)

  if hasData(diff_files, run.id): session.runs.diff = latest(run.id)
  if hasData(symbol_occurrences, run.id): session.runs.symbols = latest(run.id)
  if hasData(code_unit_snapshots, run.id): session.runs.code_units = latest(run.id)
  if hasData(doc_hits, run.id): session.runs.doc_hits = latest(run.id)
  if hasData(commits, run.id): session.runs.commits = latest(run.id)
  if hasData(symbol_refs or symbol_refs_unresolved, run.id): session.runs.gopls_refs = latest(run.id)

session.availability = boolMap(session.runs)
applyWorkspaceOverrides(session)
```

Key nuance: overrides from workspace config must be passed to `computeSessions` in every endpoint using session IDs, not only in `/sessions` route handlers.

## 7.2 Diff endpoint mismatch and final correction

### Symptom

Live UI (Diffs page) raised 404 for:

`GET /api/diff-runs?workspace_id=refactorio-foobar&session_id=refactorio-foobar:all-indexed`

### Root cause

`pkg/workbenchapi/diffs.go` used:

```go
sessions, err := computeSessions(db, ref, nil)
```

That ignored workspace overrides (including `all-indexed`) and thus session lookup failed.

### Fix

Load workspace config and pass `ws.Sessions` as overrides before `computeSessions`.

### Result

Endpoint now resolves override sessions and returns expected diff run, verified by HTTP 200 and Playwright page traversal.

## 7.3 File endpoint nullability hardening

A critical live blocker was `/api/files` failing on nullable columns (`file_exists`, `is_binary`) in `files` table.

### Design lesson

When SQLite schema allows nullability during transition, read-paths must use nullable scans (`sql.NullBool`/`sql.Null*`) and normalize outputs; strict scans on bool/int can transform data quality inconsistencies into user-visible 500s.

## 7.4 Request-level observability

Zerolog middleware introduced structured request traces including:

1. path
2. status
3. workspace_id
4. duration
5. bytes

This was instrumental in distinguishing backend logic bugs from stale-process confusion and from client mock issues.

## 8. Technical Deep Dive: Frontend State and Data Flow

## 8.1 Session-scoped query architecture

Session scoping was centralized in `ui/src/hooks/useSessionContext.ts` rather than duplicated in each page.

### Why this mattered

Without a shared hook, each page would independently derive run IDs and fallback behavior, increasing drift and stale-data risk.

### Pattern used

```ts
activeSession = sessions.find(id == activeSessionId) ?? sessions[0]
runIds = {
  symbols: activeSession?.runs.symbols,
  codeUnits: activeSession?.runs.code_units,
  commits: activeSession?.runs.commits,
  diff: activeSession?.runs.diff,
  docs: activeSession?.runs.doc_hits,
  goplsRefs: activeSession?.runs.gopls_refs,
}
```

Page queries then become `skip: !workspaceId || !runId` and rows are gated by domain availability.

## 8.2 Stale-data guard pattern

A recurring bug class: when session changes to one that lacks a run for a domain, previous query result remained visible.

### Fix pattern used across pages

```ts
const available = Boolean(runId)
const rows = available ? (data ?? []) : []

<EntityTable
  data={rows}
  emptyMessage={available ? "No items found" : "No <domain> data for this session"}
/>
```

This made state transitions explicit and prevented silent semantic leakage.

## 8.3 Selector wiring and user trust

Topbar controls initially existed as affordances but not full control planes. Session dropdown selection appeared interactive yet did not reliably drive app state.

The fix in `ui/src/App.tsx` + `ui/src/components/layout/Topbar.tsx` established controlled selectors mapped to Redux actions:

1. `setActiveWorkspace`
2. `setActiveSession`

This changed session selection from decorative to deterministic.

## 8.4 Default session policy

The default session was changed from naive first-entry selection to highest domain-coverage selection (tie-break by latest update). This reduced cognitive load and minimized initial empty states.

Pseudo-ranking:

```ts
score(session) = countTrue(session.availability)
best = sortBy(score DESC, last_updated DESC)[0]
```

## 8.5 Symbol table identity fix

### Symptom

React warnings for duplicate keys in Symbols page under live data.

### Root cause

`getItemId` used only `symbol_hash`, but a symbol can appear in multiple locations/occurrences.

### Fix

Composite row identity:

```ts
id = `${symbol_hash}:${file}:${line}:${col}:${run_id}`
```

This preserved stable row identity and eliminated duplicate-key warnings.

## 9. Storybook and MSW Alignment

Session scoping introduced implicit dependency on `/api/sessions` for many stories. Stories that previously worked with localized mocks started producing 404s after scoping logic was introduced.

### Resolution strategy

1. Add global baseline handlers in `.storybook/preview.ts`.
2. Normalize envelope shapes to backend (`{ items: ... }` etc).
3. Remove stale schema assumptions (for example deprecated stats forms in Dashboard stories).

### Engineering insight

Storybook is not just visual documentation in this architecture; it is a structural integration surface for client-side contract assumptions. It needs the same schema discipline as production code.

## 10. Runtime and Environment Hazards Encountered

## 10.1 Stale server binary confusion

A notable operational hazard occurred during validation: source no longer contained `tree_sitter` fields, but API responses still returned `tree_sitter`.

### Root cause

A stale server process continued listening on port `8080`, and at times conflicting process restarts obscured which binary was active.

### Detection method used

1. `lsof -iTCP:8080 -sTCP:LISTEN`
2. inspect PID and executable path (`/proc/<pid>/exe`)
3. inspect strings in running binary to verify stale symbols/docs
4. force-kill listener PID and relaunch from expected pane with explicit args

### Corrective practice recommended

Always verify active listener PID and binary path when runtime behavior disagrees with source.

## 10.2 Live data versus mock confidence gap

Several issues (session override mismatch in diff-runs, symbol duplicate keys) were invisible in static tests and build checks, and only surfaced under real data traversal.

This is why final REF-008 closure required explicit live playbook execution, not just unit tests and compile checks.

## 11. Tree-Sitter De-scope: Architectural Implications (`REF-015`)

Tree-sitter was intentionally removed on backend implementation side earlier, but runtime/API/UI contract still surfaced tree-sitter concepts.

## 11.1 What was removed from active surface

1. `/api/tree-sitter/captures` route registration and handler.
2. `tree_sitter` field from session run contract (`SessionRuns`).
3. `tree_sitter` availability emission from session computation.
4. `tree_sitter` feature advertisement from `/api/db/info` response.
5. UI `SessionRuns` type field and related session context plumbing.
6. Session card/story/mocks references.
7. API reference section documenting tree-sitter endpoint.

## 11.2 What intentionally remained

`ts_captures` schema definitions remain in `pkg/refactorindex/schema.go` for now. This was a deliberate “runtime surface removal, schema retention” compromise to avoid forced migration churn during cleanup.

## 11.3 Why this was the right interim cut

1. Removed active user-facing confusion and false affordances.
2. Reduced contract complexity immediately.
3. Preserved future reintroduction option without urgent data migration pressure.

## 12. TypeScript Build Blocker Cleanup

The TS cleanup pass resolved a mixed set of drift and lint-strictness issues.

### 12.1 Diff viewer type drift

Stories and component referenced legacy diff hunk fields (`hunk_id`, `old_count`, `new_count`) while canonical types used (`id`, `old_lines`, `new_lines`).

### 12.2 Generic cast strictness

EntityTable story sorting used unsafe cast shape; strict TS required explicit `unknown` intermediate.

### 12.3 Removed API type alias usage

`SessionAvailability` import mismatch fixed by direct usage of `Session['availability']`.

### 12.4 Run key mismatch

Dashboard used `run.run_id` while `Run` type uses `id`.

### Outcome

`npm --prefix ui run build` succeeded with bundled assets emitted to `pkg/workbenchapi/static/dist`.

## 13. API Contract Delta Summary (Before -> After)

## 13.1 Sessions

Before:

1. UI assumed inconsistent envelopes and partially optional naming conventions.
2. tree-sitter was still represented in availability/runs despite de-scoped backend behavior.

After:

1. canonical `SessionRuns` domains: `commits`, `diff`, `symbols`, `code_units`, `doc_hits`, `gopls_refs`.
2. UI queries consume run IDs from active session context.
3. missing domains produce explicit empty states, not stale rows.

## 13.2 DB info features

Before:

`features.tree_sitter` present and misleading relative to active feature policy.

After:

`features` reduced to active/meaningful flags (`fts`, `gopls_refs`, `doc_hits`).

## 13.3 Diff-runs with session_id

Before:

Session override IDs could not be resolved in `diff-runs` and returned 404.

After:

Override-aware session resolution returns diff run correctly.

## 13.4 Symbols table identity

Before:

Row key collisions where `symbol_hash` repeated across occurrences.

After:

Composite IDs eliminate duplicate key warnings and maintain stable table behavior.

## 14. Validation Matrix

| Layer | Command / Method | Result | Notes |
|---|---|---|---|
| Backend package tests | `GOWORK=off go test ./pkg/workbenchapi/...` | Pass | Repeated after final mismatch fixes |
| Frontend compile/build | `npm --prefix ui run build` | Pass | Includes `tsc -b` + Vite build |
| API health | `GET /api/health` backend and via Vite proxy | 200 | Confirms routing and proxy path |
| Sessions API | `GET /api/sessions?workspace_id=refactorio-foobar` | 200 | Includes override session `all-indexed` |
| Diff-runs override lookup | `GET /api/diff-runs?...session_id=...all-indexed` | 200 | Previously 404, fixed |
| Playwright route traversal | Dashboard/Runs/Symbols/Code Units/Commits/Diffs/Docs/Files/Search | Stable | 0 console errors after final fix |

## 15. Root-Cause Casebook (Issue -> Cause -> Corrective Action)

## 15.1 Storybook 404 on `/api/sessions`

1. Issue: stories broke after session scoping.
2. Cause: no global session handlers, envelope mismatch.
3. Action: add global MSW handlers and normalize response shapes.

## 15.2 Session dropdown looked interactive but did not control state

1. Issue: user selected sessions, behavior unchanged.
2. Cause: topbar had placeholder callbacks.
3. Action: controlled selectors + Redux dispatch wiring.

## 15.3 Stale rows after session switch

1. Issue: table continued showing prior session data.
2. Cause: no run-availability gating in table data binding.
3. Action: rows become empty when session lacks run ID.

## 15.4 `/api/sessions` 500 under live DB

1. Issue: session computation failed.
2. Cause: schema/data state not aligned with expected current tables/run metadata.
3. Action: migrate/init DB schema and refill run domains.

## 15.5 `/api/files` 500

1. Issue: file listing/search endpoint crashes.
2. Cause: nullable column scans treated as strict bools.
3. Action: nullable-safe scans and normalization.

## 15.6 `tree_sitter` lingering in runtime contract

1. Issue: contract advertised unavailable functionality.
2. Cause: partial de-scope left route/fields/docs exposed.
3. Action: REF-015 runtime surface removal.

## 15.7 TS compile drift in Diff viewer

1. Issue: build blocked by field mismatch.
2. Cause: story/component used legacy field names.
3. Action: align stories and component keys to current type.

## 15.8 Symbols duplicate key warnings

1. Issue: React warning and potential render instability.
2. Cause: non-unique key strategy (`symbol_hash` only).
3. Action: composite row identity.

## 15.9 Diff page 404 for override session

1. Issue: `diff-runs` rejected override session.
2. Cause: endpoint omitted session overrides during compute.
3. Action: pass workspace override set in `handleDiffRuns`.

## 15.10 Source/runtime mismatch confusion

1. Issue: API responses contradicted source code.
2. Cause: stale process bound to service port.
3. Action: PID-based listener verification and forced restart discipline.

## 16. Architecture Decisions and Their Consequences

## 16.1 Clean cutover over compatibility layering

### Decision

No backwards compatibility shims in UI contract layer unless mandated.

### Positive

1. Faster convergence.
2. Lower long-term maintenance complexity.
3. Reduced ambiguity in endpoint ownership.

### Negative

1. Temporary widespread breakage in stories/types.
2. Requires coordinated merge discipline.

## 16.2 Session as first-class scope primitive

### Decision

Session determines domain run IDs across pages.

### Positive

1. Coherent cross-page context.
2. Easier mental model for users exploring same index context.

### Negative

1. Any session endpoint inconsistency (for example overrides) propagates widely.
2. Demands consistent run-availability handling on all pages.

## 16.3 Keep schema support, remove runtime exposure (tree-sitter)

### Decision

Do not hard-drop `ts_captures` schema immediately; remove active runtime/UI surface.

### Positive

1. Fast de-risk of user-facing system.
2. Lower migration burden during stabilization window.

### Negative

1. Requires explicit documentation so retained schema is not misread as active feature.

## 17. API Reference Snapshot (Post-Cleanup)

### 17.1 Core endpoints used by UI

1. `GET /api/workspaces`
2. `GET /api/db/info?workspace_id=...`
3. `GET /api/runs?workspace_id=...`
4. `GET /api/sessions?workspace_id=...`
5. `GET /api/symbols?...`
6. `GET /api/symbols/{hash}/refs?...`
7. `GET /api/code-units?...`
8. `GET /api/commits?...`
9. `GET /api/diff-runs?...`
10. `GET /api/diff/{run_id}/files?...`
11. `GET /api/diff/{run_id}/file?path=...`
12. `GET /api/docs/terms?...`
13. `GET /api/docs/hits?...`
14. `GET /api/files?...`
15. `GET /api/search?...`

### 17.2 Non-goal surface now removed

1. `GET /api/tree-sitter/captures` (removed from active runtime surface)

## 18. Pseudocode Library for Future Contributors

## 18.1 Session-aware page query template

```ts
const { workspaceId, activeSession } = useSessionContext()
const runId = activeSession?.runs.<domain>
const available = Boolean(runId)

const query = useGetDomainQuery(
  { workspace_id: workspaceId!, run_id: runId, ...params },
  { skip: !workspaceId || !runId }
)

const rows = available ? (query.data ?? []) : []
```

## 18.2 Session selection reducer interaction

```ts
onWorkspaceSelect(workspaceId):
  dispatch(setActiveWorkspace(workspaceId))
  dispatch(clearActiveSessionOrRecomputeDefault())

onSessionSelect(sessionId):
  dispatch(setActiveSession(sessionId))
  resetLocalSelectionsAndOffsets()
```

## 18.3 Backend override-aware session lookup

```go
func resolveSession(db, workspaceRef, sessionID):
  overrides := loadOverridesIfWorkspaceConfigured(workspaceRef.id)
  sessions := computeSessions(db, workspaceRef, overrides)
  return findByID(sessions, sessionID)
```

## 18.4 React row identity strategy

```ts
function stableRowId(item): string {
  // prefer immutable unique composite, not just domain identifier
  return `${item.hash}:${item.file}:${item.line}:${item.col}:${item.run_id}`
}
```

## 19. What We Learned (Engineering Lessons)

## 19.1 Contract drift is a system problem, not a single-file problem

Misalignment propagated across types, slices, components, stories, and live APIs. The only durable approach was ticketed, cross-layer synchronization.

## 19.2 Compile-green is necessary but insufficient

TypeScript and tests did not reveal final diff-runs override mismatch or symbol key collisions. Live page traversal against real data was essential.

## 19.3 Session semantics require strict consistency

Any endpoint accepting `session_id` must apply the same session resolution strategy and include workspace overrides. Partial adoption causes user-visible inconsistencies.

## 19.4 Stale process risk is real during rapid iterations

When runtime output contradicts source, inspect listener PID and executable path before assuming source is wrong.

## 19.5 Storybook should be treated as contract test surface

When MSW fixtures drift, stories become false confidence generators. Storybook baselines must evolve with API schema changes.

## 20. Observed Outcome Metrics

The following are qualitative-quantitative outcomes from the cleanup program:

1. REF-008 tasks reached full completion (`29/29`).
2. Frontend build transitioned from failing with 12 TS errors to passing.
3. Live Playwright traversal across 9 primary UI routes reports 0 console errors post-final fixes.
4. Session-scoped diff run lookup moved from deterministic 404 for override session to deterministic 200 with expected run payload.
5. Runtime contract simplification completed for tree-sitter exposure (while retaining schema compatibility path).

## 21. Remaining Risks and Constraints

Even with cleanup complete, these risks remain and should be monitored:

1. `refactor-index ingest range` remains fragile due historical worktree compile compatibility.
2. `ts_captures` schema remains in DB model and may reappear in assumptions if documentation drifts.
3. Multi-process local development can regress into stale-binary confusion without strict process hygiene.
4. Several follow-up UX/product tickets (`REF-009` to `REF-014`) are scaffold-only and will require new task decomposition before execution.

## 22. Recommendations for REF-009 to REF-014

## 22.1 REF-009 Session Selection UX

Start by codifying session naming strategy (server-side stable naming vs client-derived labels), then add explicit selection persistence policy and shortcut flows.

## 22.2 REF-010 File Tree Lazy Loading

Current explorer data is workable for medium repos but should move to incremental loading and path-prefix pagination to avoid heavy initial payloads.

## 22.3 REF-011 Search Drill-In

Unify search result navigation semantics so each result type has deterministic open-target behavior and context breadcrumbs.

## 22.4 REF-012 UI Error Handling

Establish a consistent error-state component taxonomy (retryable, terminal, empty, unauthorized/path invalid) and standardize per-page behavior.

## 22.5 REF-013 State Persistence and Deep Links

Map the minimal state vector for useful restoration: workspace, session, route, selected entity IDs, filter/sort/search state.

## 22.6 REF-014 Spec Gap Features

Convert broad spec gap bucket into a concrete epic with independent sub-docs (plans, audits, admin) and acceptance criteria tied to `sources/ui-design.md` scenarios.

## 23. Handoff Playbook for New Engineers

A new engineer taking over this area should proceed in this order:

1. Read `REF-008` diary for step history and rationale.
2. Run backend + UI in tmux with explicit command lines and verify `/api/health`.
3. Execute `npm --prefix ui run build` and `GOWORK=off go test ./pkg/workbenchapi/...`.
4. Traverse all primary routes in browser/Playwright under live workspace.
5. Confirm session override behavior for `all-indexed` on Diffs page.
6. Confirm no duplicate key warnings on Symbols page.
7. Only then start feature tickets.

## 24. Engineering Appendix A: Commit Chronology (Relevant Slice)

1. `ee1e0bc` Align UI API contract and session scoping.
2. `c9b2846` Fix Storybook MSW handlers for session-scoped pages.
3. `b44d71e` Add zerolog request logging and Glazed CLI wiring.
4. `be28060` Stabilize live API flow after Glazed v1 upgrade.
5. `0525237` Fix topbar session switching and stale session data.
6. `2476c0c` Improve session default selection and card timestamp fallback.
7. `5031d68` Remove tree-sitter from workbench API and UI session surface.
8. `f8bd48d` Fix UI TypeScript build errors in stories and dashboard.
9. `3388bbb` Fix session-scoped diff-runs and duplicate symbol row keys.
10. `35152de` Record REF-008 live validation and mismatch closure.

## 25. Engineering Appendix B: Files Most Critical to Correctness

Backend:

1. `pkg/workbenchapi/sessions.go`
2. `pkg/workbenchapi/diffs.go`
3. `pkg/workbenchapi/files.go`
4. `pkg/workbenchapi/search.go`
5. `pkg/workbenchapi/db_info.go`

Frontend:

1. `ui/src/App.tsx`
2. `ui/src/hooks/useSessionContext.ts`
3. `ui/src/pages/SymbolsPage.tsx`
4. `ui/src/pages/DiffsPage.tsx`
5. `ui/src/components/layout/Topbar.tsx`

Documentation control plane:

1. `ttmp/.../REF-008.../tasks.md`
2. `ttmp/.../REF-008.../changelog.md`
3. `ttmp/.../REF-008.../reference/01-diary.md`
4. `ttmp/.../REF-015.../reference/01-diary.md`

## 26. Engineering Appendix C: Representative API Payloads

### 26.1 Sessions (post-cleanup)

```json
{
  "items": [
    {
      "id": "refactorio-foobar:all-indexed",
      "workspace_id": "refactorio-foobar",
      "git_from": "ALL",
      "git_to": "INDEXED",
      "runs": {
        "commits": 3,
        "diff": 4,
        "symbols": 5,
        "code_units": 6,
        "doc_hits": 7,
        "gopls_refs": 8
      },
      "availability": {
        "commits": true,
        "diff": true,
        "symbols": true,
        "code_units": true,
        "doc_hits": true,
        "gopls_refs": true
      }
    }
  ]
}
```

### 26.2 DB info features (post-cleanup)

```json
{
  "features": {
    "fts": true,
    "gopls_refs": true,
    "doc_hits": true
  }
}
```

### 26.3 Diff-runs by session (post-fix)

```json
{
  "items": [
    {
      "id": 4,
      "status": "success",
      "git_from": "HEAD~30",
      "git_to": "HEAD"
    }
  ]
}
```

## 27. Engineering Appendix D: Incident Log (Condensed)

Incident 1: Storybook page 404s after session scoping

1. Trigger: page stories loading without `/api/sessions` handler.
2. Impact: false-negative UI state in component workbench.
3. Resolution: global MSW baseline + envelope normalization.

Incident 2: `/api/sessions` 500 in live run

1. Trigger: schema/data mismatch in workspace DB context.
2. Impact: session-centric UI unusable.
3. Resolution: schema init/migration and domain re-ingestion.

Incident 3: `/api/files` 500

1. Trigger: nullable bool scanning.
2. Impact: Files page + file search unusable.
3. Resolution: nullable-safe handler logic.

Incident 4: session dropdown appears broken

1. Trigger: topbar placeholder callbacks.
2. Impact: user confusion and blocked exploration flows.
3. Resolution: controlled selectors and Redux wiring.

Incident 5: stale rows across session switch

1. Trigger: ungated table data rendering.
2. Impact: misrepresentation of active session data.
3. Resolution: run availability gating and selection resets.

Incident 6: residual tree-sitter surface

1. Trigger: incomplete de-scope.
2. Impact: misleading capability exposure.
3. Resolution: REF-015 runtime/API/UI removal.

Incident 7: TS build blockers

1. Trigger: contract/type drift in stories and page code.
2. Impact: build pipeline red.
3. Resolution: type alignment and strictness fixes.

Incident 8: diff-runs override 404

1. Trigger: `computeSessions` invoked without overrides.
2. Impact: Diffs page error for aggregate session.
3. Resolution: override-aware resolution path.

Incident 9: duplicate symbol keys

1. Trigger: non-unique row ID strategy.
2. Impact: React warnings and potential render anomalies.
3. Resolution: composite row key strategy.

Incident 10: stale backend process confusion

1. Trigger: old binary still bound to service port.
2. Impact: false interpretation of code correctness.
3. Resolution: PID-level listener ownership checks and forced restart.

## 28. Engineering Appendix E: Repro + Validation Command Set

```bash
# 1) Backend tests
GOWORK=off go test ./pkg/workbenchapi/...

# 2) Frontend build
npm --prefix ui run build

# 3) Core health checks
curl -sS http://127.0.0.1:8080/api/health
curl -sS http://127.0.0.1:3001/api/health

# 4) Session coverage
curl -sS 'http://127.0.0.1:8080/api/sessions?workspace_id=refactorio-foobar' | jq

# 5) Diff-runs override validation
curl -sS 'http://127.0.0.1:8080/api/diff-runs?workspace_id=refactorio-foobar&session_id=refactorio-foobar%3Aall-indexed' | jq

# 6) Process ownership sanity check
lsof -iTCP:8080 -sTCP:LISTEN -n -P
```

## 29. Final Assessment

The UI cleanup and API alignment effort should be considered successful by engineering criteria:

1. Contract coherence: achieved.
2. Live route operability across primary pages: achieved.
3. Build and test baseline: achieved.
4. Session-centric semantics reliability: achieved.
5. De-scoping of unsupported tree-sitter runtime surface: achieved.
6. Documentation and handoff traceability: achieved.

The remaining work is no longer “stabilization debt” within REF-008. It is now feature-forward work captured in follow-up tickets (`REF-009` to `REF-014`), which require decomposition and execution but are not blocked by unresolved core alignment problems.

## 30. Suggested Closure Statement for REF-008

`REF-008` can be closed with the following engineering summary:

"Frontend and backend contracts have been fully aligned for the current supported domains. Session-scoped page behavior is now deterministic under live data. TypeScript and backend package test pipelines are green. Final runtime mismatches discovered during live traversal were fixed and validated. The ticket has shifted from remediation to completion."


## 31. Deep Domain Analysis: Symbols Slice

The Symbols vertical is useful as a representative pattern because it touches list rendering, detail rendering, optional cross-domain linking (gopls refs), pagination behavior, and table identity semantics. The final implementation in `ui/src/pages/SymbolsPage.tsx` demonstrates the stabilized approach adopted elsewhere.

### 31.1 Data contract shape

Backend list response is consumed as a list of symbol occurrences with fields including `symbol_hash`, `name`, `kind`, `pkg`, `file`, `line`, `col`, `run_id`, and additional metadata used in detail panes. The important contract choice is that list identity is occurrence-level, not definition-level.

In earlier UI assumptions, `symbol_hash` was treated as uniquely identifying a row. In real data this is not true because multiple occurrences can exist for a single symbol hash. That mistake produced duplicate row-key warnings and could have caused unstable row reconciliation in React.

### 31.2 Query flow and session binding

The current flow:

1. Fetch active session and derive `symbolsRunId`.
2. Fetch symbol list scoped by run.
3. For selected row, optionally fetch refs using `gopls_refs` run fallbacking to `symbols` run.
4. Reset selection on session change.

Pseudocode:

```ts
const run = activeSession?.runs.symbols
const refsRun = activeSession?.runs.gopls_refs ?? run

const symbols = useGetSymbolsQuery({ workspace_id, run_id: run, ...filters }, { skip: !run })
const refs = useGetSymbolRefsQuery({ hash, workspace_id, run_id: refsRun }, { skip: !hash || !refsRun })

useEffect(() => {
  setSelectedSymbol(null)
  setOffset(0)
}, [sessionId])
```

### 31.3 Selection identity correction

The correction landed in `3388bbb`:

```ts
function symbolRowID(s: Symbol): string {
  return `${s.symbol_hash}:${s.file}:${s.line}:${s.col}:${s.run_id}`
}
```

This solved the immediate warning and codified the conceptual model that rows are occurrence rows, not abstract symbol-definition rows.

### 31.4 What this teaches

If an entity table is driven by observational or denormalized rows (occurrences, events, snapshots), key strategy must reflect row-level uniqueness, not conceptual entity uniqueness. This is especially important for paged tables where deterministic row behavior under filtering and sorting is required.

## 32. Deep Domain Analysis: Code Units Slice

Code Units is structurally similar to Symbols but with a different detail payload pattern and potential heavy data when body text is included.

### 32.1 Core contract

UI expects `code_unit` list rows and detail endpoint to include textual body (`body_text`) and optional docs. Contract alignment here required field-name synchronization and pagination semantics that match backend limits.

### 32.2 Session scoping and stale data semantics

The key correction in cleanup was not just binding by `run_id`; it was ensuring pages render a domain-specific “no data for this session” state when `run_id` is absent. Without this guard, old results persist and create a subtle but serious context integrity bug.

### 32.3 Correctness pattern

```ts
const run = activeSession?.runs.code_units
const available = Boolean(run)
const { data, isLoading } = useGetCodeUnitsQuery({ workspace_id, run_id: run, limit, offset }, { skip: !run })
const rows = available ? (data ?? []) : []
```

### 32.4 Risk note

Code unit detail payloads can be large depending on body text. Future optimization may require deferred detail fetching only on row selection (if not already done everywhere) and cache invalidation strategy when switching sessions rapidly.

## 33. Deep Domain Analysis: Commits Slice

Commits is one of the most user-visible domains because it ties code-history context to many user workflows. The cleanup around commits focused on run ID alignment and stale state behavior.

### 33.1 Contract and shape

Backend emits commit list with stable fields (`hash`, `subject`, `author_name`, `committer_date`, etc.) and commit-file details in dedicated endpoint. The UI side had to stop assuming older naming variants.

### 33.2 Session behavior

When a session lacks `runs.commits`, commits page must not silently show previous session list. This bug was present before stale-data guards were added.

### 33.3 Resulting behavior

1. Session with commits run -> commit list renders and detail requests are valid.
2. Session without commits run -> explicit empty-state message.
3. Switching between sessions updates list deterministically.

### 33.4 Engineering implication

This was one of the clearest demonstrations that “query skip conditions” alone are not enough. UI display logic must explicitly encode domain availability semantics.

## 34. Deep Domain Analysis: Diffs Slice

Diffs had the highest late-stage risk because it combined:

1. session-scoped run selection,
2. nested endpoint chain (`/diff-runs` -> `/diff/{run}/files` -> `/diff/{run}/file`),
3. local viewer rendering assumptions.

### 34.1 Runtime mismatch discovered late

The issue:

`GET /api/diff-runs?workspace_id=...&session_id=refactorio-foobar:all-indexed` returned 404.

This was not a simple frontend bug; it was a backend session resolution inconsistency.

### 34.2 Root-cause detail

`handleDiffRuns` used `computeSessions(db, ref, nil)` instead of loading workspace overrides first. Since `all-indexed` was an override session (stored in workspace config), it was invisible to this endpoint.

### 34.3 Corrective patch pattern

```go
overrides := []SessionOverride{}
if ref.ID != "" {
  cfg, _ := s.loadWorkspaceConfig()
  ws, _, ok := cfg.FindWorkspace(ref.ID)
  if ok {
    overrides = ws.Sessions
  }
}
sessions, err := computeSessions(db, ref, overrides)
```

### 34.4 Why this mattered beyond Diffs

The issue exposed a broader rule: every endpoint that accepts `session_id` must delegate to the same session-resolution policy path or equivalent helper. Inconsistent policy creates domain-specific correctness anomalies that are hard to debug from UI alone.

## 35. Deep Domain Analysis: Docs/Terms Slice

Docs/Terms was less bug-prone but still affected by session scoping and contract drift.

### 35.1 Semantics

Docs terms and hits are derived from `doc_hits` domain, and therefore should only be queried with `runs.doc_hits` from active session.

### 35.2 Stabilization behavior

As with other domains, stale docs terms from prior session were cleared when session changed to one without doc_hits run. This behavior was essential because terms appear globally meaningful even when they are run-scoped and session-specific.

### 35.3 Future extension note

If search drill-in (`REF-011`) deepens docs integration, consider preserving explicit “run context” badges near terms/hits to avoid user confusion when comparing sessions.

## 36. Deep Domain Analysis: Files Slice

Files is special because it is less directly session-scoped in user mental model but still impacted by backend robustness and workspace selection.

### 36.1 Critical production issue

`/api/files` and file search endpoints threw 500 due to nullable column scan mismatches (`file_exists`, `is_binary`).

### 36.2 Why this was high impact

Files Explorer is one of the top-level navigation anchors. When broken, user confidence in whole app degrades even if other domains function.

### 36.3 Corrective theme

Read-path resilience for nullable DB values is non-negotiable in evolving schema environments.

## 37. Deep Domain Analysis: Search Slice

Search alignment required contract normalization for request payload and result shape.

### 37.1 Core model

Search request supports run scoping via `run_ids` map and optional `session_id` context. Proper search behavior in aligned system depends on session run map assembly from `useSessionContext`.

### 37.2 Risk from mismatch

If run IDs are misnamed (`diffs` vs `diff`, etc.) or omitted, search silently produces incomplete results. This is harder to detect than hard 500 errors.

### 37.3 Recommendation

For future search work, add explicit diagnostics in dev mode indicating which run IDs were included in current query payload to reduce invisible mismatch risk.

## 38. RTK Query Slice Strategy Post-Alignment

The slice strategy matured into a consistent pattern:

1. Each domain has typed endpoints in `ui/src/api/*`.
2. Transform functions normalize backend envelopes early.
3. Pages only handle display-state concerns, not contract translation.

### 38.1 Advantages observed

1. Clearer ownership boundaries.
2. Easier contract audits.
3. Reduced ad-hoc mapping logic in components.

### 38.2 Remaining caution

Transform functions must stay synchronized with backend. A small envelope drift can invalidate multiple pages silently if transform assumptions become stale.

## 39. Component Architecture Notes

## 39.1 App shell and topbar

`ui/src/components/layout/AppShell.tsx` and `ui/src/components/layout/Topbar.tsx` now form a stable compositional shell where workspace/session selectors are first-class control inputs.

### 39.2 Selection and detail panels

Selection components (`SessionCard`, table selection, inspector panel patterns) became significantly more reliable after session reset behavior was standardized.

### 39.3 Storybook value

Component stories served dual purpose:

1. visual contract,
2. schema validation proxy.

The cleanup converted Storybook from source of false alarms (404s due missing handlers) into a more truthful fixture environment.

## 40. Backend Command and Logging Evolution

The migration to glazed command wiring and zerolog in `cmd/refactorio/*` plus `pkg/workbenchapi/logging.go` was foundational for operational debugging.

### 40.1 Why logs changed the debugging profile

Without structured request logs, 500s and missing session behavior were attributed to frontend too often. With logs, path/status/workspace correlation provided immediate narrowing.

### 40.2 Pattern introduced

1. root logging options via glazed.
2. persistent logger initialization.
3. per-request middleware logs.
4. explicit error logs at failure boundaries (for example session computation).

### 40.3 Operational guidance

When re-running validation playbooks, start backend with explicit `--log-level debug` and keep logs visible in tmux pane.

## 41. Environment and Process Discipline

Several hours of troubleshooting were consumed by stale process ambiguity. This warrants formal discipline in the postmortem.

### 41.1 Process discipline runbook

1. Identify listener PID before assuming backend code is active.
2. Kill listener PID explicitly if behavior mismatches source.
3. Relaunch from known working directory with explicit command arguments.
4. Reprobe known endpoints immediately (`/api/health`, `/api/db/info`).
5. Only then continue UI validation.

### 41.2 Why this matters in multi-tmux workflows

When one pane runs old command flags and another pane issues restarts, race conditions in manual process lifecycle are easy to create.

## 42. Data and DB Considerations

The cleanup relied on a concrete workspace DB (`foobar.db`) whose state had to be brought forward.

### 42.1 Data consistency requirements observed

1. Session computation assumes `meta_runs` and per-domain run-scoped table content.
2. File/read paths assume nullable-safe scans.
3. Session availability semantics assume truthful `tableHasRunData` behavior.

### 42.2 Migration/re-ingest practicality

Re-ingestion commands for commits, diff, symbols, code units, docs, and gopls were used to refill meaningful live data quickly. This provided a realistic validation bed without building a new synthetic environment.

## 43. API-by-API Failure Mode Catalog

### 43.1 `/api/sessions`

Failure mode:

1. `session_error` when schema/runtime mismatch.

Mitigation:

1. schema update and run data refill.
2. explicit error logging around session compute.

### 43.2 `/api/files`

Failure mode:

1. 500 scan errors on nullable bool columns.

Mitigation:

1. nullable-safe scan and normalization.

### 43.3 `/api/diff-runs` with `session_id`

Failure mode:

1. false 404 for override sessions.

Mitigation:

1. override-aware session compute path.

### 43.4 Client-side table renders

Failure mode:

1. stale rows on session switch.
2. duplicate keys with observational rows.

Mitigation:

1. availability-gated rows.
2. composite row IDs.

## 44. Documentation Operations Lessons

The use of `docmgr` and structured diaries created a high-signal change record. A few meta-lessons are worth preserving:

1. diary steps are most useful when tied to commit hashes.
2. task checkoffs should map to verifiable command output.
3. changelog entries should include both behavior and file intent.
4. ticket closure should happen only after runtime validation, not simply after code merge.

## 45. Engineering Decision Register (Compact)

| Decision | Context | Alternative Rejected | Reason |
|---|---|---|---|
| Backend as contract source | UI/backend drift | Dual compatibility in client | Would prolong ambiguity and debt |
| Session-scoped run IDs per page | multi-domain coherence | global unscoped query defaults | Violates contextual exploration model |
| Stale-data guard rows = [] | session switch behavior | show previous cached data | Misrepresents active session state |
| Remove active tree-sitter surface | backend impl removed | leave optional references | Confusing and error-prone contract |
| Composite symbol row keys | duplicate occurrence rows | key by symbol_hash | not unique in real data |
| Override-aware session lookup in diff-runs | aggregate session support | endpoint-local session map without overrides | fails for configured synthetic sessions |

## 46. Engineering Appendix F: Detailed Endpoint Contracts (Post-Cleanup)

### 46.1 Workspaces

`GET /api/workspaces`

Response envelope:

```json
{
  "items": [
    {
      "id": "refactorio-foobar",
      "name": "refactorio",
      "db_path": ".../foobar.db",
      "repo_root": ".../refactorio",
      "sessions": [
        {
          "id": "refactorio-foobar:all-indexed",
          "git_from": "ALL",
          "git_to": "INDEXED",
          "runs": {
            "commits": 3,
            "diff": 4,
            "symbols": 5,
            "code_units": 6,
            "doc_hits": 7,
            "gopls_refs": 8
          }
        }
      ]
    }
  ]
}
```

### 46.2 DB info

`GET /api/db/info?workspace_id=refactorio-foobar`

Critical post-cleanup expectations:

1. `features` contains active domains only.
2. no `tree_sitter` key.

### 46.3 Sessions

`GET /api/sessions?workspace_id=refactorio-foobar`

Critical expectations:

1. includes override session `all-indexed`.
2. availability map matches runtime supported domains.

### 46.4 Diffs

`GET /api/diff-runs?workspace_id=...&session_id=...`

Critical expectation:

1. override session IDs resolve to run list when configured.

### 46.5 Symbols

`GET /api/symbols?workspace_id=...&run_id=...`

Client rendering expectation:

1. row IDs must be unique per occurrence.

## 47. Engineering Appendix G: Frontend Build Error Retrospective

The set of TS build blockers resolved in `f8bd48d` were not random. They trace to two categories:

1. **Type drift from evolving API models**
   - Diff hunk field names changed but story/component references did not.
   - Dashboard run key assumption drifted from `run_id` to `id`.

2. **Strictness and cleanup debt**
   - Unused props/vars under strict compiler options.
   - Type import alias removed but component import not updated.

This pattern argues for a standing practice: each contract-alignment milestone should run full `tsc -b` and story compile checks, not only targeted page tests.

## 48. Engineering Appendix H: Live Validation Narrative (Detailed)

### Pass 1

1. Backend and UI reachable.
2. Page traversal mostly worked.
3. Two issues surfaced:
   - 404 on diff-runs with override session.
   - duplicate key warnings on Symbols.

### Interruption

Conflicting observations in `/api/db/info` (tree_sitter still present) indicated stale process.

### Process correction

1. Inspect listener PID.
2. Kill stale listener.
3. Relaunch expected command.
4. Reprobe endpoint.

### Pass 2

1. `features` no longer included tree_sitter.
2. Diffs page loaded run/files under aggregate session.
3. Symbols warning eliminated.
4. Route sweep produced 0 console errors.

## 49. Engineering Appendix I: Operational Anti-Patterns to Avoid

1. Running validation against unknown listener process.
2. Assuming MSW/story fixtures are representative without envelope audits.
3. Using conceptual IDs for table keys when data model is occurrence/snapshot oriented.
4. Treating endpoint-level session logic as one-off per route.
5. Closing tickets on compile-green without live traversal in session-scoped apps.

## 50. Engineering Appendix J: Proposed Next-Level Hardening

This work completed cleanup goals, but if we want further reliability hardening, these are high-value follow-ons:

1. Introduce shared backend helper for session resolution with overrides, and migrate endpoints to it.
2. Add integration test for `diff-runs` with override session ID.
3. Add UI runtime check in dev mode for duplicate row IDs in critical tables.
4. Add a CI smoke route-traversal step using Playwright against seeded local DB.
5. Create a `make ref008-validate`-style command encapsulating the key probe/build/test sequence.

## 51. Extended Closure Recommendation

Given all tasks complete and runtime validated under real data, `REF-008` closure is technically justified. Suggested close entry:

"Closed after full live UI/backend playbook completion. Session-scoped API/UI contract is aligned, runtime mismatches resolved, build/test pipelines green, and unsupported tree-sitter runtime surface removed. Follow-up work now belongs to feature tickets REF-009 through REF-014."

## 52. Final Reflection

The principal accomplishment here is not just bug count reduction. It is that the system moved from a *partially coherent prototype* to a *coherent operational baseline*.

A coherent baseline has concrete properties:

1. deterministic behavior under session changes,
2. deterministic contract interpretation across backend and frontend,
3. deterministic operator debugging path with structured logs,
4. deterministic documentation trace for handoff.

That state has now been achieved for the supported scope.


## 53. Day-by-Day Chronology (Engineering Narrative)

### Day 0: Foundation already in motion (context from REF-007)

The system entered this cleanup period with significant feature surface already built. The API route landscape was broad, and UI composition was mature enough to navigate most core pages. However, this did not mean readiness for production-like validation.

At this point, engineering confidence was mostly structural (“the routes and components exist”) rather than semantic (“the routes and components agree on what fields mean and when data is valid”). The cleanup effort should be viewed as the transition from structural confidence to semantic confidence.

### Day 1: Contract-first alignment and session scoping push

The first major push in `REF-008` aligned front-end contracts to backend contracts by replacing assumptions instead of layering compatibility code. This was intentionally disruptive but decisive.

Session scoping then became the center of gravity. Before this work, pages were often querying without strict alignment to active session domains. After this phase, the session context (`runs` mapping) became mandatory input for domain queries.

This introduced a new class of visible defects (particularly in Storybook and empty-state rendering), but those were expected and desirable because they made hidden inconsistencies explicit.

### Day 2: Storybook/MSW and developer-loop stabilization

Session-scoped fetching created missing-handler 404s in Storybook, especially for pages that now required `/api/sessions` regardless of local story parameters.

This day was less about runtime user behavior and more about developer feedback loop correctness. Global MSW defaults were introduced, and per-story fixtures were corrected to backend envelope shapes. The result was a meaningful local simulation environment.

### Day 3: Live backend bring-up, schema reconciliation, and logging

A key transition occurred when the work moved from mocks to live DB-backed execution. This immediately exposed server-side faults, including session computation failures and files endpoint scan errors.

Structured logging (`zerolog`) and command wiring updates (glazed) were critical enablers. They did not fix behavior directly, but they reduced diagnosis time by making requests and error boundaries observable.

### Day 4: Session UX credibility and stale data elimination

At this point, users could navigate but still could not trust session switching behavior fully. The topbar controls existed but had incomplete behavior, and stale rows persisted across session changes.

Controlled selectors and data gating were implemented. This changed perceived quality significantly because the UI stopped showing contradictory state after session changes.

### Day 5: Tree-sitter surface de-scope and contract simplification

The team decision to de-scope tree-sitter was already made at implementation level, but active runtime surface still advertised tree-sitter concepts.

This day reconciled policy and contract: remove tree-sitter from active API/UI surfaces while preserving schema compatibility for future reintroduction. This reduced conceptual noise and removed false capability hints.

### Day 6: TypeScript debt cleanup and full build green

Residual compile errors were then addressed. These were largely contract-drift artifacts (legacy field names, stale imports, wrong run key assumptions) mixed with strictness cleanup.

With those fixed, full UI build finally became green. This was necessary but still not sufficient for ticket closure.

### Day 7: Full live traversal and last two runtime mismatches

Final live playbook uncovered two subtle runtime issues not caught by compile/tests:

1. `diff-runs` session override lookup inconsistency.
2. duplicate key warnings in symbol occurrence table.

These were corrected immediately, then revalidated under live traversal. At this point, ticket tasks were truly complete by behavior, not only by code diff.

## 54. File-by-File Walkthrough: Backend Paths

This section explains backend files in terms of cleanup relevance and architectural role.

### 54.1 `pkg/workbenchapi/sessions.go`

This file is the semantic kernel for session-scoped UI behavior. It computes session groupings, run availability, and metadata from `meta_runs` and domain tables.

Cleanup relevance:

1. Session availability became authoritative for frontend gating.
2. Domain list was narrowed post-tree-sitter de-scope.
3. Override handling in session routes remained the model for other endpoints.

Design takeaway:

Any backend endpoint that accepts `session_id` should either call into this same logic or use a helper built on top of it.

### 54.2 `pkg/workbenchapi/diffs.go`

This file originally diverged from session-route behavior by not passing session overrides into compute path for `session_id` lookups.

Cleanup relevance:

1. Final runtime mismatch fix landed here.
2. Validated via direct HTTP and Playwright Diffs page.

Design takeaway:

Avoid endpoint-local re-derivation of session behavior when a canonical route already established override semantics.

### 54.3 `pkg/workbenchapi/db_info.go`

This file exposes DB capabilities and feature flags that frontends and operators use for capability assumptions.

Cleanup relevance:

1. Tree-sitter feature advertisement removed from active feature map.
2. Feature list now matches active runtime support.

Design takeaway:

Feature maps are operationally sensitive. Inconsistent feature flags create the same trust issues as broken endpoints.

### 54.4 `pkg/workbenchapi/files.go` and `pkg/workbenchapi/search.go`

These files handled nullable `files` columns more robustly after live failures.

Cleanup relevance:

1. Eliminated 500s in Files page and file-search path.
2. Demonstrated necessity of nullable-safe scan behavior in real datasets.

Design takeaway:

When schema evolution allows nullable transitional values, read paths must not assume strict non-null invariants unless migration guarantees are absolute.

### 54.5 `pkg/workbenchapi/routes.go`

This file encodes active API surface area.

Cleanup relevance:

1. Tree-sitter route registration removed.
2. Active route graph now aligns with current product scope.

Design takeaway:

Route registration is policy. If route graph and product policy diverge, clients will eventually encode wrong assumptions.

## 55. File-by-File Walkthrough: Frontend Paths

### 55.1 `ui/src/App.tsx`

Role:

1. top-level orchestration of workspace and session state.
2. selector options assembly.
3. default session selection behavior.

Cleanup relevance:

1. Default session moved to highest-coverage strategy.
2. Selector wiring became authoritative.

Design takeaway:

High-level orchestration should minimize hidden heuristics and make default selection logic explicit and testable.

### 55.2 `ui/src/hooks/useSessionContext.ts`

Role:

1. central derivation of active session and run IDs.

Cleanup relevance:

1. unified run-ID map for domains.
2. removed tree-sitter field post-de-scope.

Design takeaway:

Shared hook pattern prevents query-scoping drift across pages.

### 55.3 `ui/src/pages/SymbolsPage.tsx`

Role:

1. list + detail + refs behavior for symbol exploration.

Cleanup relevance:

1. stale-data guard.
2. composite row key strategy.

Design takeaway:

Data shape and row identity assumptions must be informed by real data cardinality, not by idealized domain models.

### 55.4 `ui/src/pages/DiffsPage.tsx`

Role:

1. run selection, file list, hunk view.

Cleanup relevance:

1. became the surface that exposed the `diff-runs` override bug.

Design takeaway:

Cross-layer bugs often present as UI route failures; route-level reproduction is a useful detector for backend policy mismatches.

### 55.5 `ui/src/components/layout/Topbar.tsx`

Role:

1. workspace/session selectors.
2. global search entry.

Cleanup relevance:

1. placeholder UX replaced with controlled selectors.

Design takeaway:

Selectors should never be rendered as interactive if they are not state-authoritative. Faux controls erode trust quickly.

### 55.6 `ui/src/components/code-display/DiffViewer.tsx` and `.stories.tsx`

Role:

1. diff hunk rendering with optional highlight/click.

Cleanup relevance:

1. schema field names aligned (`id`, `old_lines`, `new_lines`).
2. build blockers removed.

Design takeaway:

Story fixtures are often the first place where schema drift becomes visible. Keep fixtures coupled to canonical types, not legacy API memory.

## 56. Operational Validation Artifacts

This cleanup repeatedly validated at three layers:

1. compile/test layer,
2. endpoint probe layer,
3. full-route runtime layer.

### 56.1 Compile/test layer

Commands:

```bash
GOWORK=off go test ./pkg/workbenchapi/...
npm --prefix ui run build
```

These validated syntax, type consistency, and package-level behavior.

### 56.2 Endpoint probe layer

Representative probes:

```bash
curl -sS 'http://127.0.0.1:8080/api/health'
curl -sS 'http://127.0.0.1:8080/api/sessions?workspace_id=refactorio-foobar'
curl -sS 'http://127.0.0.1:8080/api/diff-runs?workspace_id=refactorio-foobar&session_id=refactorio-foobar%3Aall-indexed'
```

These verified concrete contract expectations quickly without UI noise.

### 56.3 Runtime route layer

Playwright traversal across primary nav pages was used to assert:

1. pages render,
2. no fatal error copy appears,
3. console remains error-free.

This caught issues that compile/tests missed.

## 57. Architecture Tradeoff Analysis (Extended)

### 57.1 Why not maintain dual contracts in UI?

Dual contract support seems safer short-term but creates complex and fragile branching in client code:

1. transformation logic bifurcates,
2. test matrix doubles,
3. stories need mirrored fixture sets,
4. deprecation horizon often slips.

Given ticket goals and local-first system assumptions, single-contract cutover was the right call.

### 57.2 Why not solve all session routing in frontend only?

A frontend-only workaround for `diff-runs` 404 (for example fallback to explicit run ID when session lookup fails) was possible but rejected.

Reason: this would paper over backend inconsistency and leave API behavior non-uniform across clients. The proper fix belongs on backend policy path.

### 57.3 Why keep schema-level tree-sitter remnants?

Hard-dropping schema objects would require migration policy decisions and potential compatibility concerns for existing DBs.

Given immediate priority to stabilize user-facing contract, runtime surface removal was sufficient and lower risk. Schema cleanup can be planned explicitly later.

## 58. Question-and-Answer Handoff Section

### Q1: Is backend or frontend now the contract source of truth?

Backend remains source of truth for active runtime API behavior. Frontend types and transforms should be updated whenever backend contract changes.

### Q2: What is a session in practical terms?

A session is a grouped context mapping each supported domain to a run ID (if available). It is not merely UI state; it is the cross-domain scope that gives all pages coherent context.

### Q3: Why do we have synthetic `all-indexed` session?

To provide a high-coverage default context that spans current indexed domains, reducing context switching and sparse initial views.

### Q4: Why were there duplicate symbol rows?

Symbol list rows represent occurrences, not unique symbol definitions. Multiple occurrences can share `symbol_hash`, so row key must include location/run dimensions.

### Q5: What should be monitored after closure?

1. endpoint behaviors involving `session_id`,
2. process restart discipline,
3. build pipeline and Storybook fixture drift,
4. follow-up ticket decomposition quality (REF-009..014).

## 59. Expanded Glossary

### API alignment

Ensuring request parameters, field names, response envelopes, and nullability semantics are consistent between producer (backend) and consumer (frontend).

### Session scoping

Using active session’s per-domain run IDs to constrain data queries so all page views represent a coherent index context.

### Availability gating

Rendering policy that binds UI table rows to domain run availability. If no run exists in session, rows are set to empty and domain-specific empty state is shown.

### Override session

A session configured in workspace settings rather than derived strictly from run grouping. Example: `all-indexed` aggregate session.

### Runtime mismatch

Behavior discrepancy that appears under live execution even though compile/tests may pass.

### Structural confidence

Confidence that components/routes exist and compile.

### Semantic confidence

Confidence that behavior and meaning are correct under real data and real user flows.

## 60. Expanded Lessons for Future Large-Scale Cleanup Programs

### 60.1 Sequence matters

Attempting visual polish before contract stabilization would have wasted time. The successful sequence was:

1. contract cutover,
2. state behavior correctness,
3. runtime stabilization,
4. toolchain/build cleanup,
5. final live validation.

### 60.2 Cross-layer ownership must be explicit

Every mismatch had at least two layers involved. Ticketing and diary discipline helped prevent “frontend bug” vs “backend bug” blame loops by forcing concrete evidence and file-level accountability.

### 60.3 Validation needs diversity

No single validation strategy was enough:

1. tests caught package regressions,
2. compile caught type drift,
3. probes caught endpoint policy bugs,
4. live traversal caught interaction-level mismatches.

### 60.4 Documentation is part of the fix

In this program, docs were not after-the-fact artifacts. They were control surfaces for continuity, handoff, and correctness audits. Without diary + changelog discipline, regressions would likely reintroduce already-solved issues.

### 60.5 Tight loops beat perfect plans

Many final fixes were discovered in short loops: observe, patch, validate immediately, document. This was more effective than speculative broad refactors in unstable zones.

## 61. Closure Checklist for This Postmortem Deliverable

1. Postmortem content stored in active ticket: yes (`analysis/03-...postmortem.md`).
2. Scope covers last relevant ticket diaries (`REF-007`, `REF-008`, `REF-015`): yes.
3. Includes prose, pseudocode, API snapshots, file references, decisions, problems, lessons: yes.
4. Includes runtime validation narrative and remaining risks: yes.
5. Suitable for teammate handoff and onboarding: yes.

## 62. End-of-Document Statement

This postmortem should be treated as the engineering source document for understanding the UI cleanup and API-alignment program completed in this cycle. If a future incident appears similar (session drift, contract mismatch, stale rows, route-specific 404 with valid session), start from sections 15, 34, 41, and 48 before proposing new architectural changes.


## 63. Extended Comparative Notes Against UI Design Intent

This section explicitly maps cleanup outcomes to the spirit of `sources/ui-design.md` so future contributors can see where implementation now aligns, where it intentionally diverges, and why.

### 63.1 Alignment achieved

1. **Session as orientation primitive**
   - Design intent emphasized context scoping. Implementation now consistently scopes domain queries by active session run IDs.
2. **Cross-domain exploration model**
   - Main pages now function under one coherent session context, which was a central product premise.
3. **Operational transparency**
   - Structured server logs and route-level behavior make the system auditable during debugging, matching the design’s emphasis on inspectability.
4. **Robust empty states**
   - Domain-absent sessions now show explicit empty-state explanations rather than misleading stale content.

### 63.2 Intentional simplification/divergence

1. **Tree-sitter feature path**
   - The design history references tree-sitter as optional/aspirational in places. Runtime surface has been intentionally removed for now.
2. **Feature breadth versus reliability**
   - Cleanup prioritized operational correctness over adding new UX feature depth from design backlog.
3. **Session naming polish**
   - Labels are functionally improved but still partly derived client-side for unnamed sessions.

### 63.3 Implication for future design-driven work

Design alignment should now proceed from a stable baseline. Follow-up tickets should avoid re-opening contract-level uncertainty while implementing design-level enhancements.

## 64. Extended Failure Taxonomy

The following taxonomy is offered to accelerate future incident triage by categorizing failure shapes seen in this campaign.

### Category A: Contract drift failures

Definition: frontend assumptions and backend payloads diverge in fields, envelopes, or semantics.

Observed examples:

1. Story fixtures using outdated response envelopes.
2. TS type drift in diff hunk fields.
3. Dashboard key mismatch for `Run` objects.

Detection modes:

1. compile errors,
2. runtime undefined access,
3. empty render with non-empty backend data.

Corrective template:

1. update canonical UI types,
2. update RTK transforms,
3. update fixtures and stories,
4. rerun build and route probes.

### Category B: Scope coherence failures

Definition: active context (session/workspace) is not consistently applied across pages and endpoints.

Observed examples:

1. stale rows after session switch.
2. diff-runs session override 404.

Detection modes:

1. user reports "selection doesn’t work",
2. route-specific 404 despite valid session object,
3. cross-page context mismatch.

Corrective template:

1. centralize context derivation,
2. ensure endpoint parity for session resolution,
3. gate display by run availability.

### Category C: Runtime data-shape failures

Definition: data in DB violates strict assumptions in read path.

Observed examples:

1. nullable bool scan errors in files endpoints.

Detection modes:

1. 500 on otherwise valid requests,
2. logs showing scan type mismatch.

Corrective template:

1. use nullable scan types,
2. normalize output values,
3. add defensive tests where practical.

### Category D: Operational process failures

Definition: actual running binary/process differs from expected code state.

Observed examples:

1. tree_sitter appearing in responses after code removal.

Detection modes:

1. endpoint output contradicts source grep,
2. multiple server processes and listener confusion.

Corrective template:

1. inspect listener PID,
2. verify executable path,
3. kill/restart deterministically.

### Category E: UI rendering identity failures

Definition: UI list rendering uses non-unique identifiers leading to unstable reconciliation.

Observed examples:

1. duplicate symbol row key warnings.

Detection modes:

1. React console warnings,
2. selection glitches in dynamic lists.

Corrective template:

1. map data model cardinality,
2. compose stable unique key from immutable dimensions.

## 65. Extended Pseudocode: End-to-End Interaction Flow

This pseudocode models the now-stable high-level user journey from app load to cross-domain navigation.

```pseudo
onAppLoad:
  workspaceOptions <- GET /api/workspaces
  if workspaceOptions not empty:
    activeWorkspace <- previouslySelected || firstWorkspace

  sessions <- GET /api/sessions?workspace_id=activeWorkspace
  if activeSession not set:
    activeSession <- chooseBestByCoverageThenTimestamp(sessions)

  renderTopbar(activeWorkspace, activeSession)
  renderRoute(currentRoute)

onRouteRender(route):
  runMap <- deriveRunMap(activeSession.runs)

  switch route:
    case "symbols":
      if runMap.symbols missing:
        showEmpty("No symbol data for this session")
      else:
        rows <- GET /api/symbols?run_id=runMap.symbols
        renderTable(rows, key=symbol_hash:file:line:col:run_id)

    case "diffs":
      diffRuns <- GET /api/diff-runs?session_id=activeSession.id
      if diffRuns empty:
        showEmpty("No diff runs")
      else:
        files <- GET /api/diff/{run}/files
        if file selected:
          hunks <- GET /api/diff/{run}/file?path=file
          renderDiff(hunks)

    case "commits":
      if runMap.commits missing:
        showEmpty("No commits data for this session")
      else:
        commits <- GET /api/commits?run_id=runMap.commits
        renderCommits(commits)

    case "docs":
      if runMap.doc_hits missing:
        showEmpty("No docs data for this session")
      else:
        terms <- GET /api/docs/terms?run_id=runMap.doc_hits
        renderTerms(terms)

onSessionChange(newSession):
  activeSession <- newSession
  clearPageSelections()
  resetPaginationOffsets()
  rerenderCurrentRouteWithNewRunMap()
```

This flow captures the intended invariant: **all route data must be derived from active session run map**.

## 66. Final Expanded Conclusion

The cleanup campaign should be interpreted as a controlled convergence process where architectural intent and runtime behavior were forced into agreement across multiple layers and multiple tickets. The largest gain is not any single fix; it is the elimination of ambiguity in how data context is selected, transformed, and rendered.

Concretely, we now have:

1. A backend contract that matches active feature policy.
2. A frontend state model that encodes session scope explicitly.
3. A route set that renders deterministically under real data.
4. A logging and validation discipline that catches regressions earlier.
5. Documentation artifacts that make handoff executable rather than interpretive.

If future work preserves these properties, feature development in `REF-009` through `REF-014` can proceed as product iteration. If these properties are allowed to regress, future efforts will re-enter stabilization mode and lose momentum to hidden contract debt.
