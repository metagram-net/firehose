package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/auth"
	"github.com/metagram-net/firehose/auth/apikey"
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
			fmt.Printf("User ID: %s\n", r.userID)
			fmt.Printf("API Key: %s\n", r.apiKey)
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "the user's email address")
	cmd.MarkFlagRequired("email")
	return cmd
}

type registration struct {
	userID uuid.UUID
	apiKey apikey.Plaintext
}

func userRegister(ctx context.Context, db *sql.DB, email string) (*registration, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	reg, err := auth.RegisterUser(ctx, db, email)
	if err != nil {
		return nil, err
	}

	return &registration{
		userID: reg.User.ID,
		apiKey: reg.Plaintext,
	}, tx.Commit()
}
