package workbenchapi

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type FileTreeItem struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Ext      string `json:"ext,omitempty"`
	Exists   bool   `json:"exists,omitempty"`
	IsBinary bool   `json:"is_binary,omitempty"`
}

type FileContentResponse struct {
	Path      string   `json:"path"`
	Ref       string   `json:"ref,omitempty"`
	Content   string   `json:"content"`
	Lines     []string `json:"lines,omitempty"`
	LineCount int      `json:"line_count"`
}

type FileHistoryRecord struct {
	Hash          string `json:"hash"`
	Subject       string `json:"subject"`
	CommitterDate string `json:"committer_date"`
	Status        string `json:"status"`
	OldPath       string `json:"old_path,omitempty"`
	NewPath       string `json:"new_path,omitempty"`
}

func (s *Server) registerFileRoutes() {
	s.apiMux.HandleFunc("/files", s.handleFiles)
	s.apiMux.HandleFunc("/file", s.handleFile)
	s.apiMux.HandleFunc("/files/history", s.handleFileHistory)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
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

	prefix := strings.TrimSpace(r.URL.Query().Get("prefix"))
	prefix = strings.Trim(prefix, "/")
	matchPrefix := prefix
	if matchPrefix != "" && !strings.HasSuffix(matchPrefix, "/") {
		matchPrefix += "/"
	}

	ext := strings.TrimSpace(r.URL.Query().Get("ext"))
	exists := parseBoolPtr(r.URL.Query().Get("exists"))
	isBinary := parseBoolPtr(r.URL.Query().Get("is_binary"))
	limit, offset := parseLimitOffset(r, 1000, 5000)

	query := "SELECT path, ext, file_exists, is_binary FROM files WHERE 1=1"
	args := []interface{}{}
	if matchPrefix != "" {
		query += " AND path LIKE ?"
		args = append(args, matchPrefix+"%")
	}
	if ext != "" {
		query += " AND ext = ?"
		args = append(args, ext)
	}
	if exists != nil {
		query += " AND file_exists = ?"
		args = append(args, boolToInt(*exists))
	}
	if isBinary != nil {
		query += " AND is_binary = ?"
		args = append(args, boolToInt(*isBinary))
	}
	query += " ORDER BY path LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query files", nil)
		return
	}
	defer rows.Close()

	dirSet := map[string]FileTreeItem{}
	fileItems := []FileTreeItem{}
	for rows.Next() {
		var path string
		var extVal sql.NullString
		var existsVal sql.NullInt64
		var binaryVal sql.NullInt64
		if err := rows.Scan(&path, &extVal, &existsVal, &binaryVal); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan files", nil)
			return
		}
		rel := path
		if matchPrefix != "" {
			rel = strings.TrimPrefix(path, matchPrefix)
		}
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) == 0 {
			continue
		}
		segment := parts[0]
		if segment == "" {
			continue
		}
		if len(parts) == 1 {
			item := FileTreeItem{
				Path:     joinPrefix(matchPrefix, segment),
				Kind:     "file",
				Exists:   existsVal.Valid && existsVal.Int64 == 1,
				IsBinary: binaryVal.Valid && binaryVal.Int64 == 1,
			}
			if extVal.Valid {
				item.Ext = extVal.String
			}
			fileItems = append(fileItems, item)
			continue
		}
		dirPath := joinPrefix(matchPrefix, segment)
		if _, ok := dirSet[dirPath]; !ok {
			dirSet[dirPath] = FileTreeItem{Path: dirPath, Kind: "dir"}
		}
	}

	items := []FileTreeItem{}
	for _, item := range dirSet {
		items = append(items, item)
	}
	items = append(items, fileItems...)
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	ref, ok := s.requireWorkspaceRef(w, r)
	if !ok {
		return
	}
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "path is required", map[string]string{"field": "path"})
		return
	}
	if strings.TrimSpace(ref.RepoRoot) == "" {
		writeError(w, http.StatusConflict, "repo_required", "repo_root is required to read file contents", nil)
		return
	}

	refName := strings.TrimSpace(r.URL.Query().Get("ref"))
	withLines := parseBool(r.URL.Query().Get("with_lines"))

	content, err := readFileContent(ref.RepoRoot, refName, path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "file_error", "failed to read file contents", map[string]string{"error": err.Error()})
		return
	}
	lines := []string{}
	if withLines {
		lines = splitLines(content)
	}
	writeJSON(w, http.StatusOK, FileContentResponse{
		Path:      path,
		Ref:       refName,
		Content:   content,
		Lines:     lines,
		LineCount: len(splitLines(content)),
	})
}

func (s *Server) handleFileHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	ref, ok := s.requireWorkspaceRef(w, r)
	if !ok {
		return
	}
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeError(w, http.StatusBadRequest, "invalid_argument", "path is required", map[string]string{"field": "path"})
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
	query := `
SELECT c.hash, c.subject, c.committer_date, cf.status, cf.old_path, cf.new_path
FROM commit_files cf
JOIN commits c ON c.id = cf.commit_id
JOIN files f ON f.id = cf.file_id
WHERE f.path = ?`
	args := []interface{}{path}
	if runID > 0 {
		query += " AND c.run_id = ?"
		args = append(args, runID)
	}
	query += " ORDER BY c.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query file history", nil)
		return
	}
	defer rows.Close()

	items := []FileHistoryRecord{}
	for rows.Next() {
		var record FileHistoryRecord
		var oldPath sql.NullString
		var newPath sql.NullString
		if err := rows.Scan(&record.Hash, &record.Subject, &record.CommitterDate, &record.Status, &oldPath, &newPath); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan file history", nil)
			return
		}
		if oldPath.Valid {
			record.OldPath = oldPath.String
		}
		if newPath.Valid {
			record.NewPath = newPath.String
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func readFileContent(repoRoot string, ref string, path string) (string, error) {
	if ref == "" {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(path))
		content, err := os.ReadFile(absPath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	cmd := exec.Command("git", "-C", repoRoot, "show", fmt.Sprintf("%s:%s", ref, path))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func splitLines(content string) []string {
	scanner := bufio.NewScanner(bytes.NewBufferString(content))
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func joinPrefix(prefix string, segment string) string {
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		return segment
	}
	return prefix + "/" + segment
}

func parseBoolPtr(value string) *bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed := parseBool(trimmed)
	return &parsed
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
