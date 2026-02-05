package workbenchapi

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type CommitRecord struct {
	RunID         int64  `json:"run_id"`
	Hash          string `json:"hash"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorEmail   string `json:"author_email,omitempty"`
	AuthorDate    string `json:"author_date,omitempty"`
	CommitterDate string `json:"committer_date,omitempty"`
	Subject       string `json:"subject,omitempty"`
	Body          string `json:"body,omitempty"`
}

type CommitFileRecord struct {
	Path    string `json:"path"`
	Status  string `json:"status"`
	OldPath string `json:"old_path,omitempty"`
	NewPath string `json:"new_path,omitempty"`
	BlobOld string `json:"blob_old,omitempty"`
	BlobNew string `json:"blob_new,omitempty"`
}

func (s *Server) registerCommitRoutes() {
	s.apiMux.HandleFunc("/commits", s.handleCommits)
	s.apiMux.HandleFunc("/commits/", s.handleCommit)
}

func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
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
	runID := parseRunID(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	author := strings.TrimSpace(r.URL.Query().Get("author"))
	after := strings.TrimSpace(r.URL.Query().Get("after"))
	before := strings.TrimSpace(r.URL.Query().Get("before"))

	query := ""
	args := []interface{}{}
	if q != "" {
		if !hasTable(db, "commits_fts") {
			writeError(w, http.StatusBadRequest, "search_error", "fts table not available: commits_fts", nil)
			return
		}
		query = `
SELECT c.run_id, c.hash, c.author_name, c.author_email, c.author_date, c.committer_date, c.subject, c.body
FROM commits_fts fts
JOIN commits c ON c.id = fts.rowid
WHERE commits_fts MATCH ?`
		args = append(args, q)
	} else {
		query = `
SELECT c.run_id, c.hash, c.author_name, c.author_email, c.author_date, c.committer_date, c.subject, c.body
FROM commits c
WHERE 1=1`
	}
	if runID > 0 {
		query += " AND c.run_id = ?"
		args = append(args, runID)
	}
	if author != "" {
		query += " AND (c.author_name LIKE ? OR c.author_email LIKE ?)"
		like := "%" + author + "%"
		args = append(args, like, like)
	}
	if after != "" {
		query += " AND c.committer_date >= ?"
		args = append(args, after)
	}
	if before != "" {
		query += " AND c.committer_date <= ?"
		args = append(args, before)
	}
	if path != "" {
		query += " AND EXISTS (SELECT 1 FROM commit_files cf JOIN files f ON f.id = cf.file_id WHERE cf.commit_id = c.id AND f.path = ?)"
		args = append(args, path)
	}
	query += " ORDER BY c.committer_date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query commits", nil)
		return
	}
	defer rows.Close()

	items := []CommitRecord{}
	for rows.Next() {
		record, err := scanCommitRecord(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan commits", nil)
			return
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/commits/")
	if path == "" {
		writeError(w, http.StatusNotFound, "not_found", "commit not found", nil)
		return
	}
	parts := strings.Split(path, "/")
	hash := parts[0]
	if hash == "" {
		writeError(w, http.StatusNotFound, "not_found", "commit not found", nil)
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		s.getCommit(w, r, hash)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "files":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
				return
			}
			s.getCommitFiles(w, r, hash)
			return
		case "diff":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
				return
			}
			s.getCommitDiff(w, r, hash)
			return
		}
	}

	writeError(w, http.StatusNotFound, "not_found", "commit endpoint not found", nil)
}

func (s *Server) getCommit(w http.ResponseWriter, r *http.Request, hash string) {
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

	row := db.QueryRowContext(r.Context(), `
SELECT run_id, hash, author_name, author_email, author_date, committer_date, subject, body
FROM commits
WHERE hash = ?
ORDER BY id DESC
LIMIT 1`, hash)

	record, err := scanCommitRecord(row)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "not_found", "commit not found", map[string]string{"hash": hash})
			return
		}
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query commit", nil)
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func (s *Server) getCommitFiles(w http.ResponseWriter, r *http.Request, hash string) {
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

	rows, err := db.QueryContext(r.Context(), `
SELECT f.path, cf.status, cf.old_path, cf.new_path, cf.blob_old, cf.blob_new
FROM commit_files cf
JOIN commits c ON c.id = cf.commit_id
JOIN files f ON f.id = cf.file_id
WHERE c.hash = ?
ORDER BY f.path`, hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query_error", "failed to query commit files", nil)
		return
	}
	defer rows.Close()

	items := []CommitFileRecord{}
	for rows.Next() {
		var record CommitFileRecord
		var oldPath sql.NullString
		var newPath sql.NullString
		var blobOld sql.NullString
		var blobNew sql.NullString
		if err := rows.Scan(&record.Path, &record.Status, &oldPath, &newPath, &blobOld, &blobNew); err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to scan commit files", nil)
			return
		}
		if oldPath.Valid {
			record.OldPath = oldPath.String
		}
		if newPath.Valid {
			record.NewPath = newPath.String
		}
		if blobOld.Valid {
			record.BlobOld = blobOld.String
		}
		if blobNew.Valid {
			record.BlobNew = blobNew.String
		}
		items = append(items, record)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) getCommitDiff(w http.ResponseWriter, r *http.Request, hash string) {
	pathFilter := strings.TrimSpace(r.URL.Query().Get("path"))
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

	runID, err := findDiffRunForCommit(db, hash)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "diff run not found for commit", map[string]string{"hash": hash})
		return
	}
	if pathFilter == "" {
		store := refactorindex.NewStore(db)
		records, err := store.ListDiffFiles(r.Context(), refactorindex.DiffFileFilter{RunID: runID})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "query_error", "failed to query diff files", nil)
			return
		}
		items := make([]DiffFileRecord, 0, len(records))
		for _, record := range records {
			items = append(items, DiffFileRecord{RunID: record.RunID, Status: record.Status, Path: record.Path, OldPath: record.OldPath, NewPath: record.NewPath})
		}
		writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "files": items})
		return
	}

	hunks, err := loadDiffHunks(db, runID, pathFilter)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "diff file not found", map[string]string{"path": pathFilter})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "path": pathFilter, "hunks": hunks})
}

func findDiffRunForCommit(db *sql.DB, hash string) (int64, error) {
	row := db.QueryRow(`
SELECT id
FROM meta_runs
WHERE git_to = ? AND (git_from = ? OR git_from = ?)
ORDER BY started_at DESC
LIMIT 1`, hash, hash, hash+"^")
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func scanCommitRecord(scanner interface {
	Scan(dest ...interface{}) error
}) (CommitRecord, error) {
	var record CommitRecord
	var authorName sql.NullString
	var authorEmail sql.NullString
	var authorDate sql.NullString
	var committerDate sql.NullString
	var subject sql.NullString
	var body sql.NullString
	if err := scanner.Scan(
		&record.RunID,
		&record.Hash,
		&authorName,
		&authorEmail,
		&authorDate,
		&committerDate,
		&subject,
		&body,
	); err != nil {
		return CommitRecord{}, err
	}
	if authorName.Valid {
		record.AuthorName = authorName.String
	}
	if authorEmail.Valid {
		record.AuthorEmail = authorEmail.String
	}
	if authorDate.Valid {
		record.AuthorDate = authorDate.String
	}
	if committerDate.Valid {
		record.CommitterDate = committerDate.String
	}
	if subject.Valid {
		record.Subject = subject.String
	}
	if body.Valid {
		record.Body = body.String
	}
	return record, nil
}
