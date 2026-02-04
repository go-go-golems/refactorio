package refactorindex

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type ReportConfig struct {
	DBPath    string
	RunID     int64
	OutputDir string
}

type ReportResult struct {
	Name     string
	Path     string
	RowCount int
}

func GenerateReports(ctx context.Context, cfg ReportConfig) ([]ReportResult, error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if cfg.RunID == 0 {
		return nil, errors.New("run id is required")
	}
	if strings.TrimSpace(cfg.OutputDir) == "" {
		cfg.OutputDir = "reports"
	}

	db, err := OpenDB(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, errors.Wrap(err, "create reports dir")
	}

	queries, err := fs.Glob(reportsFS, "reports/queries/*.sql")
	if err != nil {
		return nil, errors.Wrap(err, "list report queries")
	}
	if len(queries) == 0 {
		return nil, errors.New("no report queries found")
	}

	results := make([]ReportResult, 0, len(queries))
	for _, queryPath := range queries {
		name := strings.TrimSuffix(filepath.Base(queryPath), ".sql")
		templatePath := filepath.Join("reports", "templates", name+".md.tmpl")
		tplContent, err := reportsFS.ReadFile(templatePath)
		if err != nil {
			return nil, errors.Wrap(err, "read report template")
		}
		sqlContent, err := reportsFS.ReadFile(queryPath)
		if err != nil {
			return nil, errors.Wrap(err, "read report query")
		}

		rows, err := queryRows(ctx, db, string(sqlContent), cfg.RunID)
		if err != nil {
			return nil, err
		}

		data := map[string]interface{}{
			"RunID": cfg.RunID,
			"Rows":  rows,
		}

		reportPath := filepath.Join(cfg.OutputDir, name+".md")
		if err := renderTemplate(reportPath, string(tplContent), data); err != nil {
			return nil, err
		}

		results = append(results, ReportResult{
			Name:     name,
			Path:     reportPath,
			RowCount: len(rows),
		})
	}

	return results, nil
}

func queryRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]map[string]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "execute report query")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrap(err, "read report columns")
	}

	results := []map[string]string{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, errors.Wrap(err, "scan report row")
		}

		row := make(map[string]string, len(columns))
		for i, col := range columns {
			row[col] = fmt.Sprint(values[i])
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate report rows")
	}

	return results, nil
}

func renderTemplate(path string, tpl string, data interface{}) error {
	t, err := template.New(filepath.Base(path)).Parse(tpl)
	if err != nil {
		return errors.Wrap(err, "parse report template")
	}
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "create report file")
	}
	defer func() {
		_ = f.Close()
	}()

	if err := t.Execute(f, data); err != nil {
		return errors.Wrap(err, "execute report template")
	}
	return nil
}
