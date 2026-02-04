package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

func TestJSRunCommand(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "index.sqlite")

	db, err := refactorindex.OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	store := refactorindex.NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("init schema: %v", err)
	}

	runID, err := store.CreateRun(ctx, refactorindex.RunConfig{
		ToolVersion: "test",
		RootPath:    dir,
		SourcesDir:  dir,
	})
	if err != nil {
		_ = db.Close()
		t.Fatalf("create run: %v", err)
	}

	tx, err := store.BeginTx(ctx)
	if err != nil {
		_ = db.Close()
		t.Fatalf("begin tx: %v", err)
	}
	fileID, err := store.GetOrCreateFile(ctx, tx, "internal/api/client.go")
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert file: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE files SET file_exists = 1, is_binary = 0 WHERE id = ?", fileID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("update file flags: %v", err)
	}
	symID, err := store.GetOrCreateSymbolDef(ctx, tx, refactorindex.SymbolDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "hash-client",
	})
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert symbol def: %v", err)
	}
	if err := store.InsertSymbolOccurrence(ctx, tx, runID, nil, fileID, symID, 42, 6, true); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert symbol occurrence: %v", err)
	}
	if err := tx.Commit(); err != nil {
		_ = db.Close()
		t.Fatalf("commit tx: %v", err)
	}
	_ = db.Close()

	scriptPath := filepath.Join(dir, "script.js")
	script := `const idx = require("refactor-index"); idx.querySymbols({pkg:"github.com/acme/project/internal/api", name:"Client", kind:"type"});`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cmd := exec.Command("go", "run", "./cmd/refactorio", "js", "run", "--script", scriptPath, "--index-db", dbPath, "--run-id", fmt.Sprint(runID))
	cmd.Dir = "/home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Client") {
		t.Fatalf("expected output to contain symbol name, got: %s", output)
	}
}

func TestJSRunCommandTrace(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "index.sqlite")

	db, err := refactorindex.OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	store := refactorindex.NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("init schema: %v", err)
	}

	runID, err := store.CreateRun(ctx, refactorindex.RunConfig{
		ToolVersion: "test",
		RootPath:    dir,
		SourcesDir:  dir,
	})
	if err != nil {
		_ = db.Close()
		t.Fatalf("create run: %v", err)
	}

	tx, err := store.BeginTx(ctx)
	if err != nil {
		_ = db.Close()
		t.Fatalf("begin tx: %v", err)
	}
	fileID, err := store.GetOrCreateFile(ctx, tx, "internal/api/client.go")
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert file: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE files SET file_exists = 1, is_binary = 0 WHERE id = ?", fileID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("update file flags: %v", err)
	}
	symID, err := store.GetOrCreateSymbolDef(ctx, tx, refactorindex.SymbolDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "hash-client",
	})
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert symbol def: %v", err)
	}
	if err := store.InsertSymbolOccurrence(ctx, tx, runID, nil, fileID, symID, 42, 6, true); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert symbol occurrence: %v", err)
	}
	if err := tx.Commit(); err != nil {
		_ = db.Close()
		t.Fatalf("commit tx: %v", err)
	}
	_ = db.Close()

	scriptPath := filepath.Join(dir, "script.js")
	script := `const idx = require("refactor-index"); idx.querySymbols({pkg:"github.com/acme/project/internal/api", name:"Client", kind:"type"});`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	tracePath := filepath.Join(dir, "js_trace.jsonl")
	cmd := exec.Command("go", "run", "./cmd/refactorio", "js", "run", "--script", scriptPath, "--index-db", dbPath, "--run-id", fmt.Sprint(runID), "--trace", tracePath)
	cmd.Dir = "/home/manuel/workspaces/2026-02-04/implement-refactorio-refactoring/refactorio"
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, stderr.String())
	}

	data, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	if !strings.Contains(string(data), "\"action\":\"querySymbols\"") {
		t.Fatalf("expected trace to contain querySymbols, got: %s", string(data))
	}
}
