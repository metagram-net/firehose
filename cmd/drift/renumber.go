package main

import (
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/drift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func renumberCmd() *cobra.Command {
	var write bool

	cmd := &cobra.Command{
		Use:          "renumber",
		Short:        "Renumber migrations to fix filesystem sorting",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir := viper.GetString("migrations-dir")
			return drift.Renumber(dir, write)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&write, "write", "w", false, "Execute renames instead of just printing them")
	return cmd
}
