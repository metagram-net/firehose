package main

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultMigrationsDir = "migrations"

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "drift",
		Short:   "Manage SQL migrations",
		Version: "0.1.0",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			err := viper.ReadInConfig()
			var notFound viper.ConfigFileNotFoundError
			if errors.As(err, &notFound) {
				// No config file needed, use the defaults.
				return nil
			}
			return err
		},
	}
	flags := cmd.Flags()
	flags.String("migrations-dir", defaultMigrationsDir, "Directory containing migration files")
	viper.BindPFlags(flags)

	cmd.AddCommand(
		migrateCmd(),
		newCmd(),
		setupCmd(),
	)
	return cmd
}
