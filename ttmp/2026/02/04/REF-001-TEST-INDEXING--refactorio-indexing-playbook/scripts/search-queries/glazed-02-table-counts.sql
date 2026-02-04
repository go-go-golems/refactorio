-- Table counts for glazed DB
.headers on
.mode column
SELECT 'commits' AS table_name, COUNT(*) AS count FROM commits;
SELECT 'commit_files' AS table_name, COUNT(*) AS count FROM commit_files;
SELECT 'file_blobs' AS table_name, COUNT(*) AS count FROM file_blobs;
SELECT 'files' AS table_name, COUNT(*) AS count FROM files;
SELECT 'diff_files' AS table_name, COUNT(*) AS count FROM diff_files;
SELECT 'diff_hunks' AS table_name, COUNT(*) AS count FROM diff_hunks;
SELECT 'diff_lines' AS table_name, COUNT(*) AS count FROM diff_lines;
SELECT 'symbol_defs' AS table_name, COUNT(*) AS count FROM symbol_defs;
SELECT 'symbol_occurrences' AS table_name, COUNT(*) AS count FROM symbol_occurrences;
SELECT 'code_units' AS table_name, COUNT(*) AS count FROM code_units;
SELECT 'code_unit_snapshots' AS table_name, COUNT(*) AS count FROM code_unit_snapshots;
SELECT 'code_unit_snapshots_fts' AS table_name, COUNT(*) AS count FROM code_unit_snapshots_fts;
SELECT 'symbol_defs_fts' AS table_name, COUNT(*) AS count FROM symbol_defs_fts;
SELECT 'commits_fts' AS table_name, COUNT(*) AS count FROM commits_fts;
SELECT 'files_fts' AS table_name, COUNT(*) AS count FROM files_fts;
SELECT 'symbol_refs' AS table_name, COUNT(*) AS count FROM symbol_refs;
SELECT 'ts_captures' AS table_name, COUNT(*) AS count FROM ts_captures;
SELECT 'doc_hits' AS table_name, COUNT(*) AS count FROM doc_hits;
