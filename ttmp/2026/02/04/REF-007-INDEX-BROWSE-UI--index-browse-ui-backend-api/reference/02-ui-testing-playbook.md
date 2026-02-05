---
Title: UI Testing Playbook
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
    - Path: refactorio/ui/vite.config.ts
      Note: Vite dev server config with /api proxy to :8080
    - Path: refactorio/ui/package.json
      Note: npm scripts (dev, build, storybook)
    - Path: refactorio/pkg/workbenchapi/server.go
      Note: Go API server entry point
    - Path: refactorio/cmd/refactorio/api.go
      Note: "refactorio api serve" cobra command
    - Path: refactorio/ui/.storybook/preview.ts
      Note: Storybook config with MSW initialization
    - Path: refactorio/ui/src/stories/decorators.tsx
      Note: withPageContext decorator for page stories
ExternalSources: []
Summary: Step-by-step playbook for running the Workbench UI against a real backend, Storybook, and end-to-end testing.
LastUpdated: 2026-02-05T16:12:00-05:00
WhatFor: Reference for developers setting up and testing the UI locally.
WhenToUse: When starting local development, onboarding, or verifying the UI works end-to-end.
---

# UI Testing Playbook

## Goal

Provide copy/paste-ready instructions for running the Workbench UI in all test configurations: Storybook (isolated, mocked), Vite dev server (proxied to real backend), and production build (embedded in Go binary).

## Prerequisites

- Go toolchain
- Node.js 18+ and npm
- A Git repository to index (e.g. `glazed` at `~/code/wesen/corporate-headquarters/glazed`)
- Optional: `rg` (ripgrep) for doc-hits ingestion

All commands assume you're in the refactorio root:

```bash
cd /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio
```

## 1. Storybook (Isolated, MSW-Mocked)

Storybook renders each page component in isolation with MSW intercepting all API calls. No backend needed.

### Start Storybook

```bash
cd ui && npm run storybook
```

Opens at **http://localhost:6006**. Navigate to `Pages/*` in the sidebar.

### What to check

- **Pages/DashboardPage** — DB info cards, session list, recent runs
- **Pages/RunsPage** — Entity table with pagination, inspector panel
- **Pages/SymbolsPage** — Kind filter dropdown, symbol refs in inspector
- **Pages/CodeUnitsPage** — Kind/search filters, code unit detail
- **Pages/CommitsPage** — Search filter, commit files in inspector
- **Pages/DiffsPage** — 3-panel layout: run list → file list → diff viewer
- **Pages/DocsPage** — Term table, doc hits in inspector
- **Pages/FilesPage** — File tree sidebar, code viewer
- **Pages/SearchPage** — Search bar, results list (URL query sync)
- **Pages/WorkspacePage** — Workspace selector, create/edit form

Each page has 3 story variants:
- **Default** — Populated with mock data
- **Empty** — Empty responses, shows empty states
- **Loading** — Infinite delay, shows skeleton/placeholder states

### Build Storybook (static)

```bash
cd ui && npx storybook build
# Output: ui/storybook-static/
```

## 2. Dev Server Against Real Backend

The Vite dev server proxies `/api/*` requests to the Go backend on `:8080`. This is the primary way to test the UI end-to-end during development.

### Step 2a: Create an Index Database

Pick a target repo and build an index. This example uses the glazed repo with a 20-commit range:

```bash
ROOT=$(pwd)
DB=/tmp/refactorio-test-index.db
REPO=~/code/wesen/corporate-headquarters/glazed

# Initialize the database
GOWORK=off go run ./cmd/refactor-index init --db "$DB"

# Ingest everything in one shot (commits + diff + symbols + code-units + doc-hits)
echo -e "Processor\nCommandProcessor\nhandler\nmiddleware\nrefactor" > /tmp/terms.txt

GOWORK=off go run ./cmd/refactor-index ingest range \
  --db "$DB" \
  --repo "$REPO" \
  --from HEAD~20 --to HEAD \
  --include-diff \
  --include-symbols \
  --include-code-units \
  --include-doc-hits \
  --terms /tmp/terms.txt \
  --sources-dir /tmp/refactorio-sources \
  --ignore-package-errors
```

Verify the index has data:

```bash
GOWORK=off go run ./cmd/refactor-index list symbols --db "$DB" --limit 5
GOWORK=off go run ./cmd/refactor-index list diff-files --db "$DB" --limit 5
```

For a quick/minimal index (commits + diffs only, faster):

```bash
GOWORK=off go run ./cmd/refactor-index init --db "$DB"
GOWORK=off go run ./cmd/refactor-index ingest commits --db "$DB" --repo "$REPO" --from HEAD~20 --to HEAD
GOWORK=off go run ./cmd/refactor-index ingest diff --db "$DB" --repo "$REPO" --from HEAD~20 --to HEAD
```

### Step 2b: Start the Backend API Server

```bash
GOWORK=off go run ./cmd/refactorio api serve
```

The server starts on **http://localhost:8080** with the API at `/api`.

Flags:
- `--addr :9090` — Change listen address
- `--workspace-config /path/to/workspaces.json` — Custom config location (default: `~/.config/refactorio/workspaces.json`)

### Step 2c: Register a Workspace

With the server running, register a workspace that points to the index DB:

```bash
curl -s -X POST http://localhost:8080/api/workspaces \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "glazed",
    "db_path": "/tmp/refactorio-test-index.db",
    "repo_root": "'"$HOME"'/code/wesen/corporate-headquarters/glazed"
  }' | python3 -m json.tool
```

Verify:

```bash
curl -s http://localhost:8080/api/workspaces | python3 -m json.tool
```

### Step 2d: Start the Vite Dev Server

In a second terminal:

```bash
cd ui && npm run dev
```

Opens at **http://localhost:3000**. The Vite proxy forwards `/api/*` → `localhost:8080`.

### Step 2e: Smoke Test the UI

1. Open **http://localhost:3000** — should show the Dashboard with DB info cards
2. Click through sidebar: Runs, Symbols, Code Units, Commits, Diffs, Docs, Files, Search
3. Click a row in any table — inspector panel should open with details
4. Type in the search bar — should navigate to `/search?q=...` with results
5. Go to Workspaces — should show the registered workspace, "New Workspace" form should work

### Quick API Smoke Test (curl)

```bash
BASE=http://localhost:8080/api

# DB info
curl -s "$BASE/db/info?workspace_id=glazed" | python3 -m json.tool

# Sessions
curl -s "$BASE/sessions?workspace_id=glazed" | python3 -m json.tool

# Runs (paginated)
curl -s "$BASE/runs?workspace_id=glazed&limit=5" | python3 -m json.tool

# Symbols
curl -s "$BASE/symbols?workspace_id=glazed&limit=5" | python3 -m json.tool

# Search
curl -s -X POST "$BASE/search?workspace_id=glazed" \
  -H 'Content-Type: application/json' \
  -d '{"query":"handler"}' | python3 -m json.tool

# Files tree
curl -s "$BASE/files?workspace_id=glazed" | python3 -m json.tool
```

## 3. Production Build (go:embed)

> **Note:** The go:embed integration (task 15) is not yet wired. These are the expected steps once it is.

```bash
# Build the UI into pkg/workbenchapi/static/dist/
cd ui && npm run build

# Run the Go server (will serve the SPA from the embedded assets)
cd .. && GOWORK=off go run ./cmd/refactorio api serve
```

The server would serve the SPA at `/` and the API at `/api` from a single binary.

## Terminal Layout

For active development, use 3 terminals (or tmux panes):

```
┌──────────────────────────────────────┐
│ Terminal 1: Go backend               │
│ GOWORK=off go run ./cmd/refactorio   │
│   api serve                          │
├──────────────────────────────────────┤
│ Terminal 2: Vite dev server          │
│ cd ui && npm run dev                 │
├──────────────────────────────────────┤
│ Terminal 3: Working terminal         │
│ (curl, git, editing)                 │
└──────────────────────────────────────┘
```

For Storybook development, only 1 terminal is needed:

```
cd ui && npm run storybook
```

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| "Select a workspace to get started" on Dashboard | No workspace registered | Register one via POST `/api/workspaces` (Step 2c) |
| 502/ECONNREFUSED on page load | Backend not running | Start `go run ./cmd/refactorio api serve` |
| Empty tables (no rows) | DB has no data for that domain | Re-run ingest with `--include-symbols`, `--include-code-units`, etc. |
| "No sessions found" | Commit range has no diff runs | Use a larger `--from HEAD~N` range when ingesting |
| Storybook stories show errors | Missing MSW handlers | Check browser console for unhandled requests |
| CORS errors in browser | Accessing backend directly | Use Vite dev server (`:3000`), not backend (`:8080`) directly |
| `GOWORK=off` needed | go.work file in parent dir | Always prefix Go commands with `GOWORK=off` |

## Related

- [Diary Step 27](./01-diary.md) — RTK Query slices, page integration, Storybook stories
- [Workbench API Reference](../../../../../../pkg/doc/topics/04-workbench-api-reference.md) — Full REST API docs
- [refactor-index tutorial](refactor-index-how-to-use) — `go run ./cmd/refactor-index help refactor-index-how-to-use`
