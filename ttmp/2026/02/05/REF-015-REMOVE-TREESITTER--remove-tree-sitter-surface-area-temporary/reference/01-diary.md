---
Title: Diary
Ticket: REF-015-REMOVE-TREESITTER
Status: active
Topics:
    - backend
    - frontend
    - api
    - refactorio
    - indexing
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/doc/topics/04-workbench-api-reference.md
      Note: Removes /tree-sitter/captures endpoint and tree_sitter feature references.
    - Path: pkg/workbenchapi/db_info.go
      Note: |-
        Removes tree_sitter feature advertisement.
        Removed tree_sitter feature flag
    - Path: pkg/workbenchapi/routes.go
      Note: |-
        Removes tree-sitter route registration from API startup.
        Removed tree-sitter route registration
    - Path: pkg/workbenchapi/runs.go
      Note: Removes ts_captures summary count from run summary endpoint.
    - Path: pkg/workbenchapi/session_types.go
      Note: |-
        Removes tree_sitter from session run contract.
        Removed tree_sitter session contract field
    - Path: pkg/workbenchapi/sessions.go
      Note: |-
        Removes tree_sitter run and availability field from session computation.
        Removed tree_sitter availability/run mapping
    - Path: pkg/workbenchapi/tree_sitter.go
      Note: Deleted tree-sitter endpoint implementation.
    - Path: ttmp/2026/02/05/REF-015-REMOVE-TREESITTER--remove-tree-sitter-surface-area-temporary/changelog.md
      Note: Records step outcomes and commit linkage.
    - Path: ttmp/2026/02/05/REF-015-REMOVE-TREESITTER--remove-tree-sitter-surface-area-temporary/tasks.md
      Note: |-
        Tracks task status for removal work.
        Tracked and checked all removal tasks
    - Path: ui/src/components/selection/SessionCard.tsx
      Note: Removes tree-sitter availability label from session card.
    - Path: ui/src/hooks/useSessionContext.ts
      Note: Removes tree-sitter run-id mapping from session context hook.
    - Path: ui/src/types/api.ts
      Note: |-
        Removes tree_sitter from frontend SessionRuns type.
        Removed tree_sitter field from frontend SessionRuns
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-05T18:31:40-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Track the temporary hard removal of tree-sitter from the Refactorio backend/UI API surface so the system contract matches current implementation.

## Step 1: Create Ticket and Execution Task List

I created a dedicated ticket so this cleanup is isolated from the broader API-alignment stream and can be handed off clearly. The initial focus is on removing runtime/API references first, then synchronizing frontend contract and docs.

I also replaced the placeholder task list with explicit executable tasks, then started implementation immediately (no planning-only stop).

### Prompt Context

**User prompt (verbatim):** "Create a new ticket to remove treesitter entirely for now, add tasks, then work on it (work on task, check it off, commit, and keep a frequent diary)."

**Assistant interpretation:** Create a new docmgr ticket, define concrete tasks, execute the work directly, keep task status current, commit changes, and maintain the diary continuously.

**Inferred user intent:** Remove all active tree-sitter surface area now so backend and frontend contracts are internally consistent, while keeping documentation traceability for teammates.

### What I did
- Created `REF-015-REMOVE-TREESITTER`.
- Added a diary doc for the ticket.
- Replaced placeholder task with four concrete tasks covering backend removal, frontend removal, docs/diary, and validation.

### Why
- Existing ticket streams were already crowded; this needs an isolated changelog and clear ownership.

### What worked
- `docmgr ticket create-ticket` and `docmgr task` workflow worked cleanly.

### What didn't work
- N/A

### What I learned
- A ticket-per-cutover-step keeps implementation history clearer than mixing this into REF-008.

### What was tricky to build
- Task wording needed to be precise enough that each check-off mapped to verifiable code and command outputs.

### What warrants a second pair of eyes
- Task granularity: ensure this decomposition matches how the team wants to split review and handoff.

### What should be done in the future
- If tree-sitter returns later, create a fresh reintroduction ticket rather than reusing this removal ticket.

### Code review instructions
- Review `ttmp/2026/02/05/REF-015-REMOVE-TREESITTER--remove-tree-sitter-surface-area-temporary/tasks.md` first.

### Technical details
Commands run:
```bash
docmgr ticket create-ticket --ticket REF-015-REMOVE-TREESITTER --title "Remove Tree-Sitter Surface Area (Temporary)" --topics backend,frontend,api,refactorio,indexing
docmgr doc add --ticket REF-015-REMOVE-TREESITTER --doc-type reference --title "Diary"
docmgr task remove --ticket REF-015-REMOVE-TREESITTER --id 1
docmgr task add --ticket REF-015-REMOVE-TREESITTER --text "Remove tree-sitter API route and session/db feature surface from workbench backend"
docmgr task add --ticket REF-015-REMOVE-TREESITTER --text "Remove tree_sitter fields from frontend API types, session context, and session UI stories/mocks"
docmgr task add --ticket REF-015-REMOVE-TREESITTER --text "Update docs/changelog/diary for clean-cut temporary removal and follow-up plan"
docmgr task add --ticket REF-015-REMOVE-TREESITTER --text "Validate server and UI compile paths impacted by removal"
```

## Step 2: Remove Tree-Sitter Runtime/API/UI Surface

I removed the tree-sitter route and contract exposure from `workbenchapi`, deleted the tree-sitter handler file, and removed tree-sitter session fields from frontend types and selectors. This makes session payloads, DB features, and UI availability rendering consistent with a tree-sitter-disabled system.

I also updated the backend API reference page and ran targeted validation. Backend package tests pass. UI build still fails, but due pre-existing unrelated type issues in diff/entity table stories and dashboard/session selector types.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Execute the first concrete removal tasks immediately and verify the changed paths.

**Inferred user intent:** Land a clean, reviewable diff that strips tree-sitter from active contracts now.

**Commit (code):** 5031d68 — "Remove tree-sitter from workbench API and UI session surface"

### What I did
- Removed tree-sitter route registration from `pkg/workbenchapi/routes.go`.
- Removed `TreeSitter` from `SessionRuns` in backend session types.
- Removed tree-sitter session computation and availability emission in `pkg/workbenchapi/sessions.go`.
- Removed `tree_sitter` feature from `/db/info` response map.
- Removed `ts_captures` run-summary count branch.
- Deleted `pkg/workbenchapi/tree_sitter.go`.
- Removed `tree_sitter` from UI `SessionRuns` and session context mapping.
- Removed tree-sitter labels/references from session selection stories and mocks.
- Removed tree-sitter endpoint and feature examples from `pkg/doc/topics/04-workbench-api-reference.md`.
- Marked tasks 1, 2, and 4 done.

### Why
- Tree-sitter backend implementation is intentionally removed, so keeping API/UI surface references causes contract drift and confusion.

### What worked
- `GOWORK=off go test ./pkg/workbenchapi/...` passed.
- Search verification over touched API/UI paths no longer shows `tree_sitter`/`tree-sitter` references.

### What didn't work
- `npm --prefix ui run build` failed with pre-existing unrelated errors:
  - `DiffViewer*.stories.tsx` and `DiffViewer.tsx` `hunk_id` type issues.
  - `EntityTable.stories.tsx` Symbol-to-record cast issues.
  - `SessionSelector.tsx` importing missing `SessionAvailability`.
  - `DashboardPage.tsx` expecting `run_id` on `Run`.

### What I learned
- Backend/API cleanup is straightforward; frontend full build remains blocked by previously known type-contract drift outside this ticket.

### What was tricky to build
- Removing fields from shared session contracts requires synchronized changes across backend computation, frontend types, and story mocks to avoid latent shape mismatches.

### What warrants a second pair of eyes
- Confirm no downstream consumer still expects `/api/tree-sitter/captures` or `tree_sitter` keys in session/db-info payloads.

### What should be done in the future
- N/A

### Code review instructions
- Start with `pkg/workbenchapi/session_types.go`, `pkg/workbenchapi/sessions.go`, and deleted `pkg/workbenchapi/tree_sitter.go`.
- Then review `ui/src/types/api.ts` and `ui/src/hooks/useSessionContext.ts`.
- Finally check `pkg/doc/topics/04-workbench-api-reference.md` for contract docs alignment.

### Technical details
Commands run:
```bash
GOWORK=off go test ./pkg/workbenchapi/...
npm --prefix ui run build
```
