---
Title: JS Index API Postmortem
Ticket: REF-006-INDEX-LAYER-JS
Status: active
Topics:
    - refactorio
    - js
    - index
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-04T16:15:20-05:00
WhatFor: ""
WhenToUse: ""
---

# JS Index API Postmortem

## Summary
The JS index API implementation succeeded and tests are green. The work delivered a goja runtime with safety guards, a refactor-index JS module, a CLI runner, and unit/integration tests. Documentation and examples were added to make the API approachable.

## What Went Well
- Reused go-go-goja patterns for runtime and module registration, avoiding reinvention.
- Query results are deterministic due to explicit sorting and disabled time/random.
- The JS module returns plain JS objects, which simplifies script consumption.
- Tests caught early integration issues and validated the end-to-end runner.

## What Didn’t Go Well
- Initial tests failed because goja exported arrays as `[]map[string]interface{}` rather than `[]interface{}`; the test helper had to be adjusted.
- Fileset parsing from JS objects required a custom helper; `ExportTo` alone was insufficient.

## Root Causes
- goja’s `Export()` behavior differs depending on object shape, which affected test expectations.
- The `fileset` object structure in JS required explicit extraction of `include`/`exclude` arrays.

## Fixes Applied
- Added a `assertRowCount` helper that accepts both `[]map[string]interface{}` and `[]interface{}`.
- Implemented `parseFileset` to robustly parse JS objects with `include`/`exclude` fields.

## Tests Run
- `go test ./pkg/refactor/js/modules/refactorindex`
- `go test ./cmd/refactorio`

## Follow-Up Improvements
- Add a dedicated example script showing `--trace` and `js_trace.jsonl` outputs.
- Extend tests to cover trace file creation.
- Consider adding `queryFiles` filters by extension and binary flags for performance.

## Takeaways
The JS index API is stable enough for early adopters. The next step should be careful incremental expansion of query capabilities and trace visibility, while keeping the runtime locked down.
