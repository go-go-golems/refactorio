-- Search code unit bodies for a term (glazed DB)
.headers on
.mode column

SELECT
  cu.name,
  cu.kind,
  cu.pkg,
  f.path AS file_path,
  substr(s.body_text, 1, 120) AS body_snippet
FROM code_unit_snapshots_fts
JOIN code_unit_snapshots s ON s.id = code_unit_snapshots_fts.rowid
JOIN code_units cu ON cu.id = s.code_unit_id
JOIN files f ON f.id = s.file_id
WHERE s.run_id = 4
  AND code_unit_snapshots_fts MATCH '"errgroup"'
LIMIT 10;
