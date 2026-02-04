package main

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type IngestTreeSitterCommand struct {
	*cmds.CommandDescription
}

type IngestTreeSitterSettings struct {
	DBPath     string `glazed:"db"`
	RootDir    string `glazed:"root"`
	Language   string `glazed:"language"`
	QueriesYML string `glazed:"queries"`
	FileGlob   string `glazed:"file-glob"`
	CommitID   int64  `glazed:"commit-id"`
	SourcesDir string `glazed:"sources-dir"`
}

var _ cmds.GlazeCommand = &IngestTreeSitterCommand{}

func NewIngestTreeSitterCommand() (*IngestTreeSitterCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"tree-sitter",
		cmds.WithShort("Ingest tree-sitter query captures into the refactor index"),
		cmds.WithLong("Run Oak tree-sitter queries and store capture locations."),
		cmds.WithFlags(
			fields.New(
				"db",
				fields.TypeString,
				fields.WithHelp("Path to the SQLite database"),
				fields.WithRequired(true),
			),
			fields.New(
				"root",
				fields.TypeString,
				fields.WithHelp("Root directory to scan"),
				fields.WithRequired(true),
			),
			fields.New(
				"language",
				fields.TypeString,
				fields.WithHelp("Tree-sitter language name"),
				fields.WithRequired(true),
			),
			fields.New(
				"queries",
				fields.TypeString,
				fields.WithHelp("Path to tree-sitter query YAML"),
				fields.WithRequired(true),
			),
			fields.New(
				"file-glob",
				fields.TypeString,
				fields.WithHelp("Optional glob to filter files"),
				fields.WithDefault(""),
			),
			fields.New(
				"commit-id",
				fields.TypeInteger,
				fields.WithHelp("Optional commit id to associate with captures"),
				fields.WithDefault(0),
			),
			fields.New(
				"sources-dir",
				fields.TypeString,
				fields.WithHelp("Directory to write raw tool outputs"),
				fields.WithDefault("sources"),
			),
		),
	)

	return &IngestTreeSitterCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestTreeSitterCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestTreeSitterSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	var commitID *int64
	if settings.CommitID > 0 {
		commitID = &settings.CommitID
	}

	result, err := refactorindex.IngestTreeSitter(ctx, refactorindex.IngestTreeSitterConfig{
		DBPath:     settings.DBPath,
		RootDir:    settings.RootDir,
		Language:   settings.Language,
		QueriesYML: settings.QueriesYML,
		FileGlob:   settings.FileGlob,
		CommitID:   commitID,
		SourcesDir: settings.SourcesDir,
	})
	if err != nil {
		return err
	}

	if err := gp.AddRow(ctx, ingestTreeSitterRow(result)); err != nil {
		return errors.Wrap(err, "add ingest tree-sitter row")
	}

	return nil
}

func ingestTreeSitterRow(result *refactorindex.IngestTreeSitterResult) types.Row {
	commitID := int64(0)
	if result.CommitID != nil {
		commitID = *result.CommitID
	}

	return types.NewRow(
		types.MRP("run_id", result.RunID),
		types.MRP("files", result.Files),
		types.MRP("captures", result.Captures),
		types.MRP("queries", result.Queries),
		types.MRP("skipped", result.Skipped),
		types.MRP("language", result.Language),
		types.MRP("query_file", result.QueryFile),
		types.MRP("file_glob", result.FileGlob),
		types.MRP("commit_id", commitID),
	)
}
