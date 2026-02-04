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

type IngestCommitsCommand struct {
	*cmds.CommandDescription
}

type IngestCommitsSettings struct {
	DBPath   string `glazed:"db"`
	RepoPath string `glazed:"repo"`
	FromRef  string `glazed:"from"`
	ToRef    string `glazed:"to"`
}

var _ cmds.GlazeCommand = &IngestCommitsCommand{}

func NewIngestCommitsCommand() (*IngestCommitsCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"commits",
		cmds.WithShort("Ingest commit lineage into the refactor index"),
		cmds.WithLong("Capture commit metadata, file changes, and blob stats into SQLite."),
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
		),
	)

	return &IngestCommitsCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestCommitsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestCommitsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	result, err := refactorindex.IngestCommits(ctx, refactorindex.IngestCommitsConfig{
		DBPath:   settings.DBPath,
		RepoPath: settings.RepoPath,
		FromRef:  settings.FromRef,
		ToRef:    settings.ToRef,
	})
	if err != nil {
		return err
	}

	if err := gp.AddRow(ctx, ingestCommitsRow(result.RunID, result.CommitCount, result.FileCount, result.BlobCount)); err != nil {
		return errors.Wrap(err, "add ingest commits row")
	}

	return nil
}

func ingestCommitsRow(runID int64, commits int, files int, blobs int) types.Row {
	return types.NewRow(
		types.MRP("run_id", runID),
		types.MRP("commits", commits),
		types.MRP("files", files),
		types.MRP("blobs", blobs),
	)
}
