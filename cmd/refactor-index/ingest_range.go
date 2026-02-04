package main

import (
	"context"
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

type IngestRangeCommand struct {
	*cmds.CommandDescription
}

type IngestRangeSettings struct {
	DBPath     string `glazed:"db"`
	RepoPath   string `glazed:"repo"`
	FromRef    string `glazed:"from"`
	ToRef      string `glazed:"to"`
	SourcesDir string `glazed:"sources-dir"`

	IncludeDiff         bool `glazed:"include-diff"`
	IncludeSymbols      bool `glazed:"include-symbols"`
	IncludeCodeUnits    bool `glazed:"include-code-units"`
	IncludeDocHits      bool `glazed:"include-doc-hits"`
	IncludeGopls        bool `glazed:"include-gopls"`
	IgnorePackageErrors bool `glazed:"ignore-package-errors"`

	TermsFile             string   `glazed:"terms"`
	GoplsTargets          []string `glazed:"gopls-target"`
	GoplsTargetsFile      string   `glazed:"gopls-targets-file"`
	GoplsTargetsJSON      string   `glazed:"gopls-targets-json"`
	GoplsSkipSymbolLookup bool     `glazed:"gopls-skip-symbol-lookup"`
}

var _ cmds.GlazeCommand = &IngestRangeCommand{}

func NewIngestRangeCommand() (*IngestRangeCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"range",
		cmds.WithShort("Ingest multiple passes across a commit range"),
		cmds.WithLong("Orchestrate commit lineage plus optional diff/symbols/code units/doc hits/gopls ingestion."),
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
				"from",
				fields.TypeString,
				fields.WithHelp("Git ref for the start of the range"),
				fields.WithRequired(true),
			),
			fields.New(
				"to",
				fields.TypeString,
				fields.WithHelp("Git ref for the end of the range"),
				fields.WithRequired(true),
			),
			fields.New(
				"sources-dir",
				fields.TypeString,
				fields.WithHelp("Directory to write raw tool outputs"),
				fields.WithDefault("sources"),
			),
			fields.New(
				"include-diff",
				fields.TypeBool,
				fields.WithHelp("Include diff ingestion per commit"),
				fields.WithDefault(false),
			),
			fields.New(
				"include-symbols",
				fields.TypeBool,
				fields.WithHelp("Include symbol ingestion per commit"),
				fields.WithDefault(false),
			),
			fields.New(
				"include-code-units",
				fields.TypeBool,
				fields.WithHelp("Include code unit ingestion per commit"),
				fields.WithDefault(false),
			),
			fields.New(
				"include-doc-hits",
				fields.TypeBool,
				fields.WithHelp("Include doc hits ingestion per commit"),
				fields.WithDefault(false),
			),
			fields.New(
				"include-gopls",
				fields.TypeBool,
				fields.WithHelp("Include gopls references ingestion per commit"),
				fields.WithDefault(false),
			),
			fields.New(
				"ignore-package-errors",
				fields.TypeBool,
				fields.WithHelp("Continue symbols/code-units with partial results if go/packages reports errors"),
				fields.WithDefault(false),
			),
			fields.New(
				"terms",
				fields.TypeString,
				fields.WithHelp("Terms file for doc hits"),
				fields.WithDefault(""),
			),
			fields.New(
				"gopls-target",
				fields.TypeStringList,
				fields.WithHelp("Gopls target spec 'symbol_hash|path|line|col|commit_id'"),
				fields.WithDefault([]string{}),
			),
			fields.New(
				"gopls-targets-file",
				fields.TypeString,
				fields.WithHelp("File containing gopls target specs"),
				fields.WithDefault(""),
			),
			fields.New(
				"gopls-targets-json",
				fields.TypeString,
				fields.WithHelp("JSON file containing gopls target specs"),
				fields.WithDefault(""),
			),
			fields.New(
				"gopls-skip-symbol-lookup",
				fields.TypeBool,
				fields.WithHelp("Skip symbol hash lookup and store unresolved refs"),
				fields.WithDefault(true),
			),
		),
	)

	return &IngestRangeCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestRangeCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestRangeSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	goplsTargets, err := loadGoplsTargets(settings.GoplsTargets, settings.GoplsTargetsFile, settings.GoplsTargetsJSON)
	if err != nil {
		return err
	}

	result, err := refactorindex.IngestCommitRange(ctx, refactorindex.RangeIngestConfig{
		DBPath:                settings.DBPath,
		RepoPath:              settings.RepoPath,
		FromRef:               settings.FromRef,
		ToRef:                 settings.ToRef,
		SourcesDir:            settings.SourcesDir,
		IncludeDiff:           settings.IncludeDiff,
		IncludeSymbols:        settings.IncludeSymbols,
		IncludeCodeUnits:      settings.IncludeCodeUnits,
		IncludeDocHits:        settings.IncludeDocHits,
		IncludeGopls:          settings.IncludeGopls,
		IgnorePackageErrors:   settings.IgnorePackageErrors,
		TermsFile:             settings.TermsFile,
		GoplsTargets:          goplsTargets,
		GoplsSkipSymbolLookup: settings.GoplsSkipSymbolLookup,
	})
	if err != nil {
		return err
	}

	if len(result.Commits) == 0 {
		return gp.AddRow(ctx, ingestRangeSummaryRow(result.CommitLineageRunID, 0))
	}

	for _, commit := range result.Commits {
		row := types.NewRow(
			types.MRP("commit_lineage_run_id", result.CommitLineageRunID),
			types.MRP("commit_count", len(result.Commits)),
			types.MRP("commit_hash", commit.CommitHash),
			types.MRP("diff_run_id", commit.DiffRunID),
			types.MRP("symbols_run_id", commit.SymbolsRunID),
			types.MRP("code_units_run_id", commit.CodeUnitsRunID),
			types.MRP("doc_hits_run_id", commit.DocHitsRunID),
			types.MRP("gopls_run_id", commit.GoplsRunID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "add ingest range row")
		}
	}

	return nil
}

func ingestRangeSummaryRow(runID int64, commitCount int) types.Row {
	return types.NewRow(
		types.MRP("commit_lineage_run_id", runID),
		types.MRP("commit_count", commitCount),
		types.MRP("commit_hash", ""),
		types.MRP("diff_run_id", 0),
		types.MRP("symbols_run_id", 0),
		types.MRP("code_units_run_id", 0),
		types.MRP("doc_hits_run_id", 0),
		types.MRP("gopls_run_id", 0),
	)
}

func (c *IngestRangeCommand) Validate(values *values.Values) error {
	settings := &IngestRangeSettings{}
	if err := values.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.IncludeDocHits && strings.TrimSpace(settings.TermsFile) == "" {
		return errors.New("terms file is required when include-doc-hits is set")
	}
	return nil
}
