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

type IngestCodeUnitsCommand struct {
	*cmds.CommandDescription
}

type IngestCodeUnitsSettings struct {
	DBPath              string `glazed:"db"`
	RootDir             string `glazed:"root"`
	SourcesDir          string `glazed:"sources-dir"`
	IgnorePackageErrors bool   `glazed:"ignore-package-errors"`
}

var _ cmds.GlazeCommand = &IngestCodeUnitsCommand{}

func NewIngestCodeUnitsCommand() (*IngestCodeUnitsCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"code-units",
		cmds.WithShort("Ingest code unit snapshots into the refactor index"),
		cmds.WithLong("Capture function/type snapshots with body hashes and doc text."),
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
			fields.New(
				"ignore-package-errors",
				fields.TypeBool,
				fields.WithHelp("Continue with partial results if go/packages reports errors"),
				fields.WithDefault(false),
			),
		),
	)

	return &IngestCodeUnitsCommand{CommandDescription: cmdDesc}, nil
}

func (c *IngestCodeUnitsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &IngestCodeUnitsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	result, err := refactorindex.IngestCodeUnits(ctx, refactorindex.IngestCodeUnitsConfig{
		DBPath:              settings.DBPath,
		RootDir:             settings.RootDir,
		SourcesDir:          settings.SourcesDir,
		IgnorePackageErrors: settings.IgnorePackageErrors,
	})
	if err != nil {
		return err
	}

	if err := gp.AddRow(ctx, ingestCodeUnitsRow(
		result.RunID,
		result.CodeUnits,
		result.Snapshots,
		result.Packages,
		result.PackagesWithErrors,
		result.PackagesSkipped,
		result.Files,
		result.BodyBytes,
		result.DocEntries,
	)); err != nil {
		return errors.Wrap(err, "add ingest code-units row")
	}

	return nil
}

func ingestCodeUnitsRow(runID int64, codeUnits int, snapshots int, packages int, packagesWithErrors int, packagesSkipped int, files int, bodyBytes int, docEntries int) types.Row {
	return types.NewRow(
		types.MRP("run_id", runID),
		types.MRP("code_units", codeUnits),
		types.MRP("snapshots", snapshots),
		types.MRP("packages", packages),
		types.MRP("packages_with_errors", packagesWithErrors),
		types.MRP("packages_skipped", packagesSkipped),
		types.MRP("files", files),
		types.MRP("body_bytes", bodyBytes),
		types.MRP("doc_entries", docEntries),
	)
}
