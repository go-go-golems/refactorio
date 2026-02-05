package workbenchapi

import (
	"database/sql"
	"net/http"
	"strings"
)

type TreeSitterRecord struct {
	RunID       int64  `json:"run_id"`
	QueryName   string `json:"query_name"`
	CaptureName string `json:"capture_name"`
	NodeType    string `json:"node_type,omitempty"`
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartCol    int    `json:"start_col"`
	EndLine     int    `json:"end_line"`
	EndCol      int    `json:"end_col"`
	Snippet     string `json:"snippet"`
}

func (s *Server) registerTreeSitterRoutes() {
	s.apiMux.HandleFunc("/tree-sitter/captures", s.handleTreeSitterCaptures)
}

func (s *Server) handleTreeSitterCaptures(w http.ResponseWriter, r *http.Request) {
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

	if !hasTable(db, "ts_captures") {
		writeJSON(w, http.StatusOK, map[string]any{"items": []TreeSitterRecord{}})
		return
	}

	runID := parseRunID(r)
	queryName := strings.TrimSpace(r.URL.Query().Get("query_name"))
	captureName := strings.TrimSpace(r.URL.Query().Get("capture_name"))
	nodeType := strings.TrimSpace(r.URL.Query().Get("node_type"))
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	limit, offset := parseLimitOffset(r, 200, 2000)

	query := `
SELECT t.run_id, t.query_name, t.capture_name, t.node_type, f.path,
       t.start_line, t.start_col, t.end_line, t.end_col, t.snippet
FROM ts_captures t
JOIN files f ON f.id = t.file_id
WHERE 1=1`
	args := []interface{}{}
	if runID > 0 {
		query += " AND t.run_id = ?"
		args = append(args, runID)
	}
	if queryName != "" {
		query += " AND t.query_name = ?"
		args = append(args, queryName)
	}
	if captureName != "" {
		query += " AND t.capture_name = ?"
		args = append(args, captureName)
	}
	if nodeType != "" {
		query += " AND t.node_type = ?"
		args = append(args, nodeType)
	}
	if path != "" {
		query += " AND f.path = ?"
		args = append(args, path)
	}
	query += " ORDER BY f.path, t.start_line LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query tree-sitter captures", nil)
		return
	}
	defer rows.Close()

	items := []TreeSitterRecord{}
	for rows.Next() {
		var record TreeSitterRecord
		var nodeTypeVal sql.NullString
		if err := rows.Scan(
			&record.RunID,
			&record.QueryName,
			&record.CaptureName,
			&nodeTypeVal,
			&record.Path,
			&record.StartLine,
			&record.StartCol,
			&record.EndLine,
			&record.EndCol,
			&record.Snippet,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan tree-sitter captures", nil)
			return
		}
		if nodeTypeVal.Valid {
			record.NodeType = nodeTypeVal.String
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}
