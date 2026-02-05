package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/refactorio/pkg/refactor/js"
	"github.com/go-go-golems/refactorio/pkg/refactor/js/modules"
	refactorjs "github.com/go-go-golems/refactorio/pkg/refactor/js/modules/refactorindex"
	"github.com/go-go-golems/refactorio/pkg/refactorindex"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type JSRunCommand struct {
	*cmds.CommandDescription
}

type JSRunSettings struct {
	ScriptPath   string `glazed:"script"`
	IndexDB      string `glazed:"index-db"`
	RunID        int64  `glazed:"run-id"`
	TracePath    string `glazed:"trace"`
	OutputFormat string `glazed:"format"`
}

var _ cmds.BareCommand = &JSRunCommand{}

func NewJSRunCommand() (*JSRunCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"run",
		cmds.WithShort("Run a JavaScript file against the refactor index"),
		cmds.WithLong(`Run a JavaScript file against the refactor index.

Examples:
  refactorio js run --script ./script.js --index-db ./index.db
  refactorio js run --script ./script.js --index-db ./index.db --run-id 42
`),
		cmds.WithFlags(
			fields.New(
				"script",
				fields.TypeString,
				fields.WithHelp("Path to the JS script to execute"),
				fields.WithRequired(true),
			),
			fields.New(
				"index-db",
				fields.TypeString,
				fields.WithHelp("Path to the refactor-index SQLite DB"),
				fields.WithRequired(true),
			),
			fields.New(
				"run-id",
				fields.TypeInteger,
				fields.WithDefault(0),
				fields.WithHelp("Run ID to scope queries (0 = all)"),
			),
			fields.New(
				"trace",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Write query trace JSONL to this path"),
			),
			fields.New(
				"format",
				fields.TypeString,
				fields.WithDefault("json"),
				fields.WithHelp("Output format: json|text"),
			),
		),
	)

	return &JSRunCommand{CommandDescription: cmdDesc}, nil
}

func (c *JSRunCommand) Run(ctx context.Context, parsedLayers *values.Values) error {
	settings := &JSRunSettings{}
	if err := parsedLayers.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	log.Info().
		Str("script", settings.ScriptPath).
		Str("index_db", settings.IndexDB).
		Int64("run_id", settings.RunID).
		Str("format", settings.OutputFormat).
		Str("trace", settings.TracePath).
		Msg("Running JS script")

	db, err := refactorindex.OpenDB(ctx, settings.IndexDB)
	if err != nil {
		return err
	}
	defer db.Close()

	store := refactorindex.NewStore(db)
	module := refactorjs.NewModule(store, settings.RunID)
	if settings.TracePath != "" {
		if err := module.EnableTraceFile(settings.TracePath); err != nil {
			return err
		}
		defer module.CloseTrace()
	}

	reg := modules.NewRegistry()
	reg.Register(module)

	vm, _, err := js.NewRuntime(js.RuntimeOptions{
		Registry:      reg,
		EnableConsole: true,
		DisableTime:   true,
		DisableRandom: true,
		AllowFileJS:   false,
	})
	if err != nil {
		return err
	}

	src, err := os.ReadFile(settings.ScriptPath)
	if err != nil {
		return errors.Wrap(err, "read script")
	}

	val, err := vm.RunString(string(src))
	if err != nil {
		return err
	}
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}

	if settings.OutputFormat == "text" {
		fmt.Printf("%v\n", val.Export())
		return nil
	}

	payload, err := json.MarshalIndent(val.Export(), "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal result")
	}
	fmt.Println(string(payload))
	return nil
}

func NewJSRunCobraCommand() (*cobra.Command, error) {
	cmd, err := NewJSRunCommand()
	if err != nil {
		return nil, err
	}
	return cli.BuildCobraCommand(cmd)
}
