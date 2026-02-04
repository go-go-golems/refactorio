// list_refs.js
// Query references for a symbol hash.

const idx = require("refactor-index");

const refs = idx.queryRefs("hash-client");
console.log(JSON.stringify(refs, null, 2));
refs;
