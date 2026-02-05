package workbenchapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Workspace struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	DBPath    string            `json:"db_path"`
	RepoRoot  string            `json:"repo_root,omitempty"`
	Sessions  []SessionOverride `json:"sessions,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type WorkspaceConfig struct {
	Workspaces []Workspace `json:"workspaces"`
}

type WorkspacePatch struct {
	Name     *string `json:"name"`
	DBPath   *string `json:"db_path"`
	RepoRoot *string `json:"repo_root"`
}

func DefaultWorkspaceConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.Wrap(err, "resolve user config dir")
	}
	return filepath.Join(configDir, "refactorio", "workspaces.json"), nil
}

func LoadWorkspaceConfig(path string) (*WorkspaceConfig, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &WorkspaceConfig{}, nil
		}
		return nil, errors.Wrap(err, "read workspace config")
	}
	if len(payload) == 0 {
		return &WorkspaceConfig{}, nil
	}

	var cfg WorkspaceConfig
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return nil, errors.Wrap(err, "decode workspace config")
	}
	return &cfg, nil
}

func (c *WorkspaceConfig) Save(path string) error {
	if c == nil {
		return errors.New("workspace config is nil")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "ensure workspace config dir")
	}
	payload, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, "encode workspace config")
	}
	tmpFile, err := os.CreateTemp(dir, "workspaces-*.json")
	if err != nil {
		return errors.Wrap(err, "create temp workspace config")
	}
	if _, err := tmpFile.Write(payload); err != nil {
		_ = tmpFile.Close()
		return errors.Wrap(err, "write temp workspace config")
	}
	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "close temp workspace config")
	}
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return errors.Wrap(err, "replace workspace config")
	}
	return nil
}

func (c *WorkspaceConfig) FindWorkspace(id string) (Workspace, int, bool) {
	for idx, ws := range c.Workspaces {
		if ws.ID == id {
			return ws, idx, true
		}
	}
	return Workspace{}, -1, false
}

func (s *Server) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	path, err := s.workspaceConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to resolve workspace config path", nil)
		return
	}
	cfg, err := LoadWorkspaceConfig(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": cfg.Workspaces, "limit": 0, "offset": 0})
}

func (s *Server) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var input Workspace
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	input.ID = strings.TrimSpace(input.ID)
	if input.ID == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "id is required", map[string]string{"field": "id"})
		return
	}
	input.DBPath = strings.TrimSpace(input.DBPath)
	if input.DBPath == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "db_path is required", map[string]string{"field": "db_path"})
		return
	}
	absDB, err := filepath.Abs(input.DBPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", "db_path must be a valid path", nil)
		return
	}
	input.DBPath = absDB
	if strings.TrimSpace(input.RepoRoot) != "" {
		absRepo, err := filepath.Abs(input.RepoRoot)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "repo_root must be a valid path", nil)
			return
		}
		input.RepoRoot = absRepo
	}
	if strings.TrimSpace(input.Name) == "" {
		input.Name = input.ID
	}
	now := time.Now().UTC()
	input.CreatedAt = now
	input.UpdatedAt = now

	path, err := s.workspaceConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to resolve workspace config path", nil)
		return
	}
	cfg, err := LoadWorkspaceConfig(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
		return
	}
	if _, _, ok := cfg.FindWorkspace(input.ID); ok {
		writeError(w, http.StatusConflict, "already_exists", "workspace already exists", map[string]string{"id": input.ID})
		return
	}
	cfg.Workspaces = append(cfg.Workspaces, input)
	if err := cfg.Save(path); err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to save workspace config", nil)
		return
	}
	writeJSON(w, http.StatusCreated, input)
}

func (s *Server) getWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	path, err := s.workspaceConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to resolve workspace config path", nil)
		return
	}
	cfg, err := LoadWorkspaceConfig(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
		return
	}
	ws, _, ok := cfg.FindWorkspace(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found", map[string]string{"id": id})
		return
	}
	writeJSON(w, http.StatusOK, ws)
}

func (s *Server) patchWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	var patch WorkspacePatch
	if err := decodeJSON(w, r, &patch); err != nil {
		return
	}

	path, err := s.workspaceConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to resolve workspace config path", nil)
		return
	}
	cfg, err := LoadWorkspaceConfig(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
		return
	}
	ws, idx, ok := cfg.FindWorkspace(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found", map[string]string{"id": id})
		return
	}

	if patch.Name != nil {
		ws.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.DBPath != nil {
		value := strings.TrimSpace(*patch.DBPath)
		if value == "" {
			writeError(w, http.StatusBadRequest, "invalid_argument", "db_path cannot be empty", map[string]string{"field": "db_path"})
			return
		}
		absDB, err := filepath.Abs(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "db_path must be a valid path", nil)
			return
		}
		ws.DBPath = absDB
	}
	if patch.RepoRoot != nil {
		value := strings.TrimSpace(*patch.RepoRoot)
		if value == "" {
			ws.RepoRoot = ""
		} else {
			absRepo, err := filepath.Abs(value)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", "repo_root must be a valid path", nil)
				return
			}
			ws.RepoRoot = absRepo
		}
	}

	ws.UpdatedAt = time.Now().UTC()
	cfg.Workspaces[idx] = ws
	if err := cfg.Save(path); err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to save workspace config", nil)
		return
	}
	writeJSON(w, http.StatusOK, ws)
}

func (s *Server) deleteWorkspace(w http.ResponseWriter, r *http.Request, id string) {
	path, err := s.workspaceConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to resolve workspace config path", nil)
		return
	}
	cfg, err := LoadWorkspaceConfig(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to load workspace config", nil)
		return
	}
	_, idx, ok := cfg.FindWorkspace(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found", map[string]string{"id": id})
		return
	}
	cfg.Workspaces = append(cfg.Workspaces[:idx], cfg.Workspaces[idx+1:]...)
	if err := cfg.Save(path); err != nil {
		writeError(w, http.StatusInternalServerError, "config_error", "failed to save workspace config", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": id})
}
