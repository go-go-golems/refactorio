package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type IngestGoplsRefsCommand struct {
	*cmds.CommandDescription
}

type IngestGoplsRefsSettings struct {
	DBPath           string   `glazed:"db"`
	RepoPath         string   `glazed:"repo"`
	SourcesDir       string   `glazed:"sources-dir"`
	Targets          []string `glazed:"target"`
	TargetsFile      string   `glazed:"targets-file"`
	TargetsJSON      string   `glazed:"targets-json"`
	SkipSymbolLookup bool     `glazed:"skip-symbol-lookup"`
}

type goplsTargetJSON struct {
	SymbolHash string `json:"symbol_hash"`
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	CommitID   *int64 `json:"commit_id,omitempty"`
}

var _ cmds.GlazeCommand = &IngestGoplsRefsCommand{}

func NewIngestGoplsRefsCommand() (*IngestGoplsRefsCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"gopls-refs",
		cmds.WithShort("Ingest gopls reference locations into the refactor index"),
		cmds.WithLong("Run gopls references for symbol targets and store results."),
		cmds.WithFlags(
			fields.New(
				"db",
				fields.TypeString,
				fields.WithHelp("Path to the SQLite database"),
				fields.WithRequired(true),
			),
			fields.New(
				"repo",
				fields.TypeString,
				fields.WithHelp("Path to the git repository"),
				fields.WithRequired(true),
			),
			fields.New(
				"sources-dir",
				fields.TypeString,
				fields.WithHelp("Directory to write raw tool outputs"),
				fields.WithDefault("sources"),
			),
			fields.New(
				"target",
				fields.TypeStringList,
				fields.WithHelp("Target spec 'symbol_hash|path|line|col|commit_id' (commit_id optional)"),
				fields.WithDefault([]string{}),
			),
			fields.New(
				"targets-file",
				fields.TypeString,
				fields.WithHelp("File containing target specs, one per line"),
				fields.WithDefault(""),
			),
			fields.New(
				"targets-json",
				fields.TypeString,
				fields.WithHelp("JSON file with target objects"),
				fields.WithDefault(""),
			),
			fields.New(
				"skip-symbol-lookup",
				fields.TypeBool,
				fields.WithHelp("Skip symbol hash lookup and store unresolved refs"),
				fields.WithDefault(true),
			),
		),
	)

	return &IngestGoplsRefsCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestGoplsRefsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestGoplsRefsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	targets, err := loadGoplsTargets(settings.Targets, settings.TargetsFile, settings.TargetsJSON)
	if err != nil {
		return err
	}

	result, err := refactorindex.IngestGoplsReferences(ctx, refactorindex.IngestGoplsRefsConfig{
		DBPath:           settings.DBPath,
		RepoPath:         settings.RepoPath,
		SourcesDir:       settings.SourcesDir,
		Targets:          targets,
		SkipSymbolLookup: settings.SkipSymbolLookup,
	})
	if err != nil {
		return err
	}

	if err := gp.AddRow(ctx, ingestGoplsRefsRow(result)); err != nil {
		return errors.Wrap(err, "add ingest gopls refs row")
	}

	return nil
}

func ingestGoplsRefsRow(result *refactorindex.IngestGoplsRefsResult) types.Row {
	return types.NewRow(
		types.MRP("run_id", result.RunID),
		types.MRP("targets", result.Targets),
		types.MRP("references", result.References),
		types.MRP("raw_outputs", result.RawOutputs),
		types.MRP("skipped_files", result.SkippedFiles),
	)
}

func loadGoplsTargets(specs []string, specsFile string, jsonFile string) ([]refactorindex.GoplsRefTarget, error) {
	collected := make([]refactorindex.GoplsRefTarget, 0)

	for _, spec := range specs {
		target, err := parseGoplsTargetSpec(spec)
		if err != nil {
			return nil, err
		}
		collected = append(collected, target)
	}

	if strings.TrimSpace(specsFile) != "" {
		data, err := os.ReadFile(specsFile)
		if err != nil {
			return nil, errors.Wrap(err, "read targets file")
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			target, err := parseGoplsTargetSpec(line)
			if err != nil {
				return nil, err
			}
			collected = append(collected, target)
		}
	}

	if strings.TrimSpace(jsonFile) != "" {
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			return nil, errors.Wrap(err, "read targets json")
		}
		var items []goplsTargetJSON
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, errors.Wrap(err, "parse targets json")
		}
		for _, item := range items {
			if strings.TrimSpace(item.FilePath) == "" {
				return nil, errors.New("targets json missing file_path")
			}
			if item.Line == 0 || item.Col == 0 {
				return nil, errors.New("targets json missing line/col")
			}
			collected = append(collected, refactorindex.GoplsRefTarget{
				SymbolHash: item.SymbolHash,
				FilePath:   item.FilePath,
				Line:       item.Line,
				Col:        item.Col,
				CommitID:   item.CommitID,
			})
		}
	}

	return collected, nil
}

func parseGoplsTargetSpec(spec string) (refactorindex.GoplsRefTarget, error) {
	parts := strings.Split(spec, "|")
	if len(parts) < 4 || len(parts) > 5 {
		return refactorindex.GoplsRefTarget{}, errors.Errorf("invalid target spec: %q", spec)
	}

	symbolHash := strings.TrimSpace(parts[0])
	filePath := strings.TrimSpace(parts[1])
	lineStr := strings.TrimSpace(parts[2])
	colStr := strings.TrimSpace(parts[3])
	if filePath == "" || lineStr == "" || colStr == "" {
		return refactorindex.GoplsRefTarget{}, errors.Errorf("invalid target spec: %q", spec)
	}

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return refactorindex.GoplsRefTarget{}, errors.Wrap(err, "parse target line")
	}
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return refactorindex.GoplsRefTarget{}, errors.Wrap(err, "parse target col")
	}

	var commitID *int64
	if len(parts) == 5 {
		commitStr := strings.TrimSpace(parts[4])
		if commitStr != "" {
			value, err := strconv.ParseInt(commitStr, 10, 64)
			if err != nil {
				return refactorindex.GoplsRefTarget{}, errors.Wrap(err, "parse target commit id")
			}
			commitID = &value
		}
	}

	return refactorindex.GoplsRefTarget{
		SymbolHash: symbolHash,
		FilePath:   filePath,
		Line:       line,
		Col:        col,
		CommitID:   commitID,
	}, nil
}
