package workbenchapi

import (
	"database/sql"
	"net/http"
	"strings"
)

type DBInfoResponse struct {
	WorkspaceID   string          `json:"workspace_id,omitempty"`
	DBPath        string          `json:"db_path"`
	RepoRoot      string          `json:"repo_root,omitempty"`
	SchemaVersion int             `json:"schema_version"`
	Tables        map[string]bool `json:"tables"`
	FTSTables     map[string]bool `json:"fts_tables"`
	Features      map[string]bool `json:"features"`
	Views         map[string]bool `json:"views,omitempty"`
}

func (s *Server) registerDBRoutes() {
	s.apiMux.HandleFunc("/db/info", s.handleDBInfo)
}

func (s *Server) handleDBInfo(w http.ResponseWriter, r *http.Request) {
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

	version := 0
	if hasTable(db, "schema_versions") {
		if v, err := maxSchemaVersion(db); err == nil {
			version = v
		}
	}

	tables, views := listTablesAndViews(db)
	ftsTables := map[string]bool{}
	for name := range tables {
		if strings.HasSuffix(name, "_fts") {
			ftsTables[name] = true
		}
	}

	features := map[string]bool{
		"fts":        len(ftsTables) > 0,
		"gopls_refs": tables["symbol_refs"],
		"doc_hits":   tables["doc_hits"],
	}

	writeJSON(w, http.StatusOK, DBInfoResponse{
		WorkspaceID:   ref.ID,
		DBPath:        ref.DBPath,
		RepoRoot:      ref.RepoRoot,
		SchemaVersion: version,
		Tables:        tables,
		FTSTables:     ftsTables,
		Features:      features,
		Views:         views,
	})
}

func maxSchemaVersion(db *sql.DB) (int, error) {
	var version sql.NullInt64
	if err := db.QueryRow("SELECT max(version) FROM schema_versions").Scan(&version); err != nil {
		return 0, err
	}
	if version.Valid {
		return int(version.Int64), nil
	}
	return 0, nil
}

func hasTable(db *sql.DB, name string) bool {
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func listTablesAndViews(db *sql.DB) (map[string]bool, map[string]bool) {
	tables := map[string]bool{}
	views := map[string]bool{}
	rows, err := db.Query("SELECT name, type FROM sqlite_master WHERE type IN ('table','view')")
	if err != nil {
		return tables, views
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var kind string
		if err := rows.Scan(&name, &kind); err != nil {
			continue
		}
		if kind == "table" {
			tables[name] = true
		} else if kind == "view" {
			views[name] = true
		}
	}
	return tables, views
}
