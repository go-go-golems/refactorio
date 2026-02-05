package workbenchapi

import (
	"net/http"
	"strings"
)

type DocTermRecord struct {
	Term  string `json:"term"`
	Count int    `json:"count"`
}

type DocHitRecord struct {
	RunID     int64  `json:"run_id"`
	Term      string `json:"term"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Col       int    `json:"col"`
	MatchText string `json:"match_text"`
}

func (s *Server) registerDocRoutes() {
	s.apiMux.HandleFunc("/docs/terms", s.handleDocTerms)
	s.apiMux.HandleFunc("/docs/hits", s.handleDocHits)
}

func (s *Server) handleDocTerms(w http.ResponseWriter, r *http.Request) {
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

	if !hasTable(db, "doc_hits") {
		writeJSON(w, http.StatusOK, map[string]any{"items": []DocTermRecord{}})
		return
	}

	runID := parseRunID(r)
	prefix := strings.TrimSpace(r.URL.Query().Get("path_prefix"))
	limit, offset := parseLimitOffset(r, 100, 1000)

	query := `
SELECT h.term, count(*)
FROM doc_hits h
JOIN files f ON f.id = h.file_id
WHERE 1=1`
	args := []interface{}{}
	if runID > 0 {
		query += " AND h.run_id = ?"
		args = append(args, runID)
	}
	if prefix != "" {
		query += " AND f.path LIKE ?"
		args = append(args, prefix+"%")
	}
	query += " GROUP BY h.term ORDER BY count(*) DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query doc terms", nil)
		return
	}
	defer rows.Close()

	items := []DocTermRecord{}
	for rows.Next() {
		var record DocTermRecord
		if err := rows.Scan(&record.Term, &record.Count); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan doc terms", nil)
			return
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) handleDocHits(w http.ResponseWriter, r *http.Request) {
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

	if !hasTable(db, "doc_hits") {
		writeJSON(w, http.StatusOK, map[string]any{"items": []DocHitRecord{}})
		return
	}

	runID := parseRunID(r)
	term := strings.TrimSpace(r.URL.Query().Get("term"))
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	limit, offset := parseLimitOffset(r, 200, 2000)

	query := `
SELECT h.run_id, h.term, f.path, h.line, h.col, h.match_text
FROM doc_hits h
JOIN files f ON f.id = h.file_id
WHERE 1=1`
	args := []interface{}{}
	if runID > 0 {
		query += " AND h.run_id = ?"
		args = append(args, runID)
	}
	if term != "" {
		query += " AND h.term = ?"
		args = append(args, term)
	}
	if path != "" {
		query += " AND f.path = ?"
		args = append(args, path)
	}
	query += " ORDER BY f.path, h.line, h.col LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query doc hits", nil)
		return
	}
	defer rows.Close()

	items := []DocHitRecord{}
	for rows.Next() {
		var record DocHitRecord
		if err := rows.Scan(&record.RunID, &record.Term, &record.Path, &record.Line, &record.Col, &record.MatchText); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan doc hits", nil)
			return
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}
