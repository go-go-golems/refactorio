package workbenchapi

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
	"github.com/pkg/errors"
)

type WorkspaceRef struct {
	ID       string
	DBPath   string
	RepoRoot string
}

func (s *Server) requireWorkspaceRef(w http.ResponseWriter, r *http.Request) (WorkspaceRef, bool) {
	query := r.URL.Query()
	workspaceID := strings.TrimSpace(query.Get("workspace_id"))
	if workspaceID != "" {
		path, err := s.workspaceConfigPath()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "config_error", "failed to resolve workspace config path", nil)
			return WorkspaceRef{}, false
		}
		cfg, err := LoadWorkspaceConfig(path)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
			return WorkspaceRef{}, false
		}
		ws, _, ok := cfg.FindWorkspace(workspaceID)
		if !ok {
			writeError(w, http.StatusNotFound, "not_found", "workspace not found", map[string]string{"id": workspaceID})
			return WorkspaceRef{}, false
		}
		return WorkspaceRef{ID: ws.ID, DBPath: ws.DBPath, RepoRoot: ws.RepoRoot}, true
	}

	path := strings.TrimSpace(query.Get("db_path"))
	if path == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "workspace_id or db_path is required", nil)
		return WorkspaceRef{}, false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", "db_path must be a valid path", nil)
		return WorkspaceRef{}, false
	}
	return WorkspaceRef{DBPath: absPath}, true
}

func (s *Server) openWorkspaceDB(ctx context.Context, ref WorkspaceRef) (*sql.DB, func() error, error) {
	db, err := refactorindex.OpenDB(ctx, ref.DBPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "open workspace db")
	}
	closer := func() error { return db.Close() }
	return db, closer, nil
}
