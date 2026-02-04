package main

import (
	"github.com/spf13/cobra"
)

func NewJSCommand() (*cobra.Command, error) {
	jsCmd := &cobra.Command{
		Use:   "js",
		Short: "JavaScript tooling for refactorio",
	}

	runCmd, err := NewJSRunCommand()
	if err != nil {
		return nil, err
	}
	jsCmd.AddCommand(runCmd)

	return jsCmd, nil
}
