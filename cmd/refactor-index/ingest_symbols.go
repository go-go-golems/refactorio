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

type IngestSymbolsCommand struct {
	*cmds.CommandDescription
}

type IngestSymbolsSettings struct {
	DBPath     string `glazed:"db"`
	RootDir    string `glazed:"root"`
	SourcesDir string `glazed:"sources-dir"`
}

var _ cmds.GlazeCommand = &IngestSymbolsCommand{}

func NewIngestSymbolsCommand() (*IngestSymbolsCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"symbols",
		cmds.WithShort("Ingest Go AST symbols into the refactor index"),
		cmds.WithLong("Capture Go symbol definitions and occurrences using go/packages."),
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
				fields.WithHelp("Root directory to scan for Go packages"),
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

	return &IngestSymbolsCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestSymbolsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestSymbolsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	result, err := refactorindex.IngestSymbols(ctx, refactorindex.IngestSymbolsConfig{
		DBPath:     settings.DBPath,
		RootDir:    settings.RootDir,
		SourcesDir: settings.SourcesDir,
	})
	if err != nil {
		return err
	}

	if err := gp.AddRow(ctx, ingestSymbolsRow(result.RunID, result.Symbols, result.Occurrences, result.Packages, result.Files)); err != nil {
		return errors.Wrap(err, "add ingest symbols row")
	}

	return nil
}

func ingestSymbolsRow(runID int64, symbols int, occurrences int, packages int, files int) types.Row {
	return types.NewRow(
		types.MRP("run_id", runID),
		types.MRP("symbols", symbols),
		types.MRP("occurrences", occurrences),
		types.MRP("packages", packages),
		types.MRP("files", files),
	)
}
