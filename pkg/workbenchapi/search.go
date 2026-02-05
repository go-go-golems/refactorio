package workbenchapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

type SearchRequest struct {
	WorkspaceID string            `json:"workspace_id"`
	SessionID   string            `json:"session_id"`
	Query       string            `json:"query"`
	Types       []string          `json:"types"`
	Filters     SearchFilters     `json:"filters"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
	RunIDs      map[string]int64  `json:"run_ids"`
	Extras      map[string]string `json:"extras"`
}

type SearchFilters struct {
	Path       string `json:"path"`
	Package    string `json:"pkg"`
	SymbolKind string `json:"symbol_kind"`
	Term       string `json:"term"`
}

type SearchResult struct {
	Type       string      `json:"type"`
	Primary    string      `json:"primary"`
	Secondary  string      `json:"secondary,omitempty"`
	Path       string      `json:"path,omitempty"`
	Line       int         `json:"line,omitempty"`
	Col        int         `json:"col,omitempty"`
	Snippet    string      `json:"snippet,omitempty"`
	RunID      int64       `json:"run_id,omitempty"`
	CommitHash string      `json:"commit_hash,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
}

type SymbolSearchRecord struct {
	RunID      int64  `json:"run_id"`
	SymbolHash string `json:"symbol_hash"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Pkg        string `json:"pkg"`
	Recv       string `json:"recv,omitempty"`
	Signature  string `json:"signature,omitempty"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	Exported   bool   `json:"is_exported"`
}

type CodeUnitSearchRecord struct {
	RunID     int64  `json:"run_id"`
	UnitHash  string `json:"unit_hash"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Pkg       string `json:"pkg"`
	Recv      string `json:"recv,omitempty"`
	Signature string `json:"signature,omitempty"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	BodyText  string `json:"body_text"`
}

type DiffSearchRecord struct {
	RunID    int64  `json:"run_id"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	LineOld  int    `json:"line_no_old,omitempty"`
	LineNew  int    `json:"line_no_new,omitempty"`
	Text     string `json:"text"`
	DiffFile int64  `json:"diff_file_id"`
	HunkID   int64  `json:"hunk_id"`
}

type CommitSearchRecord struct {
	RunID         int64  `json:"run_id"`
	Hash          string `json:"hash"`
	Subject       string `json:"subject"`
	Body          string `json:"body,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorEmail   string `json:"author_email,omitempty"`
	AuthorDate    string `json:"author_date,omitempty"`
	CommitterDate string `json:"committer_date,omitempty"`
}

type DocSearchRecord struct {
	RunID     int64  `json:"run_id"`
	Term      string `json:"term"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Col       int    `json:"col"`
	MatchText string `json:"match_text"`
}

type FileSearchRecord struct {
	Path   string `json:"path"`
	Ext    string `json:"ext"`
	Exists bool   `json:"exists"`
	Binary bool   `json:"is_binary"`
}

func (s *Server) registerSearchRoutes() {
	s.apiMux.HandleFunc("/search", s.handleSearch)
	s.apiMux.HandleFunc("/search/symbols", s.handleSearchSymbols)
	s.apiMux.HandleFunc("/search/code-units", s.handleSearchCodeUnits)
	s.apiMux.HandleFunc("/search/diff", s.handleSearchDiff)
	s.apiMux.HandleFunc("/search/commits", s.handleSearchCommits)
	s.apiMux.HandleFunc("/search/docs", s.handleSearchDocs)
	s.apiMux.HandleFunc("/search/files", s.handleSearchFiles)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	var req SearchRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "query is required", map[string]string{"field": "query"})
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

	types := req.Types
	if len(types) == 0 {
		types = []string{"symbols", "code_units", "diffs", "commits", "docs", "files"}
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	results := []SearchResult{}
	for _, t := range types {
		switch t {
		case "symbols":
			records, err := querySymbolSearch(db, query, req.Filters, req.RunIDs["symbols"], limit, offset)
			if err != nil {
				writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
				return
			}
			for _, record := range records {
				results = append(results, SearchResult{
					Type:      "symbol",
					Primary:   record.Name,
					Secondary: record.Pkg,
					Path:      record.File,
					Line:      record.Line,
					Col:       record.Col,
					RunID:     record.RunID,
					Payload:   record,
				})
			}
		case "code_units":
			records, err := queryCodeUnitSearch(db, query, req.Filters, req.RunIDs["code_units"], limit, offset)
			if err != nil {
				writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
				return
			}
			for _, record := range records {
				results = append(results, SearchResult{
					Type:      "code_unit",
					Primary:   record.Name,
					Secondary: record.Pkg,
					Path:      record.File,
					Line:      record.StartLine,
					RunID:     record.RunID,
					Snippet:   record.BodyText,
					Payload:   record,
				})
			}
		case "diffs":
			records, err := queryDiffSearch(db, query, req.Filters, req.RunIDs["diffs"], limit, offset)
			if err != nil {
				writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
				return
			}
			for _, record := range records {
				results = append(results, SearchResult{
					Type:    "diff",
					Primary: record.Path,
					Path:    record.Path,
					Line:    record.LineNew,
					RunID:   record.RunID,
					Snippet: record.Text,
					Payload: record,
				})
			}
		case "commits":
			records, err := queryCommitSearch(db, query, req.RunIDs["commits"], limit, offset)
			if err != nil {
				writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
				return
			}
			for _, record := range records {
				results = append(results, SearchResult{
					Type:       "commit",
					Primary:    record.Subject,
					Secondary:  record.AuthorName,
					CommitHash: record.Hash,
					RunID:      record.RunID,
					Snippet:    record.Body,
					Payload:    record,
				})
			}
		case "docs":
			records, err := queryDocSearch(db, query, req.Filters, req.RunIDs["docs"], limit, offset)
			if err != nil {
				writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
				return
			}
			for _, record := range records {
				results = append(results, SearchResult{
					Type:    "doc",
					Primary: record.Term,
					Path:    record.Path,
					Line:    record.Line,
					Col:     record.Col,
					RunID:   record.RunID,
					Snippet: record.MatchText,
					Payload: record,
				})
			}
		case "files":
			records, err := queryFileSearch(db, query, limit, offset)
			if err != nil {
				writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
				return
			}
			for _, record := range records {
				results = append(results, SearchResult{
					Type:    "file",
					Primary: record.Path,
					Path:    record.Path,
					Payload: record,
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": results})
}

func (s *Server) handleSearchSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "q is required", map[string]string{"field": "q"})
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

	filters := SearchFilters{
		Path:       strings.TrimSpace(r.URL.Query().Get("path")),
		Package:    strings.TrimSpace(r.URL.Query().Get("pkg")),
		SymbolKind: strings.TrimSpace(r.URL.Query().Get("kind")),
	}
	runID := parseRunID(r)
	limit, offset := parseLimitOffset(r, 100, 1000)

	records, err := querySymbolSearch(db, query, filters, runID, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records, "limit": limit, "offset": offset})
}

func (s *Server) handleSearchCodeUnits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "q is required", map[string]string{"field": "q"})
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

	filters := SearchFilters{
		Path:       strings.TrimSpace(r.URL.Query().Get("path")),
		Package:    strings.TrimSpace(r.URL.Query().Get("pkg")),
		SymbolKind: strings.TrimSpace(r.URL.Query().Get("kind")),
	}
	runID := parseRunID(r)
	limit, offset := parseLimitOffset(r, 100, 1000)

	records, err := queryCodeUnitSearch(db, query, filters, runID, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records, "limit": limit, "offset": offset})
}

func (s *Server) handleSearchDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "q is required", map[string]string{"field": "q"})
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

	filters := SearchFilters{Path: strings.TrimSpace(r.URL.Query().Get("path"))}
	runID := parseRunID(r)
	limit, offset := parseLimitOffset(r, 100, 1000)

	records, err := queryDiffSearch(db, query, filters, runID, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records, "limit": limit, "offset": offset})
}

func (s *Server) handleSearchCommits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "q is required", map[string]string{"field": "q"})
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

	runID := parseRunID(r)
	limit, offset := parseLimitOffset(r, 100, 1000)

	records, err := queryCommitSearch(db, query, runID, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records, "limit": limit, "offset": offset})
}

func (s *Server) handleSearchDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "q is required", map[string]string{"field": "q"})
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

	filters := SearchFilters{
		Path: strings.TrimSpace(r.URL.Query().Get("path")),
		Term: strings.TrimSpace(r.URL.Query().Get("term")),
	}
	runID := parseRunID(r)
	limit, offset := parseLimitOffset(r, 100, 1000)

	records, err := queryDocSearch(db, query, filters, runID, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records, "limit": limit, "offset": offset})
}

func (s *Server) handleSearchFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "q is required", map[string]string{"field": "q"})
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
	records, err := queryFileSearch(db, query, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records, "limit": limit, "offset": offset})
}

func querySymbolSearch(db *sql.DB, query string, filters SearchFilters, runID int64, limit int, offset int) ([]SymbolSearchRecord, error) {
	if !hasTable(db, "symbol_defs_fts") {
		return nil, errFTSUnavailable("symbol_defs_fts")
	}
	base := `
SELECT o.run_id, d.symbol_hash, d.name, d.kind, d.pkg, d.recv, d.signature,
       f.path, o.line, o.col, o.is_exported
FROM symbol_defs_fts fts
JOIN symbol_defs d ON d.id = fts.rowid
JOIN symbol_occurrences o ON o.symbol_def_id = d.id
JOIN files f ON f.id = o.file_id
WHERE symbol_defs_fts MATCH ?`
	args := []interface{}{query}
	if runID > 0 {
		base += " AND o.run_id = ?"
		args = append(args, runID)
	}
	if filters.SymbolKind != "" {
		base += " AND d.kind = ?"
		args = append(args, filters.SymbolKind)
	}
	if filters.Package != "" {
		base += " AND d.pkg = ?"
		args = append(args, filters.Package)
	}
	if filters.Path != "" {
		base += " AND f.path = ?"
		args = append(args, filters.Path)
	}
	base += " ORDER BY d.pkg, d.name, f.path, o.line, o.col LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SymbolSearchRecord{}
	for rows.Next() {
		var record SymbolSearchRecord
		var recv sql.NullString
		var signature sql.NullString
		var exported int
		if err := rows.Scan(
			&record.RunID,
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
		); err != nil {
			return nil, err
		}
		if recv.Valid {
			record.Recv = recv.String
		}
		if signature.Valid {
			record.Signature = signature.String
		}
		record.Exported = exported == 1
		items = append(items, record)
	}
	return items, nil
}

func queryCodeUnitSearch(db *sql.DB, query string, filters SearchFilters, runID int64, limit int, offset int) ([]CodeUnitSearchRecord, error) {
	if !hasTable(db, "code_unit_snapshots_fts") {
		return nil, errFTSUnavailable("code_unit_snapshots_fts")
	}
	base := `
SELECT s.run_id, cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.end_line, s.body_text
FROM code_unit_snapshots_fts fts
JOIN code_unit_snapshots s ON s.id = fts.rowid
JOIN code_units cu ON cu.id = s.code_unit_id
JOIN files f ON f.id = s.file_id
WHERE code_unit_snapshots_fts MATCH ?`
	args := []interface{}{query}
	if runID > 0 {
		base += " AND s.run_id = ?"
		args = append(args, runID)
	}
	if filters.SymbolKind != "" {
		base += " AND cu.kind = ?"
		args = append(args, filters.SymbolKind)
	}
	if filters.Package != "" {
		base += " AND cu.pkg = ?"
		args = append(args, filters.Package)
	}
	if filters.Path != "" {
		base += " AND f.path = ?"
		args = append(args, filters.Path)
	}
	base += " ORDER BY cu.pkg, cu.name, f.path, s.start_line LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []CodeUnitSearchRecord{}
	for rows.Next() {
		var record CodeUnitSearchRecord
		var recv sql.NullString
		var signature sql.NullString
		if err := rows.Scan(
			&record.RunID,
			&record.UnitHash,
			&record.Name,
			&record.Kind,
			&record.Pkg,
			&recv,
			&signature,
			&record.File,
			&record.StartLine,
			&record.EndLine,
			&record.BodyText,
		); err != nil {
			return nil, err
		}
		if recv.Valid {
			record.Recv = recv.String
		}
		if signature.Valid {
			record.Signature = signature.String
		}
		items = append(items, record)
	}
	return items, nil
}

func queryDiffSearch(db *sql.DB, query string, filters SearchFilters, runID int64, limit int, offset int) ([]DiffSearchRecord, error) {
	if !hasTable(db, "diff_lines_fts") {
		return nil, errFTSUnavailable("diff_lines_fts")
	}
	base := `
SELECT df.run_id, f.path, dl.kind, dl.line_no_old, dl.line_no_new, dl.text, dh.id, df.id
FROM diff_lines_fts fts
JOIN diff_lines dl ON dl.id = fts.rowid
JOIN diff_hunks dh ON dh.id = dl.hunk_id
JOIN diff_files df ON df.id = dh.diff_file_id
JOIN files f ON f.id = df.file_id
WHERE diff_lines_fts MATCH ?`
	args := []interface{}{query}
	if runID > 0 {
		base += " AND df.run_id = ?"
		args = append(args, runID)
	}
	if filters.Path != "" {
		base += " AND f.path = ?"
		args = append(args, filters.Path)
	}
	base += " ORDER BY f.path, dl.id LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []DiffSearchRecord{}
	for rows.Next() {
		var record DiffSearchRecord
		var lineOld sql.NullInt64
		var lineNew sql.NullInt64
		if err := rows.Scan(
			&record.RunID,
			&record.Path,
			&record.Kind,
			&lineOld,
			&lineNew,
			&record.Text,
			&record.HunkID,
			&record.DiffFile,
		); err != nil {
			return nil, err
		}
		if lineOld.Valid {
			record.LineOld = int(lineOld.Int64)
		}
		if lineNew.Valid {
			record.LineNew = int(lineNew.Int64)
		}
		items = append(items, record)
	}
	return items, nil
}

func queryCommitSearch(db *sql.DB, query string, runID int64, limit int, offset int) ([]CommitSearchRecord, error) {
	if !hasTable(db, "commits_fts") {
		return nil, errFTSUnavailable("commits_fts")
	}
	base := `
SELECT c.run_id, c.hash, c.subject, c.body, c.author_name, c.author_email, c.author_date, c.committer_date
FROM commits_fts fts
JOIN commits c ON c.id = fts.rowid
WHERE commits_fts MATCH ?`
	args := []interface{}{query}
	if runID > 0 {
		base += " AND c.run_id = ?"
		args = append(args, runID)
	}
	base += " ORDER BY c.committer_date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []CommitSearchRecord{}
	for rows.Next() {
		var record CommitSearchRecord
		var body sql.NullString
		var authorName sql.NullString
		var authorEmail sql.NullString
		var authorDate sql.NullString
		var committerDate sql.NullString
		if err := rows.Scan(
			&record.RunID,
			&record.Hash,
			&record.Subject,
			&body,
			&authorName,
			&authorEmail,
			&authorDate,
			&committerDate,
		); err != nil {
			return nil, err
		}
		if body.Valid {
			record.Body = body.String
		}
		if authorName.Valid {
			record.AuthorName = authorName.String
		}
		if authorEmail.Valid {
			record.AuthorEmail = authorEmail.String
		}
		if authorDate.Valid {
			record.AuthorDate = authorDate.String
		}
		if committerDate.Valid {
			record.CommitterDate = committerDate.String
		}
		items = append(items, record)
	}
	return items, nil
}

func queryDocSearch(db *sql.DB, query string, filters SearchFilters, runID int64, limit int, offset int) ([]DocSearchRecord, error) {
	if !hasTable(db, "doc_hits_fts") {
		return nil, errFTSUnavailable("doc_hits_fts")
	}
	base := `
SELECT h.run_id, h.term, f.path, h.line, h.col, h.match_text
FROM doc_hits_fts fts
JOIN doc_hits h ON h.id = fts.rowid
JOIN files f ON f.id = h.file_id
WHERE doc_hits_fts MATCH ?`
	args := []interface{}{query}
	if runID > 0 {
		base += " AND h.run_id = ?"
		args = append(args, runID)
	}
	if filters.Term != "" {
		base += " AND h.term = ?"
		args = append(args, filters.Term)
	}
	if filters.Path != "" {
		base += " AND f.path = ?"
		args = append(args, filters.Path)
	}
	base += " ORDER BY f.path, h.line, h.col LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []DocSearchRecord{}
	for rows.Next() {
		var record DocSearchRecord
		if err := rows.Scan(
			&record.RunID,
			&record.Term,
			&record.Path,
			&record.Line,
			&record.Col,
			&record.MatchText,
		); err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	return items, nil
}

func queryFileSearch(db *sql.DB, query string, limit int, offset int) ([]FileSearchRecord, error) {
	if !hasTable(db, "files_fts") {
		return nil, errFTSUnavailable("files_fts")
	}
	base := `
SELECT f.path, f.ext, f.file_exists, f.is_binary
FROM files_fts fts
JOIN files f ON f.id = fts.rowid
WHERE files_fts MATCH ?
ORDER BY f.path LIMIT ? OFFSET ?`
	args := []interface{}{query, limit, offset}

	rows, err := db.Query(base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []FileSearchRecord{}
	for rows.Next() {
		var record FileSearchRecord
		var exists int
		var binary int
		if err := rows.Scan(&record.Path, &record.Ext, &exists, &binary); err != nil {
			return nil, err
		}
		record.Exists = exists == 1
		record.Binary = binary == 1
		items = append(items, record)
	}
	return items, nil
}

func parseRunID(r *http.Request) int64 {
	value := strings.TrimSpace(r.URL.Query().Get("run_id"))
	if value == "" {
		return 0
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func errFTSUnavailable(table string) error {
	return &searchError{message: "fts table not available: " + table}
}

type searchError struct {
	message string
}

func (e *searchError) Error() string {
	return e.message
}
