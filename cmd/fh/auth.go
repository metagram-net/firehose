package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication helpers",
	}
	cmd.AddCommand(
		authWhoamiCmd(),
	)
	return cmd
}

func authWhoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "whoami",
		Short:        "Ask the API who it thinks you are",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			u, err := c.Auth.Whoami(ctx)
			if err != nil {
				return err
			}
			// TODO: table encoder
			return json.NewEncoder(os.Stdout).Encode(u)
		},
	}
	return cmd
}
