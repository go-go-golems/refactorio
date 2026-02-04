package refactorindex

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const ToolVersion = "dev"

type IngestDiffConfig struct {
	DBPath      string
	RepoPath    string
	FromRef     string
	ToRef       string
	SourcesDir  string
	UseRootDiff bool
}

type IngestDiffResult struct {
	RunID  int64
	Files  int
	Hunks  int
	Lines  int
	RunDir string
}

func IngestDiff(ctx context.Context, cfg IngestDiffConfig) (_ *IngestDiffResult, err error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RepoPath) == "" {
		return nil, errors.New("repo path is required")
	}
	repoPath, err := filepath.Abs(cfg.RepoPath)
	if err != nil {
		return nil, errors.Wrap(err, "resolve repo path")
	}
	sourcesDir := cfg.SourcesDir
	if strings.TrimSpace(sourcesDir) == "" {
		sourcesDir = "sources"
	}
	sourcesDir, err = filepath.Abs(sourcesDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolve sources dir")
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
		"from":        cfg.FromRef,
		"to":          cfg.ToRef,
		"repo":        repoPath,
		"sources_dir": sourcesDir,
		"root_diff":   fmt.Sprint(cfg.UseRootDiff),
	})
	if err != nil {
		return nil, err
	}

	runID, err := store.CreateRun(ctx, RunConfig{
		ToolVersion: ToolVersion,
		GitFrom:     cfg.FromRef,
		GitTo:       cfg.ToRef,
		RootPath:    repoPath,
		SourcesDir:  sourcesDir,
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

	runDir := filepath.Join(sourcesDir, fmt.Sprintf("%d", runID))

	var nameStatusOutput []byte
	if cfg.UseRootDiff {
		nameStatusOutput, err = runGit(ctx, repoPath, "diff", "--root", "--name-status", "-z", cfg.ToRef)
	} else {
		nameStatusOutput, err = runGit(ctx, repoPath, "diff", "--name-status", "-z", cfg.FromRef, cfg.ToRef)
	}
	if err != nil {
		return nil, err
	}
	if _, err := store.WriteRawOutput(ctx, tx, runDir, runID, "git-name-status", "git-name-status.txt", nameStatusOutput); err != nil {
		return nil, err
	}

	entries, err := ParseNameStatus(nameStatusOutput)
	if err != nil {
		return nil, err
	}

	var patchOutput []byte
	if cfg.UseRootDiff {
		patchOutput, err = runGit(ctx, repoPath, "diff", "--root", "-U0", "--no-color", cfg.ToRef)
	} else {
		patchOutput, err = runGit(ctx, repoPath, "diff", "-U0", "--no-color", cfg.FromRef, cfg.ToRef)
	}
	if err != nil {
		return nil, err
	}
	if _, err := store.WriteRawOutput(ctx, tx, runDir, runID, "git-diff-u0", "git-diff-u0.patch", patchOutput); err != nil {
		return nil, err
	}

	pathToDiffFileID := make(map[string]int64)
	fileCount := 0
	for _, entry := range entries {
		primaryPath := entry.PrimaryPath()
		fileID, err := store.GetOrCreateFile(ctx, tx, primaryPath)
		if err != nil {
			return nil, err
		}
		if entry.OldPath != "" {
			_, err := store.GetOrCreateFile(ctx, tx, entry.OldPath)
			if err != nil {
				return nil, err
			}
		}
		if entry.NewPath != "" {
			_, err := store.GetOrCreateFile(ctx, tx, entry.NewPath)
			if err != nil {
				return nil, err
			}
		}
		diffFileID, err := store.InsertDiffFile(ctx, tx, runID, fileID, entry.Status, entry.OldPath, entry.NewPath)
		if err != nil {
			return nil, err
		}
		fileCount++
		pathToDiffFileID[primaryPath] = diffFileID
		if entry.OldPath != "" {
			pathToDiffFileID[entry.OldPath] = diffFileID
		}
		if entry.NewPath != "" {
			pathToDiffFileID[entry.NewPath] = diffFileID
		}
	}

	patches, err := ParseUnifiedDiff(patchOutput)
	if err != nil {
		return nil, err
	}

	hunkCount := 0
	lineCount := 0
	for _, patch := range patches {
		diffFileID := resolveDiffFileID(pathToDiffFileID, patch.OldPath, patch.NewPath)
		if diffFileID == 0 {
			continue
		}
		for _, hunk := range patch.Hunks {
			hunkID, err := store.InsertDiffHunk(ctx, tx, diffFileID, hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines)
			if err != nil {
				return nil, err
			}
			hunkCount++
			for _, line := range hunk.Lines {
				if err := store.InsertDiffLine(ctx, tx, hunkID, line.Kind, line.OldLine, line.NewLine, line.Text); err != nil {
					return nil, err
				}
				lineCount++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit diff ingestion")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestDiffResult{
		RunID:  runID,
		Files:  fileCount,
		Hunks:  hunkCount,
		Lines:  lineCount,
		RunDir: runDir,
	}, nil
}

func resolveDiffFileID(index map[string]int64, oldPath string, newPath string) int64 {
	if newPath != "" {
		if id, ok := index[newPath]; ok {
			return id
		}
	}
	if oldPath != "" {
		if id, ok := index[oldPath]; ok {
			return id
		}
	}
	return 0
}

func runGit(ctx context.Context, repoPath string, args ...string) ([]byte, error) {
	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = "git command failed"
		}
		return nil, errors.Wrap(err, msg)
	}
	return stdout, nil
}
