package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultMigrationsDir = "migrations"

func Main() error {
	root := rootCmd()
	root.AddCommand(
		migrateCmd(),
		newCmd(),
		setupCmd(),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		stop()
		log.Print("Interrupt received, cleaning up before quitting. Interrupt again to force-quit.")
	}()

	return root.ExecuteContext(ctx)
}

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
			} else if err != nil {
				return err
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.String("migrations-dir", defaultMigrationsDir, "Directory containing migration files")
	viper.BindPFlags(flags)
	return cmd
}

func init() {
	viper.SetConfigName("drift.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("DRIFT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("migrations-dir", "migrations")
}
