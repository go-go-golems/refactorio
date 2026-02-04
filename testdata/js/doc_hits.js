// doc_hits.js
// Query doc hits with include/exclude globs.

const idx = require("refactor-index");

const hits = idx.queryDocHits(["Client"], {
  include: ["docs/**/*.md"],
  exclude: ["docs/vendor/**"],
});

console.log(JSON.stringify(hits, null, 2));
hits;
