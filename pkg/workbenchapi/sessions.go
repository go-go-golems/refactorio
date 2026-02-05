package workbenchapi

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"
)

type sessionBuilder struct {
	session     *Session
	domainTimes map[string]time.Time
	updatedAt   time.Time
	updatedRaw  string
}

func (s *Server) registerSessionRoutes() {
	s.apiMux.HandleFunc("/sessions", s.handleSessions)
	s.apiMux.HandleFunc("/sessions/", s.handleSession)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSessions(w, r)
	case http.MethodPost:
		s.createSessionOverride(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/sessions/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "session not found", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getSession(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, map[string]any{"items": sessions})
}

func (s *Server) getSession(w http.ResponseWriter, r *http.Request, id string) {
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
		if session.ID == id {
			writeJSON(w, http.StatusOK, session)
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "session not found", map[string]string{"id": id})
}

func (s *Server) createSessionOverride(w http.ResponseWriter, r *http.Request) {
	ref, ok := s.requireWorkspaceRef(w, r)
	if !ok {
		return
	}
	if ref.ID == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "workspace_id is required for session overrides", nil)
		return
	}

	var input SessionOverride
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}

	input.ID = strings.TrimSpace(input.ID)
	if input.ID == "" {
		input.ID = buildSessionID(ref.ID, input.RootPath, input.GitFrom, input.GitTo, 0)
	}

	cfg, err := s.loadWorkspaceConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
		return
	}
	ws, idx, ok := cfg.FindWorkspace(ref.ID)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found", map[string]string{"id": ref.ID})
		return
	}

	replaced := false
	for i, override := range ws.Sessions {
		if override.ID == input.ID {
			ws.Sessions[i] = input
			replaced = true
			break
		}
	}
	if !replaced {
		ws.Sessions = append(ws.Sessions, input)
	}

	cfg.Workspaces[idx] = ws
	if err := s.saveWorkspaceConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to save workspace config", nil)
		return
	}

	writeJSON(w, http.StatusCreated, input)
}

func computeSessions(db *sql.DB, ref WorkspaceRef, overrides []SessionOverride) ([]Session, error) {
	runs, err := loadRunsForSessions(db)
	if err != nil {
		return nil, err
	}
	presence, _ := listTablesAndViews(db)

	builders := map[string]*sessionBuilder{}
	idMap := map[string]int{}

	for _, run := range runs {
		key := sessionKey(run)
		builder := builders[key]
		if builder == nil {
			builder = &sessionBuilder{
				session:     &Session{Runs: SessionRuns{}, Availability: map[string]bool{}},
				domainTimes: map[string]time.Time{},
			}
			sessionID := buildSessionID(ref.ID, run.RootPath, run.GitFrom, run.GitTo, run.ID)
			if count, ok := idMap[sessionID]; ok {
				count++
				idMap[sessionID] = count
				sessionID = fmt.Sprintf("%s-%d", sessionID, count)
			} else {
				idMap[sessionID] = 1
			}
			builder.session.ID = sessionID
			builders[key] = builder
		}

		if builder.session.RootPath == "" {
			builder.session.RootPath = run.RootPath
		}
		if builder.session.GitFrom == "" {
			builder.session.GitFrom = run.GitFrom
		}
		if builder.session.GitTo == "" {
			builder.session.GitTo = run.GitTo
		}
		builder.session.WorkspaceID = ref.ID

		runTime, runRaw := runTimestamp(run)
		if runTime.After(builder.updatedAt) {
			builder.updatedAt = runTime
			builder.updatedRaw = runRaw
		} else if builder.updatedRaw == "" {
			builder.updatedRaw = runRaw
		}

		if presence["diff_files"] && tableHasRunData(db, "diff_files", run.ID) {
			setRun(&builder.session.Runs.Diff, builder.domainTimes, "diff", run.ID, runTime)
		}
		if presence["symbol_occurrences"] && tableHasRunData(db, "symbol_occurrences", run.ID) {
			setRun(&builder.session.Runs.Symbols, builder.domainTimes, "symbols", run.ID, runTime)
		}
		if presence["code_unit_snapshots"] && tableHasRunData(db, "code_unit_snapshots", run.ID) {
			setRun(&builder.session.Runs.CodeUnits, builder.domainTimes, "code_units", run.ID, runTime)
		}
		if presence["doc_hits"] && tableHasRunData(db, "doc_hits", run.ID) {
			setRun(&builder.session.Runs.DocHits, builder.domainTimes, "doc_hits", run.ID, runTime)
		}
		if presence["commits"] && tableHasRunData(db, "commits", run.ID) {
			setRun(&builder.session.Runs.Commits, builder.domainTimes, "commits", run.ID, runTime)
		}
		hasRefs := false
		if presence["symbol_refs"] && tableHasRunData(db, "symbol_refs", run.ID) {
			hasRefs = true
		}
		if presence["symbol_refs_unresolved"] && tableHasRunData(db, "symbol_refs_unresolved", run.ID) {
			hasRefs = true
		}
		if hasRefs {
			setRun(&builder.session.Runs.GoplsRefs, builder.domainTimes, "gopls_refs", run.ID, runTime)
		}
		if presence["ts_captures"] && tableHasRunData(db, "ts_captures", run.ID) {
			setRun(&builder.session.Runs.TreeSitter, builder.domainTimes, "tree_sitter", run.ID, runTime)
		}
	}

	sessions := []Session{}
	for _, builder := range builders {
		builder.session.LastUpdated = builder.updatedRaw
		builder.session.Availability = availabilityFromRuns(builder.session.Runs)
		sessions = append(sessions, *builder.session)
	}

	for _, override := range overrides {
		id := strings.TrimSpace(override.ID)
		if id == "" {
			id = buildSessionID(ref.ID, override.RootPath, override.GitFrom, override.GitTo, 0)
		}
		merged := Session{
			ID:           id,
			WorkspaceID:  ref.ID,
			RootPath:     override.RootPath,
			GitFrom:      override.GitFrom,
			GitTo:        override.GitTo,
			Runs:         override.Runs,
			Availability: availabilityFromRuns(override.Runs),
		}
		updated := false
		for idx := range sessions {
			if sessions[idx].ID == id {
				sessions[idx] = merged
				updated = true
				break
			}
		}
		if !updated {
			sessions = append(sessions, merged)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastUpdated > sessions[j].LastUpdated
	})

	return sessions, nil
}

func availabilityFromRuns(runs SessionRuns) map[string]bool {
	return map[string]bool{
		"commits":     runs.Commits != nil,
		"diff":        runs.Diff != nil,
		"symbols":     runs.Symbols != nil,
		"code_units":  runs.CodeUnits != nil,
		"doc_hits":    runs.DocHits != nil,
		"gopls_refs":  runs.GoplsRefs != nil,
		"tree_sitter": runs.TreeSitter != nil,
	}
}

func runTimestamp(run RunRecord) (time.Time, string) {
	if run.FinishedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, run.FinishedAt); err == nil {
			return parsed, run.FinishedAt
		}
	}
	if run.StartedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, run.StartedAt); err == nil {
			return parsed, run.StartedAt
		}
		return time.Time{}, run.StartedAt
	}
	return time.Time{}, ""
}

func setRun(target **int64, times map[string]time.Time, key string, runID int64, runTime time.Time) {
	current, ok := times[key]
	if !ok || runTime.After(current) {
		value := runID
		*target = &value
		times[key] = runTime
	}
}

func tableHasRunData(db *sql.DB, table string, runID int64) bool {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE run_id = ? LIMIT 1", table)
	var exists int
	if err := db.QueryRow(query, runID).Scan(&exists); err != nil {
		return false
	}
	return true
}

func loadRunsForSessions(db *sql.DB) ([]RunRecord, error) {
	rows, err := db.Query("SELECT id, started_at, finished_at, status, tool_version, git_from, git_to, root_path, args_json, error_json, sources_dir FROM meta_runs ORDER BY started_at DESC")
	if err != nil {
		return nil, err
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
			return nil, err
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
	return items, nil
}

func sessionKey(run RunRecord) string {
	if strings.TrimSpace(run.GitFrom) == "" && strings.TrimSpace(run.GitTo) == "" {
		return fmt.Sprintf("run-%d", run.ID)
	}
	return fmt.Sprintf("%s|%s|%s", run.RootPath, run.GitFrom, run.GitTo)
}

func buildSessionID(prefix string, rootPath string, gitFrom string, gitTo string, runID int64) string {
	if prefix == "" {
		prefix = "db"
	}
	if strings.TrimSpace(gitFrom) == "" && strings.TrimSpace(gitTo) == "" {
		return fmt.Sprintf("%s:run-%d", prefix, runID)
	}
	key := fmt.Sprintf("%s|%s|%s", rootPath, gitFrom, gitTo)
	hash := sha1.Sum([]byte(key))
	short := hex.EncodeToString(hash[:4])
	safeFrom := sanitizeSessionLabel(gitFrom)
	safeTo := sanitizeSessionLabel(gitTo)
	return fmt.Sprintf("%s:%s..%s:%s", prefix, safeFrom, safeTo, short)
}

func sanitizeSessionLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			builder.WriteRune(r)
		case r == '-', r == '_', r == '.', r == '~':
			builder.WriteRune(r)
		default:
			builder.WriteRune('_')
		}
	}
	return builder.String()
}

func (s *Server) loadWorkspaceConfig() (*WorkspaceConfig, error) {
	path, err := s.workspaceConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadWorkspaceConfig(path)
}

func (s *Server) saveWorkspaceConfig(cfg *WorkspaceConfig) error {
	path, err := s.workspaceConfigPath()
	if err != nil {
		return err
	}
	return cfg.Save(path)
}
