package workbenchapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

type RunRecord struct {
	ID          int64  `json:"id"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at,omitempty"`
	Status      string `json:"status"`
	ToolVersion string `json:"tool_version"`
	GitFrom     string `json:"git_from"`
	GitTo       string `json:"git_to"`
	RootPath    string `json:"root_path"`
	ArgsJSON    string `json:"args_json"`
	ErrorJSON   string `json:"error_json,omitempty"`
	SourcesDir  string `json:"sources_dir"`
}

type RawOutputRecord struct {
	ID        int64  `json:"id"`
	RunID     int64  `json:"run_id"`
	Source    string `json:"source"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
}

func (s *Server) registerRunRoutes() {
	s.apiMux.HandleFunc("/runs", s.handleRuns)
	s.apiMux.HandleFunc("/runs/", s.handleRun)
	s.apiMux.HandleFunc("/raw-outputs", s.handleRawOutputs)
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
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
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	rootPath := strings.TrimSpace(r.URL.Query().Get("root_path"))
	gitFrom := strings.TrimSpace(r.URL.Query().Get("git_from"))
	gitTo := strings.TrimSpace(r.URL.Query().Get("git_to"))
	startedAfter := strings.TrimSpace(r.URL.Query().Get("started_after"))
	startedBefore := strings.TrimSpace(r.URL.Query().Get("started_before"))

	query := "SELECT id, started_at, finished_at, status, tool_version, git_from, git_to, root_path, args_json, error_json, sources_dir FROM meta_runs WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if rootPath != "" {
		query += " AND root_path = ?"
		args = append(args, rootPath)
	}
	if gitFrom != "" {
		query += " AND git_from = ?"
		args = append(args, gitFrom)
	}
	if gitTo != "" {
		query += " AND git_to = ?"
		args = append(args, gitTo)
	}
	if startedAfter != "" {
		query += " AND started_at >= ?"
		args = append(args, startedAfter)
	}
	if startedBefore != "" {
		query += " AND started_at <= ?"
		args = append(args, startedBefore)
	}
	query += " ORDER BY started_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query runs", nil)
		return
	}
	defer rows.Close()

	items := []RunRecord{}
	for rows.Next() {
		var record RunRecord
		var finished sql.NullString
		var statusVal sql.NullString
		var toolVersion sql.NullString
		var gitFromVal sql.NullString
		var gitToVal sql.NullString
		var rootPathVal sql.NullString
		var argsJSON sql.NullString
		var errorJSON sql.NullString
		var sourcesDir sql.NullString
		if err := rows.Scan(
			&record.ID,
			&record.StartedAt,
			&finished,
			&statusVal,
			&toolVersion,
			&gitFromVal,
			&gitToVal,
			&rootPathVal,
			&argsJSON,
			&errorJSON,
			&sourcesDir,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan runs", nil)
			return
		}
		if finished.Valid {
			record.FinishedAt = finished.String
		}
		if statusVal.Valid {
			record.Status = statusVal.String
		}
		if toolVersion.Valid {
			record.ToolVersion = toolVersion.String
		}
		if gitFromVal.Valid {
			record.GitFrom = gitFromVal.String
		}
		if gitToVal.Valid {
			record.GitTo = gitToVal.String
		}
		if rootPathVal.Valid {
			record.RootPath = rootPathVal.String
		}
		if argsJSON.Valid {
			record.ArgsJSON = argsJSON.String
		}
		if errorJSON.Valid {
			record.ErrorJSON = errorJSON.String
		}
		if sourcesDir.Valid {
			record.SourcesDir = sourcesDir.String
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/runs/")
	if path == "" {
		writeError(w, http.StatusNotFound, "not_found", "run not found", nil)
		return
	}
	parts := strings.Split(path, "/")
	idPart := parts[0]
	if idPart == "" {
		writeError(w, http.StatusNotFound, "not_found", "run not found", nil)
		return
	}
	runID, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", "run id must be an integer", nil)
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.getRun(w, r, runID)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "summary":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
				return
			}
			s.getRunSummary(w, r, runID)
			return
		case "raw-outputs":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
				return
			}
			s.getRunRawOutputs(w, r, runID)
			return
		default:
			writeError(w, http.StatusNotFound, "not_found", "run endpoint not found", nil)
			return
		}
	}

	writeError(w, http.StatusNotFound, "not_found", "run endpoint not found", nil)
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request, runID int64) {
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

	row := db.QueryRowContext(r.Context(), "SELECT id, started_at, finished_at, status, tool_version, git_from, git_to, root_path, args_json, error_json, sources_dir FROM meta_runs WHERE id = ?", runID)
	var record RunRecord
	var finished sql.NullString
	var statusVal sql.NullString
	var toolVersion sql.NullString
	var gitFromVal sql.NullString
	var gitToVal sql.NullString
	var rootPathVal sql.NullString
	var argsJSON sql.NullString
	var errorJSON sql.NullString
	var sourcesDir sql.NullString
	if err := row.Scan(
		&record.ID,
		&record.StartedAt,
		&finished,
		&statusVal,
		&toolVersion,
		&gitFromVal,
		&gitToVal,
		&rootPathVal,
		&argsJSON,
		&errorJSON,
		&sourcesDir,
	); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "run not found", map[string]any{"id": runID})
			return
		}
		writeError(w, http.StatusInternalServerError, "query_error", "failed to load run", nil)
		return
	}
	if finished.Valid {
		record.FinishedAt = finished.String
	}
	if statusVal.Valid {
		record.Status = statusVal.String
	}
	if toolVersion.Valid {
		record.ToolVersion = toolVersion.String
	}
	if gitFromVal.Valid {
		record.GitFrom = gitFromVal.String
	}
	if gitToVal.Valid {
		record.GitTo = gitToVal.String
	}
	if rootPathVal.Valid {
		record.RootPath = rootPathVal.String
	}
	if argsJSON.Valid {
		record.ArgsJSON = argsJSON.String
	}
	if errorJSON.Valid {
		record.ErrorJSON = errorJSON.String
	}
	if sourcesDir.Valid {
		record.SourcesDir = sourcesDir.String
	}

	writeJSON(w, http.StatusOK, record)
}

func (s *Server) getRunSummary(w http.ResponseWriter, r *http.Request, runID int64) {
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

	tables, _ := listTablesAndViews(db)
	counts := map[string]int{}

	countTable := func(key string, query string, args ...interface{}) {
		value, err := countQuery(db, query, args...)
		if err == nil {
			counts[key] = value
		}
	}

	if tables["diff_files"] {
		countTable("diff_files", "SELECT count(*) FROM diff_files WHERE run_id = ?", runID)
	}
	if tables["diff_hunks"] && tables["diff_files"] {
		countTable("diff_hunks", "SELECT count(*) FROM diff_hunks h JOIN diff_files f ON f.id = h.diff_file_id WHERE f.run_id = ?", runID)
	}
	if tables["diff_lines"] && tables["diff_hunks"] && tables["diff_files"] {
		countTable("diff_lines", "SELECT count(*) FROM diff_lines l JOIN diff_hunks h ON h.id = l.hunk_id JOIN diff_files f ON f.id = h.diff_file_id WHERE f.run_id = ?", runID)
	}
	if tables["symbol_occurrences"] {
		countTable("symbol_occurrences", "SELECT count(*) FROM symbol_occurrences WHERE run_id = ?", runID)
	}
	if tables["code_unit_snapshots"] {
		countTable("code_unit_snapshots", "SELECT count(*) FROM code_unit_snapshots WHERE run_id = ?", runID)
	}
	if tables["doc_hits"] {
		countTable("doc_hits", "SELECT count(*) FROM doc_hits WHERE run_id = ?", runID)
	}
	if tables["commits"] {
		countTable("commits", "SELECT count(*) FROM commits WHERE run_id = ?", runID)
	}
	if tables["commit_files"] && tables["commits"] {
		countTable("commit_files", "SELECT count(*) FROM commit_files cf JOIN commits c ON c.id = cf.commit_id WHERE c.run_id = ?", runID)
	}
	if tables["file_blobs"] && tables["commits"] {
		countTable("file_blobs", "SELECT count(*) FROM file_blobs fb JOIN commits c ON c.id = fb.commit_id WHERE c.run_id = ?", runID)
	}
	if tables["symbol_refs"] {
		countTable("symbol_refs", "SELECT count(*) FROM symbol_refs WHERE run_id = ?", runID)
	}
	if tables["symbol_refs_unresolved"] {
		countTable("symbol_refs_unresolved", "SELECT count(*) FROM symbol_refs_unresolved WHERE run_id = ?", runID)
	}
	if tables["raw_outputs"] {
		countTable("raw_outputs", "SELECT count(*) FROM raw_outputs WHERE run_id = ?", runID)
	}
	if tables["run_kv"] {
		countTable("run_kv", "SELECT count(*) FROM run_kv WHERE run_id = ?", runID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"run_id": runID,
		"counts": counts,
	})
}

func (s *Server) getRunRawOutputs(w http.ResponseWriter, r *http.Request, runID int64) {
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

	if !hasTable(db, "raw_outputs") {
		writeJSON(w, http.StatusOK, map[string]any{"items": []RawOutputRecord{}})
		return
	}

	rows, err := db.QueryContext(r.Context(), "SELECT id, run_id, source, path, created_at FROM raw_outputs WHERE run_id = ? ORDER BY created_at DESC", runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query raw outputs", nil)
		return
	}
	defer rows.Close()

	items := []RawOutputRecord{}
	for rows.Next() {
		var record RawOutputRecord
		if err := rows.Scan(&record.ID, &record.RunID, &record.Source, &record.Path, &record.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan raw outputs", nil)
			return
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleRawOutputs(w http.ResponseWriter, r *http.Request) {
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

	if !hasTable(db, "raw_outputs") {
		writeJSON(w, http.StatusOK, map[string]any{"items": []RawOutputRecord{}})
		return
	}

	limit, offset := parseLimitOffset(r, 100, 1000)
	query := "SELECT id, run_id, source, path, created_at FROM raw_outputs WHERE 1=1"
	args := []interface{}{}
	if source := strings.TrimSpace(r.URL.Query().Get("source")); source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}
	if runID := strings.TrimSpace(r.URL.Query().Get("run_id")); runID != "" {
		id, err := strconv.ParseInt(runID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "run_id must be an integer", nil)
			return
		}
		query += " AND run_id = ?"
		args = append(args, id)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query raw outputs", nil)
		return
	}
	defer rows.Close()

	items := []RawOutputRecord{}
	for rows.Next() {
		var record RawOutputRecord
		if err := rows.Scan(&record.ID, &record.RunID, &record.Source, &record.Path, &record.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan raw outputs", nil)
			return
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func parseLimitOffset(r *http.Request, defaultLimit int, maxLimit int) (int, int) {
	limit := defaultLimit
	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			limit = value
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			offset = value
		}
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func countQuery(db *sql.DB, query string, args ...interface{}) (int, error) {
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
