package refactorindex

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pkg/errors"
)

type DiffFileRecord struct {
	RunID   int64
	Status  string
	Path    string
	OldPath string
	NewPath string
}

type DiffFileFilter struct {
	RunID  int64
	Limit  int
	Offset int
}

type SymbolInventoryFilter struct {
	RunID        int64
	ExportedOnly bool
	Kind         string
	Name         string
	Pkg          string
	Path         string
	Limit        int
	Offset       int
}

type SymbolInventoryRecord struct {
	RunID      int64
	SymbolHash string
	Name       string
	Kind       string
	Pkg        string
	Recv       string
	Signature  string
	FilePath   string
	Line       int
	Col        int
	IsExported bool
}

type SymbolRefUnresolvedFilter struct {
	RunID  int64
	Limit  int
	Offset int
}

type SymbolRefUnresolvedRecord struct {
	RunID      int64
	CommitHash string
	SymbolHash string
	FilePath   string
	Line       int
	Col        int
	IsDecl     bool
	Source     string
}

func (s *Store) GetCommitIDByHash(ctx context.Context, runID int64, hash string) (int64, error) {
	var id int64
	if err := s.db.QueryRowContext(ctx, "SELECT id FROM commits WHERE run_id = ? AND hash = ?", runID, hash).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "fetch commit id")
	}
	return id, nil
}

func (s *Store) ListDiffFiles(ctx context.Context, filter DiffFileFilter) ([]DiffFileRecord, error) {
	query := `
		SELECT df.run_id, df.status, f.path, df.old_path, df.new_path
		FROM diff_files df
		LEFT JOIN files f ON f.id = df.file_id
		WHERE (? = 0 OR df.run_id = ?)
		ORDER BY df.run_id, f.path`

	args := []interface{}{
		filter.RunID,
		filter.RunID,
	}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	} else if filter.Offset > 0 {
		query += " LIMIT -1 OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query diff files")
	}
	defer rows.Close()

	var results []DiffFileRecord
	for rows.Next() {
		var record DiffFileRecord
		var oldPath sql.NullString
		var newPath sql.NullString
		if err := rows.Scan(&record.RunID, &record.Status, &record.Path, &oldPath, &newPath); err != nil {
			return nil, errors.Wrap(err, "scan diff file")
		}
		if oldPath.Valid {
			record.OldPath = oldPath.String
		}
		if newPath.Valid {
			record.NewPath = newPath.String
		}
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate diff files")
	}
	return results, nil
}

func (s *Store) ListSymbolInventory(ctx context.Context, filter SymbolInventoryFilter) ([]SymbolInventoryRecord, error) {
	query := `
		SELECT o.run_id, d.symbol_hash, d.name, d.kind, d.pkg, d.recv, d.signature,
		       f.path, o.line, o.col, o.is_exported
		FROM symbol_occurrences o
		JOIN symbol_defs d ON d.id = o.symbol_def_id
		JOIN files f ON f.id = o.file_id
		WHERE (? = 0 OR o.run_id = ?)
		  AND (? = '' OR d.kind = ?)
		  AND (? = '' OR d.name = ?)
		  AND (? = '' OR d.pkg = ?)
		  AND (? = '' OR f.path = ?)
		  AND (? = 0 OR o.is_exported = 1)
		ORDER BY o.run_id, d.pkg, d.name, f.path, o.line, o.col`

	args := []interface{}{
		filter.RunID,
		filter.RunID,
		filter.Kind,
		filter.Kind,
		filter.Name,
		filter.Name,
		filter.Pkg,
		filter.Pkg,
		filter.Path,
		filter.Path,
		boolToInt(filter.ExportedOnly),
	}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query symbol inventory")
	}
	defer rows.Close()

	var results []SymbolInventoryRecord
	for rows.Next() {
		var record SymbolInventoryRecord
		var recv sql.NullString
		var signature sql.NullString
		var exported int
		if err := rows.Scan(
			&record.RunID,
			&record.SymbolHash,
			&record.Name,
			&record.Kind,
			&record.Pkg,
			&recv,
			&signature,
			&record.FilePath,
			&record.Line,
			&record.Col,
			&exported,
		); err != nil {
			return nil, errors.Wrap(err, "scan symbol inventory")
		}
		if recv.Valid {
			record.Recv = recv.String
		}
		if signature.Valid {
			record.Signature = signature.String
		}
		record.IsExported = exported == 1
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate symbol inventory")
	}
	return results, nil
}

func (s *Store) ListSymbolRefsUnresolved(ctx context.Context, filter SymbolRefUnresolvedFilter) ([]SymbolRefUnresolvedRecord, error) {
	query := `
		SELECT r.run_id, c.hash, r.symbol_hash, f.path, r.line, r.col, r.is_decl, r.source
		FROM symbol_refs_unresolved r
		LEFT JOIN commits c ON c.id = r.commit_id
		JOIN files f ON f.id = r.file_id
		WHERE (? = 0 OR r.run_id = ?)
		ORDER BY r.run_id, f.path, r.line, r.col`

	args := []interface{}{
		filter.RunID,
		filter.RunID,
	}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	} else if filter.Offset > 0 {
		query += " LIMIT -1 OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query unresolved symbol refs")
	}
	defer rows.Close()

	var results []SymbolRefUnresolvedRecord
	for rows.Next() {
		var record SymbolRefUnresolvedRecord
		var commitHash sql.NullString
		var symbolHash sql.NullString
		var isDecl int
		if err := rows.Scan(
			&record.RunID,
			&commitHash,
			&symbolHash,
			&record.FilePath,
			&record.Line,
			&record.Col,
			&isDecl,
			&record.Source,
		); err != nil {
			return nil, errors.Wrap(err, "scan unresolved symbol refs")
		}
		if commitHash.Valid {
			record.CommitHash = commitHash.String
		}
		if symbolHash.Valid {
			record.SymbolHash = symbolHash.String
		}
		record.IsDecl = isDecl == 1
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate unresolved symbol refs")
	}
	return results, nil
}

type SymbolRefFilter struct {
	RunID      int64
	SymbolHash string
	Path       string
	Limit      int
	Offset     int
}

type SymbolRefRecord struct {
	RunID      int64
	CommitHash string
	SymbolHash string
	FilePath   string
	Line       int
	Col        int
	IsDecl     bool
	Source     string
}

type DocHitFilter struct {
	RunID  int64
	Terms  []string
	Path   string
	Limit  int
	Offset int
}

type DocHitRecord struct {
	RunID     int64
	Term      string
	FilePath  string
	Line      int
	Col       int
	MatchText string
}

type FileFilter struct {
	Path     string
	Ext      string
	Exists   *bool
	IsBinary *bool
	Limit    int
	Offset   int
}

type FileRecord struct {
	Path     string
	Ext      string
	Exists   bool
	IsBinary bool
}

func (s *Store) ListSymbolRefs(ctx context.Context, filter SymbolRefFilter) ([]SymbolRefRecord, error) {
	query := `
		SELECT r.run_id, c.hash, d.symbol_hash, f.path, r.line, r.col, r.is_decl, r.source
		FROM symbol_refs r
		LEFT JOIN commits c ON c.id = r.commit_id
		JOIN symbol_defs d ON d.id = r.symbol_def_id
		JOIN files f ON f.id = r.file_id
		WHERE (? = 0 OR r.run_id = ?)
		  AND (? = '' OR d.symbol_hash = ?)
		  AND (? = '' OR f.path = ?)
		ORDER BY r.run_id, f.path, r.line, r.col`

	args := []interface{}{
		filter.RunID,
		filter.RunID,
		filter.SymbolHash,
		filter.SymbolHash,
		filter.Path,
		filter.Path,
	}

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	} else if filter.Offset > 0 {
		query += " LIMIT -1 OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query symbol refs")
	}
	defer rows.Close()

	var results []SymbolRefRecord
	for rows.Next() {
		var record SymbolRefRecord
		var commitHash sql.NullString
		var isDecl int
		if err := rows.Scan(
			&record.RunID,
			&commitHash,
			&record.SymbolHash,
			&record.FilePath,
			&record.Line,
			&record.Col,
			&isDecl,
			&record.Source,
		); err != nil {
			return nil, errors.Wrap(err, "scan symbol refs")
		}
		if commitHash.Valid {
			record.CommitHash = commitHash.String
		}
		record.IsDecl = isDecl == 1
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate symbol refs")
	}
	return results, nil
}

func (s *Store) ListDocHits(ctx context.Context, filter DocHitFilter) ([]DocHitRecord, error) {
	query := `
		SELECT h.run_id, h.term, f.path, h.line, h.col, h.match_text
		FROM doc_hits h
		JOIN files f ON f.id = h.file_id
		WHERE (? = 0 OR h.run_id = ?)`

	args := []interface{}{
		filter.RunID,
		filter.RunID,
	}

	if len(filter.Terms) > 0 {
		placeholders := make([]string, len(filter.Terms))
		for i, term := range filter.Terms {
			placeholders[i] = "?"
			args = append(args, term)
		}
		query += " AND h.term IN (" + strings.Join(placeholders, ",") + ")"
	}
	if filter.Path != "" {
		query += " AND f.path = ?"
		args = append(args, filter.Path)
	}

	query += " ORDER BY h.run_id, f.path, h.line, h.col"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	} else if filter.Offset > 0 {
		query += " LIMIT -1 OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query doc hits")
	}
	defer rows.Close()

	var results []DocHitRecord
	for rows.Next() {
		var record DocHitRecord
		if err := rows.Scan(
			&record.RunID,
			&record.Term,
			&record.FilePath,
			&record.Line,
			&record.Col,
			&record.MatchText,
		); err != nil {
			return nil, errors.Wrap(err, "scan doc hit")
		}
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate doc hits")
	}
	return results, nil
}

func (s *Store) ListFiles(ctx context.Context, filter FileFilter) ([]FileRecord, error) {
	query := `
		SELECT path, ext, file_exists, is_binary
		FROM files
		WHERE (? = '' OR path = ?)
		  AND (? = '' OR ext = ?)`

	args := []interface{}{
		filter.Path,
		filter.Path,
		filter.Ext,
		filter.Ext,
	}

	if filter.Exists != nil {
		query += " AND file_exists = ?"
		args = append(args, boolToInt(*filter.Exists))
	}
	if filter.IsBinary != nil {
		query += " AND is_binary = ?"
		args = append(args, boolToInt(*filter.IsBinary))
	}

	query += " ORDER BY path"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	} else if filter.Offset > 0 {
		query += " LIMIT -1 OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query files")
	}
	defer rows.Close()

	var results []FileRecord
	for rows.Next() {
		var record FileRecord
		var exists int
		var isBinary int
		if err := rows.Scan(
			&record.Path,
			&record.Ext,
			&exists,
			&isBinary,
		); err != nil {
			return nil, errors.Wrap(err, "scan file")
		}
		record.Exists = exists == 1
		record.IsBinary = isBinary == 1
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate files")
	}
	return results, nil
}
