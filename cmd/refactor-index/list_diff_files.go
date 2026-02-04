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

type ListDiffFilesCommand struct {
	*cmds.CommandDescription
}

type ListDiffFilesSettings struct {
	DBPath string `glazed:"db"`
	RunID  int64  `glazed:"run-id"`
	Limit  int    `glazed:"limit"`
	Offset int    `glazed:"offset"`
}

var _ cmds.GlazeCommand = &ListDiffFilesCommand{}

func NewListDiffFilesCommand() (*ListDiffFilesCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"diff-files",
		cmds.WithShort("List diff files stored in the index"),
		cmds.WithLong("Query diff file metadata from the refactor index."),
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

	return &ListDiffFilesCommand{CommandDescription: cmdDesc}, nil
}

func (c *ListDiffFilesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ListDiffFilesSettings{}
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
	records, err := store.ListDiffFiles(ctx, refactorindex.DiffFileFilter{
		RunID:  settings.RunID,
		Limit:  settings.Limit,
		Offset: settings.Offset,
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := gp.AddRow(ctx, diffFileRow(record.RunID, record.Status, record.Path, record.OldPath, record.NewPath)); err != nil {
			return errors.Wrap(err, "add diff file row")
		}
	}

	return nil
}

func diffFileRow(runID int64, status string, path string, oldPath string, newPath string) types.Row {
	return types.NewRow(
		types.MRP("run_id", runID),
		types.MRP("status", status),
		types.MRP("path", path),
		types.MRP("old_path", oldPath),
		types.MRP("new_path", newPath),
	)
}
