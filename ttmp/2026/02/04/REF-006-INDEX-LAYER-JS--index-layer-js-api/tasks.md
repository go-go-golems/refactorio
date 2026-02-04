# Tasks

## TODO
- [x] Review `go-go-goja` runtime + module registry for reuse in refactorio.
- [x] Implement `pkg/refactor/js/runtime.go` with `require()` wiring and allow-list modules only.
- [x] Implement module registry interface (Name, Doc, Loader) in refactorio JS packages.
- [ ] Add `refactor-index` JS module with query APIs:
- [ ] `querySymbols(filter)`
- [ ] `queryRefs(symbolHash)`
- [ ] `queryDocHits(terms, fileset)`
- [ ] `queryFiles(fileset)`
- [ ] Normalize and sort all query results deterministically.
- [ ] Add runtime safety guards (no time/random, no fs/exec, limits).
- [ ] Add query tracing artifact `js_trace.jsonl`.
- [ ] Implement `refactorio js run` CLI command to execute scripts.
- [ ] Add fixtures and tests for each query API.
- [ ] Add integration test for `refactorio js run`.

## Done
- [ ] (none)
