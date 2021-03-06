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

	"github.com/metagram-net/firehose/client"
)

func main() {
	initViper()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		stop()
		log.Print("Interrupt received, cleaning up before quitting. Interrupt again to force-quit.")
	}()

	err := rootCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func initViper() {
	viper.SetConfigName("firehose.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("FIREHOSE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fh",
		Short:   "Interact with a Firehose server",
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
		wellknownCmd(),
		authCmd(),
		dropCmd(),
	)
	return cmd
}

func Client() (*client.Client, error) {
	return client.New(
		client.WithBaseURL(viper.GetString("url-base")),
		client.WithAuth(
			viper.GetString("user-id"),
			viper.GetString("api-key"),
		),
	)
}
