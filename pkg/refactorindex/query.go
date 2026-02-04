package refactorindex

import (
	"context"

	"github.com/pkg/errors"
)

type DiffFileRecord struct {
	RunID   int64
	Status  string
	Path    string
	OldPath string
	NewPath string
}

func (s *Store) ListDiffFiles(ctx context.Context, runID int64) ([]DiffFileRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT df.run_id, df.status, f.path, df.old_path, df.new_path
		 FROM diff_files df
		 LEFT JOIN files f ON f.id = df.file_id
		 WHERE (? = 0 OR df.run_id = ?)
		 ORDER BY df.run_id, f.path`,
		runID,
		runID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query diff files")
	}
	defer rows.Close()

	var results []DiffFileRecord
	for rows.Next() {
		var record DiffFileRecord
		if err := rows.Scan(&record.RunID, &record.Status, &record.Path, &record.OldPath, &record.NewPath); err != nil {
			return nil, errors.Wrap(err, "scan diff file")
		}
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate diff files")
	}
	return results, nil
}
