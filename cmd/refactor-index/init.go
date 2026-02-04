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

type InitCommand struct {
	*cmds.CommandDescription
}

type InitSettings struct {
	DBPath string `glazed:"db"`
}

var _ cmds.GlazeCommand = &InitCommand{}

func NewInitCommand() (*InitCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"init",
		cmds.WithShort("Initialize a refactor index database"),
		cmds.WithLong("Create the SQLite schema and metadata tables for the refactor index."),
		cmds.WithFlags(
			fields.New(
				"db",
				fields.TypeString,
				fields.WithHelp("Path to the SQLite database"),
				fields.WithRequired(true),
			),
		),
	)

	return &InitCommand{CommandDescription: cmdDesc}, nil
}

func (c *InitCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &InitSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return err
	}

	return errors.New("init not implemented")
}

func initRow(dbPath string, schemaVersion int) *types.Row {
	return types.NewRow(
		types.MRP("db_path", dbPath),
		types.MRP("schema_version", schemaVersion),
	)
}
