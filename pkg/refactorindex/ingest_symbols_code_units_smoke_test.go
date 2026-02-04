package refactorindex

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestIngestSymbolsAndCodeUnitsGolden(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n\ngo 1.25\n")

	pkgDir := filepath.Join(root, "pkg", "foo")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}

	writeFile(t, filepath.Join(pkgDir, "foo.go"), `package foo

// Person represents a user.
type Person struct {
	Name string
}

// Greet returns a greeting.
func (p *Person) Greet() string {
	return "hi"
}

// Add sums ints.
func Add(a, b int) int {
	return a + b
}

const Answer = 42
`)

	dbPath := filepath.Join(root, "index.sqlite")

	symbolsResult, err := IngestSymbols(ctx, IngestSymbolsConfig{
		DBPath:  dbPath,
		RootDir: root,
	})
	if err != nil {
		t.Fatalf("ingest symbols: %v", err)
	}
	if symbolsResult.Symbols == 0 || symbolsResult.Occurrences == 0 {
		t.Fatalf("expected symbols/occurrences > 0, got %d/%d", symbolsResult.Symbols, symbolsResult.Occurrences)
	}

	codeUnitsResult, err := IngestCodeUnits(ctx, IngestCodeUnitsConfig{
		DBPath:  dbPath,
		RootDir: root,
	})
	if err != nil {
		t.Fatalf("ingest code units: %v", err)
	}
	if codeUnitsResult.CodeUnits < 3 || codeUnitsResult.Snapshots < 3 {
		t.Fatalf("expected at least 3 code units/snapshots, got %d/%d", codeUnitsResult.CodeUnits, codeUnitsResult.Snapshots)
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	assertSymbol(t, db, "Person", "type")
	assertSymbol(t, db, "Greet", "method")
	assertSymbol(t, db, "Add", "func")
	assertSymbol(t, db, "Answer", "const")

	assertCodeUnit(t, db, "Person", "type")
	assertCodeUnit(t, db, "Greet", "method")
	assertCodeUnit(t, db, "Add", "func")

	assertSnapshotCount(t, db, codeUnitsResult.RunID, 3)
	assertSnapshotBodyLike(t, db, "type Person")
	assertSnapshotBodyLike(t, db, "func Add")
}

func assertSymbol(t *testing.T, db *sql.DB, name string, kind string) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM symbol_defs WHERE name = ? AND kind = ?", name, kind).Scan(&count); err != nil {
		t.Fatalf("query symbol %s/%s: %v", name, kind, err)
	}
	if count == 0 {
		t.Fatalf("expected symbol %s/%s", name, kind)
	}
}

func assertCodeUnit(t *testing.T, db *sql.DB, name string, kind string) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM code_units WHERE name = ? AND kind = ?", name, kind).Scan(&count); err != nil {
		t.Fatalf("query code unit %s/%s: %v", name, kind, err)
	}
	if count == 0 {
		t.Fatalf("expected code unit %s/%s", name, kind)
	}
}

func assertSnapshotCount(t *testing.T, db *sql.DB, runID int64, minCount int) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM code_unit_snapshots WHERE run_id = ?", runID).Scan(&count); err != nil {
		t.Fatalf("query snapshots: %v", err)
	}
	if count < minCount {
		t.Fatalf("expected at least %d snapshots, got %d", minCount, count)
	}
}

func assertSnapshotBodyLike(t *testing.T, db *sql.DB, needle string) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM code_unit_snapshots WHERE body_text LIKE ?", "%"+needle+"%").Scan(&count); err != nil {
		t.Fatalf("query snapshot body %s: %v", needle, err)
	}
	if count == 0 {
		t.Fatalf("expected snapshot body containing %q", needle)
	}
}
