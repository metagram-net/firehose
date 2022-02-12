package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func wellknownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "well-known",
		Short: "Fetch global and status data (for debugging)",
	}
	cmd.AddCommand(
		wellknownHealthCheckCmd(),
	)
	return cmd
}

func wellknownHealthCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "health-check",
		Short:        "See the server status report",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			res, err := c.WellKnown.HealthCheck(ctx)
			if err != nil {
				return err
			}
			// TODO: table encoder
			return json.NewEncoder(os.Stdout).Encode(res)
		},
	}
	return cmd
}
