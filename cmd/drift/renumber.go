package main

import (
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/drift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func renumberCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "renumber",
		Short:        "Renumber migrations to fix filesystem sorting",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			dir := viper.GetString("migrations-dir")
			return drift.Renumber(ctx, dir)
		},
	}
	return cmd
}
