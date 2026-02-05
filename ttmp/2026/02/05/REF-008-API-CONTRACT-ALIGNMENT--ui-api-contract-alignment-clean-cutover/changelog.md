# Changelog

## 2026-02-05

- Initial workspace created


## 2026-02-05

Upgraded refactorio CLI to Glazed v1 APIs, migrated live foobar.db to schema v17, re-ingested index domains for real UI validation, and fixed nullable files column handling that caused /api/files 500s in live mode.

### Related Files

- cmd/refactorio/api.go — Switch to v1 command decoding and remove deprecated section helpers
- cmd/refactorio/js_run.go — Switch to v1 command decoding and tags
- cmd/refactorio/root.go — Use AddLoggingSectionToRootCommand
- pkg/workbenchapi/files.go — Handle nullable file_exists/is_binary safely
- pkg/workbenchapi/search.go — Handle nullable file_exists/is_binary in file search


## 2026-02-05

Implemented real topbar workspace/session selectors (Redux-wired) and fixed stale session-scoped table data so switching sessions immediately reflects domain availability/state.

### Related Files

- ui/src/App.tsx — Provide workspace/session option lists and dispatch selection actions
- ui/src/components/layout/Topbar.tsx — Replace placeholder buttons with controlled workspace/session comboboxes
- ui/src/pages/CodeUnitsPage.tsx — Clear stale code unit table when selected session has no code-units run
- ui/src/pages/CommitsPage.tsx — Clear stale commit table when selected session has no commits run
- ui/src/pages/DiffsPage.tsx — Gate diff run list by session availability to avoid stale run selection
- ui/src/pages/DocsPage.tsx — Clear stale doc terms when selected session has no doc-hits run
- ui/src/pages/SymbolsPage.tsx — Clear stale symbol table when selected session has no symbols run

