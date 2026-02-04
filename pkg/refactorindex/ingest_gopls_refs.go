package refactorindex

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type GoplsRefTarget struct {
	SymbolHash string
	FilePath   string
	Line       int
	Col        int
	CommitID   *int64
}

type IngestGoplsRefsConfig struct {
	DBPath     string
	RepoPath   string
	SourcesDir string
	Targets    []GoplsRefTarget
}

type IngestGoplsRefsResult struct {
	RunID        int64
	Targets      int
	References   int
	RawOutputs   int
	SkippedFiles int
}

func IngestGoplsReferences(ctx context.Context, cfg IngestGoplsRefsConfig) (*IngestGoplsRefsResult, error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RepoPath) == "" {
		return nil, errors.New("repo path is required")
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
		"repo": cfg.RepoPath,
	})
	if err != nil {
		return nil, err
	}

	runID, err := store.CreateRun(ctx, RunConfig{
		ToolVersion: ToolVersion,
		RootPath:    cfg.RepoPath,
		SourcesDir:  cfg.SourcesDir,
		ArgsJSON:    argsJSON,
	})
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
	runDir := filepath.Join(sourcesDir, fmt.Sprintf("%d", runID), "gopls")

	referenceCount := 0
	rawCount := 0
	skippedFiles := 0

	for idx, target := range cfg.Targets {
		if target.SymbolHash == "" {
			continue
		}
		if target.FilePath == "" || target.Line == 0 || target.Col == 0 {
			skippedFiles++
			continue
		}
		symbolID, err := store.GetSymbolDefIDByHash(ctx, tx, target.SymbolHash)
		if err != nil {
			return nil, err
		}

		absPath := target.FilePath
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(repoPath, target.FilePath)
		}
		position := fmt.Sprintf("%s:%d:%d", absPath, target.Line, target.Col)

		if _, err := runGopls(ctx, repoPath, "prepare_rename", position); err != nil {
			return nil, errors.Wrap(err, "gopls prepare_rename")
		}

		refs, err := runGopls(ctx, repoPath, "references", "-declaration", position)
		if err != nil {
			return nil, errors.Wrap(err, "gopls references")
		}

		fileName := fmt.Sprintf("gopls-references-%d.txt", idx)
		if _, err := store.WriteRawOutput(ctx, tx, runDir, runID, "gopls-references", fileName, refs); err != nil {
			return nil, err
		}
		rawCount++

		locations := parseGoplsReferences(refs)
		for _, loc := range locations {
			relPath := loc.FilePath
			if filepath.IsAbs(relPath) {
				rel, err := filepath.Rel(repoPath, relPath)
				if err == nil {
					relPath = filepath.ToSlash(rel)
				}
			}
			fileID, err := store.GetOrCreateFile(ctx, tx, relPath)
			if err != nil {
				return nil, err
			}

			isDecl := loc.Line == target.Line && loc.Col == target.Col && sameFilePath(loc.FilePath, absPath)
			if err := store.InsertSymbolRef(ctx, tx, runID, target.CommitID, symbolID, fileID, loc.Line, loc.Col, isDecl, "gopls"); err != nil {
				return nil, err
			}
			referenceCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit gopls references")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestGoplsRefsResult{
		RunID:        runID,
		Targets:      len(cfg.Targets),
		References:   referenceCount,
		RawOutputs:   rawCount,
		SkippedFiles: skippedFiles,
	}, nil
}

type goplsLocation struct {
	FilePath string
	Line     int
	Col      int
}

func parseGoplsReferences(output []byte) []goplsLocation {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	locations := make([]goplsLocation, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		loc, err := parseGoplsLocation(line)
		if err != nil {
			continue
		}
		locations = append(locations, loc)
	}
	return locations
}

func parseGoplsLocation(line string) (goplsLocation, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return goplsLocation{}, errors.New("empty line")
	}

	parts := strings.Split(line, ":")
	if len(parts) < 3 {
		return goplsLocation{}, errors.New("invalid reference line")
	}

	// Handle path:line:col-line:col (no extra colon separators)
	if len(parts) == 4 && strings.Contains(parts[len(parts)-2], "-") {
		file := strings.Join(parts[:len(parts)-3], ":")
		lineNum, err := strconv.Atoi(parts[len(parts)-3])
		if err != nil {
			return goplsLocation{}, errors.Wrap(err, "parse line")
		}
		colPart := strings.Split(parts[len(parts)-2], "-")[0]
		colNum, err := strconv.Atoi(colPart)
		if err != nil {
			return goplsLocation{}, errors.Wrap(err, "parse col")
		}
		return goplsLocation{FilePath: file, Line: lineNum, Col: colNum}, nil
	}

	// Handle lines with start/end positions: path:line:col:line:col
	if len(parts) >= 5 {
		startLine, err1 := strconv.Atoi(parts[len(parts)-4])
		startCol, err2 := strconv.Atoi(parts[len(parts)-3])
		_, err3 := strconv.Atoi(parts[len(parts)-2])
		_, err4 := strconv.Atoi(parts[len(parts)-1])
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			file := strings.Join(parts[:len(parts)-4], ":")
			return goplsLocation{FilePath: file, Line: startLine, Col: startCol}, nil
		}
	}

	colPart := parts[len(parts)-1]
	linePart := parts[len(parts)-2]
	file := strings.Join(parts[:len(parts)-2], ":")

	if strings.Contains(colPart, "-") {
		colPart = strings.Split(colPart, "-")[0]
	}
	if strings.Contains(linePart, "-") {
		linePart = strings.Split(linePart, "-")[0]
	}

	lineNum, err := strconv.Atoi(linePart)
	if err != nil {
		return goplsLocation{}, errors.Wrap(err, "parse line")
	}
	colNum, err := strconv.Atoi(colPart)
	if err != nil {
		return goplsLocation{}, errors.Wrap(err, "parse col")
	}

	return goplsLocation{FilePath: file, Line: lineNum, Col: colNum}, nil
}

func runGopls(ctx context.Context, repoPath string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gopls", args...)
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = "gopls command failed"
		}
		return nil, errors.Wrap(err, msg)
	}
	return stdout, nil
}

func sameFilePath(a string, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}
