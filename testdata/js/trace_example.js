// trace_example.js
// Run with --trace /tmp/js_trace.jsonl to capture query traces.

const idx = require("refactor-index");

idx.querySymbols({
  pkg: "github.com/acme/project/internal/api",
  name: "Client",
  kind: "type",
});

idx.queryDocHits(["Client"], {
  include: ["docs/**/*.md"],
});

"trace example complete";
