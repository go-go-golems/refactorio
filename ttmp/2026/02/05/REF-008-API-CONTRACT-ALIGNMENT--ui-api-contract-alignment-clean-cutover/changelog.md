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

