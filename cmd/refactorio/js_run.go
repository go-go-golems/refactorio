package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/go-go-golems/refactorio/pkg/refactor/js"
	"github.com/go-go-golems/refactorio/pkg/refactor/js/modules"
	refactorjs "github.com/go-go-golems/refactorio/pkg/refactor/js/modules/refactorindex"
	"github.com/go-go-golems/refactorio/pkg/refactorindex"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewJSRunCommand() (*cobra.Command, error) {
	var scriptPath string
	var indexDB string
	var runID int64
	var tracePath string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a JavaScript file against the refactor index",
		RunE: func(cmd *cobra.Command, args []string) error {
			if scriptPath == "" {
				return errors.New("--script is required")
			}
			if indexDB == "" {
				return errors.New("--index-db is required")
			}

			ctx := context.Background()
			db, err := refactorindex.OpenDB(ctx, indexDB)
			if err != nil {
				return err
			}
			defer db.Close()

			store := refactorindex.NewStore(db)
			module := refactorjs.NewModule(store, runID)
			if tracePath != "" {
				if err := module.EnableTraceFile(tracePath); err != nil {
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

			src, err := os.ReadFile(scriptPath)
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

			if outputFormat == "text" {
				fmt.Printf("%v\n", val.Export())
				return nil
			}

			payload, err := json.MarshalIndent(val.Export(), "", "  ")
			if err != nil {
				return errors.Wrap(err, "marshal result")
			}
			fmt.Println(string(payload))
			return nil
		},
	}

	cmd.Flags().StringVar(&scriptPath, "script", "", "Path to the JS script to execute")
	cmd.Flags().StringVar(&indexDB, "index-db", "", "Path to the refactor-index SQLite DB")
	cmd.Flags().Int64Var(&runID, "run-id", 0, "Run ID to scope queries (0 = all)")
	cmd.Flags().StringVar(&tracePath, "trace", "", "Write query trace JSONL to this path")
	cmd.Flags().StringVar(&outputFormat, "format", "json", "Output format: json|text")

	return cmd, nil
}
