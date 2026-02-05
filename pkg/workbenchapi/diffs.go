package workbenchapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type DiffRunResponse struct {
	Items  []RunRecord `json:"items"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type DiffFileRecord struct {
	RunID   int64  `json:"run_id"`
	Status  string `json:"status"`
	Path    string `json:"path"`
	OldPath string `json:"old_path,omitempty"`
	NewPath string `json:"new_path,omitempty"`
}

type DiffHunkRecord struct {
	ID       int64          `json:"id"`
	OldStart int            `json:"old_start"`
	OldLines int            `json:"old_lines"`
	NewStart int            `json:"new_start"`
	NewLines int            `json:"new_lines"`
	Lines    []DiffLineItem `json:"lines"`
}

type DiffLineItem struct {
	Kind    string `json:"kind"`
	LineOld int    `json:"line_no_old,omitempty"`
	LineNew int    `json:"line_no_new,omitempty"`
	Text    string `json:"text"`
}

func (s *Server) registerDiffRoutes() {
	s.apiMux.HandleFunc("/diff-runs", s.handleDiffRuns)
	s.apiMux.HandleFunc("/diff/", s.handleDiff)
}

func (s *Server) handleDiffRuns(w http.ResponseWriter, r *http.Request) {
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

	if sessionID := strings.TrimSpace(r.URL.Query().Get("session_id")); sessionID != "" {
		overrides := []SessionOverride{}
		if ref.ID != "" {
			if cfg, err := s.loadWorkspaceConfig(); err == nil {
				if ws, _, ok := cfg.FindWorkspace(ref.ID); ok {
					overrides = ws.Sessions
				}
			}
		}

		sessions, err := computeSessions(db, ref, overrides)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "session_error", "failed to compute sessions", nil)
			return
		}
		for _, session := range sessions {
			if session.ID == sessionID {
				if session.Runs.Diff == nil {
					writeJSON(w, http.StatusOK, DiffRunResponse{Items: []RunRecord{}})
					return
				}
				run, err := queryRunByID(db, *session.Runs.Diff)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "query_error", "failed to load diff run", nil)
					return
				}
				writeJSON(w, http.StatusOK, DiffRunResponse{Items: []RunRecord{run}})
				return
			}
		}
		writeError(w, http.StatusNotFound, "not_found", "session not found", map[string]string{"id": sessionID})
		return
	}

	limit, offset := parseLimitOffset(r, 100, 1000)
	rows, err := db.QueryContext(r.Context(), `
SELECT DISTINCT m.id, m.started_at, m.finished_at, m.status, m.tool_version, m.git_from, m.git_to, m.root_path, m.args_json, m.error_json, m.sources_dir
FROM meta_runs m
JOIN diff_files df ON df.run_id = m.id
ORDER BY m.started_at DESC
LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query diff runs", nil)
		return
	}
	defer rows.Close()

	items := []RunRecord{}
	for rows.Next() {
		record, err := scanRunRecord(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan diff runs", nil)
			return
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, DiffRunResponse{Items: items, Limit: limit, Offset: offset})
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/diff/")
	if path == "" {
		writeError(w, http.StatusNotFound, "not_found", "diff not found", nil)
		return
	}
	parts := strings.Split(path, "/")
	idPart := parts[0]
	if idPart == "" {
		writeError(w, http.StatusNotFound, "not_found", "diff not found", nil)
		return
	}
	runID, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", "run_id must be an integer", nil)
		return
	}

	if len(parts) == 2 && parts[1] == "files" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.listDiffFiles(w, r, runID)
		return
	}
	if len(parts) == 2 && parts[1] == "file" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.getDiffFile(w, r, runID)
		return
	}

	writeError(w, http.StatusNotFound, "not_found", "diff endpoint not found", nil)
}

func (s *Server) listDiffFiles(w http.ResponseWriter, r *http.Request, runID int64) {
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
	store := refactorindex.NewStore(db)
	records, err := store.ListDiffFiles(r.Context(), refactorindex.DiffFileFilter{RunID: runID, Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query diff files", nil)
		return
	}

	items := make([]DiffFileRecord, 0, len(records))
	for _, record := range records {
		items = append(items, DiffFileRecord{
			RunID:   record.RunID,
			Status:  record.Status,
			Path:    record.Path,
			OldPath: record.OldPath,
			NewPath: record.NewPath,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) getDiffFile(w http.ResponseWriter, r *http.Request, runID int64) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "path is required", map[string]string{"field": "path"})
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

	hunks, err := loadDiffHunks(db, runID, path)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "diff file not found", map[string]string{"path": path})
			return
		}
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query diff file", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"path": path, "run_id": runID, "hunks": hunks})
}

func loadDiffHunks(db *sql.DB, runID int64, path string) ([]DiffHunkRecord, error) {
	rows, err := db.Query(`
SELECT dh.id, dh.old_start, dh.old_lines, dh.new_start, dh.new_lines
FROM diff_hunks dh
JOIN diff_files df ON df.id = dh.diff_file_id
JOIN files f ON f.id = df.file_id
WHERE df.run_id = ? AND f.path = ?
ORDER BY dh.id`, runID, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hunks := []DiffHunkRecord{}
	for rows.Next() {
		var hunk DiffHunkRecord
		if err := rows.Scan(&hunk.ID, &hunk.OldStart, &hunk.OldLines, &hunk.NewStart, &hunk.NewLines); err != nil {
			return nil, err
		}
		lines, err := loadDiffLines(db, hunk.ID)
		if err != nil {
			return nil, err
		}
		hunk.Lines = lines
		hunks = append(hunks, hunk)
	}
	if len(hunks) == 0 {
		return nil, sql.ErrNoRows
	}
	return hunks, nil
}

func loadDiffLines(db *sql.DB, hunkID int64) ([]DiffLineItem, error) {
	rows, err := db.Query(`
SELECT kind, line_no_old, line_no_new, text
FROM diff_lines
WHERE hunk_id = ?
ORDER BY id`, hunkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := []DiffLineItem{}
	for rows.Next() {
		var record DiffLineItem
		var oldLine sql.NullInt64
		var newLine sql.NullInt64
		if err := rows.Scan(&record.Kind, &oldLine, &newLine, &record.Text); err != nil {
			return nil, err
		}
		if oldLine.Valid {
			record.LineOld = int(oldLine.Int64)
		}
		if newLine.Valid {
			record.LineNew = int(newLine.Int64)
		}
		lines = append(lines, record)
	}
	return lines, nil
}

func queryRunByID(db *sql.DB, runID int64) (RunRecord, error) {
	row := db.QueryRow(`
SELECT id, started_at, finished_at, status, tool_version, git_from, git_to, root_path, args_json, error_json, sources_dir
FROM meta_runs
WHERE id = ?`, runID)
	return scanRunRecord(row)
}

func scanRunRecord(scanner interface {
	Scan(dest ...interface{}) error
}) (RunRecord, error) {
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
	if err := scanner.Scan(
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
		return RunRecord{}, err
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
	return record, nil
}
