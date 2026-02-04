-- Search commits by message text (glazed DB)
.headers on
.mode column

SELECT c.hash, c.committer_date, c.subject
FROM commits_fts
JOIN commits c ON c.id = commits_fts.rowid
WHERE c.run_id = 1
  AND commits_fts MATCH '"refactor"'
ORDER BY c.id DESC
LIMIT 10;
