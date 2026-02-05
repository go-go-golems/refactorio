# Tasks

## TODO - UI Implementation (MVP 1: Investigation Workbench)

- [x] Scaffold frontend project (Vite + React + Bootstrap + Zustand) in `refactorio/ui/`
- [x] Implement app shell layout (top bar, left navigation, main content area)
- [x] Implement workspace selection modal and workspace context provider
- [x] Implement DB info display and schema validation warnings
- [x] Implement runs list view + run detail view
- [x] Implement session dashboard with session cards and availability badges
- [x] Implement unified search with type toggles, filters, and preview panel
- [x] Implement symbols explorer (list, filters, pagination) + symbol detail (tabs: overview, refs, history, audit)
- [x] Implement code units explorer + code unit detail (tabs: snapshot, history, diffs, related)
- [x] Implement commits explorer + commit detail (tabs: overview, files, diff, impact)
- [x] Implement diffs explorer + diff detail (hunks, lines, search within diff)
- [x] Implement docs/terms explorer (terms-first and file-first modes)
- [x] Implement files explorer (tree) + file viewer (code view, diff overlay, history, annotations)
- [ ] Implement raw outputs view
- [ ] Wire frontend to Go backend (go:embed for production, proxy for dev)
- [ ] Add responsive layout and keyboard navigation (command palette)

## DONE - Backend API Implementation

- [x] Scaffold Workbench API server package + cobra command (refactorio api serve) with router, JSON helpers, and error model
- [x] Implement workspace registry (config file) + CRUD endpoints
- [x] Add DB open helper + /api/db/info endpoint with schema/FTS detection
- [x] Implement runs endpoints (/api/runs, /api/runs/:id, /api/runs/:id/summary, /api/runs/:id/raw-outputs, /api/raw-outputs)
- [x] Implement session resolver + /api/sessions endpoints
- [x] Implement search endpoints (FTS per type + unified /api/search)
- [x] Implement symbol endpoints (/api/symbols, /api/symbols/:hash, /api/symbols/:hash/refs)
- [x] Implement code unit endpoints (/api/code-units, /api/code-units/:hash, /api/code-units/:hash/history, /api/code-units/:hash/diff)
- [x] Implement diff endpoints (/api/diff-runs, /api/diff/:run_id/files, /api/diff/:run_id/file)
- [x] Implement commit endpoints (/api/commits, /api/commits/:hash, /api/commits/:hash/files, /api/commits/:hash/diff)
- [x] Implement docs endpoints (/api/docs/terms, /api/docs/hits)
- [x] Implement files endpoints (/api/files tree, /api/file content, /api/files/history)
- [x] Implement optional tree-sitter capture endpoint (/api/tree-sitter/captures)
- [x] Add API tests (db info, runs list, search smoke)
- [x] Expand API tests for code-units, files, and diff endpoints
- [x] Run end-to-end smoke check by starting refactorio api serve and curling core endpoints against a real index DB
- [x] Write workbench REST API reference doc as Glazed help entry
