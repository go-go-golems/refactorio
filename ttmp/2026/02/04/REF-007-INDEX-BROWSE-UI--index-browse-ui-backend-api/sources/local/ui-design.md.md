# Refactorio Workbench

*A web UI spec for investigating + assisting during code browsing and refactoring, built on the refactorio / refactor-index data model.*

This is a UX-first spec: start from how developers actually work during refactors, then derive the views, layout patterns, and data/API requirements that make those workflows fast and safe.

---

## 1. Product intent

Refactorio Workbench is a **local-first web UI** that sits on top of a **refactor-index SQLite database** and (optionally) a **checked-out repo**. It helps you:

* **Orient**: what exists, what changed, where, and why.
* **Find**: symbols, code units, docs, commits, diffs, patterns.
* **Understand impact**: references, blast radius, related code, history.
* **Plan**: define rename/migration mappings and validate feasibility.
* **Execute safely** (optional, staged): dry-run refactors, review patch, validate, audit leftovers.
* **Report**: produce shareable artifacts (markdown reports, checklists, run logs).

It is not an IDE replacement. It is a **refactor cockpit**: exploration + audit + orchestration.

---

## 2. Users & usage patterns (JTBD)

### Primary personas

1. **Refactor Driver**

    * Owns a large rename/migration.
    * Needs confidence: “Did I change everything? Did I miss docs? What’s risky?”

2. **Reviewer / Maintainer**

    * Wants a coherent story + evidence.
    * Needs: “Show me the impacts, diffs, and why this is safe.”

3. **Explorer / New Contributor**

    * Wants to understand a codebase quickly.
    * Needs: “Where is this defined? How is it used? What changed recently?”

### Core usage patterns (jobs)

These patterns drive the UI shape:

1. **Pick a scope and orient**

    * Choose repo + the “index session” (range, commit window, or last run).
    * See what data exists (diffs? symbols? code units? docs? gopls refs?).

2. **Find the thing**

    * Search across domains (symbols, code units, diffs, commits, docs).
    * Filter by run / commit / file / package / kind.

3. **Read in context**

    * Open file + highlight location.
    * View nearby diff hunks and/or commit context.
    * Jump between definition ↔ references ↔ history.

4. **Understand blast radius**

    * “If I rename X, what breaks?”
    * See semantic references (gopls), plus text hits (diff/doc).

5. **Plan & track refactor work**

    * Build a plan of symbol renames + doc/config term replacements.
    * Validate rename feasibility (`prepare_rename`) and flag risks.

6. **Audit completion**

    * Find leftover legacy terms/symbols.
    * Verify no docs or examples are stale.
    * Produce a checklist/report.

---

## 3. Core concepts (mental model)

Workbenches succeed when they hide internal complexity behind a stable user model.

### Concepts exposed in the UI

* **Workspace**

    * A named connection: `{db_path, repo_root?}`
    * One SQLite file can represent one repo over many runs.

* **Run**

    * One ingestion job (`meta_runs`) that produced rows in some domain tables.
    * Runs have status (`running|success|failed`), time bounds, args, raw outputs, and may be associated to a commit range.

* **Index Session (UI concept)**

    * A *bundle* of runs that belong together for exploration/refactoring.
    * Example: “glazed: HEAD~20 → HEAD” with:

        * commits run ✅
        * diff run ✅
        * symbols run ✅
        * code-units run ✅
        * doc-hits run ✅
        * gopls-refs run ⚠️ missing
        * tree-sitter run ⛔ failed

This abstraction is crucial because today refactor-index stores different passes as different run_ids.

* **Entity**

    * Something you browse/search:

        * File
        * Commit
        * Diff file/hunk/line
        * Symbol (definition)
        * Code Unit (function/type snapshot)
        * Doc hit
        * Tree-sitter capture (optional)

* **Plan / Refactor Run (optional orchestration)**

    * A structured plan of transformations (rename targets + doc/config replacements) and its execution log.

---

## 4. Global UI structure (derived from the jobs)

### App shell layout (consistent across the product)

* **Top bar** (always visible)

    * Workspace selector
    * Session selector
    * Global search box
    * “Command palette” button (Ctrl+K)
    * Quick actions: New Plan, Run Audit, View Runs

* **Left navigation**

    * Dashboard
    * Search
    * Explore

        * Files
        * Symbols
        * Code Units
        * Commits
        * Diffs
        * Docs / Terms
        * Tree-sitter (if enabled)
    * Refactor

        * Plans
        * Runs
        * Audits
        * Reports
    * Data / Admin

        * All Runs
        * Raw Outputs
        * Schema / DB Info
        * Settings

* **Main content**

    * Usually a **two- or three-pane** “browse + preview + inspector” pattern.

### Primary layout pattern: “List → Preview → Inspector”

Most investigation tasks are “scan results, preview, then drill in”.

**Default page layout:**

* **Left pane**: list/table (sortable, filterable, paginated)
* **Center pane**: preview/detail (diff viewer, code snippet, commit detail)
* **Right pane**: inspector/actions (add to plan, copy target spec, open in editor, related entities)

This layout should be reused across:

* Search
* Symbols
* Code units
* Commits
* Diffs
* Docs hits

---

## 5. View catalog (mapping jobs → views)

| Usage pattern           | Views that serve it                                         |
| ----------------------- | ----------------------------------------------------------- |
| Pick scope & orient     | Workspace picker, Session dashboard, Runs view              |
| Find the thing          | Unified Search                                              |
| Read in context         | File Viewer, Symbol Detail, Code Unit Detail, Commit Detail |
| Understand blast radius | Symbol Detail → References tab, Refactor Plan impact panel  |
| Plan & track refactor   | Plan Builder, Plan Detail                                   |
| Audit completion        | Audit view, Docs/Terms view, Reports                        |

---

## 6. Detailed view specs

### 6.1 Workspace selection

**Goal:** connect to one or more SQLite DBs and optionally associate a repo root.

**View:** “Workspaces” (modal or dedicated page)

**Key actions**

* Add workspace:

    * Name
    * SQLite path
    * Repo root path (optional but recommended)
* Validate:

    * Read schema version
    * Confirm required tables exist
    * Detect available FTS tables and features

**Edge cases**

* DB schema < current expected:

    * Show a banner: “DB schema is older; some features disabled.”
    * Provide a safe “Upgrade schema” action (calls schema init).

---

### 6.2 Session Dashboard

**Goal:** let the user pick “the slice of data” they want to reason about.

**View:** Dashboard for a workspace

**Layout**

* Header: Workspace name + DB info (schema version, table counts)
* Section: “Sessions” (UI-defined)

    * Each card shows:

        * scope: repo root + from/to
        * last updated
        * available passes (commits/diff/symbols/code units/doc hits/gopls/tree-sitter)
        * failures + error summaries
    * Actions: “Open”, “Edit session mapping”, “Rebuild session” (optional)

**Session creation logic (UI + backend)**

* Group runs by `(root_path, git_from, git_to)` where possible.
* If missing git_from/git_to for symbol/code-unit/doc runs, allow “attach to session” manually or infer by nearest timestamps.

**Why this view exists**

* It turns run soup into a stable user experience.

---

### 6.3 Unified Search

**Goal:** one place to “find the thing” without caring which table it lives in.

**Route:** `/search`

**Layout (3-pane)**

* Left: filters
* Center: results (grouped or unified)
* Right: preview inspector

**Search inputs**

* Query string (FTS, with simple syntax)
* Types toggles:

    * Symbols
    * Code Units
    * Diffs
    * Commits
    * Docs
    * Files
    * Tree-sitter (optional)
* Scope filters:

    * Session
    * Specific run id(s)
    * Commit hash
    * File path glob
    * Package
    * Symbol kind (func/type/const/var/method)
    * Diff kind (+/-/context)

**Results display (center)**

* Default: grouped by type with collapsible headers and per-type counts.
* Each result row includes:

    * type icon
    * primary label (symbol name, file path, commit subject…)
    * snippet (highlighted)
    * location (file:line)
    * run + commit context

**Preview (right)**

* For code-bearing types: show a snippet with highlight + “Open in file viewer”
* For commit: show commit header + changed files count
* For diff line: show hunk context and “Open diff view”
* For symbol/code unit: show signature + “Open detail”

**Interaction requirements**

* Keyboard navigation: up/down selects result; Enter opens detail.
* “Copy link” for any result.

---

### 6.4 Files Explorer + File Viewer

#### Files Explorer

**Goal:** browse by structure, not by search.

**Route:** `/explore/files`

**Layout**

* Left: file tree (by path prefix) + filter (extension/type)
* Center: file viewer
* Right: context + overlays

**Tree behavior**

* Lazy-load children
* Show badges:

    * last commit date (if commit data exists)
    * doc hits count (for selected term set)
    * diff touched (for selected session)

#### File Viewer

**Goal:** read code with refactor context.

**Center panel**

* Code view with line numbers (read-only)
* Tabs: `File`, `Diff overlay`, `History`, `Annotations`

**Right panel**

* “In this file”:

    * Symbols defined here (from symbol_occurrences)
    * Code units in this file (from code_unit_snapshots)
    * Doc hits in this file (from doc_hits)
    * Tree-sitter captures (optional)
* “Actions”:

    * Copy `path:line`
    * Copy target spec for selected symbol
    * Open in editor (configurable URL scheme)
    * Add selected symbol/code unit to plan

**Diff overlay tab**

* If diff run selected: show hunks inline with +/− line shading and ability to jump to hunk.
* If commit-range diff runs exist: allow selecting a commit to overlay.

**History tab**

* Show list of commits touching file (commit_files + v_last_commit_per_file view)
* Clicking commit shows file diff and impacted code units.

**Data requirement**

* Needs a backend “file content provider”:

    * either from repo path (`git show <ref>:<path>`)
    * or from stored blobs / snapshots
    * minimum viable: show HEAD content from filesystem + show code-unit snapshots for older commits

---

### 6.5 Symbols Explorer + Symbol Detail

#### Symbols Explorer

**Goal:** list all definitions, filter fast, pick targets for refactoring.

**Route:** `/explore/symbols`

**Layout**

* Left: filters
* Center: table list
* Right: quick preview and actions

**Filters**

* Name (prefix / contains)
* Package
* Kind
* Exported only
* File path
* Session/run selection

**Table columns**

* Name
* Kind
* Package
* Signature (truncated)
* File + line
* Exported
* Hash (copy)

**Row actions**

* Open detail
* Copy target spec (`symbol_hash|file|line|col`)
* Add to plan

#### Symbol Detail

**Goal:** “what is this and where does it go?”

**Route:** `/symbols/:symbolHash`

**Header**

* `Name` + kind badge
* Package path
* Signature
* Exported status
* Hash (copy)
* Primary definition location (file:line:col)

**Tabs**

1. **Overview**

    * Definition snippet
    * Related code unit(s) containing definition span
    * “Search in diffs” shortcut (query = symbol name)
2. **References**

    * If gopls refs exist:

        * list of all references (`symbol_refs`) grouped by file
        * count summary
        * filter to declarations / non-declarations
    * If refs missing:

        * CTA: “Compute references” (creates a gopls-refs run) *(optional feature)*
3. **History**

    * Commits that changed the defining file
    * If commit-range snapshots exist: reference count delta over time (optional)
4. **Audit**

    * Doc hits and diff hits for this symbol name (text-level evidence)
    * Useful when semantic refs are incomplete.

**Key actions**

* “Add rename to plan…”
* “Run prepare_rename” (backend gopls, returns rename range + ok/fail)
* “Copy gopls position” (abs path:line:col if repo root known)

---

### 6.6 Code Units Explorer + Detail

#### Code Units Explorer

**Goal:** browse functions/types as refactor units, not just files.

**Route:** `/explore/code-units`

**List rows**

* Kind (func/type/method)
* Name
* Receiver (if method)
* Package
* File + range
* Body hash (useful for change detection)

**Filters**

* Name / package / kind
* Full-text in body/doc (FTS)
* File path
* Commit hash (if commit-scoped snapshots exist)

#### Code Unit Detail

**Goal:** read a function/type and see how it evolved.

**Route:** `/code-units/:unitHash`

**Layout**

* Center: full body + doc comment
* Right: metadata and links

**Tabs**

* Snapshot (current selected run/commit)
* History (timeline of snapshots, if present)
* Diffs (diff between two snapshots’ bodies)
* Related symbols (defs inside, or nearest defs by file/range—best-effort)

**Key actions**

* “Add to plan” (e.g., candidate for extraction/rename/move)
* “Search related diffs” (query by function name or signature tokens)

---

### 6.7 Commits Explorer + Commit Detail

#### Commits Explorer

**Goal:** answer “when did this change?” and “who touched this area?”

**Route:** `/explore/commits`

**List**

* Commit hash (short)
* Subject
* Author
* Date
* Files changed count (commit_files)

**Search**

* FTS on subject/body
* Filters for author, date range, file path

#### Commit Detail

**Goal:** investigate one change in depth.

**Tabs**

* Overview (metadata)
* Files (list of changed files with status)
* Diff (if diff run exists per commit-range, or reconstruct from stored diff lines)
* Impact (optional):

    * symbols defined in changed files
    * code units changed (by matching file ranges)
    * doc hits introduced/removed (if per-commit doc hits available)

---

### 6.8 Diffs Explorer + Diff Detail

**Route:** `/explore/diffs`

**Goal:** navigate changes across a range.

**List view**

* Diff runs (if multiple)
* Inside a run: diff files list

    * file path
    * status (A/M/D/R)
    * hunks count
    * added/removed line counts

**Diff Detail**

* Reconstructed unified diff:

    * grouped by hunks
    * line numbers old/new
    * ability to filter to:

        * added lines only
        * removed lines only
        * context lines hidden
* Right panel:

    * “Search within this diff”
    * “Open file viewer at line”
    * “Extract symbol/code unit affected” (best-effort: find nearest code unit snapshot spans containing hunk ranges)

---

### 6.9 Docs / Terms (Doc Hits)

**Route:** `/explore/docs`

**Goal:** track doc/config term changes and cleanup.

**Two modes**

1. **Terms-first**

    * list of terms (from doc_hits.term)
    * counts by term
    * click term → list hits by file
2. **File-first**

    * list files with doc hits, grouped by path prefix

**Hit detail**

* show match_text with highlight
* open in file viewer (if file is text-based)
* “Add replacement rule to plan” (term → new term)

---

### 6.10 Tree-sitter captures (optional)

**Route:** `/explore/tree-sitter`

**Goal:** structured pattern search beyond Go symbols (YAML/MD/JSON patterns, call sites, etc.)

**Filters**

* query_name
* capture_name
* node_type
* file path
* full-text in snippet (FTS optional)

**Result**

* snippet with location range
* open file at range

---

### 6.11 Runs + Raw Outputs

#### Runs list

**Route:** `/data/runs`

**Goal:** transparency + debugging: what was indexed, when, and what failed.

**Table columns**

* run_id
* status
* started/finished
* root_path
* git_from → git_to
* detected run kind (diff/commits/symbols/…)
* row counts by domain (computed summary)

**Run detail**

* args_json (pretty-printed)
* errors (error_json)
* links to raw outputs
* “Attach run to session” action (if not auto-grouped)

#### Raw outputs

**Route:** `/data/raw-outputs`

**Goal:** provide traceability to source artifacts (gopls output, diffs, rg output).

**List**

* run_id
* source (e.g. “gopls-references”)
* file path
* created_at

**Viewer**

* download / open text
* quick parse for common formats (gopls refs) *(optional)*

---

## 7. Refactor assistance features (beyond browsing)

These are the “assist” parts that turn browsing into a refactor workflow.

### 7.1 Refactor Plans

**Route:** `/refactor/plans`

**Plan data model (UI-level)**

* Plan name + description
* Scope (include/exclude paths, repo root, optional commit range)
* Rename targets:

    * symbol hash
    * old name → new name
    * package/kind
    * target_spec (file/line/col)
    * risk flags + validation results
* Doc term replacements:

    * “from” → “to”
    * file globs / exclusions

**Plan Builder UX (wizard + editor hybrid)**

**Step 1: Scope**

* choose session
* include/exclude globs
* choose which domains must be available (warn if missing)

**Step 2: Add targets**

* search symbols and add to plan
* or paste target_specs
* bulk edit table (old → new)

**Step 3: Validate**

* for each target:

    * run `prepare_rename` (backend) to confirm feasible
    * (optional) fetch gopls references count
* show risk panel:

    * prepare_rename failed
    * too many references
    * cross-package exports touched
    * missing gopls refs data

**Step 4: Preview impact**

* show:

    * semantic refs (count, top files)
    * text hits in diffs/docs for old name
    * “likely missed areas” suggestions (e.g., docs mention old term but no code refs)

**Step 5: Export / Apply**

* export mapping.yaml and/or plan.json
* start a refactor run (dry-run by default)

---

### 7.2 Refactor Runs (pipeline execution)

**Route:** `/refactor/runs`

A refactor run is a **stage timeline**:

1. Inventory (ensure indexes exist)
2. Plan (validate targets)
3. Apply (dry-run or write)
4. Validate (gofmt/test/lint)
5. Audit (FTS/doc hits for leftovers)
6. Report (markdown outputs)

**Run detail UI**

* Timeline down the left (stages)
* Main pane: logs + artifacts for selected stage
* Right pane: summary + “next action” suggestions

**Important UX**

* Everything is reproducible:

    * show the exact CLI commands / config used
    * store artifacts and link them

---

### 7.3 Audits

**Route:** `/refactor/audits`

**Audit types**

* “Leftover term audit”
* “Leftover symbol name audit”
* “Docs fenced code audit” (optional future)
* “External API call audit” (diff-based pattern scan)

**Audit run output**

* counts, file lists, exact hit locations
* “Mark resolved” is a UI annotation only (does not change DB rows)
* exports to markdown checklist

---

### 7.4 Reports

**Route:** `/refactor/reports`

**Goal:** a shareable package.

* List of generated reports (from report generator)
* Preview markdown
* Download bundle (zip) *(optional)*

---

## 8. Cross-cutting interaction rules

### Context is explicit

Every exploration page must display:

* Workspace
* Session (or run set)
* Commit context (if applicable)

### Deep links

Every entity page must be linkable:

* `/symbols/:hash?session=…`
* `/files/:path?ref=…&session=…`
* `/search?q=…&types=…`

### Copy/share primitives everywhere

* Copy file path
* Copy `path:line:col`
* Copy symbol hash
* Copy target spec
* Copy commit hash

### “Open in editor”

Configurable:

* VS Code link scheme
* JetBrains link scheme
* Or just copy absolute path

### Pagination defaults

Large tables must paginate server-side:

* diff_lines, doc_hits, commits, symbols, code_units

### Missing data is first-class

If a session lacks gopls refs:

* don’t break the view
* show “Refs unavailable” with an action to create/ingest them

---

## 9. Backend/API requirements (what the UI needs)

The UI can’t be good if the backend forces it into N+1 queries or ambiguous scoping. This is the minimal API surface to support the above views.

### Core endpoints

**Workspace / DB info**

* `GET /api/db/info`

    * schema version
    * table availability
    * feature flags (fts tables present)

**Runs**

* `GET /api/runs`
* `GET /api/runs/:id`
* `GET /api/runs/:id/summary` (row counts per domain)
* `GET /api/runs/:id/raw-outputs`

**Sessions (UI concept)**

* `GET /api/sessions` (server computes groupings)
* `GET /api/sessions/:id` (includes run ids for each pass type)
* `POST /api/sessions` (optional: manual session definitions)

**Search**

* `POST /api/search` (unified dispatcher)
* plus optional typed endpoints:

    * `GET /api/search/symbols`
    * `GET /api/search/code-units`
    * `GET /api/search/diff`
    * `GET /api/search/commits`
    * `GET /api/search/docs`
    * `GET /api/search/files`

**Explore entities**

* `GET /api/files?prefix=...`
* `GET /api/file?path=...&ref=...` (returns content + line map)
* `GET /api/symbols?filters...`
* `GET /api/symbols/:hash`
* `GET /api/symbols/:hash/refs?run_id=...`
* `GET /api/code-units?filters...`
* `GET /api/code-units/:hash`
* `GET /api/commits?filters...`
* `GET /api/commits/:hash`
* `GET /api/diff-runs?session_id=...`
* `GET /api/diff/:run_id/files`
* `GET /api/diff/:run_id/file?path=...` (hunks + lines)
* `GET /api/docs/hits?term=...&filters...`
* `GET /api/docs/terms?filters...`

### Refactor assistance endpoints (optional stage)

* `POST /api/refactor/plans`
* `GET /api/refactor/plans/:id`
* `POST /api/refactor/plans/:id/validate` (prepare_rename, reference counts)
* `POST /api/refactor/runs` (kick off runner)
* `GET /api/refactor/runs/:id` (stages + logs + artifacts)

*(If you don’t want background jobs, keep it synchronous and return artifacts immediately, but still represent stages in the UI.)*

---

## 10. Frontend implementation constraints (so the spec is buildable)

Aligning with your existing project conventions:

* **React + TypeScript**
* **RTK Query** for data fetching/caching
* **Bootstrap** for layout and standard components
* “List → Preview → Inspector” components should be shared and parameterized.

Core reusable components to spec up front:

* `<WorkspaceSelector/>`
* `<SessionSelector/>`
* `<GlobalSearchBar/>`
* `<EntityTable/>` (server-paginated)
* `<CodeViewer/>` (line numbers + highlights)
* `<DiffViewer/>` (hunks + inline/side-by-side)
* `<InspectorPanel/>`
* `<RunStatusBadge/>`
* `<CopyButton/>`
* `<OpenInEditorButton/>`
* `<CommandPalette/>`

---

## 11. MVP scope (what to build first)

If you want a “useful quickly” sequence:

### MVP 1 — Investigation workbench

* Workspace connect + DB info
* Runs list + run detail + raw outputs
* Session dashboard (even if grouping is naive at first)
* Unified search across:

    * diffs (diff_lines_fts)
    * docs (doc_hits_fts)
    * code units (fts)
    * symbols (fts)
    * commits (fts)
    * files (fts)
* Symbols explorer + detail (without refs if not present)
* Diffs explorer + diff detail
* Docs/terms explorer
* File viewer (read HEAD from repo root)

### MVP 2 — Refactor assist (planning)

* Plan builder (add symbols, rename mapping, doc terms)
* Target validation via `prepare_rename`
* Reference ingestion trigger (create gopls-refs run)
* Impact preview panel (counts + top files)

### MVP 3 — End-to-end refactor runs (optional)

* Runner execution UI (dry-run default)
* Validation/audit stages
* Report preview and export

---

## 12. Acceptance criteria (what “done” means)

A good first release meets these user-level outcomes:

* I can select a session and **answer**:

    * “Where is X defined?”
    * “Where is X referenced?” (if gopls data exists)
    * “What changed in this range?”
    * “Which files still mention the old term?”
* I can click from any search result to a **stable, contextual view**:

    * file + line highlight, diff hunk, commit, symbol detail, etc.
* The UI makes missing data obvious and actionable:

    * “No refs for this symbol—compute refs”
    * “Tree-sitter disabled / failed—see run error”
* I can produce a shareable artifact:

    * a report, a plan export, or a run summary.

---

# Refactorio Workbench — ASCII Screen Designs

## 1. App Shell & Session Dashboard

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search symbols, files, diffs…  ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  SESSION: main: HEAD~20 → HEAD                                                │
│                    │  ═══════════════════════════════════════════════════════════════════════════  │
│  🔍 Search         │                                                                                │
│                    │  Workspace: glazed                                                             │
│  📂 Explore        │  DB Path:   /Users/dev/glazed.db                                               │
│    ├─ Files        │  Repo Root: /Users/dev/src/glazed                                              │
│    ├─ Symbols      │  Schema:    v12 ✓                                                              │
│    ├─ Code Units   │                                                                                │
│    ├─ Commits      │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│    ├─ Diffs        │  │  DATA AVAILABILITY                                                      │  │
│    ├─ Docs/Terms   │  ├─────────────────────────────────────────────────────────────────────────┤  │
│    └─ Tree-sitter  │  │                                                                         │  │
│                    │  │   ✅ Commits      1,847 commits    run #42   2 hours ago                │  │
│  🔄 Refactor       │  │   ✅ Diffs        3,291 files      run #43   2 hours ago                │  │
│    ├─ Plans        │  │   ✅ Symbols     12,456 defs       run #44   1 hour ago                 │  │
│    ├─ Runs         │  │   ✅ Code Units   8,234 funcs      run #44   1 hour ago                 │  │
│    ├─ Audits       │  │   ✅ Doc Hits       847 matches    run #45   1 hour ago                 │  │
│    └─ Reports      │  │   ⚠️ Gopls Refs   (not computed)   [Compute References]                 │  │
│                    │  │   ⛔ Tree-sitter  failed           run #41   [View Error] [Retry]       │  │
│  ⚙️ Data/Admin     │  │                                                                         │  │
│    ├─ All Runs     │  └─────────────────────────────────────────────────────────────────────────┘  │
│    ├─ Raw Outputs  │                                                                                │
│    ├─ Schema Info  │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│    └─ Settings     │  │  QUICK STATS                                                            │  │
│                    │  ├─────────────────────────────────────────────────────────────────────────┤  │
│                    │  │                                                                         │  │
│                    │  │   Files Modified     142          Packages Touched    23                │  │
│                    │  │   Lines Added      4,892          Lines Removed    2,103                │  │
│                    │  │   New Symbols        67           Deleted Symbols    12                 │  │
│                    │  │   Doc Hits (legacy)  89           Contributors        4                 │  │
│                    │  │                                                                         │  │
│                    │  └─────────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
│                    │  RECENT ACTIVITY                                                               │
│                    │  ───────────────────────────────────────────────────────────────────────────  │
│                    │   • abc1234  "Rename CommandProcessor to Handler"         @alice   3h ago     │
│                    │   • def5678  "Fix middleware registration order"          @bob     5h ago     │
│                    │   • 789abcd  "Add context propagation to grpc layer"      @alice   1d ago     │
│                    │   • cde0123  "Deprecate legacy JSON codec"                @carol   2d ago     │
│                    │                                                                                │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Unified Search (3-Pane)

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 CommandProcessor             ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │┌──────────────┐ ┌────────────────────────────────────┐ ┌─────────────────────┐│
│                    ││ FILTERS      │ │ RESULTS  "CommandProcessor"  127   │ │ PREVIEW             ││
│ ●🔍 Search         │├──────────────┤ ├────────────────────────────────────┤ ├─────────────────────┤│
│                    ││              │ │                                    │ │                     ││
│  📂 Explore        ││ Types        │ │ ▼ Symbols (4)                      │ │ 📦 Symbol           ││
│    ├─ Files        ││ ☑ Symbols    │ │ ┌────────────────────────────────┐ │ │                     ││
│    ├─ Symbols      ││ ☑ Code Units │ │ │ ● CommandProcessor        type │ │ │ CommandProcessor    ││
│    ├─ Code Units   ││ ☑ Diffs      │ │ │   pkg/handlers                 │ │ │ ─────────────────── ││
│    ├─ Commits      ││ ☑ Commits    │ │ │   handlers.go:45               │ │ │                     ││
│    ├─ Diffs        ││ ☑ Docs       │ │ └────────────────────────────────┘ │ │ type CommandProcessor││
│    ├─ Docs/Terms   ││ ☐ Files      │ │   CommandProcessorConfig    type   │ │   interface {       ││
│    └─ Tree-sitter  ││ ☐ Tree-sit.  │ │   NewCommandProcessor       func   │ │     Process(ctx     ││
│                    ││              │ │   commandProcessorImpl      type   │ │       context.Ctx,  ││
│  🔄 Refactor       ││ ───────────  │ │                                    │ │       cmd Command,  ││
│    ├─ Plans        ││              │ │ ▼ Code Units (12)                  │ │     ) (Result, err) ││
│    ├─ Runs         ││ Kind         │ │   Process()           method  L127 │ │     Validate(cmd    ││
│    ├─ Audits       ││ ☑ func       │ │   NewCommandProcessor func    L45  │ │       Command,      ││
│    └─ Reports      ││ ☑ type       │ │   commandProcessorImpl.run  m L89  │ │     ) error         ││
│                    ││ ☑ method     │ │   ... +9 more                      │ │   }                 ││
│  ⚙️ Data/Admin     ││ ☐ const      │ │                                    │ │                     ││
│    ├─ All Runs     ││ ☐ var        │ │ ▼ Diffs (38)                       │ │ ─────────────────── ││
│    ├─ Raw Outputs  ││              │ │   + type CommandProcessor    L45   │ │ pkg/handlers        ││
│    ├─ Schema Info  ││ ───────────  │ │   - type Processor          L45   │ │ handlers.go:45:1    ││
│    └─ Settings     ││              │ │   + func NewCommandProcessor L67   │ │ Exported: ✓         ││
│                    ││ Package      │ │   ctx.CommandProcessor       L112  │ │ Hash: a7b3c9f2      ││
│                    ││ [Any      ▾] │ │   ... +34 more                     │ │                     ││
│                    ││              │ │                                    │ │ ─────────────────── ││
│                    ││ File Path    │ │ ▶ Commits (8)                      │ │                     ││
│                    ││ [          ] │ │                                    │ │ [Open Detail]       ││
│                    ││              │ │ ▶ Docs (65)                        │ │ [Add to Plan]       ││
│                    ││ ───────────  │ │   README.md (12)                   │ │ [Copy Hash]         ││
│                    ││              │ │   docs/api.md (23)                 │ │ [Open in Editor]    ││
│                    ││ Run          │ │   ... +30 more                     │ │                     ││
│                    ││ [#44 sym. ▾] │ │                                    │ │                     ││
│                    │└──────────────┘ └────────────────────────────────────┘ └─────────────────────┘│
│                    │  ↑/↓ Navigate   Enter Open   ⌘C Copy path   ⌘⇧C Copy spec   Esc Clear        │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Files Explorer + File Viewer

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │┌────────────────────┐┌─────────────────────────────────────┐┌────────────────┐│
│                    ││ FILES              ││ pkg/handlers/processor.go           ││ IN THIS FILE   ││
│  🔍 Search         │├────────────────────┤│ [File] [Diff] [History] [Annotate]  │├────────────────┤│
│                    ││                    │├─────────────────────────────────────┤│                ││
│  📂 Explore        ││ 📁 cmd/            ││                                     ││ SYMBOLS (7)    ││
│   ●├─ Files        ││ 📁 internal/       ││  40 │                               ││ ● CommandProc… ││
│    ├─ Symbols      ││ ▼📁 pkg/           ││  41 │ // CommandProcessor handles   ││   type L45     ││
│    ├─ Code Units   ││   ▼📁 handlers/ ●  ││  42 │ // incoming command requests  ││ ● CommandProc… ││
│    ├─ Commits      ││     📄 doc.go      ││  43 │ // and routes them to the     ││   type L52     ││
│    ├─ Diffs        ││    ●📄 processor.go││  44 │ // appropriate handler.       ││ ● NewCommandP… ││
│    ├─ Docs/Terms   ││     📄 router.go   ││▸ 45 │ type CommandProcessor interface{││   func L67     ││
│    └─ Tree-sitter  ││     📄 middleware.…││  46 │   // Process executes cmd     ││ ● commandProc… ││
│                    ││   📁 middleware/   ││  47 │   Process(ctx context.Context,││   type L89     ││
│  🔄 Refactor       ││   📁 models/       ││  48 │     cmd Command,              ││ ● Process      ││
│    ├─ Plans        ││   📁 services/     ││  49 │   ) (Result, error)           ││   method L127  ││
│    ├─ Runs         ││ 📁 test/           ││  50 │                               ││ ● Validate     ││
│    ├─ Audits       ││ 📄 go.mod          ││  51 │   // Validate checks cmd      ││   method L156  ││
│    └─ Reports      ││ 📄 go.sum          ││  52 │   Validate(cmd Command) error ││ ● register     ││
│                    ││ 📄 README.md       ││  53 │ }                             ││   func L201    ││
│  ⚙️ Data/Admin     ││                    ││  54 │                               ││                ││
│    ├─ All Runs     ││ ─────────────────  ││  55 │ // CommandProcessorConfig     ││ ────────────── ││
│    ├─ Raw Outputs  ││ Filter: [.go    ]  ││  56 │ // holds processor settings.  ││                ││
│    ├─ Schema Info  ││                    ││  57 │ type CommandProcessorConfig   ││ CODE UNITS (5) ││
│    └─ Settings     ││ Legend:            ││  58 │   struct {                    ││ ƒ NewCommand…  ││
│                    ││  ● has diff changes││  59 │   Timeout  time.Duration      ││   L67-86       ││
│                    ││  ◐ has doc hits    ││  60 │   MaxRetry int                ││ ƒ Process      ││
│                    ││                    ││  61 │   Logger   *log.Logger        ││   L127-154     ││
│                    ││                    ││  62 │ }                             ││ ƒ Validate     ││
│                    ││                    ││  63 │                               ││   L156-198     ││
│                    ││                    ││  64 │ // NewCommandProcessor creates││ ƒ register     ││
│                    ││                    ││  65 │ // a new processor instance.  ││   L201-245     ││
│                    ││                    ││  66 │ func NewCommandProcessor(     ││ ƒ run          ││
│                    ││                    ││▸ 67 │   cfg CommandProcessorConfig, ││   L248-312     ││
│                    ││                    ││  68 │ ) *commandProcessorImpl {     ││                ││
│                    ││                    ││  69 │   return &commandProcessorIm… ││ ────────────── ││
│                    ││                    ││  70 │     config: cfg,              ││                ││
│                    ││                    ││  71 │     logger: cfg.Logger,       ││ DOC HITS (2)   ││
│                    ││                    ││  72 │   }                           ││ "Processor" L41││
│                    ││                    ││  73 │ }                             ││ "Processor" L44││
│                    ││                    │├─────────────────────────────────────┤│                ││
│                    │└────────────────────┘│ L45:1  UTF-8  Go  312 lines  8.2KB  ││ [Open in VSC]  ││
│                    │                      └─────────────────────────────────────┘└────────────────┘│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. File Viewer — Diff Overlay Tab

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │┌────────────────────┐┌─────────────────────────────────────┐┌────────────────┐│
│                    ││ FILES              ││ pkg/handlers/processor.go           ││ DIFF CONTEXT   ││
│  🔍 Search         │├────────────────────┤│ [File] [Diff●] [History] [Annotate] │├────────────────┤│
│                    ││                    │├─────────────────────────────────────┤│                ││
│  📂 Explore        ││ 📁 cmd/            ││ Commit: abc1234                     ││ Commit         ││
│   ●├─ Files        ││ 📁 internal/       ││ "Rename Processor to CommandProc…"  ││ ────────────── ││
│    ├─ Symbols      ││ ▼📁 pkg/           │├─────────────────────────────────────┤│ abc1234        ││
│    ├─ Code Units   ││   ▼📁 handlers/ ●  ││                                     ││ Rename Process…││
│    ├─ Commits      ││     📄 doc.go      ││  40 │                               ││ @alice · 3h    ││
│    ├─ Diffs        ││    ●📄 processor.go││  41 │ // CommandProcessor handles   ││                ││
│    ├─ Docs/Terms   ││     📄 router.go   ││  42 │ // incoming command requests  ││ ────────────── ││
│    └─ Tree-sitter  ││     📄 middleware.…││  43 │ // and routes them to the     ││                ││
│                    ││   📁 middleware/   ││  44 │ // appropriate handler.       ││ HUNK 1 of 4   ││
│  🔄 Refactor       ││   📁 models/       ││- 45 │ type Processor interface {    ││ @@ -45,9 +45,9││
│    ├─ Plans        ││   📁 services/     ││+ 45 │ type CommandProcessor interface{│ ────────────── ││
│    ├─ Runs         ││ 📁 test/           ││  46 │   // Process executes cmd     ││                ││
│    ├─ Audits       ││ 📄 go.mod          ││  47 │   Process(ctx context.Context,││ +9  -9  ±0     ││
│    └─ Reports      ││ 📄 go.sum          ││  48 │     cmd Command,              ││                ││
│                    ││ 📄 README.md       ││  49 │   ) (Result, error)           ││ Symbols touched││
│  ⚙️ Data/Admin     ││                    ││  50 │                               ││ ● CommandProc… ││
│    ├─ All Runs     ││                    ││  51 │   // Validate checks cmd      ││ ● NewCommand…  ││
│    ├─ Raw Outputs  ││                    ││  52 │   Validate(cmd Command) error ││ ● commandProc… ││
│    ├─ Schema Info  ││                    ││  53 │ }                             ││                ││
│    └─ Settings     ││                    ││  ...│ ─────────── 11 lines ───────  ││ ────────────── ││
│                    ││                    ││- 65 │ func NewProcessor(            ││                ││
│                    ││                    ││+ 65 │ func NewCommandProcessor(     ││ Hunks          ││
│                    ││                    ││  66 │   cfg CommandProcessorConfig, ││ [1] L45  type  ││
│                    ││                    ││- 67 │ ) *processorImpl {            ││ [2] L65  func  ││
│                    ││                    ││+ 67 │ ) *commandProcessorImpl {     ││ [3] L89  type  ││
│                    ││                    ││- 68 │   return &processorImpl{      ││ [4] L127 method││
│                    ││                    ││+ 68 │   return &commandProcessorIm… ││                ││
│                    ││                    ││  69 │     config: cfg,              ││ ────────────── ││
│                    ││                    ││  70 │     logger: cfg.Logger,       ││                ││
│                    ││                    ││  71 │   }                           ││ [Prev Hunk ↑]  ││
│                    ││                    ││  72 │ }                             ││ [Next Hunk ↓]  ││
│                    │└────────────────────┘└─────────────────────────────────────┘└────────────────┘│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Symbols Explorer

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │┌──────────────┐┌──────────────────────────────────────────────┐┌─────────────┐│
│                    ││ FILTERS      ││ SYMBOLS                          12,456 total││ QUICK VIEW  ││
│  🔍 Search         │├──────────────┤├──────────────────────────────────────────────┤├─────────────┤│
│                    ││              ││ Name ▲       Kind    Package       File    Exp││             ││
│  📂 Explore        ││ Name         ││ ─────────────────────────────────────────────││ CommandProc…││
│    ├─ Files        ││ [Command   ] ││                                              ││ ─────────── ││
│   ●├─ Symbols      ││              ││ ▸ Command       type   pkg/models   models…  ✓││             ││
│    ├─ Code Units   ││ ───────────  ││ ▸ CommandBus    type   pkg/bus      bus.go   ✓││ type Command││
│    ├─ Commits      ││              ││ ● CommandConfig type   pkg/config   conf…    ✓││ Processor   ││
│    ├─ Diffs        ││ Kind         ││ ● CommandProc…  type   pkg/handlers proc…    ✓││ interface { ││
│    ├─ Docs/Terms   ││ ☑ type       ││ ▸ CommandProc…  type   pkg/handlers proc…    ✓││   Process(… ││
│    └─ Tree-sitter  ││ ☑ func       ││ ▸ CommandQueue  type   pkg/queue    queue…   ✓││   Validate… ││
│                    ││ ☑ method     ││ ▸ CommandResult type   pkg/models   result…  ✓││ }           ││
│  🔄 Refactor       ││ ☐ const      ││ ▸ NewCommand    func   pkg/models   models…  ✓││             ││
│    ├─ Plans        ││ ☐ var        ││ ● NewCommand…   func   pkg/handlers proc…    ✓││ ─────────── ││
│    ├─ Runs         ││ ☐ interface  ││ ▸ NewCommand…   func   pkg/queue    queue…   ✓││             ││
│    ├─ Audits       ││              ││ ▸ commandImpl   type   pkg/models   models…  ✗││ pkg/handlers││
│    └─ Reports      ││ ───────────  ││ ● commandProc…  type   pkg/handlers proc…    ✗││ processor.go││
│                    ││              ││ ▸ commandQueue… type   pkg/queue    queue…   ✗││ :45:1       ││
│  ⚙️ Data/Admin     ││ Package      ││                                              ││             ││
│    ├─ All Runs     ││ [pkg/hand ▾] ││ ─────────────────────────────────────────────││ Exported ✓  ││
│    ├─ Raw Outputs  ││              ││ Showing 1-12 of 47 matching    [< 1 2 3 4 >] ││ Hash:       ││
│    ├─ Schema Info  ││ ───────────  │├──────────────────────────────────────────────┤│ a7b3c9f2    ││
│    └─ Settings     ││              ││ ● = changed in session diff                  ││             ││
│                    ││ Exported     ││ ▸ = unchanged                                ││ ─────────── ││
│                    ││ ◉ All        ││                                              ││             ││
│                    ││ ○ Yes only   │└──────────────────────────────────────────────┘│ [Detail]    ││
│                    ││ ○ No only    │                                               │ [Add to Plan]││
│                    ││              │                                               │ [Copy Hash] ││
│                    ││ ───────────  │                                               │ [Copy Spec] ││
│                    ││              │                                               │ [Open File] ││
│                    ││ File Path    │                                               │             ││
│                    ││ [          ] │                                               │             ││
│                    │└──────────────┘                                               └─────────────┘│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Symbol Detail View

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  ← Back to Symbols                                                             │
│                    │                                                                                │
│  🔍 Search         │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│                    │  │  CommandProcessor                                          type  ✓ exp  │  │
│  📂 Explore        │  │  ───────────────────────────────────────────────────────────────────────│  │
│    ├─ Files        │  │  Package:   pkg/handlers                                                │  │
│   ●├─ Symbols      │  │  Location:  pkg/handlers/processor.go:45:1                              │  │
│    ├─ Code Units   │  │  Hash:      a7b3c9f2  [Copy]                                            │  │
│    ├─ Commits      │  │                                                                         │  │
│    ├─ Diffs        │  │  [Add Rename to Plan]  [Run prepare_rename]  [Open in Editor]           │  │
│    ├─ Docs/Terms   │  └─────────────────────────────────────────────────────────────────────────┘  │
│    └─ Tree-sitter  │                                                                                │
│                    │  [Overview] [References] [History] [Audit]                                     │
│  🔄 Refactor       │  ═══════════════════════════════════════════════════════════════════════════  │
│    ├─ Plans        │                                                                                │
│    ├─ Runs         │  DEFINITION                                                                    │
│    ├─ Audits       │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│    └─ Reports      │  │  45 │ type CommandProcessor interface {                                 │  │
│                    │  │  46 │   // Process executes the given command and returns a result.     │  │
│  ⚙️ Data/Admin     │  │  47 │   Process(ctx context.Context, cmd Command) (Result, error)       │  │
│    ├─ All Runs     │  │  48 │                                                                   │  │
│    ├─ Raw Outputs  │  │  49 │   // Validate checks whether the command is valid.                │  │
│    ├─ Schema Info  │  │  50 │   Validate(cmd Command) error                                     │  │
│    └─ Settings     │  │  51 │ }                                                                 │  │
│                    │  └─────────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
│                    │  RELATED CODE UNITS                                                            │
│                    │  ───────────────────────────────────────────────────────────────────────────  │
│                    │   ƒ Process        method  pkg/handlers/processor.go:127-154   [View]         │
│                    │   ƒ Validate       method  pkg/handlers/processor.go:156-198   [View]         │
│                    │                                                                                │
│                    │  QUICK SEARCH                                                                  │
│                    │  ───────────────────────────────────────────────────────────────────────────  │
│                    │   [Search "CommandProcessor" in diffs →]                                       │
│                    │   [Search "CommandProcessor" in docs →]                                        │
│                    │                                                                                │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Symbol Detail — References Tab

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  ← Back to Symbols                                                             │
│                    │                                                                                │
│  🔍 Search         │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│                    │  │  CommandProcessor                                          type  ✓ exp  │  │
│  📂 Explore        │  └─────────────────────────────────────────────────────────────────────────┘  │
│    ├─ Files        │                                                                                │
│   ●├─ Symbols      │  [Overview] [References ●] [History] [Audit]                                   │
│    ├─ Code Units   │  ═══════════════════════════════════════════════════════════════════════════  │
│    ├─ Commits      │                                                                                │
│    ├─ Diffs        │  REFERENCES (47 total)                         ☐ Declarations only            │
│    ├─ Docs/Terms   │  ───────────────────────────────────────────────────────────────────────────  │
│    └─ Tree-sitter  │                                                                                │
│                    │  ▼ pkg/handlers/ (12 refs)                                                     │
│  🔄 Refactor       │    ┌───────────────────────────────────────────────────────────────────────┐  │
│    ├─ Plans        │    │  processor.go:45     type CommandProcessor interface {        decl    │  │
│    ├─ Runs         │    │  processor.go:68     func NewCommandProcessor() CommandProcessor      │  │
│    ├─ Audits       │    │  processor.go:89     var _ CommandProcessor = (*impl)(nil)    ref     │  │
│    └─ Reports      │    │  processor.go:127    func (p *impl) Process() // impl CommandP…       │  │
│                    │    │  router.go:34        proc CommandProcessor                     ref     │  │
│  ⚙️ Data/Admin     │    │  router.go:56        func WithProcessor(p CommandProcessor)   ref     │  │
│    ├─ All Runs     │    │  router.go:78        return r.proc.Process(ctx, cmd)          ref     │  │
│    ├─ Raw Outputs  │    │  middleware.go:23    proc CommandProcessor                     ref     │  │
│    ├─ Schema Info  │    │  middleware.go:45    func Wrap(p CommandProcessor) Handler    ref     │  │
│    └─ Settings     │    │  ... +3 more                                                          │  │
│                    │    └───────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
│                    │  ▼ pkg/services/ (18 refs)                                                     │
│                    │    ┌───────────────────────────────────────────────────────────────────────┐  │
│                    │    │  executor.go:12      proc handlers.CommandProcessor            ref     │  │
│                    │    │  executor.go:34      func New(p handlers.CommandProcessor)     ref     │  │
│                    │    │  executor.go:67      e.proc.Process(ctx, cmd)                  ref     │  │
│                    │    │  ... +15 more                                                         │  │
│                    │    └───────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
│                    │  ▶ pkg/grpc/ (8 refs)                                                          │
│                    │  ▶ cmd/server/ (5 refs)                                                        │
│                    │  ▶ test/ (4 refs)                                                              │
│                    │                                                                                │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Commits Explorer

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │┌──────────────┐┌──────────────────────────────────────────────┐┌─────────────┐│
│                    ││ FILTERS      ││ COMMITS                              20 shown ││ PREVIEW     ││
│  🔍 Search         │├──────────────┤├──────────────────────────────────────────────┤├─────────────┤│
│                    ││              ││ Hash     Subject                   Author Date││             ││
│  📂 Explore        ││ Search       ││ ─────────────────────────────────────────────││ abc1234     ││
│    ├─ Files        ││ [rename    ] ││                                              ││ ─────────── ││
│   ●├─ Commits      ││              ││ ● abc1234 Rename CommandProces…  alice  3h   ││             ││
│    ├─ Symbols      ││ ───────────  ││ ▸ def5678 Fix middleware regis…  bob    5h   ││ Rename      ││
│    ├─ Code Units   ││              ││ ▸ 789abcd Add context propagat…  alice  1d   ││ Command-    ││
│    ├─ Diffs        ││ Author       ││ ▸ cde0123 Deprecate legacy JSO…  carol  2d   ││ Processor   ││
│    ├─ Docs/Terms   ││ [All     ▾]  ││ ▸ 234ef01 Update grpc handlers   alice  2d   ││ to Handler  ││
│    └─ Tree-sitter  ││              ││ ▸ 567ab23 Fix test flakiness     bob    3d   ││             ││
│                    ││ ───────────  ││ ▸ 890cd45 Add retry logic         carol  3d   ││ ─────────── ││
│  🔄 Refactor       ││              ││ ▸ 123ef67 Refactor config load    alice  4d   ││ Author:     ││
│    ├─ Plans        ││ Date Range   ││ ▸ 456ab89 Update dependencies     bob    5d   ││ alice       ││
│    ├─ Runs         ││ From:        ││ ▸ 789cd01 Add metrics endpoint    carol  5d   ││             ││
│    ├─ Audits       ││ [2024-01-01] ││ ... +10 more                                  ││ Date:       ││
│    └─ Reports      ││ To:          │├──────────────────────────────────────────────┤│ 3 hours ago ││
│                    ││ [2024-01-20] ││                                              ││             ││
│  ⚙️ Data/Admin     ││              ││ ● = has indexed diff data                    ││ Files: 8    ││
│    ├─ All Runs     ││ ───────────  ││                                              ││ +142  -87   ││
│    ├─ Raw Outputs  ││              │└──────────────────────────────────────────────┘│             ││
│    ├─ Schema Info  ││ File Path    │                                               │ ─────────── ││
│    └─ Settings     ││ [          ] │                                               │             ││
│                    ││              │                                               │ [View Full] ││
│                    │└──────────────┘                                               │ [Open Diff] ││
│                    │                                                               │ [Copy Hash] ││
│                    │                                                               └─────────────┘│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 9. Commit Detail

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  ← Back to Commits                                                             │
│                    │                                                                                │
│  🔍 Search         │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│                    │  │  abc1234f89e0c1d2e3f4a5b6c7d8e9f0a1b2c3d4                               │  │
│  📂 Explore        │  │  ───────────────────────────────────────────────────────────────────────│  │
│    ├─ Files        │  │  Rename CommandProcessor to Handler                                     │  │
│   ●├─ Commits      │  │                                                                         │  │
│    ├─ Symbols      │  │  Author:    alice <alice@example.com>                                   │  │
│    ├─ Code Units   │  │  Date:      2024-01-15 14:32:07 -0800                                   │  │
│    ├─ Diffs        │  │  Parents:   def5678                                                     │  │
│    ├─ Docs/Terms   │  │                                                                         │  │
│    └─ Tree-sitter  │  │  [Copy Hash]  [Open in GitHub]                                          │  │
│                    │  └─────────────────────────────────────────────────────────────────────────┘  │
│  🔄 Refactor       │                                                                                │
│    ├─ Plans        │  [Overview] [Files ●] [Diff] [Impact]                                         │
│    ├─ Runs         │  ═══════════════════════════════════════════════════════════════════════════  │
│    ├─ Audits       │                                                                                │
│    └─ Reports      │  CHANGED FILES (8)                                          +142  -87         │
│                    │  ───────────────────────────────────────────────────────────────────────────  │
│  ⚙️ Data/Admin     │                                                                                │
│    ├─ All Runs     │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│    ├─ Raw Outputs  │  │ Status   Path                                             +     -       │  │
│    ├─ Schema Info  │  │ ────────────────────────────────────────────────────────────────────── │  │
│    └─ Settings     │  │ M        pkg/handlers/processor.go                        +45   -32    │  │
│                    │  │ M        pkg/handlers/router.go                           +23   -18    │  │
│                    │  │ M        pkg/handlers/middleware.go                       +12   -8     │  │
│                    │  │ M        pkg/services/executor.go                         +34   -21    │  │
│                    │  │ M        pkg/grpc/server.go                               +15   -5     │  │
│                    │  │ M        cmd/server/main.go                               +8    -3     │  │
│                    │  │ M        README.md                                        +3    -0     │  │
│                    │  │ M        docs/api.md                                      +2    -0     │  │
│                    │  └─────────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
│                    │  Click a file to view its diff                                                 │
│                    │                                                                                │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 10. Diffs Explorer

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │┌──────────────┐┌──────────────────────────────────────────────┐┌─────────────┐│
│                    ││ FILTERS      ││ DIFF FILES                        142 changed ││ PREVIEW     ││
│  🔍 Search         │├──────────────┤├──────────────────────────────────────────────┤├─────────────┤│
│                    ││              ││ Path                         Status  +    -   ││             ││
│  📂 Explore        ││ Status       ││ ─────────────────────────────────────────────││ processor.go││
│    ├─ Files        ││ ☑ Modified   ││                                              ││ ─────────── ││
│   ●├─ Diffs        ││ ☑ Added      ││ ● pkg/handlers/processor.go    M    +89  -67 ││             ││
│    ├─ Symbols      ││ ☑ Deleted    ││ ▸ pkg/handlers/router.go       M    +45  -32 ││ 4 hunks     ││
│    ├─ Code Units   ││ ☑ Renamed    ││ ▸ pkg/handlers/middleware.go   M    +23  -18 ││ +89  -67    ││
│    ├─ Commits      ││              ││ ▸ pkg/services/executor.go     M    +67  -45 ││             ││
│    ├─ Docs/Terms   ││ ───────────  ││ ▸ pkg/services/config.go       M    +12  -8  ││ ─────────── ││
│    └─ Tree-sitter  ││              ││ ▸ pkg/grpc/server.go           M    +34  -21 ││             ││
│                    ││ Path Filter  ││ ▸ pkg/grpc/handler.go          M    +28  -15 ││ @@ -45,9    ││
│  🔄 Refactor       ││ [handlers  ] ││ ▸ cmd/server/main.go           M    +15  -10 ││ +45,9 @@    ││
│    ├─ Plans        ││              ││ ▸ internal/util/helpers.go     A    +45  -0  ││             ││
│    ├─ Runs         ││ ───────────  ││ ▸ internal/legacy/old.go       D    +0   -89 ││ -type Proc  ││
│    ├─ Audits       ││              ││ ▸ README.md                    M    +8   -3  ││ +type Cmd…  ││
│    └─ Reports      ││ Change Size  ││ ▸ docs/api.md                  M    +12  -5  ││  // Process ││
│                    ││ Min: [    0] ││ ... +130 more files                          ││  Process(…  ││
│  ⚙️ Data/Admin     ││ Max: [  999] ││                                              ││             ││
│    ├─ All Runs     ││              │├──────────────────────────────────────────────┤│ ─────────── ││
│    ├─ Raw Outputs  ││ ───────────  ││                                              ││             ││
│    ├─ Schema Info  ││              ││ Session total: +4,892  -2,103  142 files     ││ [Open Full] ││
│    └─ Settings     ││ Sort By      ││                                              ││ [View File] ││
│                    ││ [Changes ▾]  │└──────────────────────────────────────────────┘│ [Copy Path] ││
│                    │└──────────────┘                                               └─────────────┘│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 11. Diff Detail (Full File View)

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  ← Back to Diffs                                                               │
│                    │                                                                                │
│  🔍 Search         │  pkg/handlers/processor.go                                     +89  -67        │
│                    │  ═══════════════════════════════════════════════════════════════════════════  │
│  📂 Explore        │                                                                                │
│    ├─ Files        │  View: [Unified ▾]    Show: [All ▾]    [◀ Prev Hunk] [Next Hunk ▶]            │
│   ●├─ Diffs        │                                                                                │
│    ├─ Symbols      │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│    ├─ Code Units   │  │  OLD   NEW │                                                            │  │
│    ├─ Commits      │  │  ──────────┼────────────────────────────────────────────────────────────│  │
│    ├─ Docs/Terms   │  │            │  @@ -40,15 +40,15 @@ package handlers                     │  │
│    └─ Tree-sitter  │  │            │                                                            │  │
│                    │  │   40    40 │  // CommandProcessor handles incoming command requests     │  │
│  🔄 Refactor       │  │   41    41 │  // and routes them to the appropriate handler.            │  │
│    ├─ Plans        │  │   42    42 │                                                            │  │
│    ├─ Runs         │  │   43       │- // Processor is the main interface for command handling.  │  │
│    ├─ Audits       │  │        43  │+ // CommandProcessor is the main interface.                │  │
│    └─  Reports     │  │   44       │- type Processor interface {                                │  │
│                    │  │        44  │+ type CommandProcessor interface {                         │  │
│  ⚙️ Data/Admin     │  │   45    45 │    // Process executes the given command.                  │  │
│    ├─ All Runs     │  │   46    46 │    Process(ctx context.Context, cmd Command) (Result, err) │  │
│    ├─ Raw Outputs  │  │   47    47 │                                                            │  │
│    ├─ Schema Info  │  │   48    48 │    // Validate checks whether the command is valid.        │  │
│    └─ Settings     │  │   49    49 │    Validate(cmd Command) error                             │  │
│                    │  │   50    50 │  }                                                         │  │
│                    │  │            │                                                            │  │
│                    │  │            │  @@ -64,12 +64,12 @@ type CommandProcessorConfig struct { │  │
│                    │  │            │                                                            │  │
│                    │  │   64    64 │  // NewCommandProcessor creates a new processor instance.  │  │
│                    │  │   65       │- func NewProcessor(cfg Config) *processorImpl {            │  │
│                    │  │        65  │+ func NewCommandProcessor(cfg Config) *commandProcImpl {   │  │
│                    │  │   66       │-   return &processorImpl{                                  │  │
│                    │  │        66  │+   return &commandProcessorImpl{                           │  │
│                    │  │   67    67 │      config: cfg,                                          │  │
│                    │  │   68    68 │      logger: cfg.Logger,                                   │  │
│                    │  │   69    69 │    }                                                       │  │
│                    │  │   70    70 │  }                                                         │  │
│                    │  └─────────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
│                    │  Hunk 1 of 4      [Open in File Viewer]  [Copy Hunk]  [Search in Hunk]        │
│                    │                                                                                │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 12. Docs / Terms Explorer

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  View: [● Terms] [Files]                                                       │
│                    │                                                                                │
│  🔍 Search         │┌──────────────┐┌──────────────────────────────────────────────┐┌─────────────┐│
│                    ││ FILTERS      ││ TERMS                              23 unique ││ HITS        ││
│  📂 Explore        │├──────────────┤├──────────────────────────────────────────────┤├─────────────┤│
│    ├─ Files        ││              ││ Term                              Hits  Files││             ││
│    ├─ Symbols      ││ Search       ││ ─────────────────────────────────────────────││ "Processor" ││
│    ├─ Code Units   ││ [          ] ││                                              ││ 89 hits     ││
│    ├─ Commits      ││              ││ ● Processor                         89    18 ││ ─────────── ││
│    ├─ Diffs        ││ ───────────  ││ ▸ CommandHandler                    45    12 ││             ││
│   ●├─ Docs/Terms   ││              ││ ▸ legacy_config                     34     8 ││ README.md:45││
│    └─ Tree-sitter  ││ File Type    ││ ▸ oldHandler                        23     6 ││ "The Proces-││
│                    ││ ☑ Markdown   ││ ▸ deprecated_api                    18     4 ││  sor interfa-│
│  🔄 Refactor       ││ ☑ Go         ││ ▸ v1_endpoint                       12     3 ││  ce handles" ││
│    ├─ Plans        ││ ☑ YAML       ││ ... +17 more                                 ││             ││
│    ├─ Runs         ││ ☐ JSON       │├──────────────────────────────────────────────┤│ ─────────── ││
│    ├─ Audits       ││              ││                                              ││             ││
│    └─ Reports      ││ ───────────  ││ ⚠ 89 hits of "Processor" may need updating   ││ docs/api.md ││
│                    ││              ││                                              ││ :23         ││
│  ⚙️ Data/Admin     ││ Path Filter  │└──────────────────────────────────────────────┘│ "Call the   ││
│    ├─ All Runs     ││ [docs/     ] │                                               │  Processor  ││
│    ├─ Raw Outputs  ││              │                                               │  method to" ││
│    ├─ Schema Info  ││              │                                               │             ││
│    └─ Settings     ││              │                                               │ ─────────── ││
│                    ││              │                                               │             ││
│                    ││              │                                               │ [View File] ││
│                    ││              │                                               │ [Add Rule]  ││
│                    ││              │                                               │ [Mark Done] ││
│                    │└──────────────┘                                               └─────────────┘│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 13. Runs List (Data/Admin)

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│                    │                                                                                │
│  📊 Dashboard      │  INDEXING RUNS                                                                 │
│                    │  ═══════════════════════════════════════════════════════════════════════════  │
│  🔍 Search         │                                                                                │
│                    │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  📂 Explore        │  │ ID   Status    Kind          Range              Started      Rows       │  │
│    ├─ Files        │  │ ────────────────────────────────────────────────────────────────────── │  │
│    ├─ Symbols      │  │ #45  ✅ success doc-hits     HEAD~20→HEAD       1h ago       847       │  │
│    ├─ Code Units   │  │ #44  ✅ success symbols      HEAD~20→HEAD       1h ago       12,456    │  │
│    ├─ Commits      │  │ #44  ✅ success code-units   HEAD~20→HEAD       1h ago       8,234     │  │
│    ├─ Diffs        │  │ #43  ✅ success diffs        HEAD~20→HEAD       2h ago       3,291     │  │
│    ├─ Docs/Terms   │  │ #42  ✅ success commits      HEAD~20→HEAD       2h ago       1,847     │  │
│    └─ Tree-sitter  │  │ #41  ⛔ failed  tree-sitter  HEAD~20→HEAD       3h ago       0         │  │
│                    │  │ #40  ✅ success gopls-refs   HEAD~30→HEAD~20    1d ago       23,456    │  │
│  🔄 Refactor       │  │ #39  ✅ success symbols      HEAD~30→HEAD~20    1d ago       11,892    │  │
│    ├─ Plans        │  │ ... +32 more                                                           │  │
│    ├─ Runs         │  └─────────────────────────────────────────────────────────────────────────┘  │
│    ├─ Audits       │                                                                                │
│    └─ Reports      │  ┌─────────────────────────────────────────────────────────────────────────┐  │
│                    │  │ RUN #41 DETAIL                                                          │  │
│  ⚙️ Data/Admin     │  │ ─────────────────────────────────────────────────────────────────────── │  │
│   ●├─ All Runs     │  │                                                                         │  │
│    ├─ Raw Outputs  │  │  Kind:       tree-sitter                                                │  │
│    ├─ Schema Info  │  │  Status:     ⛔ failed                                                  │  │
│    └─ Settings     │  │  Started:    2024-01-15 11:32:07                                        │  │
│                    │  │  Finished:   2024-01-15 11:32:12                                        │  │
│                    │  │  Duration:   5s                                                         │  │
│                    │  │                                                                         │  │
│                    │  │  Error:                                                                 │  │
│                    │  │  ┌────────────────────────────────────────────────────────────────────┐ │  │
│                    │  │  │ tree-sitter query failed: invalid capture @type.definition        │ │  │
│                    │  │  │ at queries/go.scm:45                                               │ │  │
│                    │  │  └────────────────────────────────────────────────────────────────────┘ │  │
│                    │  │                                                                         │  │
│                    │  │  [View Raw Output]  [Retry Run]  [Attach to Session]                    │  │
│                    │  └─────────────────────────────────────────────────────────────────────────┘  │
│                    │                                                                                │
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## 14. Command Palette (Modal Overlay)

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  🔧 Refactorio Workbench      [glazed ▾]  [main: HEAD~20→HEAD ▾]  [🔍 Search…                      ]  [⌘K]  │
├────────────────────┬────────────────────────────────────────────────────────────────────────────────┤
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░┌───────────────────────────────────────────────────┐░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  🔍 proc                                      ⌘K │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░├───────────────────────────────────────────────────┤░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│                                                   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  SYMBOLS                                          │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ CommandProcessor         type   pkg/handlers   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ commandProcessorImpl     type   pkg/handlers   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ NewCommandProcessor      func   pkg/handlers   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ Process                  method pkg/handlers   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│                                                   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  FILES                                            │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ pkg/handlers/processor.go                      │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ pkg/services/processor_test.go                 │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│                                                   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  COMMANDS                                         │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ Search in diffs…                         ⌘⇧D  │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ Search in docs…                          ⌘⇧O  │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ New refactor plan…                       ⌘⇧P  │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ▸ Run audit…                               ⌘⇧A  │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│                                                   │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░│  ↑↓ Navigate   Enter Select   Esc Close          │░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░└───────────────────────────────────────────────────┘░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░░░░░░░░░░░░░░░░░░░░│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
└────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
```

---

## Design Notes Summary

| Pattern | Usage |
|---------|-------|
| **3-Pane Layout** | Filters → List → Preview for all exploration views |
| **Consistent badges** | ● changed, ▸ normal, ⚠️ warning, ⛔ error, ✓ exported |
| **Keyboard-first** | ↑/↓ navigate, Enter open, ⌘K palette, Esc back |
| **Deep links everywhere** | Every entity is URL-addressable with session context |
| **Copy actions** | Hash, spec, path, path:line always one click away |
| **Missing data is visible** | Gray states with CTAs to compute/retry |
| **Session context always shown** | Top bar shows workspace + session at all times |

---

# Screen 1: App Shell & Session Dashboard

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# APP SHELL (persistent wrapper for all screens)
# ═══════════════════════════════════════════════════════════════════════════════

AppShell:
  state:
    - currentWorkspace: Workspace | null
    - currentSession: Session | null
    - commandPaletteOpen: boolean

  data:
    workspaces:
      endpoint: GET /api/workspaces
      response: Workspace[]
      cache: persist-local
    
  children:
    - TopBar:
        bindings:
          workspace: currentWorkspace
          session: currentSession
        children:
          - Logo
          - WorkspaceSelector:
              data: workspaces
              selected: currentWorkspace
              onSelect: setCurrentWorkspace
          - SessionSelector:
              endpoint: GET /api/workspaces/{workspaceId}/sessions
              depends: currentWorkspace
              selected: currentSession
              onSelect: setCurrentSession
          - GlobalSearchBar:
              onSubmit: navigate(/search?q={query})
              placeholder: "Search symbols, files, diffs…"
          - CommandPaletteButton:
              shortcut: ⌘K
              onClick: toggleCommandPalette

    - SideNav:
        sections:
          - item: Dashboard, icon: 📊, route: /
          - item: Search, icon: 🔍, route: /search
          - group: Explore
            items:
              - Files, route: /explore/files
              - Symbols, route: /explore/symbols
              - CodeUnits, route: /explore/code-units
              - Commits, route: /explore/commits
              - Diffs, route: /explore/diffs
              - Docs/Terms, route: /explore/docs
              - TreeSitter, route: /explore/tree-sitter
          - group: Refactor
            items:
              - Plans, route: /refactor/plans
              - Runs, route: /refactor/runs
              - Audits, route: /refactor/audits
              - Reports, route: /refactor/reports
          - group: Data/Admin
            items:
              - AllRuns, route: /data/runs
              - RawOutputs, route: /data/raw-outputs
              - SchemaInfo, route: /data/schema
              - Settings, route: /settings

    - MainContent:
        renders: <Outlet />  # React Router

    - CommandPalette:
        when: commandPaletteOpen
        # (detailed in Screen 14)


# ═══════════════════════════════════════════════════════════════════════════════
# SESSION DASHBOARD (route: /)
# ═══════════════════════════════════════════════════════════════════════════════

SessionDashboard:
  route: /
  params: none
  requires: [currentWorkspace, currentSession]

  data:
    dbInfo:
      endpoint: GET /api/workspaces/{workspaceId}/db/info
      response:
        schemaVersion: string
        dbPath: string
        repoRoot: string | null
        tableCounts: Record<string, number>
        ftsEnabled: string[]

    sessionDetail:
      endpoint: GET /api/sessions/{sessionId}
      response:
        id: string
        name: string
        gitFrom: string
        gitTo: string
        createdAt: timestamp
        updatedAt: timestamp
        passes: PassStatus[]

    sessionStats:
      endpoint: GET /api/sessions/{sessionId}/stats
      response:
        filesModified: number
        packagesTouched: number
        linesAdded: number
        linesRemoved: number
        newSymbols: number
        deletedSymbols: number
        docHitsLegacy: number
        contributors: number

    recentCommits:
      endpoint: GET /api/sessions/{sessionId}/commits?limit=5&sort=date:desc
      response: CommitSummary[]

  types:
    PassStatus:
      kind: enum[commits, diffs, symbols, code-units, doc-hits, gopls-refs, tree-sitter]
      status: enum[success, failed, missing, running]
      runId: string | null
      rowCount: number | null
      updatedAt: timestamp | null
      errorSummary: string | null

    CommitSummary:
      hash: string
      subject: string
      author: string
      date: timestamp

  children:
    - PageHeader:
        title: "SESSION: {session.name}"
        subtitle: "{session.gitFrom} → {session.gitTo}"

    - WorkspaceInfoCard:
        bindings:
          name: currentWorkspace.name
          dbPath: dbInfo.dbPath
          repoRoot: dbInfo.repoRoot
          schemaVersion: dbInfo.schemaVersion
        render:
          - LabelValue: Workspace, {name}
          - LabelValue: DB Path, {dbPath}
          - LabelValue: Repo Root, {repoRoot}
          - SchemaVersionBadge: {schemaVersion}

    - DataAvailabilityCard:
        title: "DATA AVAILABILITY"
        bindings:
          passes: sessionDetail.passes
        children:
          - PassStatusList:
              items: passes
              renderItem: PassStatusRow
        
        PassStatusRow:
          bindings:
            pass: PassStatus
          render:
            - StatusIcon: pass.status  # ✅ ⚠️ ⛔ 🔄
            - PassKindLabel: pass.kind
            - RowCount: pass.rowCount
            - RunLink: pass.runId
            - RelativeTime: pass.updatedAt
            - ConditionalActions:
                when: pass.status == 'missing'
                render: ComputeButton(kind: pass.kind)
                when: pass.status == 'failed'
                render: [ViewErrorButton(runId: pass.runId), RetryButton]

    - QuickStatsCard:
        title: "QUICK STATS"
        bindings: sessionStats
        layout: grid(2x4)
        children:
          - StatTile: label=Files Modified, value={filesModified}
          - StatTile: label=Packages Touched, value={packagesTouched}
          - StatTile: label=Lines Added, value={linesAdded}, color=green
          - StatTile: label=Lines Removed, value={linesRemoved}, color=red
          - StatTile: label=New Symbols, value={newSymbols}
          - StatTile: label=Deleted Symbols, value={deletedSymbols}
          - StatTile: label=Doc Hits (legacy), value={docHitsLegacy}
          - StatTile: label=Contributors, value={contributors}

    - RecentActivityCard:
        title: "RECENT ACTIVITY"
        bindings:
          commits: recentCommits
        children:
          - CommitList:
              items: commits
              renderItem: CommitRow
              onItemClick: navigate(/explore/commits/{hash})
          
        CommitRow:
          render:
            - CommitHash: short=true
            - CommitSubject: truncate=50
            - AuthorBadge
            - RelativeTime


# ═══════════════════════════════════════════════════════════════════════════════
# SHARED TYPES (referenced across screens)
# ═══════════════════════════════════════════════════════════════════════════════

SharedTypes:
  Workspace:
    id: string
    name: string
    dbPath: string
    repoRoot: string | null

  Session:
    id: string
    workspaceId: string
    name: string
    gitFrom: string
    gitTo: string
    createdAt: timestamp
    updatedAt: timestamp
```

---

# Screen 2: Unified Search

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# UNIFIED SEARCH (route: /search)
# ═══════════════════════════════════════════════════════════════════════════════

UnifiedSearchPage:
  route: /search
  queryParams:
    q: string           # search query
    types: string[]     # comma-separated: symbols,code-units,diffs,commits,docs,files
    kind: string[]      # func,type,method,const,var
    package: string
    filePath: string
    runId: string

  state:
    query: string
    selectedTypes: Set<EntityType>
    filters: SearchFilters
    selectedResult: SearchResult | null

  data:
    searchResults:
      endpoint: POST /api/search
      body:
        query: string
        sessionId: string
        types: EntityType[]
        filters:
          kind: SymbolKind[]
          package: string
          filePath: string
          runId: string
        limit: number
        offset: number
      response:
        total: number
        byType: Record<EntityType, { count: number, results: SearchResult[] }>
      debounce: 300ms

  types:
    EntityType: enum[symbols, code-units, diffs, commits, docs, files, tree-sitter]
    
    SearchResult:
      type: EntityType
      id: string
      primaryLabel: string
      secondaryLabel: string
      snippet: string          # with <mark> highlights
      location: Location | null
      runId: string
      commitHash: string | null
      metadata: Record<string, any>

    Location:
      filePath: string
      line: number
      column: number

  children:
    - ThreePaneLayout:
        leftWidth: 200px
        rightWidth: 280px

        left: SearchFiltersPanel
        center: SearchResultsPanel
        right: SearchPreviewPanel

    # ─────────────────────────────────────────────────────────────────────────
    - SearchFiltersPanel:
        children:
          - FilterSection:
              title: Types
              children:
                - CheckboxGroup:
                    options: [Symbols, CodeUnits, Diffs, Commits, Docs, Files, TreeSitter]
                    selected: selectedTypes
                    onChange: setSelectedTypes

          - FilterSection:
              title: Kind
              when: selectedTypes.has('symbols') || selectedTypes.has('code-units')
              children:
                - CheckboxGroup:
                    options: [func, type, method, const, var]
                    selected: filters.kind
                    onChange: updateFilters('kind')

          - FilterSection:
              title: Package
              children:
                - SelectDropdown:
                    endpoint: GET /api/sessions/{sessionId}/packages
                    selected: filters.package
                    onChange: updateFilters('package')
                    placeholder: "Any"

          - FilterSection:
              title: File Path
              children:
                - TextInput:
                    value: filters.filePath
                    onChange: updateFilters('filePath')
                    placeholder: "glob pattern"

          - FilterSection:
              title: Run
              children:
                - SelectDropdown:
                    endpoint: GET /api/sessions/{sessionId}/runs
                    labelKey: "#{id} {kind}"
                    selected: filters.runId
                    onChange: updateFilters('runId')

    # ─────────────────────────────────────────────────────────────────────────
    - SearchResultsPanel:
        bindings:
          results: searchResults
          selected: selectedResult
        children:
          - SearchInput:
              value: query
              onChange: setQuery
              autoFocus: true

          - ResultsHeader:
              total: results.total
              query: query

          - GroupedResultsList:
              groups: results.byType
              renderGroup: ResultTypeGroup
              
        ResultTypeGroup:
          props: { type: EntityType, count: number, results: SearchResult[] }
          state: expanded: boolean (default: true)
          render:
            - CollapsibleHeader:
                icon: entityTypeIcon(type)
                label: "{type} ({count})"
                expanded: expanded
                onToggle: setExpanded
            - when: expanded
              render:
                - VirtualList:
                    items: results
                    renderItem: SearchResultRow
                    onItemClick: setSelectedResult
                    selectedId: selectedResult?.id

        SearchResultRow:
          props: result: SearchResult
          render:
            - EntityTypeIcon: result.type
            - PrimaryLabel: result.primaryLabel
            - SecondaryLabel: result.secondaryLabel, muted=true
            - SnippetPreview: result.snippet, highlight=true
            - LocationBadge: result.location

    # ─────────────────────────────────────────────────────────────────────────
    - SearchPreviewPanel:
        bindings:
          result: selectedResult
        when: selectedResult != null
        children:
          - PreviewHeader:
              icon: entityTypeIcon(result.type)
              label: result.type

          - ConditionalPreview:
              switch: result.type
              
              case symbols:
                - SymbolQuickPreview:
                    endpoint: GET /api/symbols/{result.id}
                    response: SymbolDetail
                    render:
                      - SymbolName
                      - SymbolSignature: code=true
                      - LocationLink
                      - ExportedBadge
                      - HashCopyable

              case code-units:
                - CodeUnitQuickPreview:
                    endpoint: GET /api/code-units/{result.id}
                    render:
                      - CodeSnippet: lines=15
                      - LocationLink

              case diffs:
                - DiffLinePreview:
                    render:
                      - HunkContext: lines=5
                      - DiffKindBadge  # +/-/context

              case commits:
                - CommitQuickPreview:
                    render:
                      - CommitSubject
                      - AuthorDate
                      - FilesChangedCount

              case docs:
                - DocHitPreview:
                    render:
                      - MatchText: highlight=true
                      - FileLocationLink

          - Divider

          - ActionButtonGroup:
              - Button: "Open Detail", onClick: navigateToDetail(result)
              - Button: "Add to Plan", onClick: addToPlan(result)
              - CopyButton: "Copy Hash", value: result.id
              - Button: "Open in Editor", onClick: openInEditor(result.location)

  keyboardShortcuts:
    ArrowUp: selectPreviousResult
    ArrowDown: selectNextResult
    Enter: navigateToDetail(selectedResult)
    Cmd+C: copySelectedPath
    Cmd+Shift+C: copySelectedSpec
    Escape: clearSelection
```

---

# Screen 3: Files Explorer + File Viewer

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# FILES EXPLORER (route: /explore/files)
# ═══════════════════════════════════════════════════════════════════════════════

FilesExplorerPage:
  route: /explore/files
  queryParams:
    path: string        # selected file path
    ref: string         # git ref (default: HEAD)
    line: number        # highlight line
    tab: enum[file, diff, history, annotate]

  state:
    selectedPath: string | null
    expandedDirs: Set<string>
    activeTab: TabType
    highlightLine: number | null

  data:
    fileTree:
      endpoint: GET /api/sessions/{sessionId}/files/tree
      queryParams:
        prefix: string
        depth: 2
      response:
        nodes: FileTreeNode[]
      cache: session-scoped

    fileContent:
      endpoint: GET /api/files/content
      queryParams:
        path: string
        ref: string
      response:
        content: string
        encoding: string
        lineCount: number
        size: number
      depends: selectedPath

    fileContext:
      endpoint: GET /api/sessions/{sessionId}/files/context
      queryParams:
        path: string
      response:
        symbols: SymbolOccurrence[]
        codeUnits: CodeUnitSummary[]
        docHits: DocHit[]
        lastCommit: CommitSummary
        diffStatus: DiffStatus | null
      depends: selectedPath

  types:
    FileTreeNode:
      name: string
      path: string
      type: enum[file, directory]
      children: FileTreeNode[] | null  # lazy loaded
      badges:
        hasDiffChanges: boolean
        docHitCount: number
        lastCommitDate: timestamp | null

    SymbolOccurrence:
      symbolHash: string
      name: string
      kind: SymbolKind
      line: number
      column: number
      exported: boolean

    CodeUnitSummary:
      hash: string
      name: string
      kind: enum[func, type, method]
      startLine: number
      endLine: number

    DocHit:
      term: string
      line: number
      matchText: string

    DiffStatus:
      status: enum[A, M, D, R]
      hunksCount: number
      linesAdded: number
      linesRemoved: number

  children:
    - ThreePaneLayout:
        leftWidth: 240px
        rightWidth: 260px
        
        left: FileTreePanel
        center: FileViewerPanel
        right: FileContextPanel

    # ─────────────────────────────────────────────────────────────────────────
    - FileTreePanel:
        children:
          - TreeFilter:
              placeholder: "Filter by extension"
              value: extensionFilter
              onChange: setExtensionFilter

          - FileTree:
              data: fileTree.nodes
              expanded: expandedDirs
              selected: selectedPath
              onExpand: handleExpand
              onSelect: setSelectedPath
              renderNode: FileTreeNodeRow
              lazyLoad:
                endpoint: GET /api/sessions/{sessionId}/files/tree?prefix={path}

          - TreeLegend:
              items:
                - icon: ●, label: "has diff changes"
                - icon: ◐, label: "has doc hits"

        FileTreeNodeRow:
          props: node: FileTreeNode
          render:
            - FolderIcon | FileIcon: node.type
            - NodeName: node.name
            - BadgeGroup:
                - DiffBadge: when node.badges.hasDiffChanges
                - DocHitBadge: when node.badges.docHitCount > 0

    # ─────────────────────────────────────────────────────────────────────────
    - FileViewerPanel:
        when: selectedPath != null
        bindings:
          path: selectedPath
          content: fileContent
          context: fileContext
        children:
          - FileViewerHeader:
              path: selectedPath
              children:
                - TabBar:
                    tabs: [File, Diff, History, Annotate]
                    active: activeTab
                    onChange: setActiveTab
                    badges:
                      Diff: context.diffStatus != null

          - TabContent:
              switch: activeTab

              case file:
                - CodeViewer:
                    content: fileContent.content
                    language: inferLanguage(selectedPath)
                    lineNumbers: true
                    highlightLine: highlightLine
                    onLineClick: setHighlightLine

              case diff:
                - DiffOverlayViewer:
                    endpoint: GET /api/sessions/{sessionId}/diffs/file
                    queryParams:
                      path: selectedPath
                    render:
                      - CommitSelector:
                          when: multipleCommits
                      - InlineDiffView:
                          baseContent: fileContent.content
                          hunks: diffData.hunks
                          onHunkClick: scrollToHunk

              case history:
                - FileHistoryList:
                    endpoint: GET /api/sessions/{sessionId}/files/history
                    queryParams:
                      path: selectedPath
                    response: CommitSummary[]
                    renderItem: CommitRow
                    onItemClick: showCommitDiff

              case annotate:
                - AnnotatedCodeViewer:
                    endpoint: GET /api/files/blame
                    queryParams:
                      path: selectedPath
                      ref: ref
                    render:
                      - BlameGutter: commit per line
                      - CodeContent

          - FileViewerFooter:
              render:
                - LocationDisplay: "L{highlightLine}:1"
                - EncodingBadge: fileContent.encoding
                - LanguageBadge
                - LineCountBadge: fileContent.lineCount
                - FileSizeBadge: fileContent.size

    # ─────────────────────────────────────────────────────────────────────────
    - FileContextPanel:
        when: selectedPath != null
        bindings:
          context: fileContext
        children:
          - ContextSection:
              title: "SYMBOLS ({context.symbols.length})"
              children:
                - SymbolList:
                    items: context.symbols
                    renderItem: SymbolContextRow
                    onItemClick: scrollToLine(item.line)

          - ContextSection:
              title: "CODE UNITS ({context.codeUnits.length})"
              children:
                - CodeUnitList:
                    items: context.codeUnits
                    renderItem: CodeUnitContextRow
                    onItemClick: scrollToLine(item.startLine)

          - ContextSection:
              title: "DOC HITS ({context.docHits.length})"
              when: context.docHits.length > 0
              children:
                - DocHitList:
                    items: context.docHits
                    renderItem: DocHitContextRow
                    onItemClick: scrollToLine(item.line)

          - Divider

          - ActionButtonGroup:
              - CopyButton: "Copy path", value: selectedPath
              - CopyButton: "Copy path:line", value: "{selectedPath}:{highlightLine}"
              - OpenInEditorButton: path: selectedPath, line: highlightLine

        SymbolContextRow:
          render:
            - SymbolKindIcon
            - SymbolName
            - LineNumber: muted=true

        CodeUnitContextRow:
          render:
            - CodeUnitKindIcon
            - CodeUnitName
            - LineRange: "{startLine}-{endLine}", muted=true

        DocHitContextRow:
          render:
            - TermBadge
            - LineNumber
```

---

# Screen 4: File Viewer — Diff Overlay Tab

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# FILE VIEWER - DIFF OVERLAY (nested within FilesExplorerPage)
# ═══════════════════════════════════════════════════════════════════════════════

# This is the "Diff" tab content within FileViewerPanel
# Reuses the same route: /explore/files?path={path}&tab=diff

DiffOverlayTab:
  parent: FileViewerPanel
  activeWhen: activeTab == 'diff'

  state:
    selectedCommit: string | null
    currentHunkIndex: number

  data:
    availableCommits:
      endpoint: GET /api/sessions/{sessionId}/commits/for-file
      queryParams:
        path: selectedPath
      response: CommitSummary[]

    fileDiff:
      endpoint: GET /api/sessions/{sessionId}/diffs/file
      queryParams:
        path: selectedPath
        commitHash: selectedCommit  # optional, defaults to session range
      response:
        commitHash: string
        commitSubject: string
        author: string
        date: timestamp
        hunks: DiffHunk[]
        linesAdded: number
        linesRemoved: number
        symbolsTouched: SymbolSummary[]

  types:
    DiffHunk:
      index: number
      oldStart: number
      oldCount: number
      newStart: number
      newCount: number
      lines: DiffLine[]

    DiffLine:
      type: enum[context, add, remove]
      oldLineNo: number | null
      newLineNo: number | null
      content: string

    SymbolSummary:
      hash: string
      name: string
      kind: SymbolKind
      line: number

  children:
    - DiffOverlayHeader:
        bindings:
          diff: fileDiff
          commits: availableCommits
        children:
          - CommitSelector:
              when: availableCommits.length > 1
              options: availableCommits
              selected: selectedCommit
              onChange: setSelectedCommit
              renderOption: "{hash.short} - {subject.truncate(40)}"

          - CommitInfo:
              render:
                - CommitHashBadge: fileDiff.commitHash
                - CommitSubject: fileDiff.commitSubject
                - AuthorBadge: fileDiff.author
                - RelativeTime: fileDiff.date

    - InlineDiffViewer:
        bindings:
          hunks: fileDiff.hunks
          currentHunk: currentHunkIndex
        render:
          - ForEach: hunks
            renderItem: DiffHunkBlock
            gapBetween: CollapsedLinesIndicator

        DiffHunkBlock:
          props: hunk: DiffHunk, index: number
          render:
            - HunkHeader: "@@ -{hunk.oldStart},{hunk.oldCount} +{hunk.newStart},{hunk.newCount} @@"
            - ForEach: hunk.lines
              renderItem: DiffLineRow

        DiffLineRow:
          props: line: DiffLine
          render:
            - LineGutter:
                oldLineNo: line.oldLineNo
                newLineNo: line.newLineNo
            - LineContent:
                content: line.content
                className:
                  add: "bg-green-50 text-green-900"
                  remove: "bg-red-50 text-red-900"
                  context: ""

        CollapsedLinesIndicator:
          props: count: number
          render: "─────────── {count} lines ───────────"

    - DiffContextSidebar:
        # Rendered in the right panel (FileContextPanel) when diff tab active
        bindings:
          diff: fileDiff
        children:
          - ContextSection:
              title: "Commit"
              render:
                - CommitHashLink: diff.commitHash
                - CommitSubject
                - AuthorBadge
                - RelativeTime

          - ContextSection:
              title: "Hunk {currentHunkIndex + 1} of {diff.hunks.length}"
              render:
                - HunkRangeDisplay

          - ContextSection:
              title: "Changes"
              render:
                - StatRow: "+{diff.linesAdded}", color=green
                - StatRow: "-{diff.linesRemoved}", color=red

          - ContextSection:
              title: "Symbols Touched"
              when: diff.symbolsTouched.length > 0
              children:
                - SymbolList:
                    items: diff.symbolsTouched
                    renderItem: SymbolBadgeRow
                    onItemClick: navigate(/explore/symbols/{hash})

          - Divider

          - HunkNavigator:
              children:
                - NavigationLabel: "Hunks"
                - HunkJumpList:
                    items: fileDiff.hunks
                    renderItem: "[{index}] L{newStart}"
                    onItemClick: scrollToHunk(index)

          - ButtonGroup:
              - Button: "◀ Prev Hunk", onClick: prevHunk, disabled: currentHunkIndex == 0
              - Button: "Next Hunk ▶", onClick: nextHunk, disabled: currentHunkIndex == last

  keyboardShortcuts:
    "[": prevHunk
    "]": nextHunk
    "g": promptGoToHunk
```

---

# Screen 5: Symbols Explorer

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# SYMBOLS EXPLORER (route: /explore/symbols)
# ═══════════════════════════════════════════════════════════════════════════════

SymbolsExplorerPage:
  route: /explore/symbols
  queryParams:
    q: string           # name filter
    kind: string[]      # func,type,method,const,var
    package: string
    exported: boolean
    filePath: string
    page: number
    pageSize: number

  state:
    filters: SymbolFilters
    selectedSymbol: SymbolSummary | null
    sortColumn: string
    sortDirection: enum[asc, desc]

  data:
    symbols:
      endpoint: GET /api/sessions/{sessionId}/symbols
      queryParams:
        name: filters.name
        kind: filters.kind
        package: filters.package
        exported: filters.exported
        filePath: filters.filePath
        sort: "{sortColumn}:{sortDirection}"
        page: page
        pageSize: pageSize
      response:
        total: number
        page: number
        pageSize: number
        items: SymbolSummary[]

    packages:
      endpoint: GET /api/sessions/{sessionId}/packages
      response: string[]
      cache: session-scoped

    symbolPreview:
      endpoint: GET /api/symbols/{selectedSymbol.hash}/preview
      depends: selectedSymbol
      response:
        signature: string
        docComment: string
        snippet: string
        refsCount: number | null
        hasRefsData: boolean

  types:
    SymbolFilters:
      name: string
      kind: SymbolKind[]
      package: string
      exported: boolean | null
      filePath: string

    SymbolSummary:
      hash: string
      name: string
      kind: SymbolKind
      package: string
      signature: string
      filePath: string
      line: number
      column: number
      exported: boolean
      hasChanges: boolean  # changed in session diff

  children:
    - ThreePaneLayout:
        leftWidth: 200px
        rightWidth: 260px

        left: SymbolFiltersPanel
        center: SymbolsTablePanel
        right: SymbolPreviewPanel

    # ─────────────────────────────────────────────────────────────────────────
    - SymbolFiltersPanel:
        children:
          - FilterSection:
              title: "Name"
              children:
                - TextInput:
                    value: filters.name
                    onChange: updateFilter('name')
                    placeholder: "prefix or contains"
                    debounce: 300ms

          - FilterSection:
              title: "Kind"
              children:
                - CheckboxGroup:
                    options:
                      - { value: type, label: type }
                      - { value: func, label: func }
                      - { value: method, label: method }
                      - { value: const, label: const }
                      - { value: var, label: var }
                      - { value: interface, label: interface }
                    selected: filters.kind
                    onChange: updateFilter('kind')

          - FilterSection:
              title: "Package"
              children:
                - SearchableSelect:
                    options: packages
                    selected: filters.package
                    onChange: updateFilter('package')
                    placeholder: "Any"

          - FilterSection:
              title: "Exported"
              children:
                - RadioGroup:
                    options:
                      - { value: null, label: All }
                      - { value: true, label: Yes only }
                      - { value: false, label: No only }
                    selected: filters.exported
                    onChange: updateFilter('exported')

          - FilterSection:
              title: "File Path"
              children:
                - TextInput:
                    value: filters.filePath
                    onChange: updateFilter('filePath')
                    placeholder: "glob pattern"

    # ─────────────────────────────────────────────────────────────────────────
    - SymbolsTablePanel:
        bindings:
          data: symbols
          selected: selectedSymbol
        children:
          - TableHeader:
              total: data.total
              showing: "{page * pageSize + 1}-{min((page+1) * pageSize, total)}"

          - DataTable:
              data: data.items
              selectedRow: selectedSymbol
              onRowClick: setSelectedSymbol
              onRowDoubleClick: navigate(/explore/symbols/{hash})
              sortColumn: sortColumn
              sortDirection: sortDirection
              onSort: handleSort
              columns:
                - column: name
                  header: Name
                  sortable: true
                  render: SymbolNameCell
                - column: kind
                  header: Kind
                  sortable: true
                  render: KindBadge
                - column: package
                  header: Package
                  sortable: true
                  render: TruncatedText(30)
                - column: signature
                  header: Signature
                  sortable: false
                  render: TruncatedCode(40)
                - column: filePath
                  header: File
                  sortable: true
                  render: FileLocationCell
                - column: exported
                  header: Exp
                  sortable: true
                  render: ExportedBadge

          - TableFooter:
              children:
                - TableLegend:
                    items:
                      - "● = changed in session diff"
                      - "▸ = unchanged"
                - Pagination:
                    page: data.page
                    pageSize: data.pageSize
                    total: data.total
                    onPageChange: setPage

        SymbolNameCell:
          props: row: SymbolSummary
          render:
            - ChangeIndicator: row.hasChanges  # ● or ▸
            - SymbolName: row.name

        FileLocationCell:
          props: row: SymbolSummary
          render:
            - FileName: basename(row.filePath)
            - LineNumber: ":{row.line}", muted=true

    # ─────────────────────────────────────────────────────────────────────────
    - SymbolPreviewPanel:
        when: selectedSymbol != null
        bindings:
          symbol: selectedSymbol
          preview: symbolPreview
        children:
          - PreviewHeader:
              render:
                - SymbolName: symbol.name, size=large
                - KindBadge: symbol.kind

          - Divider

          - PreviewSection:
              render:
                - CodeBlock:
                    content: preview.signature
                    language: go

          - PreviewSection:
              title: "Location"
              render:
                - FilePathLink: symbol.filePath
                - LineColumnBadge: ":{symbol.line}:{symbol.column}"

          - PreviewSection:
              title: "Details"
              render:
                - LabelValue: Package, symbol.package
                - LabelValue: Exported, ExportedBadge(symbol.exported)
                - LabelValue: Hash, CopyableHash(symbol.hash)

          - PreviewSection:
              title: "References"
              render:
                - ConditionalContent:
                    when: preview.hasRefsData
                    then: RefsCountBadge(preview.refsCount)
                    else: RefsUnavailableNotice

          - Divider

          - ActionButtonGroup:
              vertical: true
              children:
                - Button: "Open Detail", primary=true, onClick: navigate(/explore/symbols/{symbol.hash})
                - Button: "Add to Plan", onClick: addToPlan(symbol)
                - CopyButton: "Copy Hash", value: symbol.hash
                - CopyButton: "Copy Target Spec", value: formatTargetSpec(symbol)
                - Button: "Open in Editor", onClick: openInEditor(symbol)
```

---

# Screen 6: Symbol Detail View

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# SYMBOL DETAIL (route: /explore/symbols/:hash)
# ═══════════════════════════════════════════════════════════════════════════════

SymbolDetailPage:
  route: /explore/symbols/:hash
  params:
    hash: string
  queryParams:
    tab: enum[overview, references, history, audit]

  state:
    activeTab: TabType

  data:
    symbol:
      endpoint: GET /api/symbols/{hash}
      response:
        hash: string
        name: string
        kind: SymbolKind
        package: string
        signature: string
        docComment: string
        filePath: string
        line: number
        column: number
        exported: boolean
        definitionSnippet: string
        relatedCodeUnits: CodeUnitSummary[]

    prepareRenameStatus:
      endpoint: POST /api/symbols/{hash}/prepare-rename
      lazy: true  # only called on button click
      response:
        valid: boolean
        range: Range
        placeholder: string
        error: string | null

  children:
    - BackLink:
        to: /explore/symbols
        label: "← Back to Symbols"

    - SymbolHeader:
        bindings:
          symbol: symbol
        render:
          - SymbolNameTitle: symbol.name, size=xlarge
          - KindBadge: symbol.kind
          - ExportedBadge: symbol.exported
          - Divider: vertical
          - LabelValue: Package, symbol.package
          - LabelValue: Location, FileLocationLink(symbol.filePath, symbol.line, symbol.column)
          - LabelValue: Hash, CopyableHash(symbol.hash)
          - Spacer
          - ActionButtonGroup:
              - Button: "Add Rename to Plan…", onClick: openRenamePlanModal
              - Button: "Run prepare_rename", onClick: runPrepareRename, loading: prepareRenameStatus.isLoading
              - Button: "Open in Editor", onClick: openInEditor(symbol)

    - TabBar:
        tabs:
          - { id: overview, label: Overview }
          - { id: references, label: References, badge: refsCount }
          - { id: history, label: History }
          - { id: audit, label: Audit }
        active: activeTab
        onChange: setActiveTab

    - TabContent:
        switch: activeTab

        # ─────────────────────────────────────────────────────────────────────
        case overview:
          - OverviewTab:
              children:
                - Section:
                    title: "DEFINITION"
                    children:
                      - CodeBlock:
                          content: symbol.definitionSnippet
                          language: go
                          lineNumbers: true
                          startLine: symbol.line - 5
                          highlightLines: [symbol.line]

                - Section:
                    title: "RELATED CODE UNITS"
                    when: symbol.relatedCodeUnits.length > 0
                    children:
                      - CodeUnitList:
                          items: symbol.relatedCodeUnits
                          renderItem: CodeUnitRow
                          onItemClick: navigate(/explore/code-units/{hash})

                - Section:
                    title: "QUICK SEARCH"
                    children:
                      - QuickSearchLink:
                          label: "Search \"{symbol.name}\" in diffs →"
                          onClick: navigate(/search?q={symbol.name}&types=diffs)
                      - QuickSearchLink:
                          label: "Search \"{symbol.name}\" in docs →"
                          onClick: navigate(/search?q={symbol.name}&types=docs)

        # ─────────────────────────────────────────────────────────────────────
        case references:
          - ReferencesTab  # (detailed in Screen 7)

        # ─────────────────────────────────────────────────────────────────────
        case history:
          - HistoryTab:
              data:
                commits:
                  endpoint: GET /api/sessions/{sessionId}/commits/for-file
                  queryParams:
                    path: symbol.filePath
                  response: CommitSummary[]
              children:
                - CommitTimeline:
                    items: commits
                    renderItem: CommitTimelineRow
                    onItemClick: navigate(/explore/commits/{hash})

        # ─────────────────────────────────────────────────────────────────────
        case audit:
          - AuditTab:
              data:
                docHits:
                  endpoint: GET /api/sessions/{sessionId}/docs/hits
                  queryParams:
                    term: symbol.name
                  response: DocHit[]
                diffHits:
                  endpoint: GET /api/sessions/{sessionId}/diffs/search
                  queryParams:
                    query: symbol.name
                  response: DiffHit[]
              children:
                - Section:
                    title: "DOC HITS ({docHits.length})"
                    children:
                      - DocHitList:
                          items: docHits
                          renderItem: DocHitRow
                          onItemClick: navigate(/explore/files?path={filePath}&line={line})

                - Section:
                    title: "DIFF HITS ({diffHits.length})"
                    children:
                      - DiffHitList:
                          items: diffHits
                          renderItem: DiffHitRow
                          onItemClick: navigate(/explore/diffs?file={filePath}&line={line})

    - RenamePlanModal:
        when: renamePlanModalOpen
        # (separate modal component)
```

---

# Screen 7: Symbol Detail — References Tab

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# SYMBOL REFERENCES TAB (nested within SymbolDetailPage)
# ═══════════════════════════════════════════════════════════════════════════════

ReferencesTab:
  parent: SymbolDetailPage
  activeWhen: activeTab == 'references'

  state:
    expandedFiles: Set<string>
    declarationsOnly: boolean
    selectedRef: SymbolReference | null

  data:
    refsAvailability:
      endpoint: GET /api/sessions/{sessionId}/gopls-refs/status
      response:
        available: boolean
        runId: string | null
        lastUpdated: timestamp | null

    references:
      endpoint: GET /api/symbols/{hash}/refs
      queryParams:
        sessionId: sessionId
        declarationsOnly: declarationsOnly
      response:
        total: number
        byFile: Record<string, SymbolReference[]>
      depends: refsAvailability.available == true

  types:
    SymbolReference:
      id: string
      filePath: string
      line: number
      column: number
      snippet: string
      isDeclaration: boolean

  children:
    - ConditionalContent:
        when: !refsAvailability.available
        render:
          - RefsUnavailableCard:
              children:
                - AlertIcon
                - Message: "References data not available for this session."
                - SubMessage: "Gopls references have not been computed."
                - Button: "Compute References", onClick: triggerGoplsRefsRun, primary=true

    - ConditionalContent:
        when: refsAvailability.available
        render:
          - RefsHeader:
              children:
                - TotalCount: "REFERENCES ({references.total} total)"
                - Spacer
                - Checkbox:
                    label: "Declarations only"
                    checked: declarationsOnly
                    onChange: setDeclarationsOnly

          - RefsGroupedList:
              data: references.byFile
              expandedGroups: expandedFiles
              onToggleGroup: toggleFileExpanded
              renderGroup: FileRefsGroup
              renderItem: ReferenceRow

        FileRefsGroup:
          props:
            filePath: string
            refs: SymbolReference[]
            expanded: boolean
          render:
            - CollapsibleHeader:
                icon: expanded ? "▼" : "▶"
                label: "{dirname(filePath)}/"
                emphasis: basename(filePath)
                badge: "({refs.length} refs)"
                onClick: toggleExpanded

            - when: expanded
              render:
                - ReferencesList:
                    items: refs
                    renderItem: ReferenceRow
                    selected: selectedRef
                    onItemClick: setSelectedRef
                    onItemDoubleClick: navigateToFile

        ReferenceRow:
          props: ref: SymbolReference
          render:
            - FileBasename: basename(ref.filePath)
            - LineNumber: ":{ref.line}"
            - SnippetPreview: ref.snippet, highlight=true, truncate=60
            - DeclarationBadge: when ref.isDeclaration

          onClick: setSelectedRef(ref)
          onDoubleClick: navigate(/explore/files?path={ref.filePath}&line={ref.line})

    - RefPreviewPanel:
        when: selectedRef != null
        position: bottom  # or could be a slide-over
        bindings:
          ref: selectedRef
        children:
          - CodeBlock:
              endpoint: GET /api/files/snippet
              queryParams:
                path: ref.filePath
                startLine: ref.line - 3
                endLine: ref.line + 3
              highlightLine: ref.line
              language: go

          - ActionButtonGroup:
              - Button: "Open in File Viewer", onClick: navigate(/explore/files?path={ref.filePath}&line={ref.line})
              - CopyButton: "Copy Location", value: "{ref.filePath}:{ref.line}:{ref.column}"

  actions:
    triggerGoplsRefsRun:
      endpoint: POST /api/sessions/{sessionId}/runs/gopls-refs
      body:
        symbolHash: hash  # optional: compute for specific symbol
      onSuccess: refetch(refsAvailability)
```

---

# Screen 8: Commits Explorer

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# COMMITS EXPLORER (route: /explore/commits)
# ═══════════════════════════════════════════════════════════════════════════════

CommitsExplorerPage:
  route: /explore/commits
  queryParams:
    q: string           # search in subject/body
    author: string
    fromDate: string    # ISO date
    toDate: string      # ISO date
    filePath: string
    page: number

  state:
    filters: CommitFilters
    selectedCommit: CommitSummary | null
    sortDirection: enum[asc, desc]  # default desc (newest first)

  data:
    commits:
      endpoint: GET /api/sessions/{sessionId}/commits
      queryParams:
        search: filters.search
        author: filters.author
        fromDate: filters.fromDate
        toDate: filters.toDate
        filePath: filters.filePath
        sort: "date:{sortDirection}"
        page: page
        pageSize: 20
      response:
        total: number
        page: number
        items: CommitSummary[]

    authors:
      endpoint: GET /api/sessions/{sessionId}/commits/authors
      response: string[]
      cache: session-scoped

    commitPreview:
      endpoint: GET /api/commits/{selectedCommit.hash}/preview
      depends: selectedCommit
      response:
        hash: string
        subject: string
        body: string
        author: string
        authorEmail: string
        date: timestamp
        parents: string[]
        filesChanged: number
        linesAdded: number
        linesRemoved: number
        hasDiffData: boolean

  types:
    CommitFilters:
      search: string
      author: string
      fromDate: string
      toDate: string
      filePath: string

    CommitSummary:
      hash: string
      shortHash: string
      subject: string
      author: string
      date: timestamp
      filesChanged: number
      hasDiffData: boolean

  children:
    - ThreePaneLayout:
        leftWidth: 200px
        rightWidth: 280px

        left: CommitFiltersPanel
        center: CommitsListPanel
        right: CommitPreviewPanel

    # ─────────────────────────────────────────────────────────────────────────
    - CommitFiltersPanel:
        children:
          - FilterSection:
              title: "Search"
              children:
                - TextInput:
                    value: filters.search
                    onChange: updateFilter('search')
                    placeholder: "in subject/body"
                    debounce: 300ms

          - FilterSection:
              title: "Author"
              children:
                - SearchableSelect:
                    options: authors
                    selected: filters.author
                    onChange: updateFilter('author')
                    placeholder: "All"

          - FilterSection:
              title: "Date Range"
              children:
                - DateInput:
                    label: "From"
                    value: filters.fromDate
                    onChange: updateFilter('fromDate')
                - DateInput:
                    label: "To"
                    value: filters.toDate
                    onChange: updateFilter('toDate')

          - FilterSection:
              title: "File Path"
              children:
                - TextInput:
                    value: filters.filePath
                    onChange: updateFilter('filePath')
                    placeholder: "path or glob"

    # ─────────────────────────────────────────────────────────────────────────
    - CommitsListPanel:
        bindings:
          data: commits
          selected: selectedCommit
        children:
          - ListHeader:
              total: data.total
              sortControl:
                value: sortDirection
                options: [{ value: desc, label: "Newest first" }, { value: asc, label: "Oldest first" }]
                onChange: setSortDirection

          - DataTable:
              data: data.items
              selectedRow: selectedCommit
              onRowClick: setSelectedCommit
              onRowDoubleClick: navigate(/explore/commits/{hash})
              columns:
                - column: hash
                  header: Hash
                  width: 80px
                  render: CommitHashCell
                - column: subject
                  header: Subject
                  flex: 1
                  render: SubjectCell
                - column: author
                  header: Author
                  width: 100px
                  render: AuthorBadge
                - column: date
                  header: Date
                  width: 80px
                  render: RelativeTime

          - TableFooter:
              children:
                - TableLegend:
                    - "● = has indexed diff data"
                - Pagination

        CommitHashCell:
          props: row: CommitSummary
          render:
            - DiffIndicator: row.hasDiffData  # ● or ▸
            - ShortHash: row.shortHash

        SubjectCell:
          props: row: CommitSummary
          render:
            - TruncatedText: row.subject, maxLength=50

    # ─────────────────────────────────────────────────────────────────────────
    - CommitPreviewPanel:
        when: selectedCommit != null
        bindings:
          commit: selectedCommit
          preview: commitPreview
        children:
          - PreviewHeader:
              render:
                - CommitHash: preview.shortHash, size=large
                - CopyButton: value=preview.hash

          - Divider

          - PreviewSection:
              render:
                - CommitSubject: preview.subject, size=medium
                - CommitBody: preview.body, when=preview.body

          - PreviewSection:
              render:
                - LabelValue: Author, "{preview.author} <{preview.authorEmail}>"
                - LabelValue: Date, formatDate(preview.date)
                - LabelValue: Parents, preview.parents.join(", ")

          - PreviewSection:
              title: "Changes"
              render:
                - LabelValue: Files, preview.filesChanged
                - StatRow: "+{preview.linesAdded}", color=green
                - StatRow: "-{preview.linesRemoved}", color=red

          - Divider

          - ActionButtonGroup:
              vertical: true
              children:
                - Button: "View Full Detail", primary=true, onClick: navigate(/explore/commits/{commit.hash})
                - Button: "Open Diff", onClick: navigate(/explore/commits/{commit.hash}?tab=diff), disabled: !preview.hasDiffData
                - CopyButton: "Copy Hash", value: commit.hash
```

---

# Screen 9: Commit Detail

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# COMMIT DETAIL (route: /explore/commits/:hash)
# ═══════════════════════════════════════════════════════════════════════════════

CommitDetailPage:
  route: /explore/commits/:hash
  params:
    hash: string
  queryParams:
    tab: enum[overview, files, diff, impact]

  state:
    activeTab: TabType
    selectedFile: CommitFile | null

  data:
    commit:
      endpoint: GET /api/commits/{hash}
      response:
        hash: string
        shortHash: string
        subject: string
        body: string
        author: string
        authorEmail: string
        date: timestamp
        parents: string[]
        linesAdded: number
        linesRemoved: number

    commitFiles:
      endpoint: GET /api/commits/{hash}/files
      response:
        total: number
        items: CommitFile[]

    commitDiff:
      endpoint: GET /api/commits/{hash}/diff
      lazy: true  # loaded when diff tab selected
      response:
        files: DiffFileSummary[]

    commitImpact:
      endpoint: GET /api/commits/{hash}/impact
      lazy: true  # loaded when impact tab selected
      response:
        symbolsInChangedFiles: SymbolSummary[]
        codeUnitsChanged: CodeUnitSummary[]
        docHitsIntroduced: DocHit[]
        docHitsRemoved: DocHit[]

  types:
    CommitFile:
      path: string
      status: enum[A, M, D, R]
      oldPath: string | null  # for renames
      linesAdded: number
      linesRemoved: number

    DiffFileSummary:
      path: string
      status: enum[A, M, D, R]
      hunksCount: number
      linesAdded: number
      linesRemoved: number

  children:
    - BackLink:
        to: /explore/commits
        label: "← Back to Commits"

    - CommitHeader:
        bindings:
          commit: commit
        render:
          - CommitHashTitle: commit.hash, size=xlarge
          - CopyButton: value=commit.hash
          - Divider: vertical
          - CommitSubject: commit.subject, size=large
          - Spacer
          - MetadataRow:
              - LabelValue: Author, "{commit.author} <{commit.authorEmail}>"
              - LabelValue: Date, formatDateTime(commit.date)
              - LabelValue: Parents, ParentHashLinks(commit.parents)
          - Spacer
          - ActionButtonGroup:
              - CopyButton: "Copy Hash", value: commit.hash
              - Button: "Open in GitHub", onClick: openExternalCommit(commit.hash)

    - TabBar:
        tabs:
          - { id: overview, label: Overview }
          - { id: files, label: "Files", badge: commitFiles.total }
          - { id: diff, label: Diff }
          - { id: impact, label: Impact }
        active: activeTab
        onChange: setActiveTab

    - TabContent:
        switch: activeTab

        # ─────────────────────────────────────────────────────────────────────
        case overview:
          - OverviewTab:
              children:
                - Section:
                    title: "MESSAGE"
                    children:
                      - CommitSubject: commit.subject
                      - CommitBody: commit.body, when=commit.body, markdown=true

                - Section:
                    title: "SUMMARY"
                    children:
                      - StatsGrid:
                          - StatTile: label=Files Changed, value=commitFiles.total
                          - StatTile: label=Lines Added, value=commit.linesAdded, color=green
                          - StatTile: label=Lines Removed, value=commit.linesRemoved, color=red

        # ─────────────────────────────────────────────────────────────────────
        case files:
          - FilesTab:
              bindings:
                files: commitFiles.items
                selected: selectedFile
              children:
                - FilesHeader:
                    total: commitFiles.total
                    stats: "+{commit.linesAdded}  -{commit.linesRemoved}"

                - DataTable:
                    data: files
                    selectedRow: selectedFile
                    onRowClick: setSelectedFile
                    onRowDoubleClick: navigateToFileDiff
                    columns:
                      - column: status
                        header: Status
                        width: 60px
                        render: StatusBadge  # A/M/D/R with color
                      - column: path
                        header: Path
                        flex: 1
                        render: FilePathCell
                      - column: linesAdded
                        header: "+"
                        width: 60px
                        render: GreenNumber
                      - column: linesRemoved
                        header: "-"
                        width: 60px
                        render: RedNumber

                - HelpText: "Click a file to view its diff"

        # ─────────────────────────────────────────────────────────────────────
        case diff:
          - DiffTab:
              bindings:
                files: commitDiff.files
                selected: selectedFile
              children:
                - SplitPane:
                    left:
                      - DiffFileList:
                          items: files
                          selected: selectedFile
                          onItemClick: setSelectedFile
                          renderItem: DiffFileRow

                    right:
                      - when: selectedFile != null
                        render:
                          - FileDiffViewer:
                              endpoint: GET /api/commits/{hash}/diff/file
                              queryParams:
                                path: selectedFile.path
                              # renders full unified diff

        # ─────────────────────────────────────────────────────────────────────
        case impact:
          - ImpactTab:
              bindings:
                impact: commitImpact
              children:
                - Section:
                    title: "SYMBOLS IN CHANGED FILES ({impact.symbolsInChangedFiles.length})"
                    children:
                      - SymbolList:
                          items: impact.symbolsInChangedFiles
                          renderItem: SymbolRow
                          maxVisible: 10
                          expandable: true

                - Section:
                    title: "CODE UNITS CHANGED ({impact.codeUnitsChanged.length})"
                    children:
                      - CodeUnitList:
                          items: impact.codeUnitsChanged
                          renderItem: CodeUnitRow
                          maxVisible: 10

                - Section:
                    title: "DOC HITS INTRODUCED ({impact.docHitsIntroduced.length})"
                    when: impact.docHitsIntroduced.length > 0
                    children:
                      - DocHitList:
                          items: impact.docHitsIntroduced
                          renderItem: DocHitRow

                - Section:
                    title: "DOC HITS REMOVED ({impact.docHitsRemoved.length})"
                    when: impact.docHitsRemoved.length > 0
                    children:
                      - DocHitList:
                          items: impact.docHitsRemoved
                          renderItem: DocHitRow
```

---

# Screen 10: Diffs Explorer

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# DIFFS EXPLORER (route: /explore/diffs)
# ═══════════════════════════════════════════════════════════════════════════════

DiffsExplorerPage:
  route: /explore/diffs
  queryParams:
    status: string[]      # A,M,D,R
    path: string          # filter by path
    minChanges: number
    maxChanges: number
    sort: string          # changes:desc, path:asc
    page: number

  state:
    filters: DiffFilters
    selectedFile: DiffFileSummary | null
    sortColumn: string
    sortDirection: enum[asc, desc]

  data:
    sessionDiffSummary:
      endpoint: GET /api/sessions/{sessionId}/diffs/summary
      response:
        totalFiles: number
        totalLinesAdded: number
        totalLinesRemoved: number
        byStatus: Record<string, number>  # { A: 12, M: 98, D: 5, R: 3 }

    diffFiles:
      endpoint: GET /api/sessions/{sessionId}/diffs/files
      queryParams:
        status: filters.status
        path: filters.path
        minChanges: filters.minChanges
        maxChanges: filters.maxChanges
        sort: "{sortColumn}:{sortDirection}"
        page: page
        pageSize: 30
      response:
        total: number
        items: DiffFileSummary[]

    filePreview:
      endpoint: GET /api/sessions/{sessionId}/diffs/file/preview
      queryParams:
        path: selectedFile.path
      depends: selectedFile
      response:
        hunksCount: number
        linesAdded: number
        linesRemoved: number
        firstHunk: DiffHunk  # preview of first hunk

  types:
    DiffFilters:
      status: string[]  # [A, M, D, R]
      path: string
      minChanges: number
      maxChanges: number

    DiffFileSummary:
      path: string
      status: enum[A, M, D, R]
      hunksCount: number
      linesAdded: number
      linesRemoved: number
      hasChanges: boolean  # visual indicator

  children:
    - ThreePaneLayout:
        leftWidth: 200px
        rightWidth: 280px

        left: DiffFiltersPanel
        center: DiffFilesListPanel
        right: DiffPreviewPanel

    # ─────────────────────────────────────────────────────────────────────────
    - DiffFiltersPanel:
        children:
          - FilterSection:
              title: "Status"
              children:
                - CheckboxGroup:
                    options:
                      - { value: M, label: "Modified", color: blue }
                      - { value: A, label: "Added", color: green }
                      - { value: D, label: "Deleted", color: red }
                      - { value: R, label: "Renamed", color: yellow }
                    selected: filters.status
                    onChange: updateFilter('status')

          - FilterSection:
              title: "Path Filter"
              children:
                - TextInput:
                    value: filters.path
                    onChange: updateFilter('path')
                    placeholder: "path or glob"

          - FilterSection:
              title: "Change Size"
              children:
                - NumberInput:
                    label: "Min"
                    value: filters.minChanges
                    onChange: updateFilter('minChanges')
                    min: 0
                - NumberInput:
                    label: "Max"
                    value: filters.maxChanges
                    onChange: updateFilter('maxChanges')
                    min: 0

          - FilterSection:
              title: "Sort By"
              children:
                - SelectDropdown:
                    options:
                      - { value: "changes", label: "Changes (lines)" }
                      - { value: "path", label: "Path" }
                      - { value: "hunks", label: "Hunks count" }
                    selected: sortColumn
                    onChange: setSortColumn

    # ─────────────────────────────────────────────────────────────────────────
    - DiffFilesListPanel:
        bindings:
          summary: sessionDiffSummary
          files: diffFiles
          selected: selectedFile
        children:
          - ListHeader:
              title: "DIFF FILES"
              badge: "{summary.totalFiles} changed"

          - DataTable:
              data: files.items
              selectedRow: selectedFile
              onRowClick: setSelectedFile
              onRowDoubleClick: navigate(/explore/diffs/{path})
              columns:
                - column: path
                  header: Path
                  flex: 1
                  render: DiffFilePathCell
                - column: status
                  header: Status
                  width: 70px
                  render: StatusBadge
                - column: linesAdded
                  header: "+"
                  width: 60px
                  render: GreenNumber
                - column: linesRemoved
                  header: "-"
                  width: 60px
                  render: RedNumber

          - TableFooter:
              children:
                - SessionTotalStats:
                    render: "Session total: +{summary.totalLinesAdded}  -{summary.totalLinesRemoved}  {summary.totalFiles} files"
                - Pagination

        DiffFilePathCell:
          props: row: DiffFileSummary
          render:
            - ChangeIndicator: row.hasChanges
            - FilePath: row.path

    # ─────────────────────────────────────────────────────────────────────────
    - DiffPreviewPanel:
        when: selectedFile != null
        bindings:
          file: selectedFile
          preview: filePreview
        children:
          - PreviewHeader:
              render:
                - FileName: basename(file.path)
                - StatusBadge: file.status

          - Divider

          - PreviewSection:
              title: "Summary"
              render:
                - LabelValue: Hunks, preview.hunksCount
                - StatRow: "+{preview.linesAdded}", color=green
                - StatRow: "-{preview.linesRemoved}", color=red

          - PreviewSection:
              title: "First Hunk Preview"
              children:
                - HunkPreview:
                    hunk: preview.firstHunk
                    maxLines: 10

          - Divider

          - ActionButtonGroup:
              vertical: true
              children:
                - Button: "Open Full Diff", primary=true, onClick: navigate(/explore/diffs/file?path={file.path})
                - Button: "View File", onClick: navigate(/explore/files?path={file.path}&tab=diff)
                - CopyButton: "Copy Path", value: file.path
```

---

# Screen 11: Diff Detail (Full File View)

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# DIFF DETAIL - FULL FILE (route: /explore/diffs/file)
# ═══════════════════════════════════════════════════════════════════════════════

DiffDetailPage:
  route: /explore/diffs/file
  queryParams:
    path: string          # required
    view: enum[unified, split]
    show: enum[all, added, removed]

  state:
    viewMode: enum[unified, split]
    showFilter: enum[all, added, removed]
    currentHunkIndex: number

  data:
    fileDiff:
      endpoint: GET /api/sessions/{sessionId}/diffs/file
      queryParams:
        path: path
      response:
        path: string
        status: enum[A, M, D, R]
        oldPath: string | null
        linesAdded: number
        linesRemoved: number
        hunks: DiffHunk[]

  types:
    DiffHunk:
      index: number
      header: string       # @@ -40,15 +40,15 @@
      oldStart: number
      oldCount: number
      newStart: number
      newCount: number
      lines: DiffLine[]

    DiffLine:
      type: enum[context, add, remove]
      oldLineNo: number | null
      newLineNo: number | null
      content: string

  children:
    - BackLink:
        to: /explore/diffs
        label: "← Back to Diffs"

    - DiffHeader:
        bindings:
          diff: fileDiff
        render:
          - FilePath: diff.path, size=large
          - StatusBadge: diff.status
          - when: diff.oldPath
            render: RenameIndicator: "{diff.oldPath} → {diff.path}"
          - Spacer
          - StatsBadge: "+{diff.linesAdded}  -{diff.linesRemoved}"

    - DiffToolbar:
        children:
          - ToolbarGroup:
              label: "View"
              children:
                - SegmentedControl:
                    options: [{ value: unified, label: Unified }, { value: split, label: Split }]
                    selected: viewMode
                    onChange: setViewMode

          - ToolbarGroup:
              label: "Show"
              children:
                - SelectDropdown:
                    options:
                      - { value: all, label: "All lines" }
                      - { value: added, label: "Added only" }
                      - { value: removed, label: "Removed only" }
                    selected: showFilter
                    onChange: setShowFilter

          - ToolbarGroup:
              children:
                - Button: "◀ Prev Hunk", onClick: prevHunk, disabled: currentHunkIndex == 0
                - Button: "Next Hunk ▶", onClick: nextHunk, disabled: currentHunkIndex == lastHunk

    - DiffViewer:
        bindings:
          hunks: fileDiff.hunks
          viewMode: viewMode
          showFilter: showFilter
          currentHunk: currentHunkIndex
        children:
          - ConditionalRender:
              switch: viewMode

              case unified:
                - UnifiedDiffViewer:
                    hunks: hunks
                    showFilter: showFilter
                    onHunkVisible: setCurrentHunkIndex
                    renderHunk: UnifiedHunkBlock

              case split:
                - SplitDiffViewer:
                    hunks: hunks
                    showFilter: showFilter
                    renderHunk: SplitHunkBlock

        UnifiedHunkBlock:
          props: hunk: DiffHunk
          render:
            - HunkHeader:
                content: hunk.header
                className: "bg-blue-50 text-blue-700 font-mono text-sm"
            - ForEach: hunk.lines
              filter: applyShowFilter(showFilter)
              renderItem: UnifiedDiffLine

        UnifiedDiffLine:
          props: line: DiffLine
          render:
            - LineGutter:
                columns:
                  - oldLineNo: line.oldLineNo, width: 50px
                  - newLineNo: line.newLineNo, width: 50px
            - LineContent:
                content: line.content
                prefix: lineTypePrefix(line.type)  # " ", "+", "-"
                className:
                  add: "bg-green-50"
                  remove: "bg-red-50"
                  context: ""

        SplitHunkBlock:
          props: hunk: DiffHunk
          render:
            - HunkHeader: hunk.header, span=full
            - SplitView:
                left:
                  - ForEach: extractOldLines(hunk)
                    renderItem: SplitLineOld
                right:
                  - ForEach: extractNewLines(hunk)
                    renderItem: SplitLineNew

    - DiffFooter:
        children:
          - HunkIndicator: "Hunk {currentHunkIndex + 1} of {fileDiff.hunks.length}"
          - Spacer
          - ActionButtonGroup:
              - Button: "Open in File Viewer", onClick: navigate(/explore/files?path={path}&tab=diff)
              - CopyButton: "Copy Hunk", onClick: copyCurrentHunk
              - Button: "Search in Hunk", onClick: openSearchInHunk

  keyboardShortcuts:
    "[": prevHunk
    "]": nextHunk
    "u": setViewMode('unified')
    "s": setViewMode('split')
    "/": openSearchInHunk
```

---

# Screen 12: Docs / Terms Explorer

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# DOCS / TERMS EXPLORER (route: /explore/docs)
# ═══════════════════════════════════════════════════════════════════════════════

DocsExplorerPage:
  route: /explore/docs
  queryParams:
    view: enum[terms, files]
    term: string          # filter/select specific term
    fileType: string[]    # md, go, yaml, json
    path: string
    page: number

  state:
    viewMode: enum[terms, files]
    filters: DocFilters
    selectedTerm: string | null
    selectedHit: DocHit | null

  data:
    termsSummary:
      endpoint: GET /api/sessions/{sessionId}/docs/terms
      queryParams:
        fileType: filters.fileType
        path: filters.path
      response:
        total: number
        items: TermSummary[]

    termHits:
      endpoint: GET /api/sessions/{sessionId}/docs/hits
      queryParams:
        term: selectedTerm
        fileType: filters.fileType
        path: filters.path
        page: page
        pageSize: 30
      depends: selectedTerm != null
      response:
        total: number
        items: DocHit[]

    filesSummary:
      endpoint: GET /api/sessions/{sessionId}/docs/files
      queryParams:
        fileType: filters.fileType
        path: filters.path
      response:
        total: number
        items: DocFileSummary[]

  types:
    DocFilters:
      fileType: string[]
      path: string

    TermSummary:
      term: string
      hitCount: number
      fileCount: number

    DocHit:
      id: string
      term: string
      filePath: string
      line: number
      matchText: string
      context: string  # surrounding text

    DocFileSummary:
      path: string
      hitCount: number
      terms: string[]

  children:
    - ViewModeToggle:
        options:
          - { value: terms, label: "● Terms" }
          - { value: files, label: "Files" }
        selected: viewMode
        onChange: setViewMode

    - ThreePaneLayout:
        leftWidth: 200px
        rightWidth: 280px

        left: DocFiltersPanel
        center: DocListPanel
        right: DocHitsPanel

    # ─────────────────────────────────────────────────────────────────────────
    - DocFiltersPanel:
        children:
          - FilterSection:
              title: "Search"
              children:
                - TextInput:
                    placeholder: "filter terms"
                    debounce: 300ms

          - FilterSection:
              title: "File Type"
              children:
                - CheckboxGroup:
                    options:
                      - { value: md, label: "Markdown" }
                      - { value: go, label: "Go" }
                      - { value: yaml, label: "YAML" }
                      - { value: json, label: "JSON" }
                    selected: filters.fileType
                    onChange: updateFilter('fileType')

          - FilterSection:
              title: "Path Filter"
              children:
                - TextInput:
                    value: filters.path
                    onChange: updateFilter('path')
                    placeholder: "e.g., docs/"

    # ─────────────────────────────────────────────────────────────────────────
    - DocListPanel:
        children:
          - ConditionalRender:
              switch: viewMode

              # ─────────────────────────────────────────────────────────────
              case terms:
                - TermsListView:
                    bindings:
                      terms: termsSummary.items
                      selected: selectedTerm
                    children:
                      - ListHeader:
                          title: "TERMS"
                          badge: "{termsSummary.total} unique"

                      - VirtualList:
                          items: terms
                          selectedId: selectedTerm
                          onItemClick: setSelectedTerm
                          renderItem: TermRow

                      - WarningBanner:
                          when: selectedTerm != null && termHits.total > 0
                          render: "⚠ {termHits.total} hits of \"{selectedTerm}\" may need updating"

                  TermRow:
                    props: term: TermSummary
                    render:
                      - ChangeIndicator: term.hitCount > 10  # ● for high count
                      - TermLabel: term.term
                      - HitCountBadge: term.hitCount
                      - FileCountBadge: term.fileCount, muted=true

              # ─────────────────────────────────────────────────────────────
              case files:
                - FilesListView:
                    bindings:
                      files: filesSummary.items
                    children:
                      - ListHeader:
                          title: "FILES WITH DOC HITS"
                          badge: filesSummary.total

                      - GroupedList:
                          items: files
                          groupBy: dirname(path)
                          renderGroup: FileGroupHeader
                          renderItem: DocFileRow

                  DocFileRow:
                    props: file: DocFileSummary
                    render:
                      - FileName: basename(file.path)
                      - HitCountBadge: file.hitCount
                      - TermsList: file.terms.slice(0, 3)

    # ─────────────────────────────────────────────────────────────────────────
    - DocHitsPanel:
        when: selectedTerm != null
        bindings:
          term: selectedTerm
          hits: termHits.items
          selected: selectedHit
        children:
          - PanelHeader:
              title: "HITS"
              subtitle: "\"{term}\" - {termHits.total} hits"

          - HitsList:
              items: hits
              selected: selectedHit
              onItemClick: setSelectedHit
              renderItem: DocHitRow

          - when: selectedHit != null
            render:
              - Divider
              - HitPreview:
                  bindings:
                    hit: selectedHit
                  render:
                    - FileLocationLink: "{hit.filePath}:{hit.line}"
                    - MatchTextHighlight: hit.matchText, term=term
                    - ContextSnippet: hit.context

              - Divider
              - ActionButtonGroup:
                  vertical: true
                  children:
                    - Button: "View File", onClick: navigate(/explore/files?path={hit.filePath}&line={hit.line})
                    - Button: "Add Replacement Rule", onClick: openAddRuleModal(term)
                    - Button: "Mark Done", onClick: markHitDone(hit.id)

        DocHitRow:
          props: hit: DocHit
          render:
            - FileName: basename(hit.filePath)
            - LineNumber: ":{hit.line}"
            - MatchSnippet: hit.matchText, truncate=40, highlight=true

  modals:
    - AddReplacementRuleModal:
        when: addRuleModalOpen
        fields:
          - fromTerm: string, prefilled=selectedTerm
          - toTerm: string
          - fileGlobs: string[]
          - exclusions: string[]
        onSubmit:
          endpoint: POST /api/refactor/plans/{planId}/doc-rules
          body: { fromTerm, toTerm, fileGlobs, exclusions }
```

---

# Screen 13: Runs List (Data/Admin)

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# RUNS LIST (route: /data/runs)
# ═══════════════════════════════════════════════════════════════════════════════

RunsListPage:
  route: /data/runs
  queryParams:
    status: string[]      # success, failed, running
    kind: string[]        # commits, diffs, symbols, etc.
    page: number

  state:
    filters: RunFilters
    selectedRun: RunSummary | null

  data:
    runs:
      endpoint: GET /api/workspaces/{workspaceId}/runs
      queryParams:
        status: filters.status
        kind: filters.kind
        sort: "started:desc"
        page: page
        pageSize: 20
      response:
        total: number
        items: RunSummary[]

    runDetail:
      endpoint: GET /api/runs/{selectedRun.id}
      depends: selectedRun
      response:
        id: string
        status: enum[running, success, failed]
        kind: string
        rootPath: string
        gitFrom: string | null
        gitTo: string | null
        startedAt: timestamp
        finishedAt: timestamp | null
        durationMs: number | null
        argsJson: object
        errorJson: object | null
        rowCounts: Record<string, number>
        rawOutputs: RawOutputRef[]

  types:
    RunFilters:
      status: string[]
      kind: string[]

    RunSummary:
      id: string
      status: enum[running, success, failed]
      kind: string
      gitRange: string | null  # "HEAD~20→HEAD"
      startedAt: timestamp
      rowCount: number | null

    RawOutputRef:
      id: string
      source: string
      filePath: string
      createdAt: timestamp

  children:
    - PageHeader:
        title: "INDEXING RUNS"

    - TwoPaneLayout:
        topHeight: 50%
        
        top: RunsTablePanel
        bottom: RunDetailPanel

    # ─────────────────────────────────────────────────────────────────────────
    - RunsTablePanel:
        bindings:
          runs: runs.items
          selected: selectedRun
        children:
          - DataTable:
              data: runs
              selectedRow: selectedRun
              onRowClick: setSelectedRun
              columns:
                - column: id
                  header: ID
                  width: 60px
                  render: RunIdCell  # "#45"
                - column: status
                  header: Status
                  width: 90px
                  render: RunStatusBadge  # ✅ ⛔ 🔄
                - column: kind
                  header: Kind
                  width: 120px
                  render: KindBadge
                - column: gitRange
                  header: Range
                  flex: 1
                  render: GitRangeCell
                - column: startedAt
                  header: Started
                  width: 100px
                  render: RelativeTime
                - column: rowCount
                  header: Rows
                  width: 80px
                  render: NumberCell

          - Pagination

    # ─────────────────────────────────────────────────────────────────────────
    - RunDetailPanel:
        when: selectedRun != null
        bindings:
          run: runDetail
        children:
          - DetailHeader:
              render:
                - RunIdTitle: "RUN #{run.id}"
                - RunStatusBadge: run.status, size=large

          - DetailGrid:
              columns: 2
              children:
                - LabelValue: Kind, run.kind
                - LabelValue: Status, RunStatusBadge(run.status)
                - LabelValue: Started, formatDateTime(run.startedAt)
                - LabelValue: Finished, formatDateTime(run.finishedAt)
                - LabelValue: Duration, formatDuration(run.durationMs)
                - LabelValue: Root Path, run.rootPath
                - LabelValue: Git From, run.gitFrom
                - LabelValue: Git To, run.gitTo

          - Section:
              title: "ROW COUNTS"
              when: run.rowCounts
              children:
                - KeyValueList:
                    data: run.rowCounts
                    renderItem: RowCountRow

          - Section:
              title: "ERROR"
              when: run.status == 'failed' && run.errorJson
              children:
                - ErrorDisplay:
                    error: run.errorJson
                    collapsible: true

          - Section:
              title: "ARGUMENTS"
              collapsible: true
              defaultCollapsed: true
              children:
                - JsonViewer:
                    data: run.argsJson
                    maxHeight: 200px

          - Section:
              title: "RAW OUTPUTS"
              when: run.rawOutputs.length > 0
              children:
                - RawOutputsList:
                    items: run.rawOutputs
                    renderItem: RawOutputRow
                    onItemClick: navigate(/data/raw-outputs/{id})

          - ActionButtonGroup:
              children:
                - Button: "View Raw Output", onClick: navigate(/data/raw-outputs?runId={run.id})
                - Button: "Retry Run", onClick: retryRun(run.id), when: run.status == 'failed'
                - Button: "Attach to Session", onClick: openAttachModal(run.id)

        RawOutputRow:
          render:
            - SourceBadge: source
            - FilePath: filePath, truncate=40
            - RelativeTime: createdAt

  actions:
    retryRun:
      endpoint: POST /api/runs/{runId}/retry
      onSuccess: refetch(runs)

    attachToSession:
      endpoint: POST /api/sessions/{sessionId}/runs
      body: { runId }
      onSuccess: showToast("Run attached to session")
```

---

# Screen 14: Command Palette (Modal Overlay)

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# COMMAND PALETTE (global modal, triggered by ⌘K)
# ═══════════════════════════════════════════════════════════════════════════════

CommandPalette:
  type: modal
  trigger:
    shortcut: Cmd+K
    or: click CommandPaletteButton

  state:
    query: string
    selectedIndex: number
    activeSection: string | null  # for scoped search

  data:
    # Quick search across multiple domains
    quickSearch:
      endpoint: POST /api/search/quick
      body:
        query: query
        sessionId: currentSession.id
        limit: 5  # per type
      response:
        symbols: SearchResult[]
        files: SearchResult[]
        commands: CommandResult[]
      debounce: 150ms
      depends: query.length >= 2

  types:
    SearchResult:
      type: EntityType
      id: string
      label: string
      sublabel: string
      icon: string

    CommandResult:
      id: string
      label: string
      shortcut: string | null
      action: ActionType

  children:
    - PaletteContainer:
        className: "fixed inset-0 bg-black/50 flex items-start justify-center pt-20"
        onClick: closePalette  # click outside to close
        children:
          - PaletteBox:
              className: "w-[600px] bg-white rounded-lg shadow-2xl"
              onClick: stopPropagation
              children:
                - SearchInput:
                    value: query
                    onChange: setQuery
                    onKeyDown: handleKeyDown
                    placeholder: "Search symbols, files, or type a command…"
                    autoFocus: true
                    icon: 🔍
                    suffix: "⌘K"

                - ResultsContainer:
                    maxHeight: 400px
                    overflow: scroll
                    children:
                      - when: query.length < 2
                        render:
                          - RecentCommandsSection
                          - QuickActionsSection

                      - when: query.length >= 2
                        render:
                          - SearchResultsSection

    # ─────────────────────────────────────────────────────────────────────────
    - RecentCommandsSection:
        title: "RECENT"
        children:
          - CommandList:
              items: recentCommands  # from local storage
              renderItem: CommandRow
              maxVisible: 3

    # ─────────────────────────────────────────────────────────────────────────
    - QuickActionsSection:
        title: "QUICK ACTIONS"
        children:
          - CommandList:
              items:
                - { id: search-diffs, label: "Search in diffs…", shortcut: "⌘⇧D", action: openSearchDiffs }
                - { id: search-docs, label: "Search in docs…", shortcut: "⌘⇧O", action: openSearchDocs }
                - { id: new-plan, label: "New refactor plan…", shortcut: "⌘⇧P", action: openNewPlan }
                - { id: run-audit, label: "Run audit…", shortcut: "⌘⇧A", action: openRunAudit }
                - { id: switch-session, label: "Switch session…", action: openSessionSwitcher }
                - { id: switch-workspace, label: "Switch workspace…", action: openWorkspaceSwitcher }
              renderItem: CommandRow

    # ─────────────────────────────────────────────────────────────────────────
    - SearchResultsSection:
        bindings:
          results: quickSearch
          selected: selectedIndex
        children:
          - ResultGroup:
              when: results.symbols.length > 0
              title: "SYMBOLS"
              children:
                - ResultList:
                    items: results.symbols
                    selectedIndex: selectedIndex
                    baseIndex: 0
                    renderItem: SymbolResultRow
                    onItemClick: navigateToSymbol

          - ResultGroup:
              when: results.files.length > 0
              title: "FILES"
              children:
                - ResultList:
                    items: results.files
                    selectedIndex: selectedIndex
                    baseIndex: results.symbols.length
                    renderItem: FileResultRow
                    onItemClick: navigateToFile

          - ResultGroup:
              when: results.commands.length > 0
              title: "COMMANDS"
              children:
                - ResultList:
                    items: results.commands
                    selectedIndex: selectedIndex
                    baseIndex: results.symbols.length + results.files.length
                    renderItem: CommandRow
                    onItemClick: executeCommand

        SymbolResultRow:
          props: result: SearchResult, selected: boolean
          render:
            - SymbolKindIcon: result.icon
            - ResultLabel: result.label
            - ResultSublabel: result.sublabel, muted=true
          className: selected ? "bg-blue-50" : ""

        FileResultRow:
          props: result: SearchResult, selected: boolean
          render:
            - FileIcon
            - ResultLabel: result.label
            - ResultSublabel: result.sublabel, muted=true
          className: selected ? "bg-blue-50" : ""

        CommandRow:
          props: command: CommandResult, selected: boolean
          render:
            - CommandIcon: "▸"
            - CommandLabel: command.label
            - Spacer
            - ShortcutBadge: command.shortcut, when=command.shortcut
          className: selected ? "bg-blue-50" : ""

    # ─────────────────────────────────────────────────────────────────────────
    - PaletteFooter:
        className: "border-t px-4 py-2 text-sm text-gray-500"
        render:
          - KeyHint: "↑↓", label: "Navigate"
          - KeyHint: "Enter", label: "Select"
          - KeyHint: "Esc", label: "Close"

  keyboardShortcuts:
    ArrowUp: selectPrevious
    ArrowDown: selectNext
    Enter: executeSelected
    Escape: closePalette
    Cmd+Shift+D: directAction(openSearchDiffs)
    Cmd+Shift+O: directAction(openSearchDocs)
    Cmd+Shift+P: directAction(openNewPlan)
    Cmd+Shift+A: directAction(openRunAudit)

  actions:
    navigateToSymbol:
      action: navigate(/explore/symbols/{result.id})
      then: closePalette

    navigateToFile:
      action: navigate(/explore/files?path={result.id})
      then: closePalette

    executeCommand:
      action: command.action()
      then: closePalette

    openSearchDiffs:
      action: navigate(/search?types=diffs)
      then: closePalette

    openSearchDocs:
      action: navigate(/search?types=docs)
      then: closePalette

    openNewPlan:
      action: navigate(/refactor/plans/new)
      then: closePalette

    openRunAudit:
      action: navigate(/refactor/audits/new)
      then: closePalette
```

---

# Shared Components Reference

```yaml
# ═══════════════════════════════════════════════════════════════════════════════
# SHARED / REUSABLE COMPONENTS
# ═══════════════════════════════════════════════════════════════════════════════

SharedComponents:

  # ─────────────────────────────────────────────────────────────────────────────
  # LAYOUT
  # ─────────────────────────────────────────────────────────────────────────────
  ThreePaneLayout:
    props:
      leftWidth: string | number
      rightWidth: string | number
      left: ReactNode
      center: ReactNode
      right: ReactNode
    behavior:
      - Resizable panes with drag handles
      - Collapsible left/right panels
      - Persists widths to localStorage

  TwoPaneLayout:
    props:
      orientation: enum[horizontal, vertical]
      topHeight | leftWidth: string | number
      top | left: ReactNode
      bottom | right: ReactNode

  # ─────────────────────────────────────────────────────────────────────────────
  # DATA DISPLAY
  # ─────────────────────────────────────────────────────────────────────────────
  DataTable:
    props:
      data: T[]
      columns: ColumnDef[]
      selectedRow: T | null
      onRowClick: (row: T) => void
      onRowDoubleClick: (row: T) => void
      sortColumn: string
      sortDirection: enum[asc, desc]
      onSort: (column: string) => void
    features:
      - Keyboard navigation (↑/↓)
      - Sortable headers
      - Row selection highlight
      - Virtual scrolling for large datasets

  VirtualList:
    props:
      items: T[]
      renderItem: (item: T, index: number) => ReactNode
      selectedId: string | null
      onItemClick: (item: T) => void
      itemHeight: number
    behavior:
      - Only renders visible items
      - Smooth scrolling
      - Keyboard navigation

  GroupedList:
    props:
      items: T[]
      groupBy: (item: T) => string
      renderGroup: (groupKey: string, items: T[]) => ReactNode
      renderItem: (item: T) => ReactNode

  # ─────────────────────────────────────────────────────────────────────────────
  # CODE DISPLAY
  # ─────────────────────────────────────────────────────────────────────────────
  CodeViewer:
    props:
      content: string
      language: string
      lineNumbers: boolean
      highlightLine: number | null
      highlightLines: number[]
      startLine: number
      onLineClick: (line: number) => void
    features:
      - Syntax highlighting (highlight.js or similar)
      - Line number gutter
      - Click-to-select line
      - Scroll to highlighted line

  CodeBlock:
    props:
      content: string
      language: string
      maxLines: number
    render: Simplified code display without line numbers

  DiffViewer:
    variants:
      - UnifiedDiffViewer
      - SplitDiffViewer
      - InlineDiffViewer (overlay on file)
    props:
      hunks: DiffHunk[]
      showFilter: enum[all, added, removed]
      onHunkClick: (index: number) => void

  # ─────────────────────────────────────────────────────────────────────────────
  # BADGES & INDICATORS
  # ─────────────────────────────────────────────────────────────────────────────
  StatusBadge:
    variants:
      - RunStatusBadge: ✅ success | ⛔ failed | 🔄 running
      - DiffStatusBadge: A (green) | M (blue) | D (red) | R (yellow)
      - ExportedBadge: ✓ | ✗

  KindBadge:
    props:
      kind: SymbolKind | CodeUnitKind
    colors:
      type: purple
      func: blue
      method: cyan
      const: orange
      var: gray
      interface: green

  ChangeIndicator:
    props:
      hasChanges: boolean
    render: "●" (changed) | "▸" (unchanged)

  # ─────────────────────────────────────────────────────────────────────────────
  # INPUTS & FILTERS
  # ─────────────────────────────────────────────────────────────────────────────
  SearchableSelect:
    props:
      options: string[] | { value, label }[]
      selected: string
      onChange: (value: string) => void
      placeholder: string
    features:
      - Type to filter
      - Keyboard navigation
      - Clear button

  CheckboxGroup:
    props:
      options: { value, label, color? }[]
      selected: string[]
      onChange: (values: string[]) => void

  DateInput:
    props:
      value: string  # ISO date
      onChange: (value: string) => void
      label: string

  # ─────────────────────────────────────────────────────────────────────────────
  # ACTIONS
  # ─────────────────────────────────────────────────────────────────────────────
  CopyButton:
    props:
      value: string
      label: string
    behavior:
      - Click copies to clipboard
      - Shows "Copied!" toast

  OpenInEditorButton:
    props:
      path: string
      line: number
      column: number
    behavior:
      - Constructs URL based on settings (vscode://, idea://, etc.)
      - Falls back to copy path

  ActionButtonGroup:
    props:
      vertical: boolean
      children: Button[]
    render: Consistent button group styling

  # ─────────────────────────────────────────────────────────────────────────────
  # NAVIGATION
  # ─────────────────────────────────────────────────────────────────────────────
  BackLink:
    props:
      to: string
      label: string

  TabBar:
    props:
      tabs: { id, label, badge? }[]
      active: string
      onChange: (id: string) => void

  Pagination:
    props:
      page: number
      pageSize: number
      total: number
      onPageChange: (page: number) => void


# ═══════════════════════════════════════════════════════════════════════════════
# GLOBAL TYPES (shared across all screens)
# ═══════════════════════════════════════════════════════════════════════════════

GlobalTypes:
  SymbolKind: enum[type, func, method, const, var, interface, struct, field]
  CodeUnitKind: enum[func, type, method]
  EntityType: enum[symbols, code-units, diffs, commits, docs, files, tree-sitter]
  
  Location:
    filePath: string
    line: number
    column: number

  Range:
    start: Location
    end: Location

  timestamp: string  # ISO 8601
```