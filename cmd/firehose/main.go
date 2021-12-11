package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose"
	"github.com/metagram-net/firehose/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	if err := Main(); err != nil {
		panic(err)
	}
}

func initViper() error {
	viper.SetConfigName("firehose-server.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("FIREHOSE_SERVER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("host", "0.0.0.0")
	viper.SetDefault("port", "3473")

	err := viper.ReadInConfig()
	var notFound viper.ConfigFileNotFoundError
	if errors.As(err, &notFound) {
		// No config file needed, use the defaults.
		return nil
	}
	return err
}

func Main() error {
	err := initViper()
	if err != nil {
		return err
	}

	log, err := api.NewLogger()
	if err != nil {
		return err
	}

	log.Info("Starting database connection pool")
	db, err := sql.Open("pgx", viper.GetString("database-url"))
	if err != nil {
		return err
	}
	defer db.Close()

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", viper.GetString("host"), viper.GetString("port")),
		Handler: firehose.Server(log, db),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		stop()
		log.Info("Interrupt received, cleaning up before quitting. Interrupt again to force-quit.")

		// Let the server take as long as it needs to close any open
		// connections.
		//nolint:contextcheck
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Error("Error shutting down", zap.Error(err))
		}
	}()

	log.Info("Listening", zap.String("address", srv.Addr))
	if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		log.Info("Clean shutdown. Bye! 👋")
	} else if err != nil {
		log.Fatal("Error during shutdown", zap.Error(err))
	}

	if err := log.Sync(); err != nil {
		var perr *fs.PathError
		var errno syscall.Errno
		einval := uintptr(0x16)
		// Ignore cases where sync was called with an invalid argument. This
		// can happen when logging to /dev/stderr attached to a terminal.
		if errors.As(err, &perr) &&
			perr.Path == "/dev/stderr" &&
			errors.As(perr.Err, &errno) &&
			uintptr(errno) == einval {
			return nil
		}
		return err
	}
	return nil
}
