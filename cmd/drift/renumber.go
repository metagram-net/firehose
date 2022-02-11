package main

import (
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/metagram-net/firehose/clio"
	"github.com/metagram-net/firehose/drift"
)

const renumberLong string = `Renumber migrations to fix filesystem sorting.

This command renames migration files so that string sorting matches the numeric
sorting of the IDs. This happens by adding or removing prefix zeroes on the IDs
to make all the IDs the shortest width that fits them all.

Other commands ignore zero prefixes when interpreting IDs as integers. This
renumbering is never necessary for correctness.`

func renumberCmd(io *clio.IO) *cobra.Command {
	var write bool

	cmd := &cobra.Command{
		Use:          "renumber",
		Short:        "Renumber migrations to fix filesystem sorting",
		Long:         renumberLong,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir := viper.GetString("migrations-dir")
			return drift.Renumber(io, dir, write)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&write, "write", "w", false, "Execute renames instead of just printing them")
	return cmd
}
