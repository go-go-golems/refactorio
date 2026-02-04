# Tasks

## TODO

## Task 1: Multi-column FTS helper support
- [x] Add helper to create multi-column FTS tables + triggers.
- [x] Keep existing single-column helper behavior intact.

## Task 2: FTS for code unit snapshots
- [x] Add FTS table for `code_unit_snapshots` (`body_text`, `doc_text`).

## Task 3: FTS for symbol defs
- [x] Add FTS table for `symbol_defs` (`name`, `signature`, `pkg`).

## Task 4: FTS for commits
- [x] Add FTS table for `commits` (`subject`, `body`).

## Task 5: FTS for files
- [x] Add FTS table for `files` (`path`).

## Task 6: ISO8601 commit dates
- [x] Update `ingest_commits` to store ISO8601 dates.

## Task 7: v_last_commit_per_file view
- [ ] Add `v_last_commit_per_file` view with run_id, file path, commit hash, date, status.

## Task 8: Update search SQL examples
- [ ] Refactor existing search SQL examples to use new FTS tables and view.

## Task 9: Tests
- [ ] Add/extend smoke tests for new FTS tables and view.

## Task 10: Docs
- [ ] Update search design + help/tutorials to mention new FTS tables and view.
