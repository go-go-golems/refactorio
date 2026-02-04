-- List all runs with metadata (glazed DB)
.headers on
.mode column
SELECT id, tool_version, git_from, git_to, root_path, sources_dir, started_at, finished_at
FROM meta_runs
ORDER BY id;
