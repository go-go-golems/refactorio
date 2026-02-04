package refactorindex

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestIngestSymbolsBestEffort(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n\ngo 1.25\n")
	if err := os.MkdirAll(filepath.Join(root, "good"), 0o755); err != nil {
		t.Fatalf("mkdir good: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "bad"), 0o755); err != nil {
		t.Fatalf("mkdir bad: %v", err)
	}
	writeFile(t, filepath.Join(root, "good", "good.go"), "package good\n\nfunc Good() int {\n\treturn 1\n}\n")
	writeFile(t, filepath.Join(root, "bad", "bad.go"), "package bad\n\nfunc Bad( {\n")

	dbPath := filepath.Join(root, "index.sqlite")
	result, err := IngestSymbols(ctx, IngestSymbolsConfig{
		DBPath:              dbPath,
		RootDir:             root,
		IgnorePackageErrors: true,
	})
	if err != nil {
		t.Fatalf("ingest symbols best-effort: %v", err)
	}
	if result.Symbols == 0 {
		t.Fatalf("expected symbols from valid packages")
	}
	if result.PackagesWithErrors == 0 {
		t.Fatalf("expected packages with errors")
	}
	if result.PackagesSkipped == 0 {
		t.Fatalf("expected skipped packages")
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	assertRunMetadataCount(t, db, result.RunID, "go_packages_error")
}

func assertRunMetadataCount(t *testing.T, db *sql.DB, runID int64, key string) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM run_kv WHERE run_id = ? AND key = ?", runID, key).Scan(&count); err != nil {
		t.Fatalf("count run_kv: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected run_kv entries for key %s", key)
	}
}
