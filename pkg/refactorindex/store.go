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
	if err := ensureColumn(ctx, tx, "symbol_occurrences", "commit_id", "INTEGER"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_symbol_occurrences_commit_id ON symbol_occurrences(commit_id)"); err != nil {
		return errors.Wrap(err, "create symbol_occurrences commit_id index")
	}
	if err := ensureColumn(ctx, tx, "code_unit_snapshots", "commit_id", "INTEGER"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_code_unit_snapshots_commit_id ON code_unit_snapshots(commit_id)"); err != nil {
		return errors.Wrap(err, "create code_unit_snapshots commit_id index")
	}
	if err := ensureFTS(ctx, tx, "doc_hits", "doc_hits_fts", "match_text"); err != nil {
		return err
	}
	if err := ensureFTS(ctx, tx, "diff_lines", "diff_lines_fts", "text"); err != nil {
		return err
	}
	if err := ensureFTSColumns(ctx, tx, "code_unit_snapshots", "code_unit_snapshots_fts", []string{"body_text", "doc_text"}); err != nil {
		return err
	}
	if err := ensureFTSColumns(ctx, tx, "symbol_defs", "symbol_defs_fts", []string{"name", "signature", "pkg"}); err != nil {
		return err
	}
	if err := ensureFTSColumns(ctx, tx, "commits", "commits_fts", []string{"subject", "body"}); err != nil {
		return err
	}
	if err := ensureFTSColumns(ctx, tx, "files", "files_fts", []string{"path"}); err != nil {
		return err
	}
	if err := ensureColumn(ctx, tx, "meta_runs", "status", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, tx, "meta_runs", "error_json", "TEXT"); err != nil {
		return err
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
		`INSERT INTO meta_runs (started_at, status, tool_version, git_from, git_to, root_path, args_json, sources_dir)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		startedAt,
		"running",
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
		"UPDATE meta_runs SET finished_at = ?, status = ?, error_json = NULL WHERE id = ?",
		finishedAt,
		"success",
		runID,
	)
	if err != nil {
		return errors.Wrap(err, "update run finished_at")
	}
	return nil
}

func (s *Store) MarkRunFailed(ctx context.Context, runID int64, err error) error {
	if runID == 0 || err == nil {
		return nil
	}
	finishedAt := time.Now().UTC().Format(time.RFC3339Nano)
	payload, marshalErr := json.Marshal(map[string]string{
		"message": err.Error(),
	})
	if marshalErr != nil {
		payload = []byte(`{"message":"failed to encode error"}`)
	}
	_, execErr := s.db.ExecContext(
		ctx,
		"UPDATE meta_runs SET finished_at = ?, status = ?, error_json = ? WHERE id = ?",
		finishedAt,
		"failed",
		string(payload),
		runID,
	)
	if execErr != nil {
		return errors.Wrap(execErr, "mark run failed")
	}
	return nil
}

func (s *Store) InsertRunMetadata(ctx context.Context, runID int64, key string, value string) error {
	if runID == 0 || strings.TrimSpace(key) == "" {
		return nil
	}
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO run_kv (run_id, key, value, created_at) VALUES (?, ?, ?, ?)",
		runID,
		key,
		value,
		createdAt,
	)
	if err != nil {
		return errors.Wrap(err, "insert run metadata")
	}
	return nil
}

func (s *Store) InsertRunMetadataJSON(ctx context.Context, runID int64, key string, payload interface{}) error {
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "encode run metadata")
	}
	return s.InsertRunMetadata(ctx, runID, key, string(data))
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

func (s *Store) InsertSymbolOccurrence(ctx context.Context, tx *sql.Tx, runID int64, commitID *int64, fileID int64, symbolDefID int64, line int, col int, exported bool) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO symbol_occurrences (run_id, commit_id, file_id, symbol_def_id, line, col, is_exported)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID,
		nullableInt64(commitID),
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

func (s *Store) InsertCodeUnitSnapshot(ctx context.Context, tx *sql.Tx, runID int64, commitID *int64, fileID int64, codeUnitID int64, startLine int, startCol int, endLine int, endCol int, bodyHash string, bodyText string, docText string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO code_unit_snapshots (run_id, commit_id, file_id, code_unit_id, start_line, start_col, end_line, end_col, body_hash, body_text, doc_text)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		nullableInt64(commitID),
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

func (s *Store) InsertSymbolRefUnresolved(ctx context.Context, tx *sql.Tx, runID int64, commitID *int64, symbolHash string, fileID int64, line int, col int, isDecl bool, source string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO symbol_refs_unresolved (run_id, commit_id, symbol_hash, file_id, line, col, is_decl, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		nullableInt64(commitID),
		nullIfEmpty(symbolHash),
		fileID,
		line,
		col,
		boolToInt(isDecl),
		source,
	)
	if err != nil {
		return errors.Wrap(err, "insert unresolved symbol ref")
	}
	return nil
}

func (s *Store) InsertDocHit(ctx context.Context, tx *sql.Tx, runID int64, commitID *int64, fileID int64, line int, col int, term string, matchText string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO doc_hits (run_id, commit_id, file_id, line, col, term, match_text)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID,
		nullableInt64(commitID),
		fileID,
		line,
		col,
		term,
		matchText,
	)
	if err != nil {
		return errors.Wrap(err, "insert doc hit")
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

func ensureColumn(ctx context.Context, tx *sql.Tx, table string, column string, columnDef string) error {
	rows, err := tx.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return errors.Wrap(err, "inspect table columns")
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return errors.Wrap(err, "scan table info")
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "iterate table info")
	}

	_, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN "+column+" "+columnDef)
	if err != nil {
		return errors.Wrap(err, "add column")
	}
	return nil
}

func ensureFTS(ctx context.Context, tx *sql.Tx, table string, ftsTable string, column string) error {
	return ensureFTSColumns(ctx, tx, table, ftsTable, []string{column})
}

func ensureFTSColumns(ctx context.Context, tx *sql.Tx, table string, ftsTable string, columns []string) error {
	if len(columns) == 0 {
		return errors.New("fts columns are required")
	}

	exists, err := tableExists(ctx, tx, ftsTable)
	if err != nil {
		return err
	}

	columnList := strings.Join(columns, ", ")

	if !exists {
		stmt := "CREATE VIRTUAL TABLE " + ftsTable + " USING fts5(" + columnList + ", content='" + table + "', content_rowid='id')"
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrap(err, "create fts table")
		}
	}

	if err := ensureFTSTriggersColumns(ctx, tx, table, ftsTable, columns); err != nil {
		return err
	}

	if !exists {
		rebuild := "INSERT INTO " + ftsTable + "(" + ftsTable + ") VALUES('rebuild')"
		if _, err := tx.ExecContext(ctx, rebuild); err != nil {
			return errors.Wrap(err, "rebuild fts index")
		}
	}

	return nil
}

func ensureFTSTriggersColumns(ctx context.Context, tx *sql.Tx, table string, ftsTable string, columns []string) error {
	columnList := strings.Join(columns, ", ")
	newColumns := make([]string, 0, len(columns))
	oldColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		newColumns = append(newColumns, "new."+column)
		oldColumns = append(oldColumns, "old."+column)
	}

	ai := "CREATE TRIGGER IF NOT EXISTS " + ftsTable + "_ai AFTER INSERT ON " + table +
		" BEGIN INSERT INTO " + ftsTable + "(rowid, " + columnList + ") VALUES (new.id, " + strings.Join(newColumns, ", ") + "); END;"
	ad := "CREATE TRIGGER IF NOT EXISTS " + ftsTable + "_ad AFTER DELETE ON " + table +
		" BEGIN INSERT INTO " + ftsTable + "(" + ftsTable + ", rowid, " + columnList + ") VALUES('delete', old.id, " + strings.Join(oldColumns, ", ") + "); END;"
	au := "CREATE TRIGGER IF NOT EXISTS " + ftsTable + "_au AFTER UPDATE ON " + table +
		" BEGIN " +
		"INSERT INTO " + ftsTable + "(" + ftsTable + ", rowid, " + columnList + ") VALUES('delete', old.id, " + strings.Join(oldColumns, ", ") + "); " +
		"INSERT INTO " + ftsTable + "(rowid, " + columnList + ") VALUES (new.id, " + strings.Join(newColumns, ", ") + "); " +
		"END;"

	if _, err := tx.ExecContext(ctx, ai); err != nil {
		return errors.Wrap(err, "create fts insert trigger")
	}
	if _, err := tx.ExecContext(ctx, ad); err != nil {
		return errors.Wrap(err, "create fts delete trigger")
	}
	if _, err := tx.ExecContext(ctx, au); err != nil {
		return errors.Wrap(err, "create fts update trigger")
	}

	return nil
}

func tableExists(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", name).Scan(&count); err != nil {
		return false, errors.Wrap(err, "check table exists")
	}
	return count > 0, nil
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
