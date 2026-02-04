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

type IngestDocHitsCommand struct {
	*cmds.CommandDescription
}

type IngestDocHitsSettings struct {
	DBPath     string `glazed:"db"`
	RootDir    string `glazed:"root"`
	TermsFile  string `glazed:"terms"`
	CommitID   int64  `glazed:"commit-id"`
	SourcesDir string `glazed:"sources-dir"`
}

var _ cmds.GlazeCommand = &IngestDocHitsCommand{}

func NewIngestDocHitsCommand() (*IngestDocHitsCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"doc-hits",
		cmds.WithShort("Ingest document/term hits into the refactor index"),
		cmds.WithLong("Search text files for terms and store line/column hit positions."),
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
				"terms",
				fields.TypeString,
				fields.WithHelp("Path to terms file (one term per line)"),
				fields.WithRequired(true),
			),
			fields.New(
				"commit-id",
				fields.TypeInteger,
				fields.WithHelp("Optional commit id to associate with hits"),
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

	return &IngestDocHitsCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestDocHitsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestDocHitsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	var commitID *int64
	if settings.CommitID > 0 {
		commitID = &settings.CommitID
	}

	result, err := refactorindex.IngestDocHits(ctx, refactorindex.IngestDocHitsConfig{
		DBPath:     settings.DBPath,
		RootDir:    settings.RootDir,
		TermsFile:  settings.TermsFile,
		CommitID:   commitID,
		SourcesDir: settings.SourcesDir,
	})
	if err != nil {
		return err
	}

	if err := gp.AddRow(ctx, ingestDocHitsRow(result)); err != nil {
		return errors.Wrap(err, "add ingest doc hits row")
	}

	return nil
}

func ingestDocHitsRow(result *refactorindex.IngestDocHitsResult) types.Row {
	commitID := int64(0)
	if result.CommitID != nil {
		commitID = *result.CommitID
	}

	return types.NewRow(
		types.MRP("run_id", result.RunID),
		types.MRP("terms", result.Terms),
		types.MRP("hits", result.Hits),
		types.MRP("files", result.Files),
		types.MRP("skipped", result.Skipped),
		types.MRP("terms_file", result.TermsFile),
		types.MRP("commit_id", commitID),
	)
}
