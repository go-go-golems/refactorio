package main

import (
	"log"
)

func main() {
	rootCmd, err := NewRootCommand()
	if err != nil {
		log.Fatalf("failed to build root command: %v", err)
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("refactorio failed: %v", err)
	}
}
