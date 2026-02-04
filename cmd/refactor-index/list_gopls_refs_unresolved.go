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

type ListGoplsRefsUnresolvedCommand struct {
	*cmds.CommandDescription
}

type ListGoplsRefsUnresolvedSettings struct {
	DBPath string `glazed:"db"`
	RunID  int64  `glazed:"run-id"`
	Limit  int    `glazed:"limit"`
	Offset int    `glazed:"offset"`
}

var _ cmds.GlazeCommand = &ListGoplsRefsUnresolvedCommand{}

func NewListGoplsRefsUnresolvedCommand() (*ListGoplsRefsUnresolvedCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"gopls-refs-unresolved",
		cmds.WithShort("List unresolved gopls reference records"),
		cmds.WithLong("List gopls references that could not be mapped to symbol definitions."),
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
				fields.WithHelp("Filter by a specific run id (optional)"),
				fields.WithDefault(0),
			),
			fields.New(
				"limit",
				fields.TypeInteger,
				fields.WithHelp("Limit number of rows (optional)"),
				fields.WithDefault(0),
			),
			fields.New(
				"offset",
				fields.TypeInteger,
				fields.WithHelp("Offset rows (optional)"),
				fields.WithDefault(0),
			),
		),
	)

	return &ListGoplsRefsUnresolvedCommand{CommandDescription: cmdDesc}, nil
}

func (c *ListGoplsRefsUnresolvedCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ListGoplsRefsUnresolvedSettings{}
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
	records, err := store.ListSymbolRefsUnresolved(ctx, refactorindex.SymbolRefUnresolvedFilter{
		RunID:  settings.RunID,
		Limit:  settings.Limit,
		Offset: settings.Offset,
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := gp.AddRow(ctx, unresolvedGoplsRefRow(record)); err != nil {
			return errors.Wrap(err, "add unresolved gopls ref row")
		}
	}

	return nil
}

func unresolvedGoplsRefRow(record refactorindex.SymbolRefUnresolvedRecord) types.Row {
	return types.NewRow(
		types.MRP("run_id", record.RunID),
		types.MRP("commit_hash", record.CommitHash),
		types.MRP("symbol_hash", record.SymbolHash),
		types.MRP("file", record.FilePath),
		types.MRP("line", record.Line),
		types.MRP("col", record.Col),
		types.MRP("is_decl", record.IsDecl),
		types.MRP("source", record.Source),
	)
}
