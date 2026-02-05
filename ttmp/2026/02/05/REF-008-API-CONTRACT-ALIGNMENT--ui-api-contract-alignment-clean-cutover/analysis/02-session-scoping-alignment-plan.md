---
Title: Session Scoping Alignment Plan
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
RelatedFiles: []
ExternalSources: []
Summary: "Clean-cutover plan to wire session scoping across all UI queries using the backend session/run model."
LastUpdated: 2026-02-05T16:58:00-05:00
WhatFor: "Define how session scoping should work end-to-end in the UI and API after contract alignment."
WhenToUse: "Use when implementing session scoping across pages or validating session behavior."
---

# Session Scoping Alignment Plan

## Goal
Wire session scoping across the UI so every page uses the active session to select the correct run_id(s) and filter results. This is the single highest-impact functional gap for the Workbench UX.

## Context
The UI stores `activeSessionId` in Redux but does not apply it to any API queries. The backend models an “Index Session” as a grouping of runs (per domain). A session includes `runs` and `availability` fields keyed by domain (commits, diff, symbols, code_units, doc_hits, gopls_refs, tree_sitter). Proper scoping requires mapping the active session’s run IDs to each domain query.

## Canonical Behavior (Clean Cutover)
- Selecting a session should affect **all** domain queries.
- The active session should be used to derive a per-domain `run_id` (or equivalent) for queries.
- If a session has no run for a domain, the UI should show an explicit “data not available” state.
- The session selector should be globally visible and consistent (Topbar + Dashboard).

## Required Data Sources
- `GET /sessions` to populate selectable sessions.
- `GET /sessions/:id` (or cached list entry) to read `runs` map.
- Optional: `GET /diff-runs?session_id=...` for diff scoping when `runs.diff` is missing or not provided.

## Scoping Map (Domain → Session Run ID)
Define a canonical map from UI domain to `session.runs` key:
- Dashboard: uses session context to highlight availability, no run-specific query needed.
- Runs page: can remain unscoped by session (shows all runs), or optionally filter to the session’s run set.
- Symbols: `run_id = session.runs.symbols`
- Code Units: `run_id = session.runs.code_units`
- Commits: `run_id = session.runs.commits`
- Diffs: `run_id = session.runs.diff` (or `GET /diff-runs?session_id=...` to find the run)
- Docs: `run_id = session.runs.doc_hits`
- Files: may be unscoped (files are global), but file history should use session commits if present
- Search: pass `run_id` per type (or `session_id` if the backend supports it)
- Tree-sitter: `run_id = session.runs.tree_sitter` (if/when UI is built)

## UI Implementation Strategy
1. **Session Resolver Hook**
   - Create a selector or hook, e.g. `useActiveSession()` that returns the active session object and the derived run IDs.
   - Provide `sessionRuns` as a normalized map for easy use in queries.

2. **Query Parameter Wiring**
   - Update each page’s query call to include the domain run_id if available.
   - If the run_id is missing, show a “data not available for this session” empty state (not an error).

3. **Global Session UX**
   - Implement a session selector in Topbar (dropdown or modal).
   - Keep Dashboard session cards as a secondary entry point.

4. **Search Scoping**
   - For unified search, pass `run_ids` or `session_id` in the request body.
   - For typed searches, add `run_id` where supported.

## Backend Support Requirements
- Ensure session response includes `runs` for all domains.
- Ensure any endpoints that support `run_id` are documented and consistent.
- If any endpoint cannot filter by run_id, decide whether to add it or explicitly mark it as “global”.

## Validation Plan
- Use a workspace with multiple sessions and overlapping data.
- Verify that changing the active session changes:
  - The symbols list and inspector.
  - The code units list and detail.
  - The commits list and detail.
  - The diff runs and diff files.
  - The docs terms/hits.
- Confirm that missing domains show “not available” instead of empty results.

## Deliverables
- A small session scoping utility in UI state or hooks.
- Updated queries across all pages.
- Updated UI empty states for missing domain data.
- Documentation entry describing session scoping behavior.
