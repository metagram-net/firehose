package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/auth/apikey"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func userCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}
	cmd.AddCommand(
		userRegisterCmd(),
	)
	return cmd
}

func userRegisterCmd() *cobra.Command {
	var email string
	cmd := &cobra.Command{
		Use:          "register",
		Short:        "Register a new user",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			db, err := sql.Open("pgx", viper.GetString("database-url"))
			if err != nil {
				return fmt.Errorf("could not open database connection: %w", err)
			}
			defer db.Close()
			r, err := userRegister(cmd.Context(), db, email)
			if err != nil {
				return err
			}
			log.Printf("User ID: %s", r.userID)
			log.Printf("API Key: %s", r.apiKey)
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "the user's email address")
	cmd.MarkFlagRequired("email")
	return cmd
}

type registration struct {
	userID uuid.UUID
	apiKey *apikey.Plaintext
}

func userRegister(ctx context.Context, db *sql.DB, email string) (registration, error) {
	var zero registration

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}

	u, err := user.Create(ctx, tx, email)
	if err != nil {
		return zero, fmt.Errorf("could not create user: %w", err)
	}

	key, _, err := apikey.Create(ctx, tx, "Default", u.ID)
	if err != nil {
		return zero, fmt.Errorf("could not create API key: %w", err)
	}

	return registration{
		userID: u.ID,
		apiKey: key,
	}, tx.Commit()
}
