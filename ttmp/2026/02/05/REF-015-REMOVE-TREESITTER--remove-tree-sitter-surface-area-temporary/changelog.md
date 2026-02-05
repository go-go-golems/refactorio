# Changelog

## 2026-02-05

- Initial workspace created


## 2026-02-05

Step 1-2: created ticket/task plan and removed tree-sitter from workbench API + UI session contract (commit 5031d68).

### Related Files

- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/doc/topics/04-workbench-api-reference.md — Removed tree-sitter endpoint/feature docs
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/workbenchapi/sessions.go — Removed tree_sitter run and availability exposure
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/pkg/workbenchapi/tree_sitter.go — Deleted tree-sitter endpoint handler
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ui/src/hooks/useSessionContext.ts — Removed tree-sitter run mapping
- /home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio/ui/src/types/api.ts — Removed tree_sitter from SessionRuns


## 2026-02-05

Ticket closed: tree-sitter surface removed from backend/API/UI contract for temporary cutover.

