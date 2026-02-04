package refactorindex

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/oak/pkg/api"
	"github.com/pkg/errors"
)

type IngestTreeSitterConfig struct {
	DBPath     string
	RootDir    string
	Language   string
	QueriesYML string
	FileGlob   string
	CommitID   *int64
	SourcesDir string
}

type IngestTreeSitterResult struct {
	RunID     int64
	Files     int
	Captures  int
	Queries   int
	Skipped   int
	CommitID  *int64
	Language  string
	QueryFile string
	FileGlob  string
}

func IngestTreeSitter(ctx context.Context, cfg IngestTreeSitterConfig) (_ *IngestTreeSitterResult, err error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RootDir) == "" {
		return nil, errors.New("root dir is required")
	}
	if strings.TrimSpace(cfg.Language) == "" {
		return nil, errors.New("language is required")
	}
	if strings.TrimSpace(cfg.QueriesYML) == "" {
		return nil, errors.New("queries yaml is required")
	}

	rootDir, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolve root dir")
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
		"root":      rootDir,
		"lang":      cfg.Language,
		"queries":   cfg.QueriesYML,
		"file_glob": cfg.FileGlob,
	})
	if err != nil {
		return nil, err
	}

	runID, err := store.CreateRun(ctx, RunConfig{
		ToolVersion: ToolVersion,
		RootPath:    rootDir,
		SourcesDir:  cfg.SourcesDir,
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

	qb := api.NewQueryBuilder(
		api.WithLanguage(cfg.Language),
		api.FromYAML(cfg.QueriesYML),
	)

	options := []api.RunOption{}
	if strings.TrimSpace(cfg.FileGlob) != "" {
		options = append(options, api.WithGlob(cfg.FileGlob))
	} else {
		options = append(options, api.WithDirectory(rootDir), api.WithRecursive(true))
	}

	results, err := qb.Run(ctx, options...)
	if err != nil {
		return nil, err
	}

	tx, err := store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	fileCount := 0
	captureCount := 0
	queryCount := 0
	skipCount := 0

	fileIDs := make(map[string]int64)
	for filePath, byQuery := range results {
		if filePath == "" {
			skipCount++
			continue
		}
		relPath, err := filepath.Rel(rootDir, filePath)
		if err != nil {
			return nil, errors.Wrap(err, "relativize tree-sitter file")
		}
		relPath = filepath.ToSlash(relPath)
		fileID, ok := fileIDs[relPath]
		if !ok {
			id, err := store.GetOrCreateFile(ctx, tx, relPath)
			if err != nil {
				return nil, err
			}
			fileID = id
			fileIDs[relPath] = id
			fileCount++
		}

		for queryName, result := range byQuery {
			if result == nil {
				continue
			}
			queryCount++
			for _, match := range result.Matches {
				for _, capture := range match {
					startLine := int(capture.StartPoint.Row) + 1
					startCol := int(capture.StartPoint.Column) + 1
					endLine := int(capture.EndPoint.Row) + 1
					endCol := int(capture.EndPoint.Column) + 1
					if err := store.InsertTreeSitterCapture(ctx, tx, runID, cfg.CommitID, fileID, queryName, capture.Name, capture.Type, startLine, startCol, endLine, endCol, capture.Text); err != nil {
						return nil, err
					}
					captureCount++
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit tree-sitter ingestion")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestTreeSitterResult{
		RunID:     runID,
		Files:     fileCount,
		Captures:  captureCount,
		Queries:   queryCount,
		Skipped:   skipCount,
		CommitID:  cfg.CommitID,
		Language:  cfg.Language,
		QueryFile: cfg.QueriesYML,
		FileGlob:  cfg.FileGlob,
	}, nil
}
