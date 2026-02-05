package workbenchapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type listResponse struct {
	Items []map[string]interface{} `json:"items"`
}

type dbInfoResponse struct {
	SchemaVersion int             `json:"schema_version"`
	Tables        map[string]bool `json:"tables"`
}

func TestDBInfoEndpoint(t *testing.T) {
	ctx := context.Background()
	dbPath := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/db/info?db_path="+dbPath, nil)
	rec := httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload dbInfoResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.SchemaVersion == 0 {
		t.Fatalf("expected schema_version > 0")
	}
	if !payload.Tables["meta_runs"] {
		t.Fatalf("expected meta_runs table present")
	}
}

func TestRunsListEndpoint(t *testing.T) {
	ctx := context.Background()
	dbPath := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/runs?db_path="+dbPath, nil)
	rec := httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload listResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatalf("expected at least one run")
	}
}

func TestSearchSymbolsEndpoint(t *testing.T) {
	ctx := context.Background()
	dbPath := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/search/symbols?q=Client&db_path="+dbPath, nil)
	rec := httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload listResponse
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatalf("expected symbol search results")
	}
}

func seedTestDB(t *testing.T, ctx context.Context) string {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "index.sqlite")

	db, err := refactorindex.OpenDB(ctx, dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store := refactorindex.NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	runID, err := store.CreateRun(ctx, refactorindex.RunConfig{
		ToolVersion: "test",
		GitFrom:     "HEAD~1",
		GitTo:       "HEAD",
		RootPath:    "/tmp/repo",
		ArgsJSON:    "{}",
		SourcesDir:  "/tmp/src",
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	tx, err := store.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	fileID, err := store.GetOrCreateFile(ctx, tx, "internal/api/client.go")
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	symbolID, err := store.GetOrCreateSymbolDef(ctx, tx, refactorindex.SymbolDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "symbol-hash",
	})
	if err != nil {
		t.Fatalf("create symbol def: %v", err)
	}
	if err := store.InsertSymbolOccurrence(ctx, tx, runID, nil, fileID, symbolID, 10, 5, true); err != nil {
		t.Fatalf("insert symbol occurrence: %v", err)
	}

	codeUnitID, err := store.GetOrCreateCodeUnit(ctx, tx, refactorindex.CodeUnitDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: "unit-hash",
	})
	if err != nil {
		t.Fatalf("create code unit: %v", err)
	}
	if err := store.InsertCodeUnitSnapshot(ctx, tx, runID, nil, fileID, codeUnitID, 10, 1, 12, 1, "bodyhash", "type Client struct{}", ""); err != nil {
		t.Fatalf("insert code unit snapshot: %v", err)
	}

	if err := store.InsertDocHit(ctx, tx, runID, nil, fileID, 20, 3, "Client", "Client"); err != nil {
		t.Fatalf("insert doc hit: %v", err)
	}

	diffFileID, err := store.InsertDiffFile(ctx, tx, runID, fileID, "M", "internal/api/client.go", "internal/api/client.go")
	if err != nil {
		t.Fatalf("insert diff file: %v", err)
	}
	hunkID, err := store.InsertDiffHunk(ctx, tx, diffFileID, 10, 1, 10, 1)
	if err != nil {
		t.Fatalf("insert diff hunk: %v", err)
	}
	oldLine := 10
	newLine := 10
	if err := store.InsertDiffLine(ctx, tx, hunkID, "+", &oldLine, &newLine, "+type Client struct{}"); err != nil {
		t.Fatalf("insert diff line: %v", err)
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
		t.Fatalf("insert commit: %v", err)
	}
	if err := store.InsertCommitFile(ctx, tx, commitID, fileID, "M", "internal/api/client.go", "internal/api/client.go", "", ""); err != nil {
		t.Fatalf("insert commit file: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		t.Fatalf("finish run: %v", err)
	}

	// ensure file exists for file endpoint tests if needed
	filePath := filepath.Join(tempDir, "internal", "api", "client.go")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir for test file: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("type Client struct{}"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	return dbPath
}
