// list_files.js
// List files matching a fileset.

const idx = require("refactor-index");

const files = idx.queryFiles({
  include: ["docs/**/*.md", "internal/**/*.go"],
  exclude: ["**/vendor/**"],
});

console.log(JSON.stringify(files, null, 2));
files;
