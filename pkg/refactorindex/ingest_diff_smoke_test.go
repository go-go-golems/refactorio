package refactorindex

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIngestDiffGolden(t *testing.T) {
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
	sourcesDir := filepath.Join(root, "sources")

	result, err := IngestDiff(ctx, IngestDiffConfig{
		DBPath:     dbPath,
		RepoPath:   repoPath,
		FromRef:    fromRef,
		ToRef:      toRef,
		SourcesDir: sourcesDir,
	})
	if err != nil {
		t.Fatalf("ingest diff: %v", err)
	}
	if result.Files != 4 {
		t.Fatalf("expected 4 diff files, got %d", result.Files)
	}
	if result.Hunks == 0 || result.Lines == 0 {
		t.Fatalf("expected hunks/lines > 0, got hunks=%d lines=%d", result.Hunks, result.Lines)
	}

	assertRawOutputs(t, result.RunDir)

	db, err := OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	store := NewStore(db)
	records, err := store.ListDiffFiles(ctx, result.RunID)
	if err != nil {
		t.Fatalf("list diff files: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("expected 4 diff file records, got %d", len(records))
	}

	byPath := make(map[string]DiffFileRecord)
	for _, record := range records {
		byPath[record.Path] = record
	}

	assertStatus(t, byPath, "fileA.txt", "M", "fileA.txt", "fileA.txt")
	assertStatusPrefix(t, byPath, "fileB_renamed.txt", "R", "fileB.txt", "fileB_renamed.txt")
	assertStatus(t, byPath, "fileC.txt", "D", "fileC.txt", "")
	assertStatus(t, byPath, "fileD.txt", "A", "", "fileD.txt")

	assertTableCounts(t, db, result.RunID)
	assertDiffLinesFTSCount(t, db)
}

func assertRawOutputs(t *testing.T, runDir string) {
	if _, err := os.Stat(filepath.Join(runDir, "git-name-status.txt")); err != nil {
		t.Fatalf("missing git-name-status.txt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "git-diff-u0.patch")); err != nil {
		t.Fatalf("missing git-diff-u0.patch: %v", err)
	}
}

func assertTableCounts(t *testing.T, db *sql.DB, runID int64) {
	var diffFileCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM diff_files WHERE run_id = ?", runID).Scan(&diffFileCount); err != nil {
		t.Fatalf("count diff_files: %v", err)
	}
	if diffFileCount != 4 {
		t.Fatalf("expected 4 diff_files rows, got %d", diffFileCount)
	}

	var diffHunkCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM diff_hunks").Scan(&diffHunkCount); err != nil {
		t.Fatalf("count diff_hunks: %v", err)
	}
	if diffHunkCount == 0 {
		t.Fatalf("expected diff_hunks rows > 0")
	}

	var diffLineCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM diff_lines").Scan(&diffLineCount); err != nil {
		t.Fatalf("count diff_lines: %v", err)
	}
	if diffLineCount == 0 {
		t.Fatalf("expected diff_lines rows > 0")
	}
}

func assertDiffLinesFTSCount(t *testing.T, db *sql.DB) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM diff_lines_fts").Scan(&count); err != nil {
		t.Fatalf("count diff_lines_fts: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected diff_lines_fts rows > 0")
	}
}

func assertStatus(t *testing.T, records map[string]DiffFileRecord, path string, status string, oldPath string, newPath string) {
	record, ok := records[path]
	if !ok {
		t.Fatalf("expected record for %s", path)
	}
	if record.Status != status {
		t.Fatalf("expected status %s for %s, got %s", status, path, record.Status)
	}
	if record.OldPath != oldPath {
		t.Fatalf("expected old_path %q for %s, got %q", oldPath, path, record.OldPath)
	}
	if record.NewPath != newPath {
		t.Fatalf("expected new_path %q for %s, got %q", newPath, path, record.NewPath)
	}
}

func assertStatusPrefix(t *testing.T, records map[string]DiffFileRecord, path string, prefix string, oldPath string, newPath string) {
	record, ok := records[path]
	if !ok {
		t.Fatalf("expected record for %s", path)
	}
	if !strings.HasPrefix(record.Status, prefix) {
		t.Fatalf("expected status prefix %s for %s, got %s", prefix, path, record.Status)
	}
	if record.OldPath != oldPath {
		t.Fatalf("expected old_path %q for %s, got %q", oldPath, path, record.OldPath)
	}
	if record.NewPath != newPath {
		t.Fatalf("expected new_path %q for %s, got %q", newPath, path, record.NewPath)
	}
}

func writeFile(t *testing.T, path string, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func git(t *testing.T, repoPath string, args ...string) {
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
}

func gitOut(t *testing.T, repoPath string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output)
}
