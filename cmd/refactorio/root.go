package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "refactorio",
		Short: "Refactorio control plane",
	}

	jsCmd, err := NewJSCommand()
	if err != nil {
		return nil, errors.Wrap(err, "build js command")
	}
	rootCmd.AddCommand(jsCmd)

	apiCmd, err := NewAPICommand()
	if err != nil {
		return nil, errors.Wrap(err, "build api command")
	}
	rootCmd.AddCommand(apiCmd)

	return rootCmd, nil
}
