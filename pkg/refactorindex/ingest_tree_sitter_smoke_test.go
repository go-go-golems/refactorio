package refactorindex

import (
	"context"
	"database/sql"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIngestTreeSitterGoQuery(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	goFile := filepath.Join(root, "main.go")
	writeFile(t, goFile, "package main\n\nfunc Hello() {}\n")

	queryFile := filepath.Join(root, "queries.yaml")
	writeFile(t, queryFile, "language: go\nqueries:\n  funcs: |\n    (function_declaration name: (identifier) @name)\n")

	dbPath := filepath.Join(root, "index.sqlite")
	glob := filepath.Join(root, "*.go")

	result, err := IngestTreeSitter(ctx, IngestTreeSitterConfig{
		DBPath:     dbPath,
		RootDir:    root,
		Language:   "go",
		QueriesYML: queryFile,
		FileGlob:   glob,
	})
	if err != nil {
		t.Fatalf("ingest tree-sitter: %v", err)
	}
	if result.Captures == 0 {
		t.Fatalf("expected captures > 0")
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	assertCapture(t, db, "funcs", "name")
}

func TestIngestDocHits(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("rg not available")
	}

	ctx := context.Background()
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "README.md"), "hello glazed\n")
	terms := filepath.Join(root, "terms.txt")
	writeFile(t, terms, "glazed\n")

	dbPath := filepath.Join(root, "index.sqlite")
	result, err := IngestDocHits(ctx, IngestDocHitsConfig{
		DBPath:     dbPath,
		RootDir:    root,
		TermsFile:  terms,
		SourcesDir: filepath.Join(root, "sources"),
	})
	if err != nil {
		t.Fatalf("ingest doc hits: %v", err)
	}
	if result.Hits == 0 {
		t.Fatalf("expected hits > 0")
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	assertDocHitsFTSCount(t, db)
}

func TestParseGoplsLocation(t *testing.T) {
	cases := []struct {
		line     string
		filePath string
		lineNum  int
		colNum   int
	}{
		{"/tmp/main.go:10:5", "/tmp/main.go", 10, 5},
		{"/tmp/main.go:10:5-10:8", "/tmp/main.go", 10, 5},
		{"/tmp/main.go:10:5:10:8", "/tmp/main.go", 10, 5},
	}

	for _, tc := range cases {
		loc, err := parseGoplsLocation(tc.line)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.line, err)
		}
		if loc.FilePath != tc.filePath || loc.Line != tc.lineNum || loc.Col != tc.colNum {
			t.Fatalf("unexpected parse result for %q: %+v", tc.line, loc)
		}
	}
}

func assertCapture(t *testing.T, db *sql.DB, queryName string, captureName string) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM ts_captures WHERE query_name = ? AND capture_name = ?", queryName, captureName).Scan(&count); err != nil {
		t.Fatalf("query captures: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected capture for %s/%s", queryName, captureName)
	}
}

func assertDocHitsFTSCount(t *testing.T, db *sql.DB) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM doc_hits_fts").Scan(&count); err != nil {
		t.Fatalf("count doc_hits_fts: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected doc_hits_fts rows > 0")
	}
}
