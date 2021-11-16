package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/drift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func migrateCmd() *cobra.Command {
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

			return drift.Migrate(ctx, db, dir)
		},
	}
	return cmd
}
