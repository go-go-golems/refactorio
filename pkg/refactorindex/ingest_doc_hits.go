package refactorindex

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type IngestDocHitsConfig struct {
	DBPath     string
	RootDir    string
	TermsFile  string
	CommitID   *int64
	SourcesDir string
}

type IngestDocHitsResult struct {
	RunID     int64
	Terms     int
	Hits      int
	Files     int
	Skipped   int
	CommitID  *int64
	TermsFile string
}

func IngestDocHits(ctx context.Context, cfg IngestDocHitsConfig) (*IngestDocHitsResult, error) {
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, errors.New("db path is required")
	}
	if strings.TrimSpace(cfg.RootDir) == "" {
		return nil, errors.New("root dir is required")
	}
	if strings.TrimSpace(cfg.TermsFile) == "" {
		return nil, errors.New("terms file is required")
	}

	rootDir, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolve root dir")
	}
	termsPath, err := filepath.Abs(cfg.TermsFile)
	if err != nil {
		return nil, errors.Wrap(err, "resolve terms file")
	}

	terms, err := readTermsFile(termsPath)
	if err != nil {
		return nil, err
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
		"termsFile": termsPath,
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

	tx, err := store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	sourcesDir := cfg.SourcesDir
	if strings.TrimSpace(sourcesDir) == "" {
		sourcesDir = "sources"
	}
	sourcesDir, err = filepath.Abs(sourcesDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolve sources dir")
	}
	runDir := filepath.Join(sourcesDir, fmt.Sprintf("%d", runID), "doc-hits")

	fileIDs := make(map[string]int64)
	hitCount := 0
	skipCount := 0

	for _, term := range terms {
		if term == "" {
			continue
		}
		out, err := runRipgrep(ctx, rootDir, term)
		if err != nil {
			return nil, err
		}
		fileName := fmt.Sprintf("rg-%s.txt", slugify(term))
		if _, err := store.WriteRawOutput(ctx, tx, runDir, runID, "rg", fileName, out); err != nil {
			return nil, err
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			filePath, lineNum, colNum, matchText, err := parseRipgrepLine(line)
			if err != nil {
				skipCount++
				continue
			}
			relPath := filePath
			if filepath.IsAbs(filePath) {
				rel, err := filepath.Rel(rootDir, filePath)
				if err == nil {
					relPath = filepath.ToSlash(rel)
				}
			}
			fileID, ok := fileIDs[relPath]
			if !ok {
				id, err := store.GetOrCreateFile(ctx, tx, relPath)
				if err != nil {
					return nil, err
				}
				fileID = id
				fileIDs[relPath] = id
			}
			if err := store.InsertDocHit(ctx, tx, runID, cfg.CommitID, fileID, lineNum, colNum, term, matchText); err != nil {
				return nil, err
			}
			hitCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit doc hits")
	}
	if err := store.FinishRun(ctx, runID); err != nil {
		return nil, err
	}

	return &IngestDocHitsResult{
		RunID:     runID,
		Terms:     len(terms),
		Hits:      hitCount,
		Files:     len(fileIDs),
		Skipped:   skipCount,
		CommitID:  cfg.CommitID,
		TermsFile: termsPath,
	}, nil
}

func readTermsFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read terms file")
	}
	lines := strings.Split(string(content), "\n")
	terms := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		terms = append(terms, line)
	}
	return terms, nil
}

func runRipgrep(ctx context.Context, rootDir string, term string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "rg", "--line-number", "--column", "--no-heading", "--color=never", "-F", term, rootDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// rg returns exit code 1 when no matches are found
			if exitErr.ExitCode() == 1 {
				return []byte{}, nil
			}
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = "rg command failed"
		}
		return nil, errors.Wrap(err, msg)
	}
	return stdout, nil
}

func parseRipgrepLine(line string) (string, int, int, string, error) {
	parts := strings.SplitN(line, ":", 4)
	if len(parts) < 4 {
		return "", 0, 0, "", errors.New("invalid rg line")
	}
	lineNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, 0, "", errors.Wrap(err, "parse line")
	}
	colNum, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, 0, "", errors.Wrap(err, "parse col")
	}
	return parts[0], lineNum, colNum, parts[3], nil
}

func slugify(term string) string {
	term = strings.ToLower(term)
	term = strings.ReplaceAll(term, " ", "-")
	term = strings.ReplaceAll(term, "/", "-")
	term = strings.ReplaceAll(term, "\\", "-")
	term = strings.ReplaceAll(term, ":", "-")
	term = strings.ReplaceAll(term, "#", "-")
	term = strings.ReplaceAll(term, "\t", "-")
	term = strings.Trim(term, "-")
	if term == "" {
		term = "term"
	}
	return term
}
