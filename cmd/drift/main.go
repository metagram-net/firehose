package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

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
	viper.SetConfigName("drift.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("DRIFT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("migrations-dir", "migrations")
}
