package main

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

func main() {
	initViper()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		stop()
		log.Print("Interrupt received, cleaning up before quitting. Interrupt again to force-quit.")
	}()

	err := rootCmd().ExecuteContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func initViper() {
	viper.SetConfigName("firehose-admin.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("FIREHOSE_ADMIN")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fh-admin",
		Short:   "Administration commands for a Firehose server",
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
	viper.BindPFlags(flags)

	cmd.AddCommand(
		userCmd(),
		// TODO: dev/demo command to generate "realistic" data
	)
	return cmd
}
