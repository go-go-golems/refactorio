package refactorindex

const SchemaVersion = 1

const schemaSQL = `
CREATE TABLE IF NOT EXISTS schema_versions (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS meta_runs (
    id INTEGER PRIMARY KEY,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    tool_version TEXT,
    git_from TEXT,
    git_to TEXT,
    root_path TEXT,
    args_json TEXT,
    sources_dir TEXT
);

CREATE TABLE IF NOT EXISTS raw_outputs (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    source TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id)
);

CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    ext TEXT,
    file_exists INTEGER,
    is_binary INTEGER
);

CREATE TABLE IF NOT EXISTS diff_files (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    file_id INTEGER,
    status TEXT NOT NULL,
    old_path TEXT,
    new_path TEXT,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id),
    FOREIGN KEY(file_id) REFERENCES files(id)
);

CREATE TABLE IF NOT EXISTS diff_hunks (
    id INTEGER PRIMARY KEY,
    diff_file_id INTEGER NOT NULL,
    old_start INTEGER NOT NULL,
    old_lines INTEGER NOT NULL,
    new_start INTEGER NOT NULL,
    new_lines INTEGER NOT NULL,
    FOREIGN KEY(diff_file_id) REFERENCES diff_files(id)
);

CREATE TABLE IF NOT EXISTS diff_lines (
    id INTEGER PRIMARY KEY,
    hunk_id INTEGER NOT NULL,
    kind TEXT NOT NULL,
    line_no_old INTEGER,
    line_no_new INTEGER,
    text TEXT NOT NULL,
    FOREIGN KEY(hunk_id) REFERENCES diff_hunks(id)
);

CREATE INDEX IF NOT EXISTS idx_diff_files_run_id ON diff_files(run_id);
CREATE INDEX IF NOT EXISTS idx_diff_hunks_diff_file_id ON diff_hunks(diff_file_id);
CREATE INDEX IF NOT EXISTS idx_diff_lines_hunk_id ON diff_lines(hunk_id);
`
