package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCmd() *cobra.Command {
	var (
		// Set default ID out of range in case someone wants to create
		// migration 0.
		id   migrationID = -1
		slug string
	)

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new migration file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			path, err := newFile(viper.GetString("migrations-dir"), id, slug)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(path)
		},
	}
	flags := cmd.Flags()
	flags.Var(&id, "id", "Migration ID override")
	flags.StringVar(&slug, "slug", "", "Short text describing the migration")
	must(cmd.MarkFlagRequired("slug"))
	return cmd
}

func newFile(migrationsDir string, id migrationID, slug string) (string, error) {
	if id == -1 {
		var err error
		ts := time.Now().Unix()
		id, err = newMigrationID(ts)
		if err != nil {
			return "", fmt.Errorf("invalid migration ID: %w", err)
		}
	}

	files, err := available(migrationsDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.ID == int(id) {
			return "", fmt.Errorf("%w: %d: %s", ErrDuplicateID, id, f.Name)
		}
	}

	slug = slugify(slug)
	name := fmt.Sprintf("%d-%s.sql", id, slug)
	path := filepath.Join(migrationsDir, name)

	//#nosec G306 // Normal permissions for non-sensitive files.
	return path, os.WriteFile(path, []byte(template), 0o644)
}

var template = "-- TODO: write your migration here\n"

var reSeparator = regexp.MustCompile(`[\-\s._/]+`)

func slugify(s string) string {
	return reSeparator.ReplaceAllString(s, "_")
}
