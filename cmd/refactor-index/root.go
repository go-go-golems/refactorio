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

	ingestCommitsCmd, err := NewIngestCommitsCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest commits command")
	}
	cobraIngestCommitsCmd, err := cli.BuildCobraCommand(ingestCommitsCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest commits command")
	}
	ingestCmd.AddCommand(cobraIngestCommitsCmd)

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

	ingestDocHitsCmd, err := NewIngestDocHitsCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest doc-hits command")
	}
	cobraIngestDocHitsCmd, err := cli.BuildCobraCommand(ingestDocHitsCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest doc-hits command")
	}
	ingestCmd.AddCommand(cobraIngestDocHitsCmd)

	ingestTreeSitterCmd, err := NewIngestTreeSitterCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest tree-sitter command")
	}
	cobraIngestTreeSitterCmd, err := cli.BuildCobraCommand(ingestTreeSitterCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest tree-sitter command")
	}
	ingestCmd.AddCommand(cobraIngestTreeSitterCmd)

	ingestGoplsRefsCmd, err := NewIngestGoplsRefsCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest gopls-refs command")
	}
	cobraIngestGoplsRefsCmd, err := cli.BuildCobraCommand(ingestGoplsRefsCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest gopls-refs command")
	}
	ingestCmd.AddCommand(cobraIngestGoplsRefsCmd)

	ingestRangeCmd, err := NewIngestRangeCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build ingest range command")
	}
	cobraIngestRangeCmd, err := cli.BuildCobraCommand(ingestRangeCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire ingest range command")
	}
	ingestCmd.AddCommand(cobraIngestRangeCmd)
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

	listSymbolsCmd, err := NewListSymbolsCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build list symbols command")
	}
	cobraListSymbolsCmd, err := cli.BuildCobraCommand(listSymbolsCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire list symbols command")
	}
	listCmd.AddCommand(cobraListSymbolsCmd)
	rootCmd.AddCommand(listCmd)

	reportCmd, err := NewReportCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build report command")
	}
	cobraReportCmd, err := cli.BuildCobraCommand(reportCmd)
	if err != nil {
		return nil, errors.Wrap(err, "wire report command")
	}
	rootCmd.AddCommand(cobraReportCmd)

	return rootCmd, nil
}
