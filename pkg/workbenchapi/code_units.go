package workbenchapi

import (
	"database/sql"
	"net/http"
	"strings"
)

type CodeUnitRecord struct {
	RunID     int64  `json:"run_id"`
	UnitHash  string `json:"unit_hash"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Pkg       string `json:"pkg"`
	Recv      string `json:"recv,omitempty"`
	Signature string `json:"signature,omitempty"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	StartCol  int    `json:"start_col"`
	EndLine   int    `json:"end_line"`
	EndCol    int    `json:"end_col"`
	BodyHash  string `json:"body_hash,omitempty"`
	BodyText  string `json:"body_text,omitempty"`
	DocText   string `json:"doc_text,omitempty"`
}

type CodeUnitListRecord struct {
	RunID     int64  `json:"run_id"`
	UnitHash  string `json:"unit_hash"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Pkg       string `json:"pkg"`
	Recv      string `json:"recv,omitempty"`
	Signature string `json:"signature,omitempty"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	StartCol  int    `json:"start_col"`
	EndLine   int    `json:"end_line"`
	EndCol    int    `json:"end_col"`
	BodyHash  string `json:"body_hash,omitempty"`
}

type CodeUnitDiffRequest struct {
	LeftRunID  int64 `json:"left_run_id"`
	RightRunID int64 `json:"right_run_id"`
}

func (s *Server) registerCodeUnitRoutes() {
	s.apiMux.HandleFunc("/code-units", s.handleCodeUnits)
	s.apiMux.HandleFunc("/code-units/", s.handleCodeUnit)
}

func (s *Server) handleCodeUnits(w http.ResponseWriter, r *http.Request) {
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
	filters := SearchFilters{
		Path:       strings.TrimSpace(r.URL.Query().Get("path")),
		Package:    strings.TrimSpace(r.URL.Query().Get("pkg")),
		SymbolKind: strings.TrimSpace(r.URL.Query().Get("kind")),
	}
	nameFilter := strings.TrimSpace(r.URL.Query().Get("name"))
	bodyQuery := strings.TrimSpace(r.URL.Query().Get("body_q"))

	var records []CodeUnitListRecord
	if bodyQuery != "" {
		var err error
		searchRecords, err := queryCodeUnitSearch(db, bodyQuery, filters, runID, limit, offset)
		if err != nil {
			writeError(w, http.StatusBadRequest, "search_error", err.Error(), nil)
			return
		}
		records = make([]CodeUnitListRecord, 0, len(searchRecords))
		for _, record := range searchRecords {
			records = append(records, CodeUnitListRecord{
				RunID:     record.RunID,
				UnitHash:  record.UnitHash,
				Name:      record.Name,
				Kind:      record.Kind,
				Pkg:       record.Pkg,
				Recv:      record.Recv,
				Signature: record.Signature,
				File:      record.File,
				StartLine: record.StartLine,
				StartCol:  record.StartCol,
				EndLine:   record.EndLine,
				EndCol:    record.EndCol,
			})
		}
	} else {
		rows, err := db.QueryContext(r.Context(), `
SELECT s.run_id, cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.start_col, s.end_line, s.end_col, s.body_hash
FROM code_unit_snapshots s
JOIN code_units cu ON cu.id = s.code_unit_id
JOIN files f ON f.id = s.file_id
WHERE (? = 0 OR s.run_id = ?)
  AND (? = '' OR cu.kind = ?)
  AND (? = '' OR cu.pkg = ?)
  AND (? = '' OR f.path = ?)
  AND (? = '' OR cu.name = ?)
ORDER BY cu.pkg, cu.name, f.path, s.start_line
LIMIT ? OFFSET ?`,
			runID, runID,
			filters.SymbolKind, filters.SymbolKind,
			filters.Package, filters.Package,
			filters.Path, filters.Path,
			nameFilter, nameFilter,
			limit, offset,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to query code units", nil)
			return
		}
		defer rows.Close()

		records = []CodeUnitListRecord{}
		for rows.Next() {
			var record CodeUnitListRecord
			var recv sql.NullString
			var signature sql.NullString
			var bodyHash sql.NullString
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
				&record.StartCol,
				&record.EndLine,
				&record.EndCol,
				&bodyHash,
			); err != nil {
				writeError(w, http.StatusInternalServerError, "query_error", "failed to scan code units", nil)
				return
			}
			if recv.Valid {
				record.Recv = recv.String
			}
			if signature.Valid {
				record.Signature = signature.String
			}
			if bodyHash.Valid {
				record.BodyHash = bodyHash.String
			}
			records = append(records, record)
		}
	}

	items := make([]CodeUnitListRecord, 0, len(records))
	for _, record := range records {
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) handleCodeUnit(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/code-units/")
	if path == "" {
		writeError(w, http.StatusNotFound, "not_found", "code unit not found", nil)
		return
	}
	parts := strings.Split(path, "/")
	hash := parts[0]
	if hash == "" {
		writeError(w, http.StatusNotFound, "not_found", "code unit not found", nil)
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.getCodeUnit(w, r, hash)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "history":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
				return
			}
			s.getCodeUnitHistory(w, r, hash)
			return
		case "diff":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
				return
			}
			s.diffCodeUnit(w, r, hash)
			return
		}
	}

	writeError(w, http.StatusNotFound, "not_found", "code unit endpoint not found", nil)
}

func (s *Server) getCodeUnit(w http.ResponseWriter, r *http.Request, unitHash string) {
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
SELECT s.run_id, cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.start_col, s.end_line, s.end_col, s.body_hash, s.body_text, s.doc_text
FROM code_units cu
JOIN code_unit_snapshots s ON s.code_unit_id = cu.id
JOIN files f ON f.id = s.file_id
WHERE cu.unit_hash = ?`
	args := []interface{}{unitHash}
	if runID > 0 {
		query += " AND s.run_id = ?"
		args = append(args, runID)
	}
	query += " ORDER BY s.run_id DESC LIMIT 1"

	row := db.QueryRowContext(r.Context(), query, args...)
	var record CodeUnitRecord
	var recv sql.NullString
	var signature sql.NullString
	var docText sql.NullString
	if err := row.Scan(
		&record.RunID,
		&record.UnitHash,
		&record.Name,
		&record.Kind,
		&record.Pkg,
		&recv,
		&signature,
		&record.File,
		&record.StartLine,
		&record.StartCol,
		&record.EndLine,
		&record.EndCol,
		&record.BodyHash,
		&record.BodyText,
		&docText,
	); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "code unit not found", map[string]string{"unit_hash": unitHash})
			return
		}
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query code unit", nil)
		return
	}
	if recv.Valid {
		record.Recv = recv.String
	}
	if signature.Valid {
		record.Signature = signature.String
	}
	if docText.Valid {
		record.DocText = docText.String
	}

	writeJSON(w, http.StatusOK, record)
}

func (s *Server) getCodeUnitHistory(w http.ResponseWriter, r *http.Request, unitHash string) {
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

	limit, offset := parseLimitOffset(r, 50, 500)
	rows, err := db.QueryContext(r.Context(), `
SELECT s.run_id, cu.unit_hash, cu.name, cu.kind, cu.pkg, cu.recv, cu.signature,
       f.path, s.start_line, s.start_col, s.end_line, s.end_col, s.body_hash
FROM code_units cu
JOIN code_unit_snapshots s ON s.code_unit_id = cu.id
JOIN files f ON f.id = s.file_id
WHERE cu.unit_hash = ?
ORDER BY s.run_id DESC
LIMIT ? OFFSET ?`, unitHash, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query code unit history", nil)
		return
	}
	defer rows.Close()

	items := []CodeUnitRecord{}
	for rows.Next() {
		var record CodeUnitRecord
		var recv sql.NullString
		var signature sql.NullString
		var bodyHash sql.NullString
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
			&record.StartCol,
			&record.EndLine,
			&record.EndCol,
			&bodyHash,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan code unit history", nil)
			return
		}
		if recv.Valid {
			record.Recv = recv.String
		}
		if signature.Valid {
			record.Signature = signature.String
		}
		if bodyHash.Valid {
			record.BodyHash = bodyHash.String
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) diffCodeUnit(w http.ResponseWriter, r *http.Request, unitHash string) {
	var req CodeUnitDiffRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.LeftRunID == 0 || req.RightRunID == 0 {
		writeError(w, http.StatusBadRequest, "invalid_argument", "left_run_id and right_run_id are required", nil)
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

	leftBody, err := loadCodeUnitBody(db, unitHash, req.LeftRunID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "left code unit snapshot not found", nil)
		return
	}
	rightBody, err := loadCodeUnitBody(db, unitHash, req.RightRunID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "right code unit snapshot not found", nil)
		return
	}

	diff := simpleLineDiff(leftBody, rightBody)
	writeJSON(w, http.StatusOK, map[string]any{"diff": diff})
}

func loadCodeUnitBody(db *sql.DB, unitHash string, runID int64) (string, error) {
	row := db.QueryRow(`
SELECT s.body_text
FROM code_units cu
JOIN code_unit_snapshots s ON s.code_unit_id = cu.id
WHERE cu.unit_hash = ? AND s.run_id = ?
ORDER BY s.run_id DESC
LIMIT 1`, unitHash, runID)
	var body sql.NullString
	if err := row.Scan(&body); err != nil {
		return "", err
	}
	if body.Valid {
		return body.String, nil
	}
	return "", sql.ErrNoRows
}

func simpleLineDiff(left string, right string) []string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	maxLen := len(leftLines)
	if len(rightLines) > maxLen {
		maxLen = len(rightLines)
	}
	out := []string{}
	for i := 0; i < maxLen; i++ {
		var l string
		if i < len(leftLines) {
			l = leftLines[i]
		}
		var r string
		if i < len(rightLines) {
			r = rightLines[i]
		}
		switch {
		case l == r:
			out = append(out, " "+l)
		case l == "":
			out = append(out, "+"+r)
		case r == "":
			out = append(out, "-"+l)
		default:
			out = append(out, "-"+l)
			out = append(out, "+"+r)
		}
	}
	return out
}
