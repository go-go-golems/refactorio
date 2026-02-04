package main

import "github.com/spf13/cobra"

func main() {
	rootCmd, err := NewRootCommand()
	cobra.CheckErr(err)
	cobra.CheckErr(rootCmd.Execute())
}
