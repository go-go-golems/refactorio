package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "refactor-index",
		Short: "SQLite-backed refactor index tool",
	}

	initCmd, err := NewInitCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build init command")
	}
	cobraInitCmd, err := cli.BuildCobraCommand(initCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire init command")
	}
	rootCmd.AddCommand(cobraInitCmd)

	ingestCmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest data into the refactor index",
	}
	ingestDiffCmd, err := NewIngestDiffCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest diff command")
	}
	cobraIngestDiffCmd, err := cli.BuildCobraCommand(ingestDiffCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest diff command")
	}
	ingestCmd.AddCommand(cobraIngestDiffCmd)

	ingestSymbolsCmd, err := NewIngestSymbolsCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest symbols command")
	}
	cobraIngestSymbolsCmd, err := cli.BuildCobraCommand(ingestSymbolsCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest symbols command")
	}
	ingestCmd.AddCommand(cobraIngestSymbolsCmd)

	ingestCodeUnitsCmd, err := NewIngestCodeUnitsCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest code-units command")
	}
	cobraIngestCodeUnitsCmd, err := cli.BuildCobraCommand(ingestCodeUnitsCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest code-units command")
	}
	ingestCmd.AddCommand(cobraIngestCodeUnitsCmd)
	rootCmd.AddCommand(ingestCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List data from the refactor index",
	}
	listDiffFilesCmd, err := NewListDiffFilesCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build list diff-files command")
	}
	cobraListDiffFilesCmd, err := cli.BuildCobraCommand(listDiffFilesCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire list diff-files command")
	}
	listCmd.AddCommand(cobraListDiffFilesCmd)
	rootCmd.AddCommand(listCmd)

	return rootCmd, nil
}
