---
Title: JS Index API Deep Review
Ticket: REF-006-INDEX-LAYER-JS
Status: active
Topics:
    - refactorio
    - js
    - index
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: refactorio/README.md
      Note: User-facing README
    - Path: refactorio/cmd/refactorio/js_run.go
      Note: CLI runner
    - Path: refactorio/pkg/refactor/js/modules/common.go
      Note: Module registry
    - Path: refactorio/pkg/refactor/js/modules/refactorindex/refactorindex.go
      Note: JS query module
    - Path: refactorio/pkg/refactor/js/runtime.go
      Note: Runtime creation and guards
    - Path: refactorio/testdata/js
      Note: Example scripts
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-04T16:18:20-05:00
WhatFor: ""
WhenToUse: ""
---


# JS Index API Deep Review

## Purpose
This is a multi‑angle review and critique of the JS index API work (runtime, module, CLI, tests, and docs). It is intentionally opinionated and detailed to surface tradeoffs, risks, and improvement opportunities.

## Angle 1: Architectural Cohesion
The code aligns with the “query‑only” mission: JS can read the index but cannot mutate files. The architecture cleanly separates runtime setup (`pkg/refactor/js`), module registry (`pkg/refactor/js/modules`), and the `refactor-index` module itself. This layering mirrors go‑go‑goja and reduces the risk of coupling JS features to refactorio’s apply pipeline.

**Strengths**
- Clear separation of runtime, registry, and module implementation.
- Minimal surface area for JS modules.
- CLI runner is thin and uses the same runtime APIs.

**Weaknesses / Risks**
- The runtime uses both “explicit allow list” (`Modules`) and registry (`Registry`) options. This is flexible but can cause configuration ambiguity.
- Module lifecycle (trace open/close) is not enforced by an interface, so callers can forget to close trace files.

**Recommendation**
Add a small `RuntimeBuilder` that only takes a registry + options. This removes the `Modules` vs `Registry` ambiguity and centralizes lifecycle hooks.

## Angle 2: API Ergonomics for Script Authors
The JS API is minimal and easy to remember: `querySymbols`, `queryRefs`, `queryDocHits`, `queryFiles`. The return shapes are stable and plain objects, which is appropriate for JS.

**Strengths**
- Clear API names and query scope.
- Plain objects avoid goja type surprises.
- Example scripts demonstrate typical patterns.

**Weaknesses / Risks**
- `queryRefs` uses a symbol hash, which requires a two‑step query pattern. That is fine for advanced usage but harder for newcomers.
- `queryDocHits` exposes fileset filtering but only after fetching and filtering in Go. This is correct but might surprise users who expect DB‑level filtering.

**Recommendation**
Provide helper JS functions (or examples) that bundle “symbol selection + refs” into a common snippet, and clarify in docs that fileset filtering is done after DB queries.

## Angle 3: Determinism Guarantees
Determinism is essential for reproducible plans. Sorting in the module ensures stable ordering even when DB ordering changes.

**Strengths**
- Explicit `sort.Slice` in each query path.
- Runtime can disable `Date.now()` and `Math.random()`.

**Weaknesses / Risks**
- Script authors can still create non‑deterministic order by iterating over object keys or by manually sorting with custom comparators.
- Determinism is not enforced by the runner (no linting or static checks).

**Recommendation**
Add a “determinism checklist” section to the guide (one paragraph) and optionally a `--strict-determinism` mode that rejects use of non‑deterministic globals.

## Angle 4: Safety and Sandboxing
The runner blocks filesystem module loading and avoids `fs`/`exec` by default. This is an appropriate baseline.

**Strengths**
- `AllowFileJS=false` prevents arbitrary `require()` from disk.
- JS can only access allowed native modules.

**Weaknesses / Risks**
- The script itself is still read from disk by Go, so an untrusted script is still executed. This is fine in a dev‑only environment but should be called out as a trust boundary.
- There is no time limit or memory limit enforced in the runtime.

**Recommendation**
Add a runtime guard for execution time (e.g., context with deadline or goja interrupt) and document that scripts are trusted inputs for now.

## Angle 5: Data Access and Performance
The module pulls data via refactor‑index query helpers and then filters/sorts in Go. That’s a good compromise for a small API surface.

**Strengths**
- Uses existing refactor‑index queries rather than raw SQL.
- Result limit prevents unbounded memory growth.

**Weaknesses / Risks**
- Fileset filtering for doc hits and files is done in Go after loading all rows. This can be expensive for large repos.
- `ListFiles` is called with only a limit; if the DB grows, results may be truncated in unexpected ways.

**Recommendation**
Extend query helpers with path‑prefix or glob‑aware SQL filters, or allow paging in JS to avoid large in‑memory slices.

## Angle 6: Error Handling and Diagnostics
Errors are surfaced to JS by panicking with the error string; in Go, the CLI returns errors cleanly.

**Strengths**
- Early validation for required inputs (e.g., symbol hash).
- Clear error messages for missing args.

**Weaknesses / Risks**
- Panic‑based error handling means the JS side can’t catch structured error objects.
- Some errors (e.g., file glob syntax) are only visible in logs, not enriched with context in JS.

**Recommendation**
Expose errors as JS exceptions with structured fields (code/message), or add a `refactor.index.errors` helper to introspect error types.

## Angle 7: Logging and Traceability
The `js_trace.jsonl` hook is the strongest audit mechanism. It is simple and easy to consume.

**Strengths**
- Trace entries include action, args, and result counts.
- Trace generation is optional and low overhead.

**Weaknesses / Risks**
- There is no “run metadata” emitted alongside trace files (e.g., script path, run ID). Users must infer it externally.

**Recommendation**
Add a small header record in the trace file with metadata, or emit a `js_trace.meta.json` alongside the JSONL.

## Angle 8: Testing Strategy
Tests cover both module‑level behavior and end‑to‑end CLI execution.

**Strengths**
- Unit tests exercise goja integration with real DB data.
- Integration test ensures `refactorio js run` works end‑to‑end.
- Trace test ensures `--trace` produces output.

**Weaknesses / Risks**
- Tests use local DB fixtures created in Go, not ingest outputs. This is fine but may miss ingestion edge cases.
- No performance tests or large‑dataset tests.

**Recommendation**
Add a fixture DB generated from actual ingest (even a small test repo) to validate real‑world ingestion patterns.

## Angle 9: Documentation Quality
Docs are comprehensive and aligned with the style guide. They include quick start, API reference, examples, and troubleshooting.

**Strengths**
- Clear, runnable examples.
- Troubleshooting sections at multiple levels.
- Trace guidance is explicit.

**Weaknesses / Risks**
- The JS guide is in `design/` rather than a more discoverable docs directory.
- README may duplicate some content from the guide.

**Recommendation**
If this becomes stable, move the JS guide to a `docs/` folder and link from the root README to reduce duplication.

## Angle 10: Maintainability and Future Extension
The current design should scale to more query endpoints and later refactoring APIs.

**Strengths**
- Module registry pattern scales cleanly.
- Runner is minimal and can be extended.

**Weaknesses / Risks**
- The module’s configuration (limits, trace behavior) is hardcoded.
- There is no versioning of the JS API surface.

**Recommendation**
Add a `refactor-index.version()` call and a configurable `ModuleOptions` struct for limits and trace settings.

## Angle 11: Dependency Management
The work added goja and doublestar dependencies to refactorio.

**Strengths**
- Dependencies are standard and well‑known.

**Weaknesses / Risks**
- goja versions can shift; if goja is updated elsewhere, versions might drift.

**Recommendation**
Pin goja/goja_nodejs in a tools or dependency management file and document compatibility.

## Angle 12: UX and CLI Consistency
The CLI uses Cobra but not the glazed command builder. This keeps the JS runner simple but may diverge from other refactorio commands.

**Strengths**
- Simple, readable command implementation.

**Weaknesses / Risks**
- If refactorio standardizes on glazed, this CLI might need refactoring later.

**Recommendation**
Document that `js run` is currently a minimal cobra command and revisit when the broader CLI is stabilized.

## Overall Assessment
The implementation is a solid “first slice” of the JS index API: it’s safe, deterministic, test‑backed, and documented. The largest opportunities are in scaling performance (filtering at the DB layer), exposing a small versioned API contract, and improving trace metadata.

## Priority Follow‑Ups
1. Add query paging and tighter DB‑side filtering.
2. Add trace metadata record for script/run context.
3. Add API version helper in JS module.
4. Consider a runtime time‑limit guard.
