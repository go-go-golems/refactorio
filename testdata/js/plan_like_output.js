// plan_like_output.js
// Build a plan-like object from query results (no apply step).

const idx = require("refactor-index");

const symbols = idx.querySymbols({
  pkg: "github.com/acme/project/internal/api",
  name: "Client",
  kind: "type",
});

if (symbols.length !== 1) {
  throw new Error(`ambiguous symbol count: ${symbols.length}`);
}

const sym = symbols[0];

const plan = {
  plan_version: 1,
  ops: [
    {
      type: "go.gopls.rename",
      selector: { symbol_hash: sym.symbol_hash },
      resolved: {
        def_span: sym.def_span,
        old_name: sym.name,
        new_name: "APIClient",
      },
    },
    {
      type: "text.replace",
      selector: {
        mode: "ident",
        from: "Client",
        to: "APIClient",
        include: ["**/*.md", "**/*.yaml", "**/*.json"],
      },
    },
  ],
};

console.log(JSON.stringify(plan, null, 2));
plan;
