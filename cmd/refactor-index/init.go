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
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	db, err := refactorindex.OpenDB(ctx, settings.DBPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	store := refactorindex.NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		return err
	}

	if err := gp.AddRow(ctx, initRow(settings.DBPath, refactorindex.SchemaVersion)); err != nil {
		return errors.Wrap(err, "add init row")
	}

	return nil
}

func initRow(dbPath string, schemaVersion int) types.Row {
	return types.NewRow(
		types.MRP("db_path", dbPath),
		types.MRP("schema_version", schemaVersion),
	)
}
