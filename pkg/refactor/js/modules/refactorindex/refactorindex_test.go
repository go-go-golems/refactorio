package refactorindex

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/refactorio/pkg/refactor/js"
	"github.com/go-go-golems/refactorio/pkg/refactor/js/modules"
	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

func setupStore(t *testing.T) (*refactorindex.Store, int64, func()) {
	t.Helper()

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
		_ = db.Close()
		t.Fatalf("insert file: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE files SET file_exists = 1, is_binary = 0 WHERE id = ?", fileID); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("update file flags: %v", err)
	}
	docID, err := store.GetOrCreateFile(ctx, tx, "docs/api.md")
	if err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("insert doc file: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE files SET file_exists = 1, is_binary = 0 WHERE id = ?", docID); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("update doc flags: %v", err)
	}

	symID, err := store.GetOrCreateSymbolDef(ctx, tx, refactorindex.SymbolDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "hash-client",
	})
	if err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("insert symbol def: %v", err)
	}

	if err := store.InsertSymbolOccurrence(ctx, tx, runID, nil, fileID, symID, 42, 6, true); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("insert symbol occurrence: %v", err)
	}
	if err := store.InsertSymbolRef(ctx, tx, runID, nil, symID, fileID, 42, 6, true, "gopls"); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("insert symbol ref: %v", err)
	}
	if err := store.InsertDocHit(ctx, tx, runID, nil, docID, 12, 2, "Client", "Client"); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("insert doc hit: %v", err)
	}

	if err := tx.Commit(); err != nil {
		_ = db.Close()
		t.Fatalf("commit tx: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}
	return store, runID, cleanup
}

func newRuntime(t *testing.T, store *refactorindex.Store, runID int64) *gojaRuntime {
	t.Helper()
	reg := modules.NewRegistry()
	reg.Register(NewModule(store, runID))
	vm, _, err := js.NewRuntime(js.RuntimeOptions{Registry: reg})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	return &gojaRuntime{vm: vm}
}

type gojaRuntime struct {
	vm *goja.Runtime
}

func (r *gojaRuntime) Run(t *testing.T, src string) interface{} {
	t.Helper()
	val, err := r.vm.RunString(src)
	if err != nil {
		t.Fatalf("run script: %v", err)
	}
	return val.Export()
}

func TestQuerySymbols(t *testing.T) {
	store, runID, cleanup := setupStore(t)
	defer cleanup()
	vm := newRuntime(t, store, runID)

	result := vm.Run(t, `const idx = require("refactor-index"); idx.querySymbols({pkg:"github.com/acme/project/internal/api", name:"Client", kind:"type"});`)
	assertRowCount(t, result, 1)
}

func TestQueryRefs(t *testing.T) {
	store, runID, cleanup := setupStore(t)
	defer cleanup()
	vm := newRuntime(t, store, runID)

	result := vm.Run(t, `const idx = require("refactor-index"); idx.queryRefs("hash-client");`)
	assertRowCount(t, result, 1)
}

func TestQueryDocHits(t *testing.T) {
	store, runID, cleanup := setupStore(t)
	defer cleanup()
	vm := newRuntime(t, store, runID)

	result := vm.Run(t, `const idx = require("refactor-index"); idx.queryDocHits(["Client"], {include:["docs/**/*.md"]});`)
	assertRowCount(t, result, 1)
}

func TestQueryFiles(t *testing.T) {
	store, runID, cleanup := setupStore(t)
	defer cleanup()
	vm := newRuntime(t, store, runID)

	result := vm.Run(t, `const idx = require("refactor-index"); idx.queryFiles({include:["docs/**/*.md"]});`)
	assertRowCount(t, result, 1)
}

func assertRowCount(t *testing.T, result interface{}, expected int) {
	t.Helper()

	switch rows := result.(type) {
	case []map[string]interface{}:
		if len(rows) != expected {
			t.Fatalf("expected %d rows, got %d", expected, len(rows))
		}
	case []interface{}:
		if len(rows) != expected {
			t.Fatalf("expected %d rows, got %d", expected, len(rows))
		}
	default:
		t.Fatalf("expected slice, got %T", result)
	}
}
