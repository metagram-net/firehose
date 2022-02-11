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
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/metagram-net/firehose/server"
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

	app, err := NewApp(Config{
		DevelopmentLogger: viper.GetBool("development-logger"),
		DatabaseURL:       viper.GetString("database-url"),
		Host:              viper.GetString("host"),
		Port:              viper.GetString("port"),
	})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		stop()
		app.log.Info("Interrupt received, cleaning up before quitting. Interrupt again to force-quit.")

		// Let the app take as long as it needs to close any open connections.
		//nolint:contextcheck
		if err := app.Shutdown(context.Background()); err != nil {
			app.log.Error("Error shutting down", zap.Error(err))
		}
	}()

	app.Run()
	return nil
}

type Config struct {
	DevelopmentLogger bool
	DatabaseURL       string
	Host              string
	Port              string
}

type App struct {
	log  *zap.Logger
	db   *sql.DB
	srv  *http.Server
	done chan struct{}
}

func NewApp(cfg Config) (*App, error) {
	log, err := logger(cfg.DevelopmentLogger)
	if err != nil {
		return nil, err
	}

	log.Info("Starting database connection pool")
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: server.New(log, db),
	}

	done := make(chan struct{})

	return &App{log, db, srv, done}, nil
}

func (a *App) Run() {
	a.log.Info("Listening", zap.String("address", a.srv.Addr))
	if err := a.srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		a.log.Info("Clean shutdown. Bye! 👋")
	} else if err != nil {
		a.log.Fatal("Error during shutdown", zap.Error(err))
	}

	// Block until Shutdown finishes.
	<-a.done
}

func (a *App) Shutdown(ctx context.Context) error {
	// No matter what happens, let Run return.
	defer close(a.done)

	a.log.Info("Shutting down HTTP server")
	if err := a.srv.Shutdown(ctx); err != nil {
		a.log.Error("Error shutting down HTTP server", zap.Error(err))
	}

	a.log.Info("Closing database connection")
	if err := a.db.Close(); err != nil {
		a.log.Error("Error closing database connection", zap.Error(err))
	}

	a.log.Info("Log closed")
	return sync(a.log)
}

func logger(dev bool) (*zap.Logger, error) {
	if dev {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func sync(log *zap.Logger) error {
	err := log.Sync()

	var perr *fs.PathError
	var errno syscall.Errno
	einval := uintptr(0x16)
	// Ignore cases where sync was called with an invalid argument. This can
	// happen when logging to /dev/stderr attached to a terminal.
	if errors.As(err, &perr) &&
		perr.Path == "/dev/stderr" &&
		errors.As(perr.Err, &errno) &&
		uintptr(errno) == einval {
		return nil
	}
	return err
}
