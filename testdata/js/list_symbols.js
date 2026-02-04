// list_symbols.js
// Query a specific symbol by package + name + kind.

const idx = require("refactor-index");

const symbols = idx.querySymbols({
  pkg: "github.com/acme/project/internal/api",
  name: "Client",
  kind: "type",
});

console.log(JSON.stringify(symbols, null, 2));
symbols;
