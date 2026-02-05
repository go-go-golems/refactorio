package workbenchapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type SymbolRecord struct {
	SymbolHash string `json:"symbol_hash"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Pkg        string `json:"pkg"`
	Recv       string `json:"recv,omitempty"`
	Signature  string `json:"signature,omitempty"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	IsExported bool   `json:"is_exported"`
	RunID      int64  `json:"run_id,omitempty"`
}

type SymbolRefRecord struct {
	RunID      int64  `json:"run_id"`
	CommitHash string `json:"commit_hash,omitempty"`
	SymbolHash string `json:"symbol_hash"`
	Path       string `json:"path"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	IsDecl     bool   `json:"is_decl"`
	Source     string `json:"source"`
}

func (s *Server) registerSymbolRoutes() {
	s.apiMux.HandleFunc("/symbols", s.handleSymbols)
	s.apiMux.HandleFunc("/symbols/", s.handleSymbol)
}

func (s *Server) handleSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	ref, ok := s.requireWorkspaceRef(w, r)
	if !ok {
		return
	}
	db, closer, err := s.openWorkspaceDB(r.Context(), ref)
	if err != nil {
		writeError(w, http.StatusBadRequest, "db_error", "failed to open workspace database", nil)
		return
	}
	defer func() { _ = closer() }()

	limit, offset := parseLimitOffset(r, 100, 1000)
	runID := parseRunID(r)
	filters := refactorindex.SymbolInventoryFilter{
		RunID:        runID,
		ExportedOnly: parseBool(r.URL.Query().Get("exported_only")),
		Kind:         strings.TrimSpace(r.URL.Query().Get("kind")),
		Name:         strings.TrimSpace(r.URL.Query().Get("name")),
		Pkg:          strings.TrimSpace(r.URL.Query().Get("pkg")),
		Path:         strings.TrimSpace(r.URL.Query().Get("path")),
		Limit:        limit,
		Offset:       offset,
	}

	store := refactorindex.NewStore(db)
	records, err := store.ListSymbolInventory(r.Context(), filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query symbols", nil)
		return
	}

	items := make([]SymbolRecord, 0, len(records))
	for _, record := range records {
		items = append(items, SymbolRecord{
			SymbolHash: record.SymbolHash,
			Name:       record.Name,
			Kind:       record.Kind,
			Pkg:        record.Pkg,
			Recv:       record.Recv,
			Signature:  record.Signature,
			File:       record.FilePath,
			Line:       record.Line,
			Col:        record.Col,
			IsExported: record.IsExported,
			RunID:      record.RunID,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) handleSymbol(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/symbols/")
	if path == "" {
		writeError(w, http.StatusNotFound, "not_found", "symbol not found", nil)
		return
	}
	parts := strings.Split(path, "/")
	hash := parts[0]
	if hash == "" {
		writeError(w, http.StatusNotFound, "not_found", "symbol not found", nil)
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.getSymbol(w, r, hash)
		return
	}
	if len(parts) == 2 && parts[1] == "refs" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.getSymbolRefs(w, r, hash)
		return
	}

	writeError(w, http.StatusNotFound, "not_found", "symbol endpoint not found", nil)
}

func (s *Server) getSymbol(w http.ResponseWriter, r *http.Request, symbolHash string) {
	ref, ok := s.requireWorkspaceRef(w, r)
	if !ok {
		return
	}
	db, closer, err := s.openWorkspaceDB(r.Context(), ref)
	if err != nil {
		writeError(w, http.StatusBadRequest, "db_error", "failed to open workspace database", nil)
		return
	}
	defer func() { _ = closer() }()

	runID := parseRunID(r)
	query := `
SELECT d.symbol_hash, d.name, d.kind, d.pkg, d.recv, d.signature,
       f.path, o.line, o.col, o.is_exported, o.run_id
FROM symbol_defs d
JOIN symbol_occurrences o ON o.symbol_def_id = d.id
JOIN files f ON f.id = o.file_id
WHERE d.symbol_hash = ?`
	args := []interface{}{symbolHash}
	if runID > 0 {
		query += " AND o.run_id = ?"
		args = append(args, runID)
	}
	query += " ORDER BY o.line ASC LIMIT 1"

	row := db.QueryRowContext(r.Context(), query, args...)
	var record SymbolRecord
	var recv sql.NullString
	var signature sql.NullString
	var exported int
	if err := row.Scan(
		&record.SymbolHash,
		&record.Name,
		&record.Kind,
		&record.Pkg,
		&recv,
		&signature,
		&record.File,
		&record.Line,
		&record.Col,
		&exported,
		&record.RunID,
	); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "symbol not found", map[string]string{"symbol_hash": symbolHash})
			return
		}
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query symbol", nil)
		return
	}
	if recv.Valid {
		record.Recv = recv.String
	}
	if signature.Valid {
		record.Signature = signature.String
	}
	record.IsExported = exported == 1

	writeJSON(w, http.StatusOK, record)
}

func (s *Server) getSymbolRefs(w http.ResponseWriter, r *http.Request, symbolHash string) {
	ref, ok := s.requireWorkspaceRef(w, r)
	if !ok {
		return
	}
	db, closer, err := s.openWorkspaceDB(r.Context(), ref)
	if err != nil {
		writeError(w, http.StatusBadRequest, "db_error", "failed to open workspace database", nil)
		return
	}
	defer func() { _ = closer() }()

	if !hasTable(db, "symbol_refs") {
		writeJSON(w, http.StatusOK, map[string]any{"items": []SymbolRefRecord{}, "refs_available": false})
		return
	}

	limit, offset := parseLimitOffset(r, 200, 2000)
	runID := parseRunID(r)

	store := refactorindex.NewStore(db)
	records, err := store.ListSymbolRefs(r.Context(), refactorindex.SymbolRefFilter{
		RunID:      runID,
		SymbolHash: symbolHash,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query symbol refs", nil)
		return
	}

	items := make([]SymbolRefRecord, 0, len(records))
	for _, record := range records {
		items = append(items, SymbolRefRecord{
			RunID:      record.RunID,
			CommitHash: record.CommitHash,
			SymbolHash: record.SymbolHash,
			Path:       record.FilePath,
			Line:       record.Line,
			Col:        record.Col,
			IsDecl:     record.IsDecl,
			Source:     record.Source,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":          items,
		"refs_available": len(items) > 0,
		"limit":          limit,
		"offset":         offset,
	})
}

func parseBool(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}
