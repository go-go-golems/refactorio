package refactorindex

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type RangeIngestConfig struct {
	DBPath     string
	RepoPath   string
	FromRef    string
	ToRef      string
	SourcesDir string

	IncludeDiff         bool
	IncludeSymbols      bool
	IncludeCodeUnits    bool
	IncludeDocHits      bool
	IncludeTreeSitter   bool
	IncludeGopls        bool
	IgnorePackageErrors bool

	TermsFile             string
	TreeSitterLanguage    string
	TreeSitterQueries     string
	TreeSitterGlob        string
	GoplsTargets          []GoplsRefTarget
	GoplsSkipSymbolLookup bool
}

type CommitRunInfo struct {
	CommitHash      string
	WorktreePath    string
	DiffRunID       int64
	SymbolsRunID    int64
	CodeUnitsRunID  int64
	DocHitsRunID    int64
	TreeSitterRunID int64
	GoplsRunID      int64
}

type RangeIngestResult struct {
	CommitLineageRunID int64
	Commits            []CommitRunInfo
}

func IngestCommitRange(ctx context.Context, cfg RangeIngestConfig) (*RangeIngestResult, error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RepoPath) == "" {
		return nil, errors.New("repo path is required")
	}
	if strings.TrimSpace(cfg.FromRef) == "" || strings.TrimSpace(cfg.ToRef) == "" {
		return nil, errors.New("from/to refs are required")
	}

	lineageResult, err := IngestCommits(ctx, IngestCommitsConfig{
		DBPath:   cfg.DBPath,
		RepoPath: cfg.RepoPath,
		FromRef:  cfg.FromRef,
		ToRef:    cfg.ToRef,
	})
	if err != nil {
		return nil, err
	}

	commits := lineageResult.CommitHashes
	if len(commits) == 0 {
		return &RangeIngestResult{CommitLineageRunID: lineageResult.RunID}, nil
	}

	db, err := OpenDB(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()
	store := NewStore(db)

	commitIDs := make(map[string]int64, len(commits))
	for _, hash := range commits {
		commitID, err := store.GetCommitIDByHash(ctx, lineageResult.RunID, hash)
		if err != nil {
			return nil, err
		}
		commitIDs[hash] = commitID
	}

	worktreeRoot, err := os.MkdirTemp("", "refactor-index-worktrees-*")
	if err != nil {
		return nil, errors.Wrap(err, "create worktree root")
	}
	defer func() {
		_ = os.RemoveAll(worktreeRoot)
	}()

	results := make([]CommitRunInfo, 0, len(commits))
	for _, hash := range commits {
		worktreePath := filepath.Join(worktreeRoot, hash)
		if err := addWorktree(ctx, cfg.RepoPath, worktreePath, hash); err != nil {
			return nil, err
		}

		commitRun := CommitRunInfo{CommitHash: hash, WorktreePath: worktreePath}
		commitID := commitIDs[hash]
		if cfg.IncludeDiff {
			diffResult, err := IngestDiff(ctx, IngestDiffConfig{
				DBPath:     cfg.DBPath,
				RepoPath:   cfg.RepoPath,
				FromRef:    hash + "^",
				ToRef:      hash,
				SourcesDir: cfg.SourcesDir,
			})
			if err != nil {
				_ = removeWorktree(ctx, cfg.RepoPath, worktreePath)
				return nil, err
			}
			commitRun.DiffRunID = diffResult.RunID
		}

		if cfg.IncludeSymbols {
			symbolsResult, err := IngestSymbols(ctx, IngestSymbolsConfig{
				DBPath:              cfg.DBPath,
				RootDir:             worktreePath,
				SourcesDir:          cfg.SourcesDir,
				CommitID:            &commitID,
				IgnorePackageErrors: cfg.IgnorePackageErrors,
			})
			if err != nil {
				_ = removeWorktree(ctx, cfg.RepoPath, worktreePath)
				return nil, err
			}
			commitRun.SymbolsRunID = symbolsResult.RunID
		}

		if cfg.IncludeCodeUnits {
			codeUnitsResult, err := IngestCodeUnits(ctx, IngestCodeUnitsConfig{
				DBPath:              cfg.DBPath,
				RootDir:             worktreePath,
				SourcesDir:          cfg.SourcesDir,
				CommitID:            &commitID,
				IgnorePackageErrors: cfg.IgnorePackageErrors,
			})
			if err != nil {
				_ = removeWorktree(ctx, cfg.RepoPath, worktreePath)
				return nil, err
			}
			commitRun.CodeUnitsRunID = codeUnitsResult.RunID
		}

		if cfg.IncludeDocHits && strings.TrimSpace(cfg.TermsFile) != "" {
			docResult, err := IngestDocHits(ctx, IngestDocHitsConfig{
				DBPath:     cfg.DBPath,
				RootDir:    worktreePath,
				TermsFile:  cfg.TermsFile,
				SourcesDir: cfg.SourcesDir,
			})
			if err != nil {
				_ = removeWorktree(ctx, cfg.RepoPath, worktreePath)
				return nil, err
			}
			commitRun.DocHitsRunID = docResult.RunID
		}

		if cfg.IncludeTreeSitter && strings.TrimSpace(cfg.TreeSitterLanguage) != "" && strings.TrimSpace(cfg.TreeSitterQueries) != "" {
			tsResult, err := IngestTreeSitter(ctx, IngestTreeSitterConfig{
				DBPath:     cfg.DBPath,
				RootDir:    worktreePath,
				Language:   cfg.TreeSitterLanguage,
				QueriesYML: cfg.TreeSitterQueries,
				FileGlob:   cfg.TreeSitterGlob,
				SourcesDir: cfg.SourcesDir,
			})
			if err != nil {
				_ = removeWorktree(ctx, cfg.RepoPath, worktreePath)
				return nil, err
			}
			commitRun.TreeSitterRunID = tsResult.RunID
		}

		if cfg.IncludeGopls && len(cfg.GoplsTargets) > 0 {
			goplsResult, err := IngestGoplsReferences(ctx, IngestGoplsRefsConfig{
				DBPath:           cfg.DBPath,
				RepoPath:         worktreePath,
				SourcesDir:       cfg.SourcesDir,
				Targets:          cfg.GoplsTargets,
				SkipSymbolLookup: cfg.GoplsSkipSymbolLookup,
			})
			if err != nil {
				_ = removeWorktree(ctx, cfg.RepoPath, worktreePath)
				return nil, err
			}
			commitRun.GoplsRunID = goplsResult.RunID
		}

		if err := removeWorktree(ctx, cfg.RepoPath, worktreePath); err != nil {
			return nil, err
		}

		results = append(results, commitRun)
	}

	return &RangeIngestResult{
		CommitLineageRunID: lineageResult.RunID,
		Commits:            results,
	}, nil
}

func addWorktree(ctx context.Context, repoPath string, path string, commit string) error {
	_, err := runGit(ctx, repoPath, "worktree", "add", "--force", path, commit)
	if err != nil {
		return errors.Wrap(err, "add worktree")
	}
	return nil
}

func removeWorktree(ctx context.Context, repoPath string, path string) error {
	_, err := runGit(ctx, repoPath, "worktree", "remove", "--force", path)
	if err != nil {
		return errors.Wrap(err, "remove worktree")
	}
	_, _ = runGit(ctx, repoPath, "worktree", "prune")
	return nil
}
