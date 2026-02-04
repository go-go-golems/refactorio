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

type ReportCommand struct {
	*cmds.CommandDescription
}

type ReportSettings struct {
	DBPath    string `glazed:"db"`
	RunID     int64  `glazed:"run-id"`
	OutputDir string `glazed:"out"`
}

var _ cmds.GlazeCommand = &ReportCommand{}

func NewReportCommand() (*ReportCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"report",
		cmds.WithShort("Generate SQL-backed reports"),
		cmds.WithLong("Render markdown reports from SQL queries."),
		cmds.WithFlags(
			fields.New(
				"db",
				fields.TypeString,
				fields.WithHelp("Path to the SQLite database"),
				fields.WithRequired(true),
			),
			fields.New(
				"run-id",
				fields.TypeInteger,
				fields.WithHelp("Run id to report on"),
				fields.WithRequired(true),
			),
			fields.New(
				"out",
				fields.TypeString,
				fields.WithHelp("Directory to write reports"),
				fields.WithDefault("reports"),
			),
		),
	)

	return &ReportCommand{CommandDescription: cmdDesc}, nil
}

func (c *ReportCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ReportSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	results, err := refactorindex.GenerateReports(ctx, refactorindex.ReportConfig{
		DBPath:    settings.DBPath,
		RunID:     settings.RunID,
		OutputDir: settings.OutputDir,
	})
	if err != nil {
		return err
	}

	for _, result := range results {
		if err := gp.AddRow(ctx, reportRow(result.Name, result.Path, result.RowCount)); err != nil {
			return errors.Wrap(err, "add report row")
		}
	}

	return nil
}

func reportRow(name string, path string, rowCount int) types.Row {
	return types.NewRow(
		types.MRP("name", name),
		types.MRP("path", path),
		types.MRP("rows", rowCount),
	)
}
