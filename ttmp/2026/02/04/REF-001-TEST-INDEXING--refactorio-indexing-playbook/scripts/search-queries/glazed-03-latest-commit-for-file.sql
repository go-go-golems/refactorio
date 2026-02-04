-- Latest commit that touched a specific file (glazed DB)
-- Update file match as needed.
.headers on
.mode column

WITH matched_files AS (
  SELECT rowid
  FROM files_fts
  WHERE files_fts MATCH '"pkg/help/help.go"'
)
SELECT
  v.hash,
  v.committer_date,
  v.status,
  v.old_path,
  v.new_path,
  v.file_path
FROM v_last_commit_per_file v
JOIN matched_files mf ON mf.rowid = v.file_id
WHERE v.run_id = 1
ORDER BY v.commit_id DESC
LIMIT 1;
