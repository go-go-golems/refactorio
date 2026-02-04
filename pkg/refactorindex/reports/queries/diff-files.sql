SELECT
  df.status AS status,
  f.path AS path,
  df.old_path AS old_path,
  df.new_path AS new_path
FROM diff_files df
LEFT JOIN files f ON f.id = df.file_id
WHERE df.run_id = ?
ORDER BY f.path;
