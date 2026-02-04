# Tasks

## TODO
- [x] Review `go-go-goja` runtime + module registry for reuse in refactorio.
- [x] Implement `pkg/refactor/js/runtime.go` with `require()` wiring and allow-list modules only.
- [x] Implement module registry interface (Name, Doc, Loader) in refactorio JS packages.
- [x] Add `refactor-index` JS module with query APIs:
- [x] `querySymbols(filter)`
- [x] `queryRefs(symbolHash)`
- [x] `queryDocHits(terms, fileset)`
- [x] `queryFiles(fileset)`
- [x] Normalize and sort all query results deterministically.
- [x] Add runtime safety guards (no time/random, no fs/exec, limits).
- [x] Add query tracing artifact `js_trace.jsonl`.
- [x] Implement `refactorio js run` CLI command to execute scripts.
- [x] Add fixtures and tests for each query API.
- [ ] Add integration test for `refactorio js run`.

## Done
- [ ] (none)
