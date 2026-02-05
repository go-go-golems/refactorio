# Tasks

## Active

- [ ] Run UI against live backend (playbook) and confirm all pages render with real data
- [ ] Capture any remaining contract mismatches and update alignment notes

## Done

- [x] Define canonical API contract per endpoint (based on backend) and lock field names used by UI
- [x] Update UI API types to match backend response shapes (ids, field names, map vs array)
- [x] Align workspaces + DB info contract and update Dashboard/Workspace flows
- [x] Align runs + run summary contract and update Runs/Dashboard renderers
- [x] Align sessions contract and session list rendering
- [x] Align symbols endpoints (list/detail/refs) and update Symbols page + detail
- [x] Align code units endpoints (list/detail/history) and update Code Units page + detail
- [x] Align commits endpoints (list/detail/files) and update Commits page + detail
- [x] Align diffs endpoints (diff-runs/files/file) and update Diffs page + viewer
- [x] Align docs endpoints (terms/hits) and update Docs page
- [x] Align files endpoints (tree/content/history) and update Files page
- [x] Align unified + typed search endpoints and update Search page + results
- [x] Update all pages/components to new field names (runs, symbols, code units, commits, diffs, docs, files, search)
- [x] Implement session scoping hook and per-domain run-id map
- [x] Wire session run ids into symbols, code units, commits, diffs, docs, and search queries
- [x] Wire gopls refs run id into symbol refs query
- [x] Reset page selections on session change to avoid stale detail panels
- [x] Add global MSW handlers in Storybook preview to prevent 404s
- [x] Fix Storybook story handlers to match API response shapes (sessions/workspaces)
- [x] Audit page stories for missing session handlers after session scoping
- [x] Add zerolog request/error logging to workbench API (incl. session computation errors)
- [x] Convert refactorio CLI to glazed command wiring with logging layer (--log-level)
- [x] Migrate/refill live workspace DB (meta_runs schema + ingest domains) to unblock /api/sessions
- [x] Fix /api/files and file search handling for nullable file_exists/is_binary
- [x] Wire topbar workspace/session selectors to Redux state for real session switching
- [x] Prevent stale domain tables from persisting across session switches
- [x] Fix frontend TypeScript build blockers after API contract cutover
