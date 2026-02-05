package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

func main() {
	var dbPath string
	var repoRoot string
	flag.StringVar(&dbPath, "db", "", "path to sqlite db")
	flag.StringVar(&repoRoot, "repo", "", "repo root for file content")
	flag.Parse()

	if dbPath == "" || repoRoot == "" {
		log.Fatal("--db and --repo are required")
	}

	ctx := context.Background()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("mkdir db dir: %v", err)
	}

	db, err := refactorindex.OpenDB(ctx, dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	store := refactorindex.NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	runID, err := store.CreateRun(ctx, refactorindex.RunConfig{
		ToolVersion: "smoke",
		GitFrom:     "HEAD~1",
		GitTo:       "HEAD",
		RootPath:    repoRoot,
		ArgsJSON:    "{}",
		SourcesDir:  repoRoot,
	})
	if err != nil {
		log.Fatalf("create run: %v", err)
	}

	tx, err := store.BeginTx(ctx)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	filePath := "internal/api/client.go"
	fileID, err := store.GetOrCreateFile(ctx, tx, filePath)
	if err != nil {
		log.Fatalf("create file: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE files SET file_exists = 1, is_binary = 0 WHERE id = ?", fileID); err != nil {
		log.Fatalf("update file metadata: %v", err)
	}

	symbolID, err := store.GetOrCreateSymbolDef(ctx, tx, refactorindex.SymbolDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "symbol-hash",
	})
	if err != nil {
		log.Fatalf("create symbol def: %v", err)
	}
	if err := store.InsertSymbolOccurrence(ctx, tx, runID, nil, fileID, symbolID, 10, 5, true); err != nil {
		log.Fatalf("insert symbol occurrence: %v", err)
	}

	codeUnitID, err := store.GetOrCreateCodeUnit(ctx, tx, refactorindex.CodeUnitDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "unit-hash",
	})
	if err != nil {
		log.Fatalf("create code unit: %v", err)
	}
	if err := store.InsertCodeUnitSnapshot(ctx, tx, runID, nil, fileID, codeUnitID, 10, 1, 12, 1, "bodyhash", "type Client struct{}", ""); err != nil {
		log.Fatalf("insert code unit snapshot: %v", err)
	}

	if err := store.InsertDocHit(ctx, tx, runID, nil, fileID, 20, 3, "Client", "Client"); err != nil {
		log.Fatalf("insert doc hit: %v", err)
	}

	diffFileID, err := store.InsertDiffFile(ctx, tx, runID, fileID, "M", filePath, filePath)
	if err != nil {
		log.Fatalf("insert diff file: %v", err)
	}
	hunkID, err := store.InsertDiffHunk(ctx, tx, diffFileID, 10, 1, 10, 1)
	if err != nil {
		log.Fatalf("insert diff hunk: %v", err)
	}
	oldLine := 10
	newLine := 10
	if err := store.InsertDiffLine(ctx, tx, hunkID, "+", &oldLine, &newLine, "+type Client struct{}"); err != nil {
		log.Fatalf("insert diff line: %v", err)
	}

	commitID, err := store.InsertCommit(ctx, tx, runID, refactorindex.CommitInfo{
		Hash:          "abc123",
		AuthorName:    "Dev",
		AuthorEmail:   "dev@example.com",
		AuthorDate:    "2026-02-01T00:00:00Z",
		CommitterDate: "2026-02-01T00:00:00Z",
		Subject:       "Add Client",
		Body:          "body",
	})
	if err != nil {
		log.Fatalf("insert commit: %v", err)
	}
	if err := store.InsertCommitFile(ctx, tx, commitID, fileID, "M", filePath, filePath, "", ""); err != nil {
		log.Fatalf("insert commit file: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit tx: %v", err)
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		log.Fatalf("finish run: %v", err)
	}

	fileOnDisk := filepath.Join(repoRoot, "internal", "api", "client.go")
	if err := os.MkdirAll(filepath.Dir(fileOnDisk), 0o755); err != nil {
		log.Fatalf("mkdir repo file dir: %v", err)
	}
	if err := os.WriteFile(fileOnDisk, []byte("type Client struct{}"), 0o644); err != nil {
		log.Fatalf("write repo file: %v", err)
	}
}
