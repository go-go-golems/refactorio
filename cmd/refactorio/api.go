package main

import (
	"github.com/go-go-golems/refactorio/pkg/workbenchapi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewAPICommand() (*cobra.Command, error) {
	var addr string
	var basePath string
	var workspaceConfigPath string

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Workbench API server",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the workbench API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := workbenchapi.NewServer(workbenchapi.Config{
				Addr:                addr,
				BasePath:            basePath,
				WorkspaceConfigPath: workspaceConfigPath,
			})
			if err := srv.ListenAndServe(); err != nil {
				return errors.Wrap(err, "run api server")
			}
			return nil
		},
	}

	serveCmd.Flags().StringVar(&addr, "addr", ":8080", "Address to listen on")
	serveCmd.Flags().StringVar(&basePath, "base-path", "/api", "Base path for API routes")
	serveCmd.Flags().StringVar(&workspaceConfigPath, "workspace-config", "", "Path to workspace config file (default: OS config dir)")

	apiCmd.AddCommand(serveCmd)
	return apiCmd, nil
}
