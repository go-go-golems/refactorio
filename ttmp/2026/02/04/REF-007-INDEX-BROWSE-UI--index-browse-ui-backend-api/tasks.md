# Tasks

## TODO

- [ ] Add tasks here

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
- [ ] Implement files endpoints (/api/files tree, /api/file content, /api/files/history)
- [ ] Implement optional tree-sitter capture endpoint (/api/tree-sitter/captures)
- [ ] Add API tests (db info, runs list, search smoke)
