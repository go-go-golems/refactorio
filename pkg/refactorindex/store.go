package refactorindex

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type RunConfig struct {
	ToolVersion string
	GitFrom     string
	GitTo       string
	RootPath    string
	SourcesDir  string
	ArgsJSON    string
}

type RawOutput struct {
	Source string
	Path   string
}

type SymbolDef struct {
	Pkg       string
	Name      string
	Kind      string
	Recv      string
	Signature string
	Hash      string
}

type CodeUnitDef struct {
	Pkg       string
	Name      string
	Kind      string
	Recv      string
	Signature string
	Hash      string
}

type CommitInfo struct {
	Hash          string
	AuthorName    string
	AuthorEmail   string
	AuthorDate    string
	CommitterDate string
	Subject       string
	Body          string
}

func OpenDB(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, errors.Wrap(err, "open sqlite db")
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, errors.Wrap(err, "ping sqlite db")
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, errors.Wrap(err, "enable foreign keys")
	}
	return db, nil
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InitSchema(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin schema transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, schemaSQL); err != nil {
		return errors.Wrap(err, "apply schema")
	}
	if err := insertSchemaVersion(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit schema")
	}
	return nil
}

func insertSchemaVersion(ctx context.Context, tx *sql.Tx) error {
	appliedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := tx.ExecContext(
		ctx,
		"INSERT OR IGNORE INTO schema_versions (version, applied_at) VALUES (?, ?)",
		SchemaVersion,
		appliedAt,
	)
	if err != nil {
		return errors.Wrap(err, "insert schema version")
	}
	return nil
}

func (s *Store) CreateRun(ctx context.Context, cfg RunConfig) (int64, error) {
	startedAt := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO meta_runs (started_at, tool_version, git_from, git_to, root_path, args_json, sources_dir)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		startedAt,
		cfg.ToolVersion,
		cfg.GitFrom,
		cfg.GitTo,
		cfg.RootPath,
		cfg.ArgsJSON,
		cfg.SourcesDir,
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert run")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "read run id")
	}
	return id, nil
}

func (s *Store) FinishRun(ctx context.Context, runID int64) error {
	finishedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(
		ctx,
		"UPDATE meta_runs SET finished_at = ? WHERE id = ?",
		finishedAt,
		runID,
	)
	if err != nil {
		return errors.Wrap(err, "update run finished_at")
	}
	return nil
}

func (s *Store) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin transaction")
	}
	return tx, nil
}

func (s *Store) GetOrCreateFile(ctx context.Context, tx *sql.Tx, path string) (int64, error) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	_, err := tx.ExecContext(
		ctx,
		"INSERT OR IGNORE INTO files (path, ext) VALUES (?, ?)",
		path,
		ext,
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert file")
	}
	var id int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM files WHERE path = ?", path).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "fetch file id")
	}
	return id, nil
}

func (s *Store) InsertDiffFile(ctx context.Context, tx *sql.Tx, runID int64, fileID int64, status string, oldPath string, newPath string) (int64, error) {
	res, err := tx.ExecContext(
		ctx,
		"INSERT INTO diff_files (run_id, file_id, status, old_path, new_path) VALUES (?, ?, ?, ?, ?)",
		runID,
		fileID,
		status,
		nullIfEmpty(oldPath),
		nullIfEmpty(newPath),
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert diff file")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "read diff file id")
	}
	return id, nil
}

func (s *Store) InsertDiffHunk(ctx context.Context, tx *sql.Tx, diffFileID int64, oldStart int, oldLines int, newStart int, newLines int) (int64, error) {
	res, err := tx.ExecContext(
		ctx,
		"INSERT INTO diff_hunks (diff_file_id, old_start, old_lines, new_start, new_lines) VALUES (?, ?, ?, ?, ?)",
		diffFileID,
		oldStart,
		oldLines,
		newStart,
		newLines,
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert diff hunk")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "read diff hunk id")
	}
	return id, nil
}

func (s *Store) InsertDiffLine(ctx context.Context, tx *sql.Tx, hunkID int64, kind string, lineNoOld *int, lineNoNew *int, text string) error {
	_, err := tx.ExecContext(
		ctx,
		"INSERT INTO diff_lines (hunk_id, kind, line_no_old, line_no_new, text) VALUES (?, ?, ?, ?, ?)",
		hunkID,
		kind,
		nullableInt(lineNoOld),
		nullableInt(lineNoNew),
		text,
	)
	if err != nil {
		return errors.Wrap(err, "insert diff line")
	}
	return nil
}

func (s *Store) InsertRawOutput(ctx context.Context, tx *sql.Tx, runID int64, source string, path string) error {
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := tx.ExecContext(
		ctx,
		"INSERT INTO raw_outputs (run_id, source, path, created_at) VALUES (?, ?, ?, ?)",
		runID,
		source,
		path,
		createdAt,
	)
	if err != nil {
		return errors.Wrap(err, "insert raw output")
	}
	return nil
}

func (s *Store) GetOrCreateSymbolDef(ctx context.Context, tx *sql.Tx, def SymbolDef) (int64, error) {
	if def.Hash == "" {
		return 0, errors.New("symbol hash is required")
	}
	_, err := tx.ExecContext(
		ctx,
		`INSERT OR IGNORE INTO symbol_defs (pkg, name, kind, recv, signature, symbol_hash)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		def.Pkg,
		def.Name,
		def.Kind,
		nullIfEmpty(def.Recv),
		nullIfEmpty(def.Signature),
		def.Hash,
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert symbol def")
	}
	var id int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM symbol_defs WHERE symbol_hash = ?", def.Hash).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "fetch symbol def id")
	}
	return id, nil
}

func (s *Store) InsertSymbolOccurrence(ctx context.Context, tx *sql.Tx, runID int64, fileID int64, symbolDefID int64, line int, col int, exported bool) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO symbol_occurrences (run_id, file_id, symbol_def_id, line, col, is_exported)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		runID,
		fileID,
		symbolDefID,
		line,
		col,
		boolToInt(exported),
	)
	if err != nil {
		return errors.Wrap(err, "insert symbol occurrence")
	}
	return nil
}

func (s *Store) GetOrCreateCodeUnit(ctx context.Context, tx *sql.Tx, def CodeUnitDef) (int64, error) {
	if def.Hash == "" {
		return 0, errors.New("code unit hash is required")
	}
	_, err := tx.ExecContext(
		ctx,
		`INSERT OR IGNORE INTO code_units (kind, name, pkg, recv, signature, unit_hash)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		def.Kind,
		def.Name,
		def.Pkg,
		nullIfEmpty(def.Recv),
		nullIfEmpty(def.Signature),
		def.Hash,
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert code unit")
	}
	var id int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM code_units WHERE unit_hash = ?", def.Hash).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "fetch code unit id")
	}
	return id, nil
}

func (s *Store) InsertCodeUnitSnapshot(ctx context.Context, tx *sql.Tx, runID int64, fileID int64, codeUnitID int64, startLine int, startCol int, endLine int, endCol int, bodyHash string, bodyText string, docText string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO code_unit_snapshots (run_id, file_id, code_unit_id, start_line, start_col, end_line, end_col, body_hash, body_text, doc_text)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		fileID,
		codeUnitID,
		startLine,
		startCol,
		endLine,
		endCol,
		bodyHash,
		bodyText,
		nullIfEmpty(docText),
	)
	if err != nil {
		return errors.Wrap(err, "insert code unit snapshot")
	}
	return nil
}

func (s *Store) InsertCommit(ctx context.Context, tx *sql.Tx, runID int64, info CommitInfo) (int64, error) {
	if info.Hash == "" {
		return 0, errors.New("commit hash is required")
	}
	res, err := tx.ExecContext(
		ctx,
		`INSERT INTO commits (run_id, hash, author_name, author_email, author_date, committer_date, subject, body)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		info.Hash,
		nullIfEmpty(info.AuthorName),
		nullIfEmpty(info.AuthorEmail),
		nullIfEmpty(info.AuthorDate),
		nullIfEmpty(info.CommitterDate),
		nullIfEmpty(info.Subject),
		nullIfEmpty(info.Body),
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert commit")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "read commit id")
	}
	return id, nil
}

func (s *Store) InsertCommitFile(ctx context.Context, tx *sql.Tx, commitID int64, fileID int64, status string, oldPath string, newPath string, blobOld string, blobNew string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO commit_files (commit_id, file_id, status, old_path, new_path, blob_old, blob_new)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		commitID,
		fileID,
		status,
		nullIfEmpty(oldPath),
		nullIfEmpty(newPath),
		nullIfEmpty(blobOld),
		nullIfEmpty(blobNew),
	)
	if err != nil {
		return errors.Wrap(err, "insert commit file")
	}
	return nil
}

func (s *Store) InsertFileBlob(ctx context.Context, tx *sql.Tx, commitID int64, fileID int64, blobSHA string, sizeBytes *int64, lineCount *int) error {
	if blobSHA == "" {
		return errors.New("blob sha is required")
	}
	_, err := tx.ExecContext(
		ctx,
		`INSERT OR IGNORE INTO file_blobs (commit_id, file_id, blob_sha, size_bytes, line_count)
		 VALUES (?, ?, ?, ?, ?)`,
		commitID,
		fileID,
		blobSHA,
		nullableInt64(sizeBytes),
		nullableInt(lineCount),
	)
	if err != nil {
		return errors.Wrap(err, "insert file blob")
	}
	return nil
}

func (s *Store) GetSymbolDefIDByHash(ctx context.Context, tx *sql.Tx, hash string) (int64, error) {
	if hash == "" {
		return 0, errors.New("symbol hash is required")
	}
	var id int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM symbol_defs WHERE symbol_hash = ?", hash).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "fetch symbol def id")
	}
	return id, nil
}

func (s *Store) InsertSymbolRef(ctx context.Context, tx *sql.Tx, runID int64, commitID *int64, symbolDefID int64, fileID int64, line int, col int, isDecl bool, source string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO symbol_refs (run_id, commit_id, symbol_def_id, file_id, line, col, is_decl, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		nullableInt64(commitID),
		symbolDefID,
		fileID,
		line,
		col,
		boolToInt(isDecl),
		source,
	)
	if err != nil {
		return errors.Wrap(err, "insert symbol ref")
	}
	return nil
}

func (s *Store) InsertTreeSitterCapture(ctx context.Context, tx *sql.Tx, runID int64, commitID *int64, fileID int64, queryName string, captureName string, nodeType string, startLine int, startCol int, endLine int, endCol int, snippet string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO ts_captures (run_id, commit_id, file_id, query_name, capture_name, node_type, start_line, start_col, end_line, end_col, snippet)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		nullableInt64(commitID),
		fileID,
		queryName,
		captureName,
		nullIfEmpty(nodeType),
		startLine,
		startCol,
		endLine,
		endCol,
		snippet,
	)
	if err != nil {
		return errors.Wrap(err, "insert tree-sitter capture")
	}
	return nil
}

func (s *Store) WriteRawOutput(ctx context.Context, tx *sql.Tx, runDir string, runID int64, source string, fileName string, content []byte) (string, error) {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", errors.Wrap(err, "create sources dir")
	}
	path := filepath.Join(runDir, fileName)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", errors.Wrap(err, "write raw output")
	}
	if err := s.InsertRawOutput(ctx, tx, runID, source, path); err != nil {
		return "", err
	}
	return path, nil
}

func EncodeArgsJSON(args map[string]string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return "", errors.Wrap(err, "encode args json")
	}
	return string(payload), nil
}

func nullableInt(value *int) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*value), Valid: true}
}

func nullableInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func nullIfEmpty(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
