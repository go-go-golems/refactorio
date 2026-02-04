package refactorindex

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIngestCommitsGolden(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repoPath := filepath.Join(root, "repo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	git(t, repoPath, "init")
	git(t, repoPath, "config", "user.email", "test@example.com")
	git(t, repoPath, "config", "user.name", "Refactor Index")

	writeFile(t, filepath.Join(repoPath, "fileA.txt"), "alpha\nbeta\n")
	writeFile(t, filepath.Join(repoPath, "fileB.txt"), "one\n")
	writeFile(t, filepath.Join(repoPath, "fileC.txt"), "gone\n")
	git(t, repoPath, "add", "-A")
	git(t, repoPath, "commit", "-m", "initial")
	fromRef := strings.TrimSpace(gitOut(t, repoPath, "rev-parse", "HEAD"))

	writeFile(t, filepath.Join(repoPath, "fileA.txt"), "alpha\nbeta2\n")
	git(t, repoPath, "mv", "fileB.txt", "fileB_renamed.txt")
	if err := os.Remove(filepath.Join(repoPath, "fileC.txt")); err != nil {
		t.Fatalf("remove fileC: %v", err)
	}
	writeFile(t, filepath.Join(repoPath, "fileD.txt"), "new\n")
	git(t, repoPath, "add", "-A")
	git(t, repoPath, "commit", "-m", "update")
	toRef := strings.TrimSpace(gitOut(t, repoPath, "rev-parse", "HEAD"))

	dbPath := filepath.Join(root, "index.sqlite")

	result, err := IngestCommits(ctx, IngestCommitsConfig{
		DBPath:   dbPath,
		RepoPath: repoPath,
		FromRef:  fromRef,
		ToRef:    toRef,
	})
	if err != nil {
		t.Fatalf("ingest commits: %v", err)
	}
	if result.CommitCount != 2 {
		t.Fatalf("expected 2 commits, got %d", result.CommitCount)
	}
	if result.FileCount < 5 {
		t.Fatalf("expected at least 5 commit files, got %d", result.FileCount)
	}
	if result.BlobCount < 3 {
		t.Fatalf("expected at least 3 blobs, got %d", result.BlobCount)
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	assertCommitCounts(t, db, result.RunID, 2, 5, 3)
}

func TestIngestCommitRangeDiffAndSymbols(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repoPath := filepath.Join(root, "repo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	git(t, repoPath, "init")
	git(t, repoPath, "config", "user.email", "test@example.com")
	git(t, repoPath, "config", "user.name", "Refactor Index")

	writeFile(t, filepath.Join(repoPath, "go.mod"), "module example.com/test\n\ngo 1.25\n")
	pkgDir := filepath.Join(repoPath, "pkg", "foo")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	writeFile(t, filepath.Join(pkgDir, "foo.go"), "package foo\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n")

	git(t, repoPath, "add", "-A")
	git(t, repoPath, "commit", "-m", "initial")
	fromRef := strings.TrimSpace(gitOut(t, repoPath, "rev-parse", "HEAD"))

	writeFile(t, filepath.Join(pkgDir, "foo.go"), "package foo\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n\nfunc Sub(a, b int) int {\n\treturn a - b\n}\n")
	git(t, repoPath, "add", "-A")
	git(t, repoPath, "commit", "-m", "update")
	toRef := strings.TrimSpace(gitOut(t, repoPath, "rev-parse", "HEAD"))

	dbPath := filepath.Join(root, "index.sqlite")

	result, err := IngestCommitRange(ctx, RangeIngestConfig{
		DBPath:            dbPath,
		RepoPath:          repoPath,
		FromRef:           fromRef,
		ToRef:             toRef,
		SourcesDir:        filepath.Join(root, "sources"),
		IncludeDiff:       true,
		IncludeSymbols:    true,
		IncludeCodeUnits:  true,
		IncludeDocHits:    false,
		IncludeTreeSitter: false,
		IncludeGopls:      false,
	})
	if err != nil {
		t.Fatalf("ingest range: %v", err)
	}
	if result.CommitLineageRunID == 0 {
		t.Fatalf("expected commit lineage run id")
	}
	if len(result.Commits) != 2 {
		t.Fatalf("expected 2 commit runs, got %d", len(result.Commits))
	}

	for _, commit := range result.Commits {
		if commit.DiffRunID == 0 || commit.SymbolsRunID == 0 || commit.CodeUnitsRunID == 0 {
			t.Fatalf("expected diff/symbols/code-units run ids > 0, got diff=%d symbols=%d code=%d", commit.DiffRunID, commit.SymbolsRunID, commit.CodeUnitsRunID)
		}
	}

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	assertCommitCounts(t, db, result.CommitLineageRunID, 2, 1, 1)
	assertSymbol(t, db, "Sub", "func")
	latestCommit := result.Commits[len(result.Commits)-1]
	assertSymbolOccurrenceCommitID(t, db, latestCommit.CommitHash)
	assertCodeUnitSnapshotCommitID(t, db, latestCommit.CommitHash)
}

func assertCommitCounts(t *testing.T, db *sql.DB, runID int64, commits int, commitFiles int, blobs int) {
	var commitCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM commits WHERE run_id = ?", runID).Scan(&commitCount); err != nil {
		t.Fatalf("count commits: %v", err)
	}
	if commitCount != commits {
		t.Fatalf("expected %d commits, got %d", commits, commitCount)
	}

	var commitFileCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM commit_files").Scan(&commitFileCount); err != nil {
		t.Fatalf("count commit_files: %v", err)
	}
	if commitFileCount < commitFiles {
		t.Fatalf("expected at least %d commit_files, got %d", commitFiles, commitFileCount)
	}

	var blobCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM file_blobs").Scan(&blobCount); err != nil {
		t.Fatalf("count file_blobs: %v", err)
	}
	if blobCount < blobs {
		t.Fatalf("expected at least %d file_blobs, got %d", blobs, blobCount)
	}
}

func assertSymbolOccurrenceCommitID(t *testing.T, db *sql.DB, commitHash string) {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*)
		 FROM symbol_occurrences o
		 JOIN commits c ON c.id = o.commit_id
		 WHERE c.hash = ?`,
		commitHash,
	).Scan(&count); err != nil {
		t.Fatalf("count symbol_occurrences by commit: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected symbol_occurrences with commit_id for %s", commitHash)
	}
}

func assertCodeUnitSnapshotCommitID(t *testing.T, db *sql.DB, commitHash string) {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*)
		 FROM code_unit_snapshots s
		 JOIN commits c ON c.id = s.commit_id
		 WHERE c.hash = ?`,
		commitHash,
	).Scan(&count); err != nil {
		t.Fatalf("count code_unit_snapshots by commit: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected code_unit_snapshots with commit_id for %s", commitHash)
	}
}
