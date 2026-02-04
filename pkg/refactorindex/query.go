package refactorindex

import (
	"context"
	"database/sql"

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
