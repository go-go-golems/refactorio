package main

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/refactorio/pkg/workbenchapi"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type APIServeCommand struct {
	*cmds.CommandDescription
}

type APIServeSettings struct {
	Addr                string `glazed.parameter:"addr"`
	BasePath            string `glazed.parameter:"base-path"`
	WorkspaceConfigPath string `glazed.parameter:"workspace-config"`
}

var _ cmds.BareCommand = &APIServeCommand{}

func NewAPIServeCommand() (*APIServeCommand, error) {
	glazedLayer, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"serve",
		cmds.WithShort("Start the workbench API server"),
		cmds.WithLong(`Start the Refactorio Workbench API server.

Examples:
  refactorio api serve --addr :8080 --base-path /api
  refactorio api serve --workspace-config ~/.config/refactorio/workspaces.json
`),
		cmds.WithFlags(
			fields.New(
				"addr",
				fields.TypeString,
				fields.WithDefault(":8080"),
				fields.WithHelp("Address to listen on"),
			),
			fields.New(
				"base-path",
				fields.TypeString,
				fields.WithDefault("/api"),
				fields.WithHelp("Base path for API routes"),
			),
			fields.New(
				"workspace-config",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Path to workspace config file (default: OS config dir)"),
			),
		),
		cmds.WithLayersList(glazedLayer, commandSettingsLayer),
	)

	return &APIServeCommand{CommandDescription: cmdDesc}, nil
}

func (c *APIServeCommand) Run(ctx context.Context, parsedLayers *values.Values) error {
	_ = ctx
	settings := &APIServeSettings{}
	if err := values.DecodeSectionInto(parsedLayers, schema.DefaultSlug, settings); err != nil {
		return err
	}

	log.Info().
		Str("addr", settings.Addr).
		Str("base_path", settings.BasePath).
		Str("workspace_config", settings.WorkspaceConfigPath).
		Msg("Starting workbench API server")

	srv := workbenchapi.NewServer(workbenchapi.Config{
		Addr:                settings.Addr,
		BasePath:            settings.BasePath,
		WorkspaceConfigPath: settings.WorkspaceConfigPath,
	})
	if err := srv.ListenAndServe(); err != nil {
		return errors.Wrap(err, "run api server")
	}
	return nil
}

func NewAPICommand() (*cobra.Command, error) {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Workbench API server",
	}

	serveCmd, err := NewAPIServeCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build api serve command")
	}
	cobraServeCmd, err := cli.BuildCobraCommand(serveCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire api serve command")
	}
	apiCmd.AddCommand(cobraServeCmd)

	return apiCmd, nil
}
