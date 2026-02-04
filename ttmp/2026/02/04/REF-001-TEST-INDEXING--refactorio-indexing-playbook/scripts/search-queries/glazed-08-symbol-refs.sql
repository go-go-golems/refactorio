-- List gopls references for symbols matched by FTS (glazed DB)
.headers on
.mode column

WITH matched_symbols AS (
  SELECT rowid
  FROM symbol_defs_fts
  WHERE symbol_defs_fts MATCH '"Store"'
)
SELECT sd.name, sd.pkg, f.path AS file_path, sr.line, sr.col, sr.is_decl, sr.source
FROM matched_symbols ms
JOIN symbol_defs sd ON sd.id = ms.rowid
JOIN symbol_refs sr ON sr.symbol_def_id = sd.id
JOIN files f ON f.id = sr.file_id
WHERE sr.run_id = 8
ORDER BY sr.id;
