package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"

	"github.com/go-go-golems/refactorio/pkg/refactorindex"
)

type ListSymbolsCommand struct {
	*cmds.CommandDescription
}

type ListSymbolsSettings struct {
	DBPath       string `glazed:"db"`
	RunID        int64  `glazed:"run-id"`
	ExportedOnly bool   `glazed:"exported-only"`
	Kind         string `glazed:"kind"`
	Name         string `glazed:"name"`
	Pkg          string `glazed:"pkg"`
	Path         string `glazed:"path"`
	Limit        int    `glazed:"limit"`
	Offset       int    `glazed:"offset"`
}

var _ cmds.GlazeCommand = &ListSymbolsCommand{}

func NewListSymbolsCommand() (*ListSymbolsCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"symbols",
		cmds.WithShort("List symbol inventory records"),
		cmds.WithLong("List symbol definitions and occurrences stored in the refactor index."),
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
				"exported-only",
				fields.TypeBool,
				fields.WithHelp("Only include exported symbols"),
				fields.WithDefault(false),
			),
			fields.New(
				"kind",
				fields.TypeString,
				fields.WithHelp("Filter by symbol kind (optional)"),
				fields.WithDefault(""),
			),
			fields.New(
				"name",
				fields.TypeString,
				fields.WithHelp("Filter by symbol name (optional)"),
				fields.WithDefault(""),
			),
			fields.New(
				"pkg",
				fields.TypeString,
				fields.WithHelp("Filter by package path (optional)"),
				fields.WithDefault(""),
			),
			fields.New(
				"path",
				fields.TypeString,
				fields.WithHelp("Filter by file path (optional)"),
				fields.WithDefault(""),
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

	return &ListSymbolsCommand{CommandDescription: cmdDesc}, nil
}

func (c *ListSymbolsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ListSymbolsSettings{}
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
	records, err := store.ListSymbolInventory(ctx, refactorindex.SymbolInventoryFilter{
		RunID:        settings.RunID,
		ExportedOnly: settings.ExportedOnly,
		Kind:         settings.Kind,
		Name:         settings.Name,
		Pkg:          settings.Pkg,
		Path:         settings.Path,
		Limit:        settings.Limit,
		Offset:       settings.Offset,
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := gp.AddRow(ctx, symbolInventoryRow(record)); err != nil {
			return errors.Wrap(err, "add symbol inventory row")
		}
	}

	return nil
}

func symbolInventoryRow(record refactorindex.SymbolInventoryRecord) types.Row {
	targetSpec := fmt.Sprintf("%s|%s|%d|%d", record.SymbolHash, record.FilePath, record.Line, record.Col)
	return types.NewRow(
		types.MRP("run_id", record.RunID),
		types.MRP("symbol_hash", record.SymbolHash),
		types.MRP("name", record.Name),
		types.MRP("kind", record.Kind),
		types.MRP("pkg", record.Pkg),
		types.MRP("recv", record.Recv),
		types.MRP("signature", record.Signature),
		types.MRP("file", record.FilePath),
		types.MRP("line", record.Line),
		types.MRP("col", record.Col),
		types.MRP("is_exported", record.IsExported),
		types.MRP("target_spec", targetSpec),
	)
}
