package refactorindex

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIngestGoplsReferences(t *testing.T) {
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls not available")
	}

	ctx := context.Background()
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n\ngo 1.25\n")
	pkgDir := filepath.Join(root, "pkg", "foo")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}

	writeFile(t, filepath.Join(pkgDir, "foo.go"), "package foo\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n\nimport \"example.com/test/pkg/foo\"\n\nfunc main() {\n\t_ = foo.Add(1, 2)\n}\n")

	dbPath := filepath.Join(root, "index.sqlite")
	symbolsResult, err := IngestSymbols(ctx, IngestSymbolsConfig{
		DBPath:  dbPath,
		RootDir: root,
	})
	if err != nil {
		t.Fatalf("ingest symbols: %v", err)
	}
	if symbolsResult.Symbols == 0 {
		t.Fatalf("expected symbols")
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	target, err := loadGoplsTargetFromDB(t, db, "Add", "func")
	if err != nil {
		t.Fatalf("load target: %v", err)
	}

	result, err := IngestGoplsReferences(ctx, IngestGoplsRefsConfig{
		DBPath:           dbPath,
		RepoPath:         root,
		SourcesDir:       filepath.Join(root, "sources"),
		Targets:          []GoplsRefTarget{target},
		SkipSymbolLookup: false,
	})
	if err != nil {
		t.Fatalf("ingest gopls refs: %v", err)
	}
	if result.References == 0 {
		t.Fatalf("expected references > 0")
	}

	assertSymbolRefsCount(t, db, result.RunID)
}

func loadGoplsTargetFromDB(t *testing.T, db *sql.DB, name string, kind string) (GoplsRefTarget, error) {
	row := db.QueryRow(`
		SELECT d.symbol_hash, f.path, o.line, o.col
		FROM symbol_occurrences o
		JOIN symbol_defs d ON d.id = o.symbol_def_id
		JOIN files f ON f.id = o.file_id
		WHERE d.name = ? AND d.kind = ?
		LIMIT 1`, name, kind)

	var symbolHash string
	var filePath string
	var line int
	var col int
	if err := row.Scan(&symbolHash, &filePath, &line, &col); err != nil {
		return GoplsRefTarget{}, err
	}

	return GoplsRefTarget{
		SymbolHash: symbolHash,
		FilePath:   filePath,
		Line:       line,
		Col:        col,
	}, nil
}

func assertSymbolRefsCount(t *testing.T, db *sql.DB, runID int64) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM symbol_refs WHERE run_id = ?", runID).Scan(&count); err != nil {
		t.Fatalf("count symbol_refs: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected symbol_refs rows")
	}
}
