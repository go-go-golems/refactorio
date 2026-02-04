package refactorindex

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// IngestCommitsConfig controls commit lineage ingestion.
type IngestCommitsConfig struct {
	DBPath   string
	RepoPath string
	FromRef  string
	ToRef    string
}

// IngestCommitsResult reports commit ingestion counts.
type IngestCommitsResult struct {
	RunID        int64
	CommitCount  int
	FileCount    int
	BlobCount    int
	CommitHashes []string
}

func IngestCommits(ctx context.Context, cfg IngestCommitsConfig) (_ *IngestCommitsResult, err error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RepoPath) == "" {
		return nil, errors.New("repo path is required")
	}
	if strings.TrimSpace(cfg.FromRef) == "" || strings.TrimSpace(cfg.ToRef) == "" {
		return nil, errors.New("from/to refs are required")
	}

	db, err := OpenDB(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	store := NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		return nil, err
	}

	argsJSON, err := EncodeArgsJSON(map[string]string{
		"from": cfg.FromRef,
		"to":   cfg.ToRef,
		"repo": cfg.RepoPath,
	})
	if err != nil {
		return nil, err
	}

	runID, err := store.CreateRun(ctx, RunConfig{
		ToolVersion: ToolVersion,
		GitFrom:     cfg.FromRef,
		GitTo:       cfg.ToRef,
		RootPath:    cfg.RepoPath,
		ArgsJSON:    argsJSON,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = store.MarkRunFailed(ctx, runID, err)
		}
	}()

	tx, err := store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	fromHash, err := resolveCommitHash(ctx, cfg.RepoPath, cfg.FromRef)
	if err != nil {
		return nil, err
	}
	toHash, err := resolveCommitHash(ctx, cfg.RepoPath, cfg.ToRef)
	if err != nil {
		return nil, err
	}

	commitList, err := runGit(ctx, cfg.RepoPath, "rev-list", "--reverse", fmt.Sprintf("%s..%s", fromHash, toHash))
	if err != nil {
		return nil, err
	}
	commits := splitLines(commitList)
	rootCommits, err := loadRootCommits(ctx, cfg.RepoPath)
	if err != nil {
		return nil, err
	}
	if isRootCommitHash(rootCommits, fromHash) {
		commits = append([]string{fromHash}, commits...)
	}
	if len(commits) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, errors.Wrap(err, "commit empty commit ingestion")
		}
		if err := store.FinishRun(ctx, runID); err != nil {
			return nil, err
		}
		return &IngestCommitsResult{RunID: runID}, nil
	}

	fileCount := 0
	blobCount := 0
	for _, hash := range commits {
		info, err := loadCommitInfo(ctx, cfg.RepoPath, hash)
		if err != nil {
			return nil, err
		}
		commitID, err := store.InsertCommit(ctx, tx, runID, info)
		if err != nil {
			return nil, err
		}

		nameStatus, err := runGit(ctx, cfg.RepoPath, "diff-tree", "--no-commit-id", "-r", "--name-status", "-z", hash)
		if err != nil {
			return nil, err
		}
		entries, err := ParseNameStatus(nameStatus)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			primaryPath := entry.PrimaryPath()
			fileID, err := store.GetOrCreateFile(ctx, tx, primaryPath)
			if err != nil {
				return nil, err
			}
			if entry.OldPath != "" && entry.OldPath != primaryPath {
				if _, err := store.GetOrCreateFile(ctx, tx, entry.OldPath); err != nil {
					return nil, err
				}
			}
			if entry.NewPath != "" && entry.NewPath != primaryPath {
				if _, err := store.GetOrCreateFile(ctx, tx, entry.NewPath); err != nil {
					return nil, err
				}
			}

			blobNew := ""
			if entry.NewPath != "" {
				blobNew, _ = gitBlobSHA(ctx, cfg.RepoPath, hash, entry.NewPath)
			}
			blobOld := ""
			parent := hash + "^"
			if entry.OldPath != "" {
				blobOld, _ = gitBlobSHA(ctx, cfg.RepoPath, parent, entry.OldPath)
			}

			if err := store.InsertCommitFile(ctx, tx, commitID, fileID, entry.Status, entry.OldPath, entry.NewPath, blobOld, blobNew); err != nil {
				return nil, err
			}
			fileCount++

			if blobNew != "" {
				sizeBytes, lineCount := blobStats(ctx, cfg.RepoPath, blobNew)
				if err := store.InsertFileBlob(ctx, tx, commitID, fileID, blobNew, sizeBytes, lineCount); err != nil {
					return nil, err
				}
				blobCount++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit commit ingestion")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestCommitsResult{
		RunID:        runID,
		CommitCount:  len(commits),
		FileCount:    fileCount,
		BlobCount:    blobCount,
		CommitHashes: commits,
	}, nil
}

func loadCommitInfo(ctx context.Context, repoPath string, hash string) (CommitInfo, error) {
	format := "%H%x1f%an%x1f%ae%x1f%ad%x1f%cd%x1f%s%x1f%b"
	out, err := runGit(ctx, repoPath, "show", "-s", "--date=iso-strict", "--format="+format, hash)
	if err != nil {
		return CommitInfo{}, err
	}
	parts := strings.Split(string(bytes.TrimSpace(out)), "\x1f")
	if len(parts) < 7 {
		return CommitInfo{}, errors.New("unexpected commit format")
	}
	return CommitInfo{
		Hash:          parts[0],
		AuthorName:    parts[1],
		AuthorEmail:   parts[2],
		AuthorDate:    parts[3],
		CommitterDate: parts[4],
		Subject:       parts[5],
		Body:          strings.TrimSpace(parts[6]),
	}, nil
}

func gitBlobSHA(ctx context.Context, repoPath string, commit string, path string) (string, error) {
	if commit == "" || path == "" {
		return "", nil
	}
	out, err := runGit(ctx, repoPath, "rev-parse", fmt.Sprintf("%s:%s", commit, path))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func blobStats(ctx context.Context, repoPath string, blobSHA string) (*int64, *int) {
	if blobSHA == "" {
		return nil, nil
	}
	out, err := runGit(ctx, repoPath, "cat-file", "-s", blobSHA)
	if err != nil {
		return nil, nil
	}
	sizeValue := strings.TrimSpace(string(out))
	if sizeValue == "" {
		return nil, nil
	}
	sizeParsed, err := strconv.ParseInt(sizeValue, 10, 64)
	if err != nil {
		return nil, nil
	}

	content, err := runGit(ctx, repoPath, "cat-file", "-p", blobSHA)
	if err != nil {
		return &sizeParsed, nil
	}
	lineCount := bytes.Count(content, []byte{'\n'})
	return &sizeParsed, &lineCount
}

func splitLines(data []byte) []string {
	raw := strings.Split(strings.TrimSpace(string(data)), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func resolveCommitHash(ctx context.Context, repoPath string, ref string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", errors.New("ref is required")
	}
	out, err := runGit(ctx, repoPath, "rev-parse", ref)
	if err != nil {
		return "", err
	}
	hash := strings.TrimSpace(string(out))
	if hash == "" {
		return "", errors.New("empty commit hash")
	}
	return hash, nil
}

func loadRootCommits(ctx context.Context, repoPath string) (map[string]struct{}, error) {
	out, err := runGit(ctx, repoPath, "rev-list", "--max-parents=0", "--all")
	if err != nil {
		return nil, err
	}
	roots := splitLines(out)
	rootSet := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		rootSet[root] = struct{}{}
	}
	return rootSet, nil
}

func isRootCommitHash(rootSet map[string]struct{}, hash string) bool {
	if hash == "" {
		return false
	}
	_, ok := rootSet[hash]
	return ok
}
