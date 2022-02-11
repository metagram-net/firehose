package main

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/metagram-net/firehose/clio"
)

const defaultMigrationsDir = "migrations"

func rootCmd(io *clio.IO) *cobra.Command {
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
			io.SetVerbosity(clio.Verbosity(viper.GetInt("verbose")))
			return err
		},
	}
	flags := cmd.PersistentFlags()
	flags.String("migrations-dir", defaultMigrationsDir, "Directory containing migration files")
	flags.CountP("verbose", "v", "Log verbosity")
	viper.BindPFlags(flags)

	cmd.AddCommand(
		migrateCmd(io),
		newCmd(io),
		setupCmd(io),
		renumberCmd(io),
	)
	return cmd
}
