# Tasks

## Completed

- [x] Write detailed analysis + implementation guide (`design/01-search-drill-in-detailed-analysis-and-implementation-guide.md`).
- [x] Add shared drill-in utility module under `ui/src/features/search-drill-in/`:
  - [x] Typed payload and param models (`types.ts`)
  - [x] Result-to-route deep-link builder (`linkBuilder.ts`)
  - [x] URL param parsers for destination hydration (`parseDrillInParams.ts`)
  - [x] Barrel export (`index.ts`)
- [x] Wire Search UI drill-in actions:
  - [x] `SearchPage` opens result links via router navigation.
  - [x] `SearchPage` supports open-in-new-tab and copy-link actions.
  - [x] `SearchResults` supports `onOpen`, `onOpenInNewTab`, `onCopyLink` action callbacks.
- [x] Implement URL-driven drill-in hydration in all target pages:
  - [x] `SymbolsPage` (`symbol_hash`, optional `run_id`)
  - [x] `CodeUnitsPage` (`unit_hash`, optional `run_id`)
  - [x] `CommitsPage` (`commit_hash`, optional `run_id`)
  - [x] `DiffsPage` (`run_id`, `path`, optional `line_new`/`line_old`/`hunk_id`)
  - [x] `DocsPage` (`term`, optional `path`/`line`/`col`/`run_id`)
  - [x] `FilesPage` (`path`, optional `line`) with ancestor prefix hydration
- [x] Extend file API usage to support lazy prefix loading from UI:
  - [x] Export `useLazyGetFilesQuery` in `ui/src/api/files.ts`
  - [x] Re-export `useLazyGetFilesQuery` in `ui/src/api/client.ts`
- [x] Add diff-target visual highlighting support in `DiffViewer`:
  - [x] Highlight specific old/new line numbers.
  - [x] Highlight full hunk via `hunk_id`.
- [x] Run non-Playwright validation:
  - [x] `ui` production build (`npm run build`) passes.

## Remaining

- [ ] Add focused automated tests for link generation and URL param parsing.
- [ ] Add focused component/integration tests for destination hydration behavior.
- [ ] Execute manual QA pass for all six result types and record outcomes.
- [ ] Playwright E2E validation (explicitly deferred per request).
