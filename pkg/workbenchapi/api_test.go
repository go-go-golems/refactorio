package workbenchapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type listResponse struct {
	Items []map[string]interface{} `json:"items"`
}

type fileListResponse struct {
	Items []FileTreeItem `json:"items"`
}

type fileHistoryResponse struct {
	Items []FileHistoryRecord `json:"items"`
}

type codeUnitListResponse struct {
	Items []CodeUnitListRecord `json:"items"`
}

type diffRunsResponse struct {
	Items []RunRecord `json:"items"`
}

type diffFileResponse struct {
	Path  string           `json:"path"`
	RunID int64            `json:"run_id"`
	Hunks []DiffHunkRecord `json:"hunks"`
}

type dbInfoResponse struct {
	SchemaVersion int             `json:"schema_version"`
	Tables        map[string]bool `json:"tables"`
}

type seedResult struct {
	DBPath   string
	RunID    int64
	UnitHash string
	FilePath string
}

func TestDBInfoEndpoint(t *testing.T) {
	ctx := context.Background()
	seed := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/db/info?db_path="+seed.DBPath, nil)
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
	seed := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/runs?db_path="+seed.DBPath, nil)
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
	seed := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/search/symbols?q=Client&db_path="+seed.DBPath, nil)
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

func TestCodeUnitEndpoints(t *testing.T) {
	ctx := context.Background()
	seed := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/code-units?db_path="+seed.DBPath, nil)
	rec := httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var listPayload codeUnitListResponse
	if err := json.NewDecoder(rec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listPayload.Items) == 0 {
		t.Fatalf("expected code unit list results")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/code-units/"+seed.UnitHash+"?db_path="+seed.DBPath, nil)
	rec = httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var detail CodeUnitRecord
	if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if detail.UnitHash != seed.UnitHash {
		t.Fatalf("expected unit hash %s, got %s", seed.UnitHash, detail.UnitHash)
	}
}

func TestFileEndpoints(t *testing.T) {
	ctx := context.Background()
	seed := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/files?db_path="+seed.DBPath, nil)
	rec := httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var treePayload fileListResponse
	if err := json.NewDecoder(rec.Body).Decode(&treePayload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(treePayload.Items) == 0 {
		t.Fatalf("expected file tree results")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/files/history?path="+seed.FilePath+"&db_path="+seed.DBPath, nil)
	rec = httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var historyPayload fileHistoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&historyPayload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(historyPayload.Items) == 0 {
		t.Fatalf("expected file history results")
	}
}

func TestDiffEndpoints(t *testing.T) {
	ctx := context.Background()
	seed := seedTestDB(t, ctx)

	srv := NewServer(Config{BasePath: "/api"})
	req := httptest.NewRequest(http.MethodGet, "/api/diff-runs?db_path="+seed.DBPath, nil)
	rec := httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var runsPayload diffRunsResponse
	if err := json.NewDecoder(rec.Body).Decode(&runsPayload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(runsPayload.Items) == 0 {
		t.Fatalf("expected diff runs results")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/diff/"+strconv.FormatInt(seed.RunID, 10)+"/file?path="+seed.FilePath+"&db_path="+seed.DBPath, nil)
	rec = httptest.NewRecorder()
	srv.rootMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var diffPayload diffFileResponse
	if err := json.NewDecoder(rec.Body).Decode(&diffPayload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(diffPayload.Hunks) == 0 {
		t.Fatalf("expected diff hunks")
	}
}

func seedTestDB(t *testing.T, ctx context.Context) seedResult {
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

	filePath := "internal/api/client.go"
	fileID, err := store.GetOrCreateFile(ctx, tx, filePath)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE files SET file_exists = 1, is_binary = 0 WHERE id = ?", fileID); err != nil {
		t.Fatalf("update file metadata: %v", err)
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

	codeUnitHash := "unit-hash"
	codeUnitID, err := store.GetOrCreateCodeUnit(ctx, tx, refactorindex.CodeUnitDef{
		Pkg:  "github.com/acme/project/internal/api",
		Name: "Client",
		Kind: "type",
		Hash: codeUnitHash,
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
	fileOnDisk := filepath.Join(tempDir, "internal", "api", "client.go")
	if err := os.MkdirAll(filepath.Dir(fileOnDisk), 0o755); err != nil {
		t.Fatalf("mkdir for test file: %v", err)
	}
	if err := os.WriteFile(fileOnDisk, []byte("type Client struct{}"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	return seedResult{
		DBPath:   dbPath,
		RunID:    runID,
		UnitHash: codeUnitHash,
		FilePath: filePath,
	}
}
