# JS API Examples

These scripts are runnable examples for the refactor-index JS API.

## Usage

Run a script with the refactorio JS runner:

```bash
go run ./cmd/refactorio js run \
  --script testdata/js/list_symbols.js \
  --index-db /path/to/index.sqlite \
  --run-id 1
```

## Scripts

- `list_symbols.js` — query symbols by package/name/kind
- `list_refs.js` — query references by symbol hash
- `doc_hits.js` — query doc hits with fileset globs
- `list_files.js` — query files using include/exclude globs
- `plan_like_output.js` — build a plan-like object from query results
