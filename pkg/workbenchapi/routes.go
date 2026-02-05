package workbenchapi

import (
	"net/http"
	"strings"
)

func (s *Server) registerRoutes() {
	s.registerBaseRoutes()
	s.registerDBRoutes()
	s.registerWorkspaceRoutes()
	s.registerRunRoutes()
	s.registerSessionRoutes()
	s.registerSearchRoutes()
	s.registerSymbolRoutes()
	s.registerCodeUnitRoutes()
	s.registerDiffRoutes()
	s.registerCommitRoutes()
	s.registerDocRoutes()
}

func (s *Server) registerBaseRoutes() {
	s.apiMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
}

func (s *Server) registerWorkspaceRoutes() {
	s.apiMux.HandleFunc("/workspaces", s.handleWorkspaces)
	s.apiMux.HandleFunc("/workspaces/", s.handleWorkspace)
}

func (s *Server) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listWorkspaces(w, r)
	case http.MethodPost:
		s.createWorkspace(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
}

func (s *Server) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/workspaces/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getWorkspace(w, r, id)
	case http.MethodPatch:
		s.patchWorkspace(w, r, id)
	case http.MethodDelete:
		s.deleteWorkspace(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
}
