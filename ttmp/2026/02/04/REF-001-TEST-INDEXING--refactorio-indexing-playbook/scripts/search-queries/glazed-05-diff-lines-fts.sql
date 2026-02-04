-- Search diff lines for a pattern using FTS (glazed DB)
.headers on
.mode column

SELECT
  dl.id,
  dl.kind,
  dl.line_no_new,
  substr(dl.text, 1, 120) AS text_snippet,
  f.path AS file_path
FROM diff_lines_fts
JOIN diff_lines dl ON dl.id = diff_lines_fts.rowid
JOIN diff_hunks dh ON dh.id = dl.hunk_id
JOIN diff_files df ON df.id = dh.diff_file_id
JOIN files f ON f.id = df.file_id
WHERE df.run_id = 2
  AND diff_lines_fts MATCH '"context.Background"'
LIMIT 10;
