package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/metagram-net/firehose/clio"
	"github.com/metagram-net/firehose/drift"
)

func migrateCmd(io *clio.IO) *cobra.Command {
	// Set the default ID out of range to distinguish explicit zero.
	untilID := drift.MigrationID(-1)

	cmd := &cobra.Command{
		Use:          "migrate",
		Short:        "Run migrations",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			dir := viper.GetString("migrations-dir")

			db, err := sql.Open("pgx", viper.GetString("database-url"))
			if err != nil {
				return fmt.Errorf("could not open database connection: %w", err)
			}
			defer db.Close()

			var until *drift.MigrationID
			if untilID >= 0 {
				until = &untilID
			}
			return drift.Migrate(ctx, io, db, dir, until)
		},
	}

	flags := cmd.Flags()
	flags.Var(&untilID, "until", "Maximum migration ID to run (default: run all migrations)")
	return cmd
}
