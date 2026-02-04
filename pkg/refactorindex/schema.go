package refactorindex

const SchemaVersion = 6

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

CREATE TABLE IF NOT EXISTS symbol_defs (
    id INTEGER PRIMARY KEY,
    pkg TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    recv TEXT,
    signature TEXT,
    symbol_hash TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS symbol_occurrences (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    symbol_def_id INTEGER NOT NULL,
    line INTEGER NOT NULL,
    col INTEGER NOT NULL,
    is_exported INTEGER NOT NULL,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id),
    FOREIGN KEY(file_id) REFERENCES files(id),
    FOREIGN KEY(symbol_def_id) REFERENCES symbol_defs(id)
);

CREATE TABLE IF NOT EXISTS code_units (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    pkg TEXT NOT NULL,
    recv TEXT,
    signature TEXT,
    unit_hash TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS code_unit_snapshots (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    code_unit_id INTEGER NOT NULL,
    start_line INTEGER NOT NULL,
    start_col INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    end_col INTEGER NOT NULL,
    body_hash TEXT NOT NULL,
    body_text TEXT NOT NULL,
    doc_text TEXT,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id),
    FOREIGN KEY(file_id) REFERENCES files(id),
    FOREIGN KEY(code_unit_id) REFERENCES code_units(id)
);

CREATE TABLE IF NOT EXISTS commits (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    hash TEXT NOT NULL,
    author_name TEXT,
    author_email TEXT,
    author_date TEXT,
    committer_date TEXT,
    subject TEXT,
    body TEXT,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id)
);

CREATE TABLE IF NOT EXISTS commit_files (
    id INTEGER PRIMARY KEY,
    commit_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    old_path TEXT,
    new_path TEXT,
    blob_old TEXT,
    blob_new TEXT,
    FOREIGN KEY(commit_id) REFERENCES commits(id),
    FOREIGN KEY(file_id) REFERENCES files(id)
);

CREATE TABLE IF NOT EXISTS file_blobs (
    id INTEGER PRIMARY KEY,
    commit_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    blob_sha TEXT NOT NULL,
    size_bytes INTEGER,
    line_count INTEGER,
    FOREIGN KEY(commit_id) REFERENCES commits(id),
    FOREIGN KEY(file_id) REFERENCES files(id)
);

CREATE TABLE IF NOT EXISTS symbol_refs (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    commit_id INTEGER,
    symbol_def_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    line INTEGER NOT NULL,
    col INTEGER NOT NULL,
    is_decl INTEGER NOT NULL,
    source TEXT NOT NULL,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id),
    FOREIGN KEY(commit_id) REFERENCES commits(id),
    FOREIGN KEY(symbol_def_id) REFERENCES symbol_defs(id),
    FOREIGN KEY(file_id) REFERENCES files(id)
);

CREATE TABLE IF NOT EXISTS ts_captures (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    commit_id INTEGER,
    file_id INTEGER NOT NULL,
    query_name TEXT NOT NULL,
    capture_name TEXT NOT NULL,
    node_type TEXT,
    start_line INTEGER NOT NULL,
    start_col INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    end_col INTEGER NOT NULL,
    snippet TEXT NOT NULL,
    FOREIGN KEY(run_id) REFERENCES meta_runs(id),
    FOREIGN KEY(commit_id) REFERENCES commits(id),
    FOREIGN KEY(file_id) REFERENCES files(id)
);

CREATE INDEX IF NOT EXISTS idx_diff_files_run_id ON diff_files(run_id);
CREATE INDEX IF NOT EXISTS idx_diff_hunks_diff_file_id ON diff_hunks(diff_file_id);
CREATE INDEX IF NOT EXISTS idx_diff_lines_hunk_id ON diff_lines(hunk_id);
CREATE INDEX IF NOT EXISTS idx_symbol_defs_hash ON symbol_defs(symbol_hash);
CREATE INDEX IF NOT EXISTS idx_symbol_occurrences_run_id ON symbol_occurrences(run_id);
CREATE INDEX IF NOT EXISTS idx_symbol_occurrences_symbol_id ON symbol_occurrences(symbol_def_id);
CREATE INDEX IF NOT EXISTS idx_code_units_hash ON code_units(unit_hash);
CREATE INDEX IF NOT EXISTS idx_code_unit_snapshots_run_id ON code_unit_snapshots(run_id);
CREATE INDEX IF NOT EXISTS idx_commits_run_id ON commits(run_id);
CREATE INDEX IF NOT EXISTS idx_commits_hash ON commits(hash);
CREATE INDEX IF NOT EXISTS idx_commit_files_commit_id ON commit_files(commit_id);
CREATE INDEX IF NOT EXISTS idx_file_blobs_commit_id ON file_blobs(commit_id);
CREATE INDEX IF NOT EXISTS idx_symbol_refs_run_id ON symbol_refs(run_id);
CREATE INDEX IF NOT EXISTS idx_symbol_refs_symbol_id ON symbol_refs(symbol_def_id);
CREATE INDEX IF NOT EXISTS idx_symbol_refs_commit_id ON symbol_refs(commit_id);
CREATE INDEX IF NOT EXISTS idx_ts_captures_run_id ON ts_captures(run_id);
CREATE INDEX IF NOT EXISTS idx_ts_captures_commit_id ON ts_captures(commit_id);
`
