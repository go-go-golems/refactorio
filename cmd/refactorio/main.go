package main

import (
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/refactorio/pkg/doc"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd, err := NewRootCommand()
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	err = doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	cobra.CheckErr(rootCmd.Execute())
}
