-- Search diff lines for HTTP API calls in the last 20 commits (range runs 10-29)
.headers on
.mode column

SELECT
  df.run_id,
  mr.git_to AS commit_hash,
  f.path AS file_path,
  dl.kind,
  substr(dl.text, 1, 120) AS text_snippet
FROM diff_lines_fts
JOIN diff_lines dl ON dl.id = diff_lines_fts.rowid
JOIN diff_hunks dh ON dh.id = dl.hunk_id
JOIN diff_files df ON df.id = dh.diff_file_id
JOIN meta_runs mr ON mr.id = df.run_id
JOIN files f ON f.id = df.file_id
WHERE df.run_id BETWEEN 10 AND 29
  AND diff_lines_fts MATCH '"http.NewRequest" OR "grpc.Dial"'
LIMIT 25;
