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
)

type IngestDiffCommand struct {
	*cmds.CommandDescription
}

type IngestDiffSettings struct {
	DBPath     string `glazed:"db"`
	RepoPath   string `glazed:"repo"`
	FromRef    string `glazed:"from"`
	ToRef      string `glazed:"to"`
	SourcesDir string `glazed:"sources-dir"`
}

var _ cmds.GlazeCommand = &IngestDiffCommand{}

func NewIngestDiffCommand() (*IngestDiffCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"diff",
		cmds.WithShort("Ingest git diff metadata into the refactor index"),
		cmds.WithLong("Capture git diff name-status and unified patch data into SQLite."),
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
				fields.WithHelp("Git ref for the start of the diff"),
				fields.WithRequired(true),
			),
			fields.New(
				"to",
				fields.TypeString,
				fields.WithHelp("Git ref for the end of the diff"),
				fields.WithRequired(true),
			),
			fields.New(
				"sources-dir",
				fields.TypeString,
				fields.WithHelp("Directory to write raw tool outputs"),
				fields.WithDefault("sources"),
			),
		),
	)

	return &IngestDiffCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestDiffCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestDiffSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return err
	}

	return errors.New("ingest diff not implemented")
}

func ingestDiffRow(runID int64, files int, hunks int, lines int) *types.Row {
	return types.NewRow(
		types.MRP("run_id", runID),
		types.MRP("files", files),
		types.MRP("hunks", hunks),
		types.MRP("lines", lines),
	)
}
